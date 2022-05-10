package client

import (
	"fmt"
	"github.com/cloudflare/cfssl/log"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ssbcV2/global"
	"net/http"
	"time"
)

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 使用WebSocket向前端推送合约执行信息
func getLog(c *gin.Context) {
	// 升级请求为WebSocket协议
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)

	// 清空历史合约日志
	for len(global.ContractLog) != 0 {
		select {
		case <-global.ContractLog:
		default:
		}
	}

	if err != nil {
		log.Info("Upgrade failed")
		return
	}
	//defer ws.Close()
	// 5秒后断开websocket连接
	time.AfterFunc(5 * time.Second, func() {
		ws.Close()
	})
	for {
		result := <- global.ContractLog
		err = ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprint(result)))
		if err != nil {
			log.Info(err)
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}
