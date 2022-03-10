package pbft

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/cloudflare/cfssl/log"
	common2 "github.com/rjkris/go-jellyfish-merkletree/common"
	"github.com/ssbcV2/account"
	"github.com/ssbcV2/chain"
	"github.com/ssbcV2/common"
	"github.com/ssbcV2/contract"
	"github.com/ssbcV2/event"
	"github.com/ssbcV2/global"
	"github.com/ssbcV2/merkle"
	"github.com/ssbcV2/meta"
	"github.com/ssbcV2/network"
	"github.com/ssbcV2/util"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

//本地消息池（模拟持久化层），只有确认提交成功后才会存入此池
var localMessagePool = []Message{}

type node struct {
	//节点ID
	nodeID string
	//节点监听地址
	addr string
	//RSA私钥
	rsaPrivKey []byte
	//RSA公钥
	rsaPubKey []byte
}

type pbft struct {
	//节点信息
	node node
	//每笔请求自增序号
	sequenceID int
	//锁
	lock sync.Mutex
	//节点的交易池
	transPool map[int][]meta.Transaction
	//临时消息池，消息摘要对应消息本体
	messagePool map[string]Request
	//存放收到的prepare数量(至少需要收到并确认2f个)，根据摘要来对应
	prePareConfirmCount map[string]map[string]bool
	//存放收到的commit数量（至少需要收到并确认2f+1个），根据摘要来对应
	commitConfirmCount map[string]map[string]bool
	//该笔消息是否已进行Commit广播
	isCommitBordcast map[string]bool
	//该笔消息是否已对客户端进行Reply
	isReply map[string]bool
}

func NewPBFT(nodeID, addr string) *pbft {
	p := new(pbft)
	p.node.nodeID = nodeID
	p.node.addr = addr
	p.node.rsaPrivKey = p.getPivKey(nodeID) //从生成的私钥文件处读取
	p.node.rsaPubKey = p.getPubKey(nodeID)  //从生成的私钥文件处读取
	p.sequenceID = 0
	p.transPool = make(map[int][]meta.Transaction)
	p.messagePool = make(map[string]Request)
	p.prePareConfirmCount = make(map[string]map[string]bool)
	p.commitConfirmCount = make(map[string]map[string]bool)
	p.isCommitBordcast = make(map[string]bool)
	p.isReply = make(map[string]bool)
	return p
}

//节点使用的tcp监听
func (p *pbft) TcpListen() {
	listen, err := net.Listen("tcp", p.node.addr)
	if err != nil {
		log.Error(err)
	}
	log.Info("节点开启监听，地址：", p.node.addr)
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		log.Info("新连接：", conn.LocalAddr().String(), " -- ", conn.RemoteAddr().String())
		if err != nil {
			log.Error(err)
		}
		p.handleNewConn(conn)
	}
}

func (p *pbft) handleNewConn(conn net.Conn) {
	b, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Error("[handleNewConn] err:", err)
	}
	p.handleRequest(b, conn)
}

//处理tcp请求
func (p *pbft) handleRequest(data []byte, conn net.Conn) {
	//先解析消息
	msg := network.ParseTCPMsg(data)
	//根据消息类别选择handle函数
	if strings.HasPrefix(msg.Type, common.PBFT) {
		p.handlePBFTMsg(msg)
	}
	//主节点会接收到其他节点的区块链同步消息
	if msg.Type == common.BlockSynReqMsg {
		network.HandleBlockSynReqMsg(msg, conn)
	}
	//其他节点会接收到主节点的区块链同步回复
	if msg.Type == common.BlockSynResMsg {
		network.HandleBlockSynResMsg(msg, conn)
	}

	// 只有client节点处理这个case
	if msg.Type == common.PBFTReply && p.node.nodeID == global.Client {
		bcs := chain.GetCurrentBlockChain()
		newBC := new(meta.Block)
		bc := msg.Content
		err := json.Unmarshal(bc, newBC)
		util.DealJsonErr("clientHandleTcpMsg", err)
		//判断是否已将reply的新区块入库了
		if newBC.Height == len(bcs) {
			bcs = append(bcs, *newBC)
			chain.StoreBlockChain(bcs)
			//状态更新
			p.refreshState(newBC, len(bcs)-1)
		}
	}
}

