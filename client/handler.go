package client

import (
	"encoding/hex"
	"encoding/json"
	"github.com/cloudflare/cfssl/log"
	"github.com/gin-gonic/gin"
	"github.com/ssbcV2/account"
	"github.com/ssbcV2/chain"
	"github.com/ssbcV2/common"
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
		hr := warpGoodHttpResponse("发起地址不存在")
		ctx.JSON(http.StatusOK, hr)
		return
	}

	// 获取合约名称
	contractName := postC.Name
	if contractName == "" {
		log.Error("合约名称不能为空")
		hr := warpGoodHttpResponse("合约名称不能为空")
		ctx.JSON(http.StatusOK, hr)
		return
	}
	if account.ContainsAddress(contractName) {
		log.Error("该合约已存在")
		hr := warpGoodHttpResponse("同名合约已存在")
		ctx.JSON(http.StatusOK, hr)
		return
	}

	// 封装为交易发送至主节点，经共识后真正部署
	go sendNewContract(postC)
	hr := warpGoodHttpResponse(common.Success)
	ctx.JSON(http.StatusOK, hr)
}

//将部署封装为交易发送至主节点
func sendNewContract(c meta.ContractPost) {
	data := meta.TransactionData{}
	data.Code = c.Code
	t := meta.Transaction{
		From:      c.Account,
		To:        common.ContractDeployAddress,
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
	res := struct {
		PrivateKey     string
		PublicKey      string
		AccountAddress string
	}{
		string(priKey),
		string(pubKey),
		account,
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
	hr := warpGoodHttpResponse(res)
	ctx.JSON(http.StatusOK, hr)
}

//链上信息query服务
func query(ctx *gin.Context) {
	//测试用，之后需要删掉(库也需要删)
	tests := make([]meta.AbstractBlockHeader, 0)
	test := meta.AbstractBlockHeader{
		ChainId:    "ssbc2",
		Height:     0,
		Hash:       []byte("hello"),
		PreHash:    []byte("hello"),
		MerkleRoot: []byte("hello"),
	}
	tests = append(tests, test)
	tests = append(tests, test)
	testKey := "abstract_block_header_store_key_ssbc2"
	testsByte, _ := json.Marshal(tests)
	levelDB.DBPut(testKey, testsByte)

	//获取到查询参数
	queryKey := ctx.Query("queryKey")
	//根据查询key去库中查询数据
	val := levelDB.DBGet(queryKey)
	log.Info("链上数据服务查询结果:", string(val))
	hr := warpGoodHttpResponse(val)
	ctx.JSON(http.StatusOK, hr)
}

//获取全部的交易
func getAllTrans(ctx *gin.Context) {
	all := chain.GetAllTransactions()
	hr := warpGoodHttpResponse(all)
	ctx.JSON(http.StatusOK, hr)
}

func getAllAccounts(ctx *gin.Context) {
	all := []meta.Account{}
	for _, address := range account.GetTotalAddress() {
		account := account.GetAccount(address)
		// 私钥从 client 本地获取
		account.PrivateKey = string(levelDB.DBGet(address + common.AccountsPrivateKeySuffix))
		all = append(all, account)
	}
	hr := warpGoodHttpResponse(all)
	ctx.JSON(http.StatusOK, hr)
}

func postEvent(ctx *gin.Context) {
	b, _ := ctx.GetRawData()
	params := meta.EventMessageParams{}
	err := json.Unmarshal(b, &params)
	if err != nil {
		log.Errorf("[postEvent], json decode err: %s", err)
		return
	}
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
	hr := warpGoodHttpResponse(common.Success)
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
		hr := warpGoodHttpResponse(msg)
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
	if t.Type == meta.Invoke {
		t.To = account.GetAccount(t.Contract).Address
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
	hr := warpGoodHttpResponse(common.Success)
	ctx.JSON(http.StatusOK, hr)
}

//用户查询当前所有区块-->获取当前的区块链
func getBlockChain(ctx *gin.Context) {
	//获取当前区块链
	bcs := chain.GetCurrentBlockChain()
	hr := warpGoodHttpResponse(bcs)
	ctx.JSON(http.StatusOK, hr)
}

//用户根据区块高度获取到某一个区块
func getBlock(ctx *gin.Context) {
	//获取到请求参数中的区块高度
	h := ctx.Query("height")
	hInt64, err := strconv.ParseInt(h, 10, 64)
	if err != nil {
		log.Error("[getBlock],parseInt err:", err)
		panic(err)
	}
	hInt := int(hInt64)
	bc := chain.GetBlock(hInt)
	if bc == nil {
		hr := warpGoodHttpResponse("Invalid param")
		ctx.JSON(http.StatusBadRequest, hr)
	} else {
		hr := warpGoodHttpResponse(bc)
		ctx.JSON(http.StatusOK, hr)
	}
}

//用户获取到某一区块中的所有交易
func getBlockTrans(ctx *gin.Context) {
	//先解析出请求中的区块高度
	h := ctx.Query("height")
	hInt64, err := strconv.ParseInt(h, 10, 64)
	if err != nil {
		log.Error("[getBlock],parseInt err:", err)
		panic(err)
	}
	hInt := int(hInt64)
	bc := chain.GetBlock(hInt)
	if bc == nil {
		hr := warpGoodHttpResponse("Invalid param")
		ctx.JSON(http.StatusBadRequest, hr)
	} else {
		trans := bc.TX
		hr := warpGoodHttpResponse(trans)
		ctx.JSON(http.StatusOK, hr)
	}
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
		hr := warpGoodHttpResponse(msg)
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
	hr := warpGoodHttpResponse(common.Success)
	ctx.JSON(http.StatusOK, hr)
}

func warpGoodHttpResponse(data interface{}) meta.HttpResponse {
	res := meta.HttpResponse{
		StatusCode: http.StatusOK,
		Data:       data,
		Code:       20000,
	}
	return res
}
func warpBadHttpResponse(data interface{}) meta.HttpResponse {
	res := meta.HttpResponse{
		StatusCode: http.StatusBadRequest,
		Data:       data,
	}
	return res
}

func warpHttpResponse(status int, data interface{}) meta.HttpResponse {
	res := meta.HttpResponse{
		StatusCode: status,
		Data:       data,
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
