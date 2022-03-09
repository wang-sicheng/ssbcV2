package smart_contract

import (
	"errors"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/account"
	"github.com/ssbcV2/global"
	"plugin"
)

/*
 * 区块链提供给合约的接口
 */

// 本次调用是谁发起的，即调用者地址（合约账户、外部账户）
func Caller() string {
	return curContext.Caller
}

// 返回交易发起者的地址。这是一个不随着调用深度变化的值
// 例：账户alice调用了A合约，A合约调用了B合约，B合约调用了C合约…，无论调用层次多深，执行GetOrigin()得到的都是alice的账户地址。
func Origin() string {
	return curContext.Origin
}

// 本次调用调用者给了本合约多少资产
func Value() int {
	return curContext.Value
}

// 当前合约拥有多少资产
func Balance() int {
	return curContext.Balance
}

// 返回本合约的地址
func Self() string {
	contractAccount := account.GetAccount(curContext.Name)
	return contractAccount.Address
}

// 合约向 to 账户转账
func Transfer(to string, amount int) error {
	if amount <= 0 {
		return nil
	}
	if !account.CanTransfer(curContext.Name, amount) {
		return errors.New("合约账户余额不足，无法转账")
	}
	global.ChangedAccounts = append(global.ChangedAccounts, account.SubBalance(curContext.Name, amount))
	global.ChangedAccounts = append(global.ChangedAccounts, account.AddBalance(to, amount))
	return nil
}

// 调用智能合约
func CallContract(name string, method string, args map[string]string) (interface{}, error) {
	SetRecurContext(name, method, args)
	PrintContext()

	defer func() {
		stack.Pop()
		curContext = stack.Top() // CallContract结束后获取上一层context
	}()

	// 参数校验
	if name == "" || method == "" {
		return nil, errors.New("invalid call params")
	}

	dir := "./smart_contract/contract/" + global.NodeID + "/" + name + "/"
	log.Info("call contract: " + dir)
	p, err := plugin.Open(dir + name + ".so")
	if err != nil {
		return nil, err
	}
	f, err := p.Lookup(method)
	if err != nil {
		log.Infof("找不到目标方法：%v，执行FallBack方法", method)
		f, err := p.Lookup("Fallback")
		if err != nil {
			log.Info("没有提供Fallback方法")
			return nil, err
		}
		a, _ := f.(func(map[string]string) (interface{}, error))(args)
		log.Infof("执行结果：%v\n", a)
		return a, nil
	}

	a, _ := f.(func(map[string]string) (interface{}, error))(args)
	log.Infof("执行结果：%v\n", a)
	return a, nil
}