func (p *pbft) handlePBFTMsg(msg meta.TCPMessage) {
	//根据消息命令调用不同的功能
	switch msg.Type {
	case common.PBFTRequest:
		p.handleClientRequest(msg.Content)
	case common.PBFTPrePrepare:
		p.handlePrePrepare(msg.Content)
	case common.PBFTPrepare:
		p.handlePrepare(msg.Content)
	case common.PBFTCommit:
		p.handleCommit(msg.Content)
	}
}

//处理客户端发来的请求
func (p *pbft) handleClientRequest(content []byte) {
	log.Info("主节点已接收到客户端发来的request ...")
	log.Infof("request info: %s", string(content))
	//Step1:使用json解析出Request结构体
	r := new(Request)
	err := json.Unmarshal(content, r)
	if err != nil {
		log.Error(err)
	}

	var transList []meta.Transaction
	// 收到事件消息，转化成交易放入交易池
	if r.Type == 1 {
		var message meta.EventMessage
		err := json.Unmarshal([]byte(r.Content), &message)
		if err != nil {
			log.Errorf("event message decode error: %s", err)
		} else {
			eventTrans, err := event.EventToTransaction(message)
			if err != nil {
				log.Errorf("event to trans error: %s", err)
			}
			transList = append(transList, eventTrans...)
		}
	} else {
		//Step2：主节点需要先将交易存储至临时的交易池，待交易池满，打包为区块进行PBFT共识
		transMsg := r.Content
		trans := meta.Transaction{}
		log.Infof("交易信息：%v\n", transMsg)
		err = json.Unmarshal([]byte(transMsg), &trans)
		util.DealJsonErr("handleClientRequest", err)

		// 检查交易能否执行，没问题就打包成块
		if ok := checkTran(trans); !ok {
			// 将消息反馈至前端
			return
		}
		trans.Timestamp = time.Now().String()
		trans.Id, _ = util.CalculateHash([]byte(trans.Timestamp))
		transList = append(transList, trans)
	}

	//step3：主节点对交易进行验签，验签不通过的丢弃
	//if !RsaVerySignWithSha256(trans.Hash,trans.Sign,[]byte(trans.PublicKey)){
	//	log.Error("[handleClientRequest] 验签失败!!")
	//	return
	//}
	//*******************************************************************

	//待注释内容--测试专用(主节点作恶：篡改交易内容，hash值变化，但是无法篡改客户端签名)
	//trans.Value=100
	//tB,_:=json.Marshal(trans)
	//tH,_:=util.CalculateHash(tB)
	//trans.Hash=tH

	//*******************************************************************

	//解析交易、执行交易步骤根据交易的input生成output

	bc := chain.GetCurrentBlockChain()
	index := len(bc)
	p.transPool[index] = append(p.transPool[index], transList...)
	//满足交易数则打包新区块
	if len(p.transPool[index]) == common.TxsThreshold {
		//主节点接收到的交易已经到达阈值，打包新区块进行PBFT共识
		newBlock := chain.GenerateNewBlock(p.transPool[index])
		//主节点对打包区块进行签名
		blockSign := util.RsaSignWithSha256(newBlock.Hash, p.node.rsaPrivKey)
		newBlock.Signature = blockSign
		newBlockMsg, err := json.Marshal(newBlock)
		util.DealJsonErr("handleClientRequest", err)
		r.Content = string(newBlockMsg)
		//添加信息序号
		p.sequenceIDAdd()
		//获取消息摘要
		digest := getDigest(*r)
		log.Info("已将request存入临时消息池")
		//存入临时消息池
		p.messagePool[digest] = *r
		//主节点对消息摘要进行签名
		digestByte, _ := hex.DecodeString(digest)
		signInfo := util.RsaSignWithSha256(digestByte, p.node.rsaPrivKey)
		//拼接成PrePrepare，准备发往follower节点
		pp := PrePrepare{*r, digest, p.sequenceID, signInfo}
		b, err := json.Marshal(pp)
		if err != nil {
			log.Error(err)
		}
		log.Info("正在向其他节点进行进行PrePrepare广播 ...")
		//进行PrePrepare广播
		msg := meta.TCPMessage{
			Type:    common.PBFTPrePrepare,
			Content: b,
		}
		p.broadcast(msg)
		log.Info("PrePrepare广播完成")
	} else {
		log.Infof("主节点已将交易存储至交易池，交易详情：%+v", transList)
	}
}

