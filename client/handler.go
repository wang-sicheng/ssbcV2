package client

import (
	"encoding/hex"
	"encoding/json"
	"github.com/cloudflare/cfssl/log"
	"github.com/gin-gonic/gin"
	"github.com/ssbcV2/account"
	"github.com/ssbcV2/chain"
	"github.com/ssbcV2/common"
	"github.com/ssbcV2/contract"
	"github.com/ssbcV2/event"
	"github.com/ssbcV2/global"
	"github.com/ssbcV2/levelDB"
	"github.com/ssbcV2/meta"
	"github.com/ssbcV2/pbft"
	"github.com/ssbcV2/util"
	"net/http"
	"strconv"
	"time"
)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		origin := c.Request.Header.Get("Origin")

		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization") //自定义 Header
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")

		}

		if method == "OPTIONS" {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization") //自定义 Header
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.AbortWithStatus(http.StatusNoContent)
		}

		c.Next()
	}
}

// 提交智能合约代码
func postContract(ctx *gin.Context) {
	postC := meta.ContractPost{}
	_ = ctx.ShouldBind(&postC)

	from := postC.Account
	if !account.ContainsAddress(from) {
		log.Error("发起地址不存在")
		hr := errResponse("发起地址不存在")
		ctx.JSON(http.StatusOK, hr)
		return
	}

	// 获取合约名称
	contractName := postC.Name
	if contractName == "" {
		log.Error("合约名称不能为空")
		hr := errResponse("合约名称不能为空")
		ctx.JSON(http.StatusOK, hr)
		return
	}
	if account.ContainsContract(contractName) {
		log.Error("该合约已存在")
		hr := errResponse("同名合约已存在")
		ctx.JSON(http.StatusOK, hr)
		return
	}

	// 静态代码检测和模型检测
	result, err := check(postC.Code)
	if err != nil {
		hr := errResponse(result)
		ctx.JSON(http.StatusOK, hr)
		return
	}

	// 封装为交易发送至主节点，经共识后真正部署
	go sendNewContract(postC)
	hr := goodResponse("")
	ctx.JSON(http.StatusOK, hr)
}

//将部署封装为交易发送至主节点
func sendNewContract(c meta.ContractPost) {
	data := meta.TransactionData{}
	data.Code = c.Code
	t := meta.Transaction{
		From:      c.Account,
		To:        generateContractAddress(),
		Dest:      "",
		Contract:  c.Name,
		Method:    "",
		Args:      nil,
		Data:      data,
		Value:     0,
		Id:        nil,
		Timestamp: "",
		Hash:      nil,
		PublicKey: c.PublicKey,
		Sign:      nil,
		Type:      meta.Publish,
	}
	//客户端在转发交易之前需要对交易进行签名
	//先将交易进行hash
	tByte, _ := json.Marshal(t)
	t.Hash, _ = util.CalculateHash(tByte)
	//t.Sign=RsaSignWithSha256(t.Hash,[]byte(c.PrivateKey))
	//客户端需要把交易信息发送给主节点
	r := new(pbft.Request)
	r.Timestamp = time.Now().UnixNano()
	r.ClientAddr = global.ClientToNodeAddr
	r.Message.ID = util.GetRandom()
	r.Type = 0
	tb, _ := json.Marshal(t)
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
	//默认N0为主节点，直接把请求信息发送至N0
	util.TCPSend(msg, global.NodeTable[global.Master])
}

