#### 智能合约接口

在合约中需要引入 contract 包
```golang
import "github.com/ssbcV2/contract"
```

```golang
// 调用其他智能合约，需要自行封装参数和解析结果
func Call(name string, method string, args map[string]string) (interface{}, error)

// 调用智能合约同时向合约转账
func CallWithValue(name string, method string, args map[string]string) (interface{}, error)

// 本次调用是谁发起的，即调用者地址（合约账户、外部账户）
func Caller() string

// 返回交易发起者的地址。这是一个不随着调用深度变化的值
// 例：账户alice调用了A合约，A合约调用了B合约，B合约调用了C合约…，无论调用层次多深，执行Origin()得到的都是alice的账户地址。
func Origin() string

// 本次调用调用者给了本合约多少资产
func Value() int

// 返回本合约的地址
func Self() string

// 合约向账户 to 转账 amount 资产
func Transfer(to string, amount int) (interface{}, error)
```

#### 智能合约模板（golang plugin）
```golang

package main	// 包名必须为main

import (
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/contract" // 使用内置功能时引入
)

// 参数必须为 map[string]string, 返回结果必须为 (interface{}, error)
func Multiply(args map[string]string) (interface{}, error) {
	// 调用其他合约，自行封装参数
	num, err := contract.Call("random", "GetRandom", map[string]string{})
	if err != nil {
		log.Infof("[Multiply] 调用random失败")
		return nil, err
	}
	caller := contract.Caller()
	origin := contract.Origin()
	value  := contract.Value()
	balance:= contract.Balance()
	self   := contract.Self()
	contract.Transfer(caller, value)
	
	a := num.(int)
	log.Infof("[Multiply] 调用 random.GetRandom 成功，结果：%v\n", a)
	ans := a * a
	log.Infof("[Multiply] 结果：%v\n", ans)
	return ans, nil
}
```
