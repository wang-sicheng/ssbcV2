package contract

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/account"
	"github.com/ssbcV2/global"
	"github.com/ssbcV2/meta"
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
func Call(name string, method string, args map[string]interface{}) (interface{}, error) {
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
func CallWithValue(name string, method string, args map[string]interface{}, value int) (interface{}, error) {
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

// 调用目标链的合约，结果通过回调函数传入
func CrossCall(sourceContract, targetChain, targetContract, method string, args map[string]interface{}, callback string) (interface{}, error) {
	argsBytes , _ := json.Marshal(args)
	argsStr := string(argsBytes)

	data := map[string]string{
		"sourceChain": global.ChainID,
		//"targetChain": targetContract,
		"caller": Caller(),
		"sourceContract": sourceContract,
		"targetContract": targetContract,
		"method": method,
		"args": argsStr,
		"callback": callback,
	}
	tdataBytes, _ := json.Marshal(data)
	reqArgs := map[string]interface{}{
		"type": "chain",
		"chainName": targetChain,
		"contract": "cross",
		"function": "Call",
		"tData": string(tdataBytes),
		"address": "",
	}

	res, err := Call("oracle", "TransferData", reqArgs)
	if err != nil {
		log.Errorf("CrossCall调用TransferData预言机合约失败：%s", err)
		return nil, err
	}
	return res, nil
}

func Info(info ...interface{}) {
	log.Info(info)
	if global.NodeID == global.Client {
		global.ContractLog <- fmt.Sprint(info...)
	}
}

func Infof(format string, info ...interface{}) {
	log.Infof(format, info)
	if global.NodeID == global.Client {
		global.ContractLog <- fmt.Sprintf(format, info...)
	}
}

/* 获取跨链合约数据
targetChain: 目标链名
contractName: 合约名称
dataName: 目标数据
callback: 数据传回时调用的方法（确保当前合约存在该方法）
*/
func GetCrossContractData(targetChain, contractName, dataName, callback string) (interface{}, error) {
	cb := meta.Callback{
		Caller:   "",
		Value:    0,
		Contract: Name(),
		Method:   callback,
		Args:     nil,
		Address:  "",
	}
	cbBytes, _ := json.Marshal(cb)
	reqArgs := map[string]interface{}{
		"type":     "chain", // "api":第三方接口，"chain":"跨链数据"
		"callback": string(cbBytes),
		"name":     targetChain,
		"dataType": "contractData",
		"params":   contractName + "," + dataName,
	}

	// 调用QueryData预言机合约请求外部数据
	res, err := Call("oracle", "QueryData", reqArgs)
	if err != nil {
		Infof("call QueryData contract error: %s", err)
		return nil, err
	}
	return res, nil
}

func ToBytes(s interface{}) []byte {
	data, err := json.Marshal(s)
	if err != nil {
		Infof("json.Marshal error %v:", err)
	}
	return data
}
