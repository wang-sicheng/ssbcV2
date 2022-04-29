package main

import (
	"encoding/json"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/contract"
)

func Call(args map[string]interface{}) (interface{}, error) {
	sourceChain := args["sourceChain"].(string)
	caller := args["caller"].(string)
	sourceContract := args["sourceContract"].(string)
	targetContract := args["targetContract"].(string)
	method := args["method"].(string)
	argsStr := args["args"].(string)
	callback := args["callback"].(string)


	var arg map[string]interface{}
	err := json.Unmarshal([]byte(argsStr), &arg)
	if err != nil {
		log.Error("CrossCall Unmarshal error", err)
	}
	res, err := contract.Call(targetContract, method, arg)

	log.Infof("链%v的%v调用%v合约的%v()方法，参数为%v，回调方法%v\n", sourceChain, caller, targetContract, method, arg, callback)

	result := map[string]interface{}{}
	result["result"] = res

	tdataBytes, _ := json.Marshal(result)
	reqArgs := map[string]interface{}{
		"type": "chain",
		"chainName": sourceChain,
		"contract": sourceContract,
		"function": callback,
		"tData": string(tdataBytes),
		"address": "",
	}
	res, err = contract.Call("oracle", "TransferData", reqArgs)
	if err != nil {
		log.Errorf("Call调用TransferData预言机合约失败：%s", err)
		return nil, err
	}

	return res, nil
}
