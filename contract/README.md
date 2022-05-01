#### 智能合约接口
建议在后端项目ssbcV2/contract目录下编辑合约，之后再复制到前端编辑框内发布。

在合约中需要引入 contract 包
```golang
import "github.com/ssbcV2/contract"
```

```golang
// 调用其他智能合约，需要自行封装参数和解析结果
func Call(name string, method string, args map[string]interface{}) (interface{}, error)

// 调用智能合约同时向合约转账
func CallWithValue(name string, method string, args map[string]interface{}, value int) (interface{}, error)

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

// 合约操作 from 账户向 to 账户转账
func TransferFrom(from, to string, amount int) error

// 返回当前合约拥有多少资产
func Balance() int

// 根据地址获取对应账户的余额
func GetBalance(address string) int

// 记录执行情况，调用时会推送到前端
func Info(info ...interface{})
```

#### 智能合约模板（golang plugin）
```golang

package main	// 包名必须为main

import (
	"github.com/ssbcV2/contract" // 使用内置功能时引入
)

var Caller string   // 首字母大写对外可见，小写不可见，方法同理
var Origin string
var Value int
var Balance int
var Self string

// 参数必须为 map[string]interface{}, 返回结果必须为 (interface{}, error)
func Multiply(args map[string]interface{}) (interface{}, error) {
	Caller = contract.Caller()
	Origin = contract.Origin()
	Value  = contract.Value()
	Balance= contract.Balance()
	Self   = contract.Self()

	// 调用其他合约，自行封装参数
	num, err := contract.Call("random", "GetRandom", map[string]interface{}{})
	if err != nil {
		contract.Info("[Multiply] 调用random失败")
		return nil, err
	}
	a := num.(int)
	contract.Info("[Multiply] 调用 random.GetRandom 成功，结果：%v\n", a)
	ans := a * a
	contract.Info("[Multiply] 结果：%v\n", ans)
	return ans, nil
}
```
