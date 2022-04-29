package main

import (
	"encoding/json"
	"errors"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/contract" // 调用其他智能合约时引入
	"github.com/ssbcV2/meta"
	"strconv"
)

/*
碳交易链
*/

var localReputation string

func init()  {
	localReputation = "0.1"
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
		log.Errorf("call QueryData contract error: %s", err)
		return meta.ContractUpdateData{}, err
	}
	return res, nil
}

func StartCarbonTrade(args map[string]interface{}) (interface{}, error) {
	log.Infof("StartCarbonTrade方法收到参数：%+v", args)
	data, ok := args["data"].(string)
	if !ok {
		log.Error("data参数不存在")
		return meta.ContractUpdateData{}, errors.New("data参数不存在")
	}
	log.Infof("得到碳征信链的全局信誉值：%v", data)
	dataMap := map[string]interface{}{}
	err := json.Unmarshal([]byte(data), &dataMap)
	if err != nil {
		log.Error("data解析失败：%s", err)
		return meta.ContractUpdateData{}, err
	}
	credit, err := strconv.ParseFloat(dataMap["GlobalReputation"].(string), 32)
	if err != nil {
		log.Error("credit解析失败：%s", err)
		return meta.ContractUpdateData{}, err
	}
	if credit >= 0.5 {
		log.Infof("模拟碳交易执行......")
	}
	return nil, nil
}

// push碳交易链的局部信誉给碳征信链
func PushLocalReputation(args map[string]interface{}) (interface{}, error) {
	tdata := map[string]string{
		"localRep": localReputation,
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
		log.Errorf("PushLocalReputation调用TransferData预言机合约失败：%s", err)
		return nil, err
	}
	return res, nil
}

// 回退函数，当没有方法匹配时执行此方法
func Fallback(args map[string]interface{}) (interface{}, error) {
	return meta.ContractUpdateData{}, nil
}
