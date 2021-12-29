package global

import "github.com/ssbcV2/meta"

/*
 *	节点用到的全局变量
 */

var ChangedAccounts = []meta.JFTreeData{}		// 当前区块需要更新到状态树的account
var TreeData = []meta.JFTreeData{}			// 当前区块需要更新的event，sub
var TaskList = []meta.ContractTask{}		// 当前区块智能合约执行队列
