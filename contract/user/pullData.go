package main

import (
	"encoding/json"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/contract" // 调用其他智能合约时引入
	"github.com/ssbcV2/meta"
)

var externalData string

// pull外部数据，回调updateData
func NewRequest(args map[string]string) (interface{}, error) {
	cb := meta.Callback{
		Caller:   "",
		Value:    0,
		Contract: "pullData",
		Method:   "UpdateData",
		Args:     nil,
		Address:  "",
	}
	cbBytes, _ := json.Marshal(cb)
	reqArgs := map[string]string{
		"type":     "api", // "api":第三方接口，"chain":"跨链数据"
		"url":      "http://localhost:7777/testApi",
		"callback": string(cbBytes),
	}
	// 日志事件
	recordArgs := map[string]string{
		"state": "success",
	}
	// 调用QueryData预言机合约请求外部数据
	res, err := contract.Call("oracle", "QueryData", reqArgs)
	if err != nil {
		log.Errorf("call QueryData contract error: %s", err)
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
func UpdateData(args map[string]string) (interface{}, error) {
	recordArgs := map[string]string{
		"state": "success",
	}
	log.Infof("updateData方法收到参数：%+v", args)
	newData, ok := args["data"]
	if !ok {
		recordArgs["state"] = "fail"
		_, err := contract.Call("oracle", "RecordEvent", recordArgs)
		return meta.ContractUpdateData{}, err
	}
	externalData = newData
	log.Infof("externalData更新成功：%s", externalData)
	_, _ = contract.Call("oracle", "RecordEvent", recordArgs)
	return meta.ContractUpdateData{}, nil
}

// pull跨链数据,回调UseChainData
func PullChainData(args map[string]string) (interface{}, error) {
	cb := meta.Callback{
		Caller:   "",
		Value:    0,
		Contract: "pullData",
		Method:   "UseChainData",
		Args:     nil,
		Address:  "",
	}
	cbBytes, _ := json.Marshal(cb)
	reqArgs := map[string]string{
		"type":     "chain", // "api":第三方接口，"chain":"跨链数据"
		"callback": string(cbBytes),
		"name": "ssbc2",
		"dataType": "getBlockChain",
		"params": "",
	}
	// 日志事件
	recordArgs := map[string]string{
		"state": "success",
	}
	// 调用QueryData预言机合约请求外部数据
	res, err := contract.Call("oracle", "QueryData", reqArgs)
	if err != nil {
		log.Errorf("call QueryData contract error: %s", err)
		recordArgs["state"] = "fail"
		_, err = contract.Call("oracle", "RecordEvent", recordArgs)
		return meta.ContractUpdateData{}, err
	}
	resBytes, _ := json.Marshal(res)
	recordArgs["res"] = string(resBytes)
	_, err = contract.Call("oracle", "RecordEvent", recordArgs)
	return res, nil
}

func UseChainData(args map[string]string) (interface{}, error) {
	recordArgs := map[string]string{
		"state": "success",
	}
	log.Infof("UseChainData方法收到参数：%+v", args)
	data, ok := args["data"]
	if !ok {
		recordArgs["state"] = "fail"
		_, err := contract.Call("oracle", "RecordEvent", recordArgs)
		return meta.ContractUpdateData{}, err
	}
	log.Infof("PullChainData跨链数据：%s", data)
	_, _ = contract.Call("oracle", "RecordEvent", recordArgs)
	return nil, nil
}

// 回退函数，当没有方法匹配时执行此方法
func Fallback(args map[string]string) (interface{}, error) {
	return meta.ContractUpdateData{}, nil
}
