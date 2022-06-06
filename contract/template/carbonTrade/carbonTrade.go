package main

import (
	"errors"
	"github.com/ssbcV2/contract"
	"strconv"
)

// 碳交易订单
type CreditOrder struct {
	From string // 卖方地址
	To string  // 买方地址
	Amount uint16 // 碳额数量
	Status int // 订单状态（0:未交易;1:已交易;2:已取消）
}

// 账户信息
type CarbonInfo struct {
	Locked uint16 // 锁定碳权额度
	Credit uint16 // 碳权额度
	Balance uint16 // 账户资金余额
}

// 全部碳交易订单
var OrderList []CreditOrder
var CreditMap map[string]CarbonInfo

// 部署合约时初始化
func init() {
	OrderList = []CreditOrder{}
	CreditMap = make(map[string]CarbonInfo)
}

// 用户交易前执行，分配初始碳额和资金
func Deposit(args map[string]interface{}) (interface{}, error) {
	CreditMap[contract.Caller()] = CarbonInfo{
		Locked: 0,
		Credit: 100,
		Balance: 100,
	}
	return nil, nil
}

// 发布订单方法
func MakeOrder(args map[string]interface{}) (interface{}, error) {
	Sender := contract.Caller() // 合约调用方
	amountStr, ok := args["amount"].(string)
	if !ok {
		return nil, errors.New("缺少amount参数")
	}
	amountInt, _ := strconv.Atoi(amountStr)
	amount := uint16(amountInt)
	if CreditMap[Sender].Credit- amount < 0 { // 存在整数溢出漏洞  0-1=65535
		return nil, errors.New("碳权余额不足")
	}
	info := CreditMap[Sender]
	info.Locked += amount // 锁定amount数量的credit
	info.Credit -= amount
	CreditMap[Sender] = info
	OrderList = append(OrderList, CreditOrder{
		From:   Sender,
		To:     "",
		Amount: amount,
		Status: 0,
	})
	return nil, nil
}

// 购买碳权方法
func BuyOrder(args map[string]interface{}) (interface{}, error) {
	Sender := contract.Caller() // 合约调用方
	indexStr, ok := args["index"].(string)
	index, _ := strconv.Atoi(indexStr)
	if !ok { // 使用if判断执行条件，存在假充值漏洞
		return nil, errors.New("缺少index参数")
	}
	if index < 0 || index >= len(OrderList) {
		return nil, errors.New("index超出范围")
	}
	toInfo := CreditMap[Sender] // 碳权购买方
	order := OrderList[index] // 当前交易的订单
	fromInfo := CreditMap[order.From] // 碳权出售方
	if order.Status != 0 {
		return nil, errors.New("无效的订单编号")
	}
	if toInfo.Balance < order.Amount {
		return nil, errors.New("买方余额不足")
	}
	toInfo.Credit += order.Amount
	fromInfo.Locked -= order.Amount
	order.To = Sender
	order.Status = 1
	// 更新合约数据
	CreditMap[Sender] = toInfo
	CreditMap[order.From] = fromInfo
	OrderList[index] = order
	return nil, nil
}

// 卖方取消订单
func CancelOrder(args map[string]interface{}) (interface{}, error) {
	Sender := contract.Caller()
	indexStr, ok := args["index"].(string)
	index, _ := strconv.Atoi(indexStr)
	if !ok { // 使用if判断执行条件，存在假充值漏洞
		return nil, errors.New("缺少index参数")
	}
	if index < 0 || index >= len(OrderList) {
		return nil, errors.New("index超出范围")
	}
	order := OrderList[index]
	if order.From != Sender {
		return nil, errors.New("账户无权限")
	}
	fromInfo := CreditMap[Sender]
	order.Status = 2
	fromInfo.Credit += order.Amount
	fromInfo.Locked -= order.Amount

	CreditMap[Sender] = fromInfo
	OrderList[index] = order
	return nil, nil
}

// 买方向卖方转账
func Transfer(args map[string]interface{}) (interface{}, error) {
	Sender := contract.Caller() // 合约调用方
	to, ok := args["to"].(string)
	if !ok {
		return nil, errors.New("缺少to参数")
	}
	amountStr, ok := args["amount"].(string)
	if !ok {
		return nil, errors.New("缺少amount参数")
	}
	amountInt, _ := strconv.Atoi(amountStr)
	amount := uint16(amountInt)

	fromInfo := CreditMap[Sender]
	toInfo := CreditMap[to]
	fromInfo.Balance -= amount // 余额在BuyOrder()中经过校验
	toInfo.Balance += amount
	CreditMap[Sender] = fromInfo
	CreditMap[to] = toInfo
	return nil, nil
}


// 回退函数，当没有方法匹配时执行此方法
func Fallback(args map[string]interface{}) (interface{}, error) {
	contract.Transfer(contract.Caller(), contract.Value()) // 将转账退回
	return nil, nil
}