//处理预准备消息
func (p *pbft) handlePrePrepare(content []byte) {
	log.Info("本节点已接收到主节点发来的PrePrepare ...")
	//	//使用json解析出PrePrepare结构体
	pp := new(PrePrepare)
	err := json.Unmarshal(content, pp)
	if err != nil {
		log.Error(err)
	}
	//获取主节点的公钥，用于数字签名验证
	primaryNodePubKey := p.getPubKey(global.Master)
	digestByte, _ := hex.DecodeString(pp.Digest)
	//首先检查所有的交易客户端签名（防止主节点作恶）
	//step1先获取到全部的交易
	block := meta.Block{}
	err = json.Unmarshal([]byte(pp.RequestMessage.Content), &block)
	if err != nil {
		log.Error("[handlePrePrepare] json err:", err)
	}
	//for _,tx:=range block.TX{
	//	//验签,只要有一笔交易的验签不通过则拒绝进行prepare广播
	//	if !RsaVerySignWithSha256(tx.Hash,tx.Sign,[]byte(tx.PublicKey)){
	//		log.Info("交易签名验证失败，怀疑主节点篡改交易信息，拒绝进行prepare广播")
	//		return
	//	}
	//}
	if digest := getDigest(pp.RequestMessage); digest != pp.Digest {
		log.Info("信息摘要对不上，拒绝进行prepare广播")
	} else if p.sequenceID+1 != pp.SequenceID {
		log.Info("消息序号对不上，拒绝进行prepare广播")
	} else if !util.RsaVerySignWithSha256(digestByte, pp.Sign, primaryNodePubKey) {
		log.Info("主节点签名验证失败！,拒绝进行prepare广播")
	} else {
		//序号赋值
		p.sequenceID = pp.SequenceID
		//将信息存入临时消息池
		log.Info("已将消息存入临时节点池")
		p.messagePool[pp.Digest] = pp.RequestMessage
		//节点使用私钥对其签名
		sign := util.RsaSignWithSha256(digestByte, p.node.rsaPrivKey)
		//拼接成Prepare
		pre := Prepare{pp.Digest, pp.SequenceID, p.node.nodeID, sign}

		//*******************************************************************

		//待注释--测试专用（从节点作恶：篡改消息摘要）
		//if p.node.nodeID=="N1"{
		//	log.Info("从节点N1作恶：篡改消息摘要")
		//	pre.Digest="就是玩儿"
		//}
		//
		//if p.node.nodeID=="N2"{
		//	log.Info("从节点N2作恶：篡改消息摘要")
		//	pre.Digest="就是玩儿"
		//}

		//*******************************************************************
		bPre, err := json.Marshal(pre)
		if err != nil {
			log.Error(err)
		}
		//进行准备阶段的广播
		log.Info("正在进行Prepare广播 ...")
		msg := meta.TCPMessage{
			Type:    common.PBFTPrepare,
			Content: bPre,
		}
		p.broadcast(msg)
		log.Info("Prepare广播完成")
	}
}