//账户注册
func registerAccount(ctx *gin.Context) {
	//首先生成公私钥
	priKey, pubKey := util.GetKeyPair()
	//账户地址
	//将公钥进行hash
	pubHash, _ := util.CalculateHash(pubKey)
	log.Infof("public hash len: %d", len(pubHash))
	//将公钥hash作为账户地址,256位
	account := hex.EncodeToString(pubHash)
	log.Infof("account address len: %d", len(account))
	res := meta.ChainAccount{
		AccountAddress: account,
		PublicKey:      string(pubKey),
		PrivateKey:     string(priKey),
	}
	// client 存储账户的私钥
	levelDB.DBPut(account+common.AccountsPrivateKeySuffix, priKey)

	// 将交易类型设置为Register
	t := meta.Transaction{
		From:      account,
		To:        account,
		Dest:      "",
		Contract:  "",
		Method:    "",
		Args:      nil,
		Data:      meta.TransactionData{},
		Value:     common.InitBalance,
		Id:        nil,
		Timestamp: "",
		Hash:      nil,
		PublicKey: string(pubKey),
		Sign:      nil,
		Type:      meta.Register,
	}
	//客户端在转发交易之前需要对交易进行签名
	//先将交易进行hash
	tByte, _ := json.Marshal(t)
	t.Hash, _ = util.CalculateHash(tByte)
	//t.Sign=RsaSignWithSha256(t.Hash,[]byte(pt.PrivateKey))
	//客户端需要把交易信息发送给主节点
	r := new(pbft.Request)
	r.Timestamp = time.Now().UnixNano()
	r.ClientAddr = global.ClientToNodeAddr
	r.Message.ID = util.GetRandom()
	r.Type = 0

	tb, _ := json.Marshal(t)
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
	//默认N0为主节点，直接把请求信息发送至N0
	util.TCPSend(msg, global.NodeTable[global.Master])
	//返回提交成功
	hr := goodResponse(res)
	ctx.JSON(http.StatusOK, hr)
}

//链上信息query服务
func query(ctx *gin.Context) {
	data, _ := ctx.GetRawData()
	log.Infof("[client] 收到查询请求: %s\n", string(data))

	q := meta.Query{}
	err := json.Unmarshal(data, &q)
	if err != nil {
		log.Error("[query],json decode err:", err)
	}

	var response meta.HttpResponse
	switch q.Type {
	case "getBlockChain": // 获取区块链
		bcs := chain.GetCurrentBlockChain()
		response = goodResponse(bcs)

	case "getBlock": // 获取指定高度的区块
		height := q.Parameters[0]
		hInt64, err := strconv.ParseInt(height, 10, 64)
		if err != nil {
			log.Error("[getBlock],parseInt err:", err)
			panic(err)
		}
		hInt := int(hInt64)
		bc := chain.GetBlock(hInt)
		if bc == nil {
			response = errResponse("Invalid param")
		} else {
			response = goodResponse(bc)
		}

	case "getAllTxs": // 获取所有的交易
		all := chain.GetAllTransactions()
		response = goodResponse(all)

	case "getAllAccounts": // 获取所有的账户
		all := []meta.Account{}
		for _, address := range account.GetTotalAddress() {
			account := account.GetAccount(address)
			// 私钥从 client 本地获取
			account.PrivateKey = string(levelDB.DBGet(address + common.AccountsPrivateKeySuffix))
			all = append(all, account)
		}
		response = goodResponse(all)

	case "getOneBlockTxs": // 获取指定高度的区块的所有交易
		h := q.Parameters[0]
		hInt64, err := strconv.ParseInt(h, 10, 64)
		if err != nil {
			log.Error("[getBlock],parseInt err:", err)
			panic(err)
		}
		hInt := int(hInt64)
		bc := chain.GetBlock(hInt)
		if bc == nil {
			response = errResponse("Invalid param")
		} else {
			trans := bc.TX
			response = goodResponse(trans)
		}
	case "contractData":  // 获取合约内的数据
		if q.Parameters == nil || len(q.Parameters) < 2 {
			response = errResponse("参数错误")
			log.Info("获取合约内数据失败")
			break
		}
		name := q.Parameters[0]
		target := q.Parameters[1:]

		res, err := contract.Get(name, target)
		if err != nil {
			response = errResponse("获取合约内数据失败")
			log.Info("获取合约内数据失败", err)
		} else {
			response = goodResponse(res)
		}
	case "getEvent":
		data, _ := event.GetAllEventData()
		response = goodResponse(data)
	default:
		log.Info("Query参数有误!")
		response = errResponse("Query参数有误!")
	}

	ctx.JSON(http.StatusOK, response)
}

