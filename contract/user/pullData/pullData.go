package main

import (
	"encoding/json"
	"github.com/ssbcV2/contract" // 调用其他智能合约时引入
	"github.com/ssbcV2/meta"
)

var ExternalData interface{}

// pull外部数据，回调updateData
func NewRequest(args map[string]interface{}) (interface{}, error) {
	cb := meta.Callback{
		Caller:   "",
		Value:    0,
		Contract: contract.Name(),
		Method:   "UpdateData",
		Args:     nil,
		Address:  "",
	}
	cbBytes, _ := json.Marshal(cb)
	reqArgs := map[string]interface{}{
		"type":     "api", // "api":第三方接口，"chain":"跨链数据"
		"url":      "http://localhost:7777/testApi",
		"callback": string(cbBytes),
	}
	// 日志事件
	recordArgs := map[string]interface{}{
		"state": "success",
	}
	// 调用QueryData预言机合约请求外部数据
	res, err := contract.Call("oracle", "QueryData", reqArgs)
	if err != nil {
		contract.Info("call QueryData contract error: %s", err)
		recordArgs["state"] = "fail"
		_, err = contract.Call("oracle", "RecordEvent", recordArgs)
		return meta.ContractUpdateData{}, err
	}
	resBytes, _ := json.Marshal(res)
	recordArgs["res"] = string(resBytes)
	_, err = contract.Call("oracle", "RecordEvent", recordArgs)
	return res, nil
}

// 回调函数,更新externalData
func UpdateData(args map[string]interface{}) (interface{}, error) {
	recordArgs := map[string]interface{}{
		"state": "success",
	}
	contract.Info("updateData方法收到参数：%+v", args)
	newData, ok := args["data"]
	if !ok {
		recordArgs["state"] = "fail"
		_, err := contract.Call("oracle", "RecordEvent", recordArgs)
		return meta.ContractUpdateData{}, err
	}
	ExternalData = newData
	contract.Info("externalData更新成功：%s", ExternalData)
	_, _ = contract.Call("oracle", "RecordEvent", recordArgs)
	return meta.ContractUpdateData{}, nil
}

// pull跨链数据,回调UseChainData
func PullChainData(args map[string]interface{}) (interface{}, error) {
	cb := meta.Callback{
		Caller:   "",
		Value:    0,
		Contract: contract.Name(),
		Method:   "UseChainData",
		Args:     nil,
		Address:  "",
	}
	cbBytes, _ := json.Marshal(cb)
	reqArgs := map[string]interface{}{
		"type":     "chain", // "api":第三方接口，"chain":"跨链数据"
		"callback": string(cbBytes),
		"name": "ssbc2",
		"dataType": "getBlockChain",
		"params": "",
	}
	// 日志事件
	recordArgs := map[string]interface{}{
		"state": "success",
	}
	// 调用QueryData预言机合约请求外部数据
	res, err := contract.Call("oracle", "QueryData", reqArgs)
	if err != nil {
		contract.Info("call QueryData contract error: %s", err)
		recordArgs["state"] = "fail"
		_, err = contract.Call("oracle", "RecordEvent", recordArgs)
		return meta.ContractUpdateData{}, err
	}
	resBytes, _ := json.Marshal(res)
	recordArgs["res"] = string(resBytes)
	_, err = contract.Call("oracle", "RecordEvent", recordArgs)
	return res, nil
}

func UseChainData(args map[string]interface{}) (interface{}, error) {
	recordArgs := map[string]interface{}{
		"state": "success",
	}
	contract.Info("UseChainData方法收到参数：%+v", args)
	data, ok := args["data"]
	if !ok {
		recordArgs["state"] = "fail"
		_, err := contract.Call("oracle", "RecordEvent", recordArgs)
		return meta.ContractUpdateData{}, err
	}
	contract.Info("PullChainData跨链数据：%s", data)
	_, _ = contract.Call("oracle", "RecordEvent", recordArgs)
	return nil, nil
}

// 回退函数，当没有方法匹配时执行此方法
func Fallback(args map[string]interface{}) (interface{}, error) {
	return meta.ContractUpdateData{}, nil
}
