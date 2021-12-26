package client

import (
	"github.com/cloudflare/cfssl/log"
	"github.com/gin-gonic/gin"
	"github.com/ssbcV2/common"
)

// 监听用户请求
func ListenRequest() {
	r := gin.Default()
	//使用跨域组件
	r.Use(Cors())
	//获取到当前区块链
	r.GET("/getBlockChain", getBlockChain)
	//获取到指定高度的区块
	r.GET("/getBlock", getBlock)
	//获取到指定高度区块的交易列表
	r.GET("/getOneBlockTrans", getBlockTrans)
	//提交一笔交易
	r.POST("/postTran", postTran)
	//获取全部交易
	r.GET("/getAllTrans", getAllTrans)
	r.GET("/getAllAccounts", getAllAccounts)
	//注册账户
	r.GET("/registerAccount", registerAccount)
	//提交一笔跨链交易
	//http.HandleFunc("/postCrossTran", server.postCrossTran)
	//提交智能合约
	r.POST("/postContract", postContract)
	//提供链上query服务--既能服务于普通节点也能服务于智能合约
	r.GET("/query", query)
	//r.POST("/decrypt",decrypt)
	// 发起事件
	r.POST("/postEvent", postEvent)
	r.Run(common.ClientToUserAddr)

	log.Info(" ---------------------------------------------------------------------------------")
	log.Info("|  已启动PBFT客户端，请启动全部节点后再发送消息！  |")
	log.Info(" ---------------------------------------------------------------------------------")
}
