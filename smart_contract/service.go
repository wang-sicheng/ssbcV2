package smart_contract

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/ssbcV2/account"
	"github.com/ssbcV2/global"
	"github.com/ssbcV2/meta"
	"log"
)

/*
 * 区块链提供给合约的接口
 */

type context struct {
	Name    string            // 当前执行的合约的名称
	Method  string            // 被调用的方法
	Args    map[string]string // 参数
	Balance int               // 合约账户的余额
	Caller  string            // 调用者地址（合约账户、外部账户）
	Origin  string            // 最初调用者（外部账户），如果不涉及合约调用合约，那么 Caller == Origin
	Value   int               // 调用合约交易的转账金额（如果是非调用合约则暂时没有）
}

// 合约调用栈
type contextStack struct {
	contexts []context
}

func (t *contextStack) Push(c context) {
	t.contexts = append(t.contexts, c)
}

func (t *contextStack) Pop() context {
	if !t.IsEmpty() {
		top := t.contexts[len(t.contexts)-1]
		t.contexts = t.contexts[:len(t.contexts)-1]
		return top
	}
	return context{}
}

func (t *contextStack) Top() context {
	if !t.IsEmpty() {
		return t.contexts[len(t.contexts)-1]
	}
	return context{}
}
func (t *contextStack) IsEmpty() bool {
	if len(t.contexts) > 0 {
		return false
	}
	return true
}

var stack contextStack // 合约调用栈
var curContext context // 当前调用的context

func init() {
	curContext = context{}
	stack = contextStack{[]context{}}
}

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

// 返回本合约的地址
func Self() string {
	contractAccount := account.GetAccount(curContext.Name)
	return contractAccount.Address
}

// 合约向 to 账户转账
func Transfer(to string, amount int) (interface{}, error) {
	if amount <= 0 {
		return nil, nil
	}
	if !account.CanTransfer(curContext.Name, amount) {
		return nil, errors.New("合约账户余额不足，无法转账")
	}
	global.ChangedAccounts = append(global.ChangedAccounts, account.SubBalance(curContext.Name, amount))
	global.ChangedAccounts = append(global.ChangedAccounts, account.AddBalance(to, amount))
	return nil, nil
}

func PrintContext() {
	bs, _ := json.Marshal(curContext)
	var out bytes.Buffer
	_ = json.Indent(&out, bs, "", "\t")
	log.Printf("当前合约调用的context: %v\n", out.String())
}

// 第一次调用合约前加载合约信息
func SetContext(task meta.ContractTask) {
	contractAccount := account.GetAccount(task.Name)
	curContext.Name = task.Name
	curContext.Balance = contractAccount.Balance
	curContext.Caller = task.Caller
	curContext.Origin = task.Caller
	curContext.Value = task.Value
	curContext.Method = task.Method
}

// 合约调用合约时设置合约信息
func SetRecurContext(name string, method string, args map[string]string) {
	if len(stack.contexts) == 0 { // 用户调用合约时（第一次调用）不执行该函数
		stack.Push(curContext) // context设置完毕，入栈
		return
	}
	curContext.Caller = curContext.Name // 调用者为上一个合约
	curContext.Name = name
	curContext.Method = method
	curContext.Args = args

	curContext.Value = 0 // 目前暂不支持合约转合约

	contract := account.GetAccount(name)
	curContext.Balance = contract.Balance

	stack.Push(curContext) // context设置完毕，入栈
}
