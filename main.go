package main

import (
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/account"
	"github.com/ssbcV2/chain"
	"github.com/ssbcV2/client"
	"github.com/ssbcV2/common"
	"github.com/ssbcV2/global"
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
		return
	}
	nodeID := os.Args[1]

	// 删除 levelDB/path 和 smart_contract/contract 目录
	if nodeID == "clear" {
		clear()
		return
	}

	// 不存在该节点编号
	if !util.Contains(common.Ssbc1Nodes, nodeID) &&
		!util.Contains(common.Ssbc2Nodes, nodeID) {
		log.Info("无此节点编号！")
		return
	}

	merkle.AccountStatePath = "./levelDB/db/path/account/" + nodeID // 账户状态和事件状态分开存储
	merkle.EventStatePath = "./levelDB/db/path/event/" + nodeID
	// 数据库连接
	levelDB.InitDB(nodeID)

	// 从levelDB读取账户信息（必须在数据库建立连接后，所以不能在init()完成）
	account.GetFromDisk()

	if util.Contains(common.Ssbc1Nodes, nodeID) {
		global.ChainID = "ssbc1"
		global.Master = "N0"
		global.Client = "client1"
		global.NodeTable = common.NodeTable1
		global.ClientToNodeAddr = common.Client1ToNodeAddr
		global.ClientToUserAddr = common.Client1ToUserAddr

		if nodeID == "client1" {
			go client.ListenRequest() // 启动客户端程序
			p := pbft.NewPBFT(nodeID, common.Client1ToNodeAddr)
			go p.TcpListen()
		} else if addr, ok := common.NodeTable1[nodeID]; ok {
			p := pbft.NewPBFT(nodeID, addr)
			go p.TcpListen() //启动节点
		}
	}

	if util.Contains(common.Ssbc2Nodes, nodeID) {
		global.ChainID = "ssbc2"
		global.Master = "N4"
		global.Client = "client2"
		global.NodeTable = common.NodeTable2
		global.ClientToNodeAddr = common.Client2ToNodeAddr
		global.ClientToUserAddr = common.Client2ToUserAddr

		if nodeID == "client2" {
			go client.ListenRequest() // 启动客户端程序
			p := pbft.NewPBFT(nodeID, common.Client2ToNodeAddr)
			go p.TcpListen()
		} else if addr, ok := common.NodeTable2[nodeID]; ok {
			p := pbft.NewPBFT(nodeID, addr)
			go p.TcpListen() //启动节点
		}
	}

	// 初始化
	initBlockChain(nodeID)
	global.NodeID = nodeID

	select {}
}

//初始化
func initBlockChain(ID string) {
	chain.BlockChain = make([]meta.Block, 0)
	var accounts []meta.JFTreeData
	var events []meta.JFTreeData

	bc := chain.GetCurrentBlockChain()
	// 初始化创世区块前，所有节点先初始化0版本的state
	if len(bc) == 0 {
		merkle.InitAccount.Address = "init account address"
		merkle.InitEvent.EventID = "init event id"
		accounts = append(accounts, merkle.InitAccount)
		events = append(events, merkle.InitEvent)
		stateRootHash, _ := merkle.UpdateStateTree(accounts, 0, merkle.AccountStatePath)
		eventRootHash, _ := merkle.UpdateStateTree(events, 0, merkle.EventStatePath)
		if ID == global.Master {
			gb := chain.GenerateGenesisBlock()
			gb.StateRoot = stateRootHash.Bytes()
			gb.EventRoot = eventRootHash.Bytes()
			chain.BlockChain = append(chain.BlockChain, gb)
			chain.StoreBlockChain(chain.BlockChain)
		}
	}
	if ID == global.Master {
		if len(bc) != 0 {
			chain.BlockChain = bc
		}
	} else {
		//非主节点初始化时需要和主节点进行区块信息同步
		//先生成区块链同步请求消息，再发送
		msg := network.GenBlockSynReqMsg(ID)
		log.Info("发送区块链同步消息,msg:", msg, "addr:", global.NodeTable[global.Master])
		util.TCPSend(msg, global.NodeTable[global.Master])
	}
}

// 清空数据
func clear() {
	os.RemoveAll("./levelDB/db/path/")
	os.RemoveAll("./smart_contract/contract/")
	log.Info("成功删除数据!")
}
