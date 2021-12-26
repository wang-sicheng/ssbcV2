package pbft

import (
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/common"
)

func ClientSendMessageAndListen() {
	//开始用户请求的监听
	go clientHttpListenV2()
	//开启客户端的本地监听（主要用来接收节点的reply信息）
	go clientTcpListen()
	log.Info("客户端开启监听，地址：", common.ClientToNodeAddr)
	log.Info(" ---------------------------------------------------------------------------------")
	log.Info("|  已进入PBFT客户端，请启动全部节点后再发送消息！  |")
	log.Info(" ---------------------------------------------------------------------------------")
}
