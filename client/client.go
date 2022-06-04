package client

import (
	"github.com/cloudflare/cfssl/log"
	"github.com/gin-gonic/gin"
	"github.com/ssbcV2/global"
	"github.com/unrolled/secure"
)

// 监听用户请求
func ListenRequest() {
	r := gin.Default()
	r.Use(Cors()) // 使用跨域组件
	//r.Use(TlsHandler()) // 重定向为https
	r.POST("/postTran", postTran)              // 提交一笔交易
	r.GET("/registerAccount", registerAccount) // 注册账户
	r.POST("/postCrossTran", postCrossTran)    // 提交一笔跨链交易
	r.POST("/postContract", postContract)      // 提交智能合约
	r.POST("/query", query)                    // 提供链上查询服务
	r.POST("/postEvent", postEvent)            // 发起事件
	r.GET("/getLog", getLog)                   // 与前端建立websocket
	r.POST("/modelUpload", modelUpload) // 上传模型
	r.GET("/genCode", genCode)     //生成代码
	r.Run(global.ClientToUserAddr)


	log.Info(" ---------------------------------------------------------------------------------")
	log.Info("|  已启动PBFT客户端，请启动全部节点后再发送消息！  |")
	log.Info(" ---------------------------------------------------------------------------------")
}

func TlsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     "localhost:8080",
		})
		err := secureMiddleware.Process(c.Writer, c.Request)

		// If there was an error, do not continue.
		if err != nil {
			c.Abort()
			return
		}
		// Avoid header rewrite if response is a redirection.
		//if status := c.Writer.Status(); status > 300 && status < 399 {
		//	c.Abort()
		//}
		c.Next()
	}
}