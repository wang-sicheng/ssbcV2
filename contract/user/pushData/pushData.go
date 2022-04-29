package main

import (
	"encoding/json"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/contract"
	"github.com/ssbcV2/meta"
)

// push数据测试，ssbc1链生成数据传输给ssbc2链的ConsumeData
func ProductData(args map[string]interface{}) (interface{}, error) {
	tdata := map[string]interface{}{
		"msg": "data from ssbc1:pushData:ProductData",
	}
	tdataBytes, _ := json.Marshal(tdata)
	reqArgs := map[string]interface{}{
		"type": "chain",
		"chainName": "ssbc2",
		"contract": "pushData",
		"function": "ConsumeData",
		"tData": string(tdataBytes),
		"address": "",
	}
	res, err := contract.Call("oracle", "TransferData", reqArgs)
	if err != nil {
		log.Errorf("productData调用TransferData预言机合约失败：%s", err)
		return nil, err
	}
	return res, nil
}

// 提前部署在ssbc2链上
func ConsumeData(args map[string]interface{}) (interface{}, error){
	log.Infof("consumeData收到数据：%+v", args)
	return nil, nil
}

// 回退函数，当没有方法匹配时执行此方法
func Fallback(args map[string]interface{}) (interface{}, error) {
	return meta.ContractUpdateData{}, nil
}