//处理准备消息
func (p *pbft) handlePrepare(content []byte) {
	//使用json解析出Prepare结构体
	pre := new(Prepare)
	err := json.Unmarshal(content, pre)
	if err != nil {
		log.Error(err)
	}
	log.Infof("本节点已接收到%s节点发来的Prepare ... \n", pre.NodeID)
	//获取消息源节点的公钥，用于数字签名验证
	MessageNodePubKey := p.getPubKey(pre.NodeID)
	digestByte, _ := hex.DecodeString(pre.Digest)
	if _, ok := p.messagePool[pre.Digest]; !ok {
		log.Info("当前临时消息池无此摘要，拒绝执行commit广播")
	} else if p.sequenceID != pre.SequenceID {
		log.Info("消息序号对不上，拒绝执行commit广播")
	} else if !util.RsaVerySignWithSha256(digestByte, pre.Sign, MessageNodePubKey) {
		log.Info("节点签名验证失败！,拒绝执行commit广播")
	} else {
		p.setPrePareConfirmMap(pre.Digest, pre.NodeID, true)
		count := 0
		for range p.prePareConfirmCount[pre.Digest] {
			count++
		}
		//因为主节点不会发送Prepare，所以不包含自己
		specifiedCount := 0
		if p.node.nodeID == global.Master {
			specifiedCount = common.NodeCount / 3 * 2
		} else {
			specifiedCount = (common.NodeCount / 3 * 2) - 1
		}
		//如果节点至少收到了2f个prepare的消息（包括自己）,并且没有进行过commit广播，则进行commit广播
		p.lock.Lock()
		//获取消息源节点的公钥，用于数字签名验证
		if count >= specifiedCount && !p.isCommitBordcast[pre.Digest] {
			log.Info("本节点已收到至少2f个节点(包括本地节点)发来的Prepare信息 ...")

			//*******************************************************************

			//待注释--测试专用（节点作恶：即使全部验证通过也拒绝广播）
			//if p.node.nodeID==common.Master{
			//	log.Info("主节点作恶：全部验证通过，但是拒绝广播")
			//	p.lock.Unlock()
			//	return
			//}
			//
			//if p.node.nodeID=="N1"{
			//	log.Info("从节点N1作恶：全部验证通过，但是拒绝广播")
			//	p.lock.Unlock()
			//	return
			//}
			//
			//*******************************************************************

			//节点使用私钥对其签名
			sign := util.RsaSignWithSha256(digestByte, p.node.rsaPrivKey)
			c := Commit{pre.Digest, pre.SequenceID, p.node.nodeID, sign}
			bc, err := json.Marshal(c)
			if err != nil {
				log.Error(err)
			}
			//进行提交信息的广播
			log.Info("正在进行commit广播")
			msg := meta.TCPMessage{
				Type:    common.PBFTCommit,
				Content: bc,
			}
			p.broadcast(msg)
			p.isCommitBordcast[pre.Digest] = true
			log.Info("commit广播完成")
		}
		p.lock.Unlock()
	}
}

