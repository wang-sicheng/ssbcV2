package main

import (
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/account"
	"github.com/ssbcV2/chain"
	"github.com/ssbcV2/client"
	"github.com/ssbcV2/common"
	"github.com/ssbcV2/levelDB"
	"github.com/ssbcV2/merkle"
	"github.com/ssbcV2/meta"
	"github.com/ssbcV2/network"
	"github.com/ssbcV2/pbft"
	"github.com/ssbcV2/util"
	"os"
)

func main() {
	// 为四个节点生成公私钥
	util.GenRsaKeys()

	if len(os.Args) != 2 {
		log.Error("输入的参数有误！")
	}
	nodeID := os.Args[1]
	merkle.StatePath = "./levelDB/db/path/statedb/" + nodeID // 账户数据暂时使用单独的数据库存储
	// 数据库连接
	levelDB.InitDB(nodeID)

	// 从levelDB读取账户信息（必须在数据库建立连接后，所以不能在init()完成）
	account.GetFromDisk()

	if nodeID == "client" {
		go client.ListenRequest() 	// 启动客户端程序
		p := pbft.NewPBFT(nodeID, common.ClientToNodeAddr)
		go p.TcpListen()

		//初始化
		initBlockChain(nodeID)
	} else if addr, ok := common.NodeTable[nodeID]; ok {
		p := pbft.NewPBFT(nodeID, addr)
		go p.TcpListen() //启动节点
		//初始化
		initBlockChain(nodeID)
	} else {
		log.Fatal("无此节点编号！")
	}
	select {}
}

//初始化
func initBlockChain(ID string) {
	chain.BlockChain = make([]meta.Block, 0)
	//先判断节点是否为主节点
	if ID == "N0" {
		//判断是否无创世区块
		bc := chain.GetCurrentBlockChain()
		if len(bc) == 0 {
			gb := chain.GenerateGenesisBlock()
			chain.BlockChain = append(chain.BlockChain, gb)
			chain.StoreBlockChain(chain.BlockChain)
		} else {
			chain.BlockChain = bc
		}
	} else {
		//非主节点初始化时需要和主节点进行区块信息同步
		//先生成区块链同步请求消息，再发送
		msg := network.GenBlockSynReqMsg(ID)
		log.Info("发送区块链同步消息,msg:", msg, "addr:", common.NodeTable["N0"])
		network.TCPSend(msg, common.NodeTable["N0"])
	}
}
