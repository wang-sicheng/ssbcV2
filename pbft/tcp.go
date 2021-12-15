package pbft

import (
	"encoding/json"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/account"
	"github.com/ssbcV2/chain"
	"github.com/ssbcV2/common"
	"github.com/ssbcV2/merkle"
	"github.com/ssbcV2/meta"
	"github.com/ssbcV2/network"
	"github.com/ssbcV2/util"
	"io/ioutil"
	"net"
)

//客户端使用的tcp监听
func clientTcpListen() {
	listen, err := net.Listen("tcp", commonconst.ClientToNodeAddr)
	if err != nil {
		log.Error(err)
	}
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Error(err)
		}
		log.Info("新连接：", conn.LocalAddr().String(), " -- ", conn.RemoteAddr().String())
		clientHandleNewConn(conn)
	}
}

func clientHandleNewConn(conn net.Conn) {
	//buf := make([]byte, 4096)
	//n, err := conn.Read(buf) //从conn读取
	//log.Info("接收到消息：", string(buf[:n]))
	//if err == nil {
	//	clientHandleTcpMsg(buf[:n], conn)
	//} else if err != io.EOF {
	//	log.Error(err)
	//}
	//defer conn.Close()

	//Version2
	b, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Error("[clientHandleNewConn] err:", err)
	}
	clientHandleTcpMsg(b, conn)

}

//客户端对其他节点发来的tcp消息进行处理
func clientHandleTcpMsg(content []byte, conn net.Conn) {
	//先解析消息
	msg := network.ParseTCPMsg(content)
	//根据消息类别选择handle函数
	switch msg.Type {
	case commonconst.PBFTReply:
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
			refreshState(*newBC)
		}
	//接收来自主节点的区块链同步回复
	case commonconst.BlockSynResMsg:
		network.HandleBlockSynResMsg(msg, conn)
	default:
		log.Error("[clientHandleTcpMsg] invalid tcp msg type:", msg.Type)
	}
}

// 该refreshState是client使用的，client不参与共识，但是其他节点达成共识后会把区块发送给client
// 逻辑有点奇怪，后续应该会重新设计
func refreshState(b meta.Block) {
	//ste1：首先取出本区块中所有的交易
	txs := b.TX
	// 状态树的版本是区块的高度，版本号从0开始
	ver := b.Height-1
	// 需要更新到状态树的account
	var accounts []meta.Account
	// 执行每一笔交易
	for _, tx := range txs {
		clientExecute(tx, &accounts)
	}
	stateRootHash, err := merkle.UpdateAccountState(accounts, uint64(ver))
	if err != nil {
		log.Error(err)
	}
	b.StateRoot = stateRootHash.Bytes()
}

func clientExecute(tx meta.Transaction, accounts *[]meta.Account) {
	switch tx.Type {
	case meta.Register:
		*accounts = append(*accounts, account.CreateAccount(tx.To, tx.PublicKey, commonconst.InitBalance))
	case meta.Transfer:
		*accounts = append(*accounts, account.SubBalance(tx.From, tx.Value), account.AddBalance(tx.To, tx.Value))
	case meta.Publish:
		*accounts  = append(*accounts, account.CreateContract(tx.To, "", tx.Data.Code, tx.Contract))
	case meta.Invoke:
	default:
		log.Infof("未知的交易类型")
	}
}

