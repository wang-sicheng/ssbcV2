package smart_contract

import (
	"errors"
	"github.com/ssbcV2/account"
	"github.com/ssbcV2/global"
	"github.com/ssbcV2/meta"
)

/*
 * 区块链提供给合约的接口
 */

var Name string  			// 当前执行的合约的名称
var Balance int				// 合约账户的余额
var Caller	string			// 调用者地址（合约账户、外部账户、事件）
var Value	int				// 调用合约交易的转账金额（如果是非调用合约则没有）

// 合约向 to 账户转账
func Transfer(to string, amount int) (interface{}, error){
	if amount <= 0 {
		return nil, nil
	}
	if !account.CanTransfer(Name, amount) {
		return nil, errors.New("合约账户余额不足，无法转账")
	}
	global.ChangedAccounts = append(global.ChangedAccounts, account.SubBalance(Name, amount))
	global.ChangedAccounts = append(global.ChangedAccounts, account.AddBalance(to, amount))
	return nil, nil
}

// 合约调用合约（暂时不用：因为合约调用合约需要同步拿到结果，但是将task放到合约队列里是异步执行的）
func Call(callee string, method string, args map[string]string) (interface{}, error) {
	// 添加到合约队列
	global.TaskList = append(global.TaskList, meta.ContractTask{
		Caller: Name,
		Name:   callee,
		Method: method,
		Args:   args,
	})
	return nil, nil
}

// 调用合约前加载合约信息，暂时先放这里
func LoadInfo(task meta.ContractTask) {
	contractAccount := account.GetAccount(task.Name)
	Name = task.Name
	Balance = contractAccount.Balance
	Caller = task.Caller
	Value = task.Value
}