func postEvent(ctx *gin.Context) {
	b, _ := ctx.GetRawData()
	params := meta.EventMessageParams{}
	err := json.Unmarshal(b, &params)
	if err != nil {
		log.Errorf("[postEvent], json decode err: %s", err)
		return
	}
	log.Infof("postEvent params: %+v", params)
	var args map[string]string
	err = json.Unmarshal([]byte(params.Args), &args)
	if err != nil {
		log.Errorf("[event args], json decode err: %s", err)
		return
	}
	em := meta.EventMessage{
		From:      params.From,
		EventID:   params.EventKey,
		Data:      args,
		Sign:      nil, // TODO:增加签名
		PublicKey: params.PublicKey,
		TimeStamp: "",
		Hash:      nil,
	}
	req := pbft.Request{
		Message:    pbft.Message{},
		Timestamp:  time.Now().UnixNano(),
		ClientAddr: global.ClientToNodeAddr,
	}
	emBytes, _ := json.Marshal(em)
	req.Message.Content = string(emBytes)
	req.Message.ID = util.GetRandom()
	req.Type = 1
	reqBytes, _ := json.Marshal(req)
	msg := meta.TCPMessage{
		Type:    common.PBFTRequest,
		Content: reqBytes,
		From:    "",
		To:      "",
	}
	util.TCPSend(msg, global.NodeTable[global.Master])
	hr := goodResponse("")
	ctx.JSON(http.StatusOK, hr)
}

//提交一笔交易
func postTran(ctx *gin.Context) {
	b, _ := ctx.GetRawData()
	log.Infof("[client] 收到一笔交易: %s\n", string(b))

	pt := meta.PostTran{}
	//err := ctx.ShouldBind(&pt)
	err := json.Unmarshal(b, &pt)
	//err := ctx.BindJSON(&pt)
	if err != nil {
		log.Error("[postTran],json decode err:", err)
	}

	// 检查交易参数
	if msg, ok := checkTranParameters(&pt); !ok {
		hr := errResponse(msg)
		log.Infof(msg + "\n")
		ctx.JSON(http.StatusOK, hr)
		return
	}

	//将args解析
	args := make(map[string]string)
	err = json.Unmarshal([]byte(pt.Args), &args)
	if err != nil {
		log.Error("[postTran] json err:", err)
	}
	log.Infof("合约参数：%v\n", args)
	t := meta.Transaction{
		From:      pt.From,
		To:        pt.To,
		Dest:      pt.Dest,
		Contract:  pt.Contract,
		Method:    pt.Method,
		Args:      args,
		Data:      meta.TransactionData{},
		Value:     pt.Value,
		Id:        nil,
		Timestamp: "",
		Hash:      nil,
		PublicKey: pt.PublicKey,
		Sign:      nil,
		Type:      pt.Type,
	}
	//客户端在转发交易之前需要对交易进行签名
	//先将交易进行hash
	tByte, _ := json.Marshal(t)
	t.Hash, _ = util.CalculateHash(tByte)
	//t.Sign=RsaSignWithSha256(t.Hash,[]byte(pt.PrivateKey))
	//客户端需要把交易信息发送给主节点
	r := new(pbft.Request)
	r.Timestamp = time.Now().UnixNano()
	r.ClientAddr = global.ClientToNodeAddr
	r.Message.ID = util.GetRandom()
	r.Type = 0

	tb, _ := json.Marshal(t)
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
	//默认N0为主节点，直接把请求信息发送至N0
	util.TCPSend(msg, global.NodeTable[global.Master])
	//返回提交成功
	hr := goodResponse("")
	ctx.JSON(http.StatusOK, hr)
}