//处理提交确认消息
func (p *pbft) handleCommit(content []byte) {
	//使用json解析出Commit结构体
	c := new(Commit)
	err := json.Unmarshal(content, c)
	if err != nil {
		log.Error(err)
	}
	log.Infof("本节点已接收到%s节点发来的Commit ... \n", c.NodeID)
	//获取消息源节点的公钥，用于数字签名验证
	MessageNodePubKey := p.getPubKey(c.NodeID)
	digestByte, _ := hex.DecodeString(c.Digest)
	if _, ok := p.prePareConfirmCount[c.Digest]; !ok {
		log.Info("当前prepare池无此摘要，拒绝将信息持久化到本地消息池")
	} else if p.sequenceID != c.SequenceID {
		log.Info("消息序号对不上，拒绝将信息持久化到本地消息池")
	} else if !util.RsaVerySignWithSha256(digestByte, c.Sign, MessageNodePubKey) {
		log.Info("节点签名验证失败！,拒绝将信息持久化到本地消息池")
	} else {
		p.setCommitConfirmMap(c.Digest, c.NodeID, true)
		count := 0
		for range p.commitConfirmCount[c.Digest] {
			count++
		}
		//如果节点至少收到了2f+1个commit消息（包括自己）,并且节点没有回复过,并且已进行过commit广播，则提交信息至本地消息池，并reply成功标志至客户端！
		p.lock.Lock()
		if count >= common.NodeCount/3*2 && !p.isReply[c.Digest] && p.isCommitBordcast[c.Digest] {
			log.Info("本节点已收到至少2f + 1 个节点(包括本地节点)发来的Commit信息 ...")
			//将消息信息，提交到本地消息池中！
			localMessagePool = append(localMessagePool, p.messagePool[c.Digest].Message)
			info := p.node.nodeID + "节点已将msgid:" + strconv.Itoa(p.messagePool[c.Digest].ID) + "存入本地消息池中,消息内容为：" + p.messagePool[c.Digest].Content
			log.Info(info)
			//既然已经得到共识，新区块上链，落库
			bcs := chain.GetCurrentBlockChain()
			newBCMsg := p.messagePool[c.Digest].Content
			newBC := new(meta.Block)
			err = json.Unmarshal([]byte(newBCMsg), newBC)
			util.DealJsonErr("handleCommit", err)
			//先更新状态树，得到stateRoot，再上链
			p.refreshState(newBC, len(bcs))

			bcs = append(bcs, *newBC)
			chain.StoreBlockChain(bcs)
			//给客户端reply
			log.Info("正在reply客户端 ...")
			tcpMsg := meta.TCPMessage{
				Type:    common.PBFTReply,
				Content: []byte(newBCMsg),
			}
			util.TCPSend(tcpMsg, global.ClientToNodeAddr)
			p.isReply[c.Digest] = true
			log.Info("reply完毕")
		}
		p.lock.Unlock()
	}
}

// 状态数据库更新
func (p *pbft) refreshState(b *meta.Block, height int) {
	//ste1：首先取出本区块中所有的交易
	txs := b.TX

	// 执行每一笔交易
	for _, tx := range txs {
		p.execute(tx)
	}
	if len(global.ChangedAccounts) != 0 && len(global.TreeData) != 0 {
		return
	}
	// state和event版本同步,和区块高度相同
	var stateRootHash, eventRootHash common2.HashValue
	if len(global.ChangedAccounts) != 0 {
		stateRootHash, _ = merkle.UpdateStateTree(global.ChangedAccounts, uint64(height), merkle.AccountStatePath)
	} else { // 需要更新的账户为空时，更新初始账户
		stateRootHash, _ = merkle.UpdateStateTree([]meta.JFTreeData{merkle.InitAccount}, uint64(height), merkle.AccountStatePath)
	}
	b.StateRoot = stateRootHash.Bytes()
	if len(global.TreeData) != 0 {
		eventRootHash, _ = merkle.UpdateStateTree(global.TreeData, uint64(height), merkle.EventStatePath)
	} else {
		eventRootHash, _ = merkle.UpdateStateTree([]meta.JFTreeData{merkle.InitEvent}, uint64(height), merkle.EventStatePath)
	}
	b.EventRoot = eventRootHash.Bytes()

	global.ChangedAccounts = []meta.JFTreeData{} // 本轮区块d的状态修改已持久化，清空列表
	global.TreeData = []meta.JFTreeData{}        // 本轮区块的event，sub已持久化，清空列表
}

