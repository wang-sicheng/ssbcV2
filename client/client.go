package client

import (
	"github.com/cloudflare/cfssl/log"
	"github.com/gin-gonic/gin"
	"github.com/ssbcV2/global"
)

// 监听用户请求
func ListenRequest() {
	r := gin.Default()
	r.Use(Cors()) // 使用跨域组件

	r.POST("/postTran", postTran)              // 提交一笔交易
	r.GET("/registerAccount", registerAccount) // 注册账户
	r.POST("/postCrossTran", postCrossTran)    // 提交一笔跨链交易
	r.POST("/postContract", postContract)      // 提交智能合约
	r.POST("/query", query)                    // 提供链上查询服务
	r.POST("/postEvent", postEvent)            // 发起事件
	r.GET("/getLog", getLog)                   // 与前端建立websocket
	r.Run(global.ClientToUserAddr)
	r.POST("modelUpload", modelUpload) // 上传模型
	r.GET("/genCode", modelUpload)     //生成代码

	log.Info(" ---------------------------------------------------------------------------------")
	log.Info("|  已启动PBFT客户端，请启动全部节点后再发送消息！  |")
	log.Info(" ---------------------------------------------------------------------------------")
}