func postCrossTran(ctx *gin.Context) {
	b, _ := ctx.GetRawData()
	log.Infof("[client] 收到跨链交易: %s\n", string(b))

	pt := meta.PostCrossTran{}
	//err := ctx.ShouldBind(&pt)
	err := json.Unmarshal(b, &pt)
	//err := ctx.BindJSON(&pt)
	if err != nil {
		log.Error("[postTran],json decode err:", err)
	}

	// 检查交易参数
	if msg, ok := checkCrossTranParameters(&pt); !ok {
		hr := errResponse(msg)
		log.Infof(msg + "\n")
		ctx.JSON(http.StatusOK, hr)
		return
	}

	t := meta.Transaction{
		SourceChainId: pt.SourceChain,
		DestChainId:   pt.DestChain,
		From:          pt.From,
		To:            pt.To,
		Value:         pt.Value,
		Timestamp:     "",
		PublicKey:     pt.PublicKey,
		Type:          meta.CrossTransfer,
	}
	//客户端在转发交易之前需要对交易进行签名
	//先将交易进行hash
	//tByte, _ := json.Marshal(t)
	//t.Hash, _ = util.CalculateHash(tByte)
	//t.Sign=RsaSignWithSha256(t.Hash,[]byte(pt.PrivateKey))
	//客户端需要把交易信息发送给主节点
	r := new(pbft.Request)
	r.Timestamp = time.Now().UnixNano()
	r.ClientAddr = global.ClientToNodeAddr
	r.Message.ID = util.GetRandom()
	r.Type = 0

	tb, _ := json.Marshal(t)
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
	//默认N0为主节点，直接把请求信息发送至N0
	util.TCPSend(msg, global.NodeTable[global.Master])
	//返回提交成功
	hr := goodResponse("")
	ctx.JSON(http.StatusOK, hr)
}

// 正常响应，返回数据
func goodResponse(data interface{}) meta.HttpResponse {
	res := meta.HttpResponse{
		Data: data,
		Code: 20000,
	}
	return res
}

// 出现异常，返回异常信息
func errResponse(errMsg string) meta.HttpResponse {
	res := meta.HttpResponse{
		Error: errMsg,
		Data:  "",
		Code:  20000,
	}
	return res
}

// 检查交易参数
func checkTranParameters(pt *meta.PostTran) (string, bool) {
	if pt.From == "" {
		return "发起地址不能为空", false
	}

	if pt.From == pt.To {
		return "发起地址和接收地址不能相同", false
	}

	if pt.PublicKey == "" {
		return "公钥不能为空", false
	}

	// 调用合约
	if pt.Contract != "" {
		if pt.Method == "" {
			return "方法不能为空", false
		}
		return "", true
	}

	// 转账交易
	if pt.Value <= 0 {
		return "转账金额必须为正整数", false
	}

	// 确保发起地址已存在
	if !account.ContainsAddress(pt.From) {
		return "发起地址不存在", false
	}

	// 确保接收地址已存在
	if !account.ContainsAddress(pt.To) {
		return "接收地址不存在", false
	}
	return "", true
}

// 检查跨链交易参数
func checkCrossTranParameters(pt *meta.PostCrossTran) (string, bool) {
	if pt.From == "" {
		return "发起地址不能为空", false
	}

	if pt.From == pt.To {
		return "发起地址和接收地址不能相同", false
	}

	if pt.PublicKey == "" {
		return "公钥不能为空", false
	}

	// 调用合约
	if pt.Contract != "" {
		if pt.Method == "" {
			return "方法不能为空", false
		}
		return "", true
	}

	// 转账交易
	if pt.Value <= 0 {
		return "转账金额必须为正整数", false
	}

	// 确保发起地址已存在，接收地址本链无法确定
	if !account.ContainsAddress(pt.From) {
		return "发起地址不存在", false
	}

	if len(pt.To) == 0 {
		return "接收地址不能为空", false
	}
	return "", true
}

// 生成合约地址（虽然合约地址不应该由公私钥生成）
func generateContractAddress() string {
	//首先生成公私钥
	_, pubKey := util.GetKeyPair()
	//账户地址
	//将公钥进行hash
	pubHash, _ := util.CalculateHash(pubKey)
	//将公钥hash作为账户地址,256位
	address := hex.EncodeToString(pubHash)
	log.Infof("contract account address len: %d", len(address))
	return address
}
