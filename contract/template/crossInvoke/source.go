package main

import (
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/contract"
)

var (
	A float64
	B float64
	Result float64
)

func init() {
	A = 3.5
	B = 4.6
	Result = 0
}

func CrossInvokeAdd(args map[string]interface{}) (interface{}, error) {
	arg := map[string]interface{}{
		"A": A,
		"B": B,
	}
	res, _ := contract.CrossCall(contract.Name(), "ssbc2", "math", "Add", arg, "ReceiveResult")
	return res, nil
}

func SetAB(args map[string]interface{}) (interface{}, error) {
	a, ok := args["A"].(float64)
	if !ok {
		log.Info("参数A格式错误")
	} else {
		A = a
	}

	b, ok := args["B"].(float64)
	if !ok {
		log.Info("参数B格式错误")
	} else {
		B = b
	}
	return nil, nil
}

func ReceiveResult(args map[string]interface{}) (interface{}, error) {
	result, ok := args["result"].(float64)
	if !ok {
		log.Info("获取跨链结果失败")
	}
	Result = result
	return Result, nil
}
