package contract

import (
	"errors"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/account"
	"github.com/ssbcV2/global"
)

/*
 * 区块链提供给合约的接口
 */

// 返回调用者地址（合约账户、外部账户）
func Caller() string {
	return curContext.Caller
}

// 返回交易发起者的地址。这是一个不随着调用深度变化的值
// 例：账户alice调用了A合约，A合约调用了B合约，B合约调用了C合约，Origin()得到的都是alice的账户地址。
func Origin() string {
	return curContext.Origin
}

// 返回调用合约时转入了多少资产
func Value() int {
	return curContext.Value
}

// 返回当前合约拥有多少资产
func Balance() int {
	return curContext.Balance
}

// 返回当前合约的地址
func Self() string {
	return curContext.Address
}

// 返回合约的名称
func Name() string {
	return curContext.Name
}

// 根据地址获取对应账户的余额
func GetBalance(address string) int {
	return account.GetAccount(address).Balance
}

// 当前合约向 to 账户转账
func Transfer(to string, amount int) error {
	if amount <= 0 {
		return nil
	}
	if !account.CanTransfer(curContext.Address, amount) {
		return errors.New("合约账户余额不足，无法转账")
	}
	global.ChangedAccounts = append(global.ChangedAccounts, account.SubBalance(curContext.Address, amount))
	global.ChangedAccounts = append(global.ChangedAccounts, account.AddBalance(to, amount))
	return nil
}

// 由 from 向 to 账户转账 （后续添加 Authorize()）
func TransferFrom(from, to string, amount int) error {
	if amount <= 0 {
		return nil
	}
	if !account.CanTransfer(from, amount) {
		return errors.New("合约账户余额不足，无法转账")
	}
	global.ChangedAccounts = append(global.ChangedAccounts, account.SubBalance(from, amount))
	global.ChangedAccounts = append(global.ChangedAccounts, account.AddBalance(to, amount))
	return nil
}

// 调用智能合约
func Call(name string, method string, args map[string]string) (interface{}, error) {
	log.Infof("调用 %v 合约的 %v() 方法\n", name, method)
	SetRecurContext(name, method, args, 0)
	PrintContext()

	defer func() {
		stack.Pop()
		if !stack.IsEmpty() {
			curContext = stack.Top() // Call结束后获取上一层context
		}
	}()

	res, err := execute(name, method, args)
	if err != nil {
		return nil, err
	}
	return res, err
}

// 调用智能合约的同时向合约转账
func CallWithValue(name string, method string, args map[string]string, value int) (interface{}, error) {
	log.Infof("调用 %v 合约的 %v() 方法\n", name, method)
	targetAcc := account.GetContractByName(name)
	err := Transfer(targetAcc.Address, value) // 向合约转账
	if err != nil {
		return nil, err
	}

	SetRecurContext(name, method, args, value)
	PrintContext()

	defer func() {
		stack.Pop()
		if !stack.IsEmpty() {
			curContext = stack.Top() // Call结束后获取上一层context
		}
	}()

	res, err := execute(name, method, args)
	if err != nil {
		return nil, err
	}
	return res, err
}