func (p *pbft) execute(tx meta.Transaction) {
	switch tx.Type {
	case meta.Register:
		global.ChangedAccounts = append(global.ChangedAccounts, account.CreateAccount(tx.To, tx.PublicKey, common.InitBalance))
	case meta.Transfer:
		global.ChangedAccounts = append(global.ChangedAccounts, account.SubBalance(tx.From, tx.Value), account.AddBalance(tx.To, tx.Value))
	case meta.Publish:
		contract.SetContext(meta.ContractTask{
			Caller: tx.From, // 部署时加载发布人的地址，用于智能合约init
		})

		err := p.deployContract(tx.Contract, tx.Data.Code)
		if err != nil {
			log.Error("节点部署合约出错: ", err)
			return
		}
		newAccount := account.CreateContract(tx.Contract, tx.Data.Code, tx.From)
		global.ChangedAccounts = append(global.ChangedAccounts, newAccount)
		//更新事件数据，每个节点都执行
		contractName := tx.Contract
		eList, _ := event.ExecuteInitEvent(contractName, newAccount.Address, tx.From)
		global.TreeData = append(global.TreeData, eList...)
	case meta.Invoke:
		// 调用合约的同时向合约账户转账
		if tx.Value > 0 {
			global.ChangedAccounts = append(global.ChangedAccounts, account.SubBalance(tx.From, tx.Value))
			global.ChangedAccounts = append(global.ChangedAccounts, account.AddBalance(tx.Contract, tx.Value))
		}
		// 每个节点都会去执行智能合约，需要确保智能合约执行的确定性（暂时没有做合约执行后的共识）
		global.TaskList = append(global.TaskList, meta.ContractTask{
			tx.From,
			tx.Value,
			tx.Contract,
			tx.Method,
			tx.Args,
		})
		for len(global.TaskList) != 0 {
			err := event.HandleContractTask()
			if err != nil {
				log.Errorf("contract task handle error: %s", err)
				continue
			}
		}
	case meta.CrossTransfer:
		if global.ChainID == tx.SourceChainId {
			global.ChangedAccounts = append(global.ChangedAccounts, account.SubBalance(tx.From, tx.Value))

			// 由client节点向目标链client发送交易和proof
			if p.node.nodeID == global.Client {
				postCT2Dest(tx)
			}

		} else if global.ChainID == tx.DestChainId {
			// todo: 校验proof
			//proof := tx.Proof
			global.ChangedAccounts = append(global.ChangedAccounts, account.AddBalance(tx.To, tx.Value))
		}

	default:
		log.Infof("未知的交易类型")
	}
}

// 交易打包前检测
func checkTran(tx meta.Transaction) bool {
	switch tx.Type {
	case meta.Transfer:
		// 判断能否转账
		if ok := account.CanTransfer(tx.From, tx.Value); !ok {
			log.Info("余额不足，无法转账")
			return false
		}
	case meta.Register:
		// 检查能否创建账户
	case meta.Publish:
		// 判断合约能否部署（并没有真正部署）
	case meta.Invoke:
		// 判断合约能否调用（并没有真正调用）
	case meta.CrossTransfer:
		// 判断能否跨链转账
	default:
		log.Infof("未知的交易类型")
		return false
	}
	return true
}

// 向目标链client发送转账交易
func postCT2Dest(ct meta.Transaction) {
	cBcs := chain.GetCurrentBlockChain()
	height := len(cBcs) - 1
	// 获取到区块中所有的交易
	txs := cBcs[height].TX
	// 生成该交易的merkle proof
	tranHash, merklePath, merkleIndex := merkle.GetTranMerklePath(txs, 0) // 交易index都是0，因为目前每个区块只有一个交易
	proof := meta.CrossTranProof{
		MerklePath:  merklePath,
		TransHash:   tranHash,
		Height:      height,
		MerkleIndex: merkleIndex,
	}
	ct.Proof = proof

	r := new(Request)
	r.Timestamp = time.Now().UnixNano()
	r.ClientAddr = global.ClientToNodeAddr
	r.Message.ID = util.GetRandom()
	r.Type = 0

	tb, _ := json.Marshal(ct)
	r.Message.Content = string(tb)
	br, err := json.Marshal(r)
	if err != nil {
		log.Error(err)
	}
	//log.Info(string(br))
	msg := meta.TCPMessage{
		Type:    common.PBFTRequest,
		Content: br,
	}
	var targetAddress string
	if global.ChainID == common.ChainId1 {
		targetAddress = common.NodeTable2["N4"]
	}
	if global.ChainID == common.ChainId2 {
		targetAddress = common.NodeTable1["N0"]
	}
	log.Infof("target: %v", targetAddress)
	util.TCPSend(msg, targetAddress)
}

