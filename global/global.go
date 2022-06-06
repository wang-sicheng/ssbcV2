package global

import (
	"github.com/ssbcV2/meta"
)

/*
 *	节点用到的全局变量
 */

var ChangedAccounts = []meta.JFTreeData{} // 当前区块需要更新到状态树的account
var TreeData = []meta.JFTreeData{}        // 当前区块需要更新的event，sub
var TaskList = []meta.ContractTask{}      // 当前区块智能合约执行队列

var ContractLog = make(chan interface{}, 20)	// 智能合约执行日志，会通过客户端推送到前端

/*
 * 以下参数根据命令行参数确定，不要重新赋值
 */
var RootDir string // 项目根目录
var NodeID string  // 当前节点的 nodeID，用于单机多节点运行时区分目录
var ChainID string
var Master = "" // master节点ID
var Client = "" // client节点ID
var ClientToNodeAddr = ""
var ClientToUserAddr = ""
var NodeTable = map[string]string{}
