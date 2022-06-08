package main

import (
	"encoding/json"
	"errors"
	"github.com/ssbcV2/contract" // 调用其他智能合约时引入
	"github.com/ssbcV2/meta"
)

/*
碳交易链
*/

var LocalReputation float64
var GlobalReputation float64

func init()  {
	LocalReputation = 0.1
	GlobalReputation = 0
}

// pull碳征信链的征信数据,回调StartCarbonTrade
func PullCarbonCreditData(args map[string]interface{}) (interface{}, error) {
	cb := meta.Callback{
		Caller:   "",
		Value:    0,
		Contract: contract.Name(),
		Method:   "StartCarbonTrade",
		Args:     nil,
		Address:  "",
	}
	cbBytes, _ := json.Marshal(cb)
	reqArgs := map[string]interface{}{
		"type":     "chain", // "chain":"跨链数据"
		"callback": string(cbBytes),
		"name": "ssbc2", // 碳征信链
		"dataType": "contractData",
		"params": "GlobalCredit,GlobalReputation",
	}
	// 调用QueryData预言机合约请求外部数据
	res, err := contract.Call("oracle", "QueryData", reqArgs)
	if err != nil {
		contract.Info("call QueryData contract error: %s", err)
		return meta.ContractUpdateData{}, err
	}
	return res, nil
}

func StartCarbonTrade(args map[string]interface{}) (interface{}, error) {
	contract.Info("StartCarbonTrade方法收到参数：%+v", args)
	data, ok := args["data"].(string)
	if !ok {
		contract.Info("data参数不存在")
		return meta.ContractUpdateData{}, errors.New("data参数不存在")
	}
	contract.Info("得到碳征信链的全局信誉值：%v", data)
	dataMap := map[string]interface{}{}
	err := json.Unmarshal([]byte(data), &dataMap)
	if err != nil {
		contract.Info("data解析失败：%s", err)
		return meta.ContractUpdateData{}, err
	}
	credit, ok := dataMap["GlobalReputation"].(float64)
	GlobalReputation = credit
	if !ok {
		contract.Info("credit解析失败：%s", err)
		return meta.ContractUpdateData{}, nil
	}
	if credit >= 0.5 {
		contract.Info("模拟碳交易执行......")
	}
	return nil, nil
}

// push碳交易链的局部信誉给碳征信链
func PushLocalReputation(args map[string]interface{}) (interface{}, error) {
	tdata := map[string]interface{}{
		"localRep": LocalReputation,
	}
	tdataBytes, _ := json.Marshal(tdata)
	reqArgs := map[string]interface{}{
		"type": "chain",
		"chainName": "ssbc2",
		"contract": "GlobalCredit",
		"function": "UpdateGlobalReputation",
		"tData": string(tdataBytes),
		"address": "",
	}
	res, err := contract.Call("oracle", "TransferData", reqArgs)
	if err != nil {
		contract.Info("PushLocalReputation调用TransferData预言机合约失败：%s", err)
		return nil, err
	}
	return res, nil
}

// 回退函数，当没有方法匹配时执行此方法
func Fallback(args map[string]interface{}) (interface{}, error) {
	return meta.ContractUpdateData{}, nil
}