//序号累加
func (p *pbft) sequenceIDAdd() {
	p.lock.Lock()
	p.sequenceID++
	p.lock.Unlock()
}

//向除自己外的其他节点进行广播
func (p *pbft) broadcast(msg meta.TCPMessage) {
	for i := range global.NodeTable {
		if i == p.node.nodeID {
			continue
		}
		go util.TCPSend(msg, global.NodeTable[i])
	}
}

//为多重映射开辟赋值
func (p *pbft) setPrePareConfirmMap(val, val2 string, b bool) {
	if _, ok := p.prePareConfirmCount[val]; !ok {
		p.prePareConfirmCount[val] = make(map[string]bool)
	}
	p.prePareConfirmCount[val][val2] = b
}

//为多重映射开辟赋值
func (p *pbft) setCommitConfirmMap(val, val2 string, b bool) {
	if _, ok := p.commitConfirmCount[val]; !ok {
		p.commitConfirmCount[val] = make(map[string]bool)
	}
	p.commitConfirmCount[val][val2] = b
}

//传入节点编号， 获取对应的公钥
func (p *pbft) getPubKey(nodeID string) []byte {
	key, err := ioutil.ReadFile("Keys/" + nodeID + "/" + nodeID + "_RSA_PUB")
	if err != nil {
		log.Error(err)
	}
	return key
}

//传入节点编号， 获取对应的私钥
func (p *pbft) getPivKey(nodeID string) []byte {
	key, err := ioutil.ReadFile("Keys/" + nodeID + "/" + nodeID + "_RSA_PIV")
	if err != nil {
		log.Error(err)
	}
	return key
}

func (p *pbft) deployContract(name, code string) error {
	dir := "./contract/contract/" + p.node.nodeID + "/" + name + "/"
	if util.FileExists(dir) {
		log.Error("该合约已存在")
		return errors.New("该合约已存在")
	} else {
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			log.Error(err)
		}
		// 创建保存文件
		destFile, err := os.Create(dir + name + ".go")
		if err != nil {
			log.Error("Create failed: %s\n", err)
			return err
		}
		defer destFile.Close()
		_, _ = destFile.WriteString(code)

		err = contract.GoBuildPlugin(name)
		if err != nil {
			//将文件夹删除
			err1 := os.RemoveAll(dir)
			if err1 != nil {
				log.Error(err1)
			}
			return err
		}
	}
	return nil
}

// 部署预言机系统智能合约
func (p *pbft) DeploySysContract() error {
	r, _ := regexp.Compile("(.*).go")
	fds, err := ioutil.ReadDir("./contract/system")
	if err != nil {
		log.Errorf("遍历contract/system失败: %s", err)
		return err
	}
	for _, fi := range fds {
		if !fi.IsDir() {
			res := r.FindStringSubmatch(fi.Name())
			if len(res) == 2 {
				contractName := res[1]
				code, _ := ioutil.ReadFile("./contract/system/" + fi.Name())
				//log.Infof("name: %s", contractName)
				err := p.deployContract(contractName, string(code))
				if err != nil {
					log.Errorf("系统合约部署失败:%s,%s", fi.Name(), err)
					continue
				} else {
					log.Infof("系统合约%s部署成功", contractName)
				}
			}
		}
	}
	return nil
}
