package main

import (
	"encoding/json"
	"errors"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/contract"
	"github.com/ssbcV2/meta"
	"github.com/ssbcV2/util"
)

var statistics map[string]interface{}		// 接收来自A链的统计数据
var total int							// 要分发的总额度
var amount int							// 众筹的总数额
var ready bool							// 数据是否已经准备好

func init() {
	statistics = map[string]interface{}{}
	total = contract.Value()
	ready = false
	amount = 0
}

func Add(args map[string]string) (interface{}, error) {
	total += contract.Value()
	return nil, nil
}

func GetCoin(args map[string]string) (interface{}, error) {
	if !ready {
		cb := meta.Callback{
			Caller:   "",
			Value:    0,
			Contract: contract.Name(),
			Method:   "ReceiveData",
			Args:     nil,
			Address:  "",
		}
		cbBytes, _ := json.Marshal(cb)
		reqArgs := map[string]string{
			"type":     "chain", // "api":第三方接口，"chain":"跨链数据"
			"callback": string(cbBytes),
			"name": "ssbc2",
			"dataType": "contractData",
			"params": "deposit,Money",
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

	a_addr, ok := args["a_addr"]
	if !ok {
		return nil, errors.New("缺少a_addr参数")
	}

	if len(statistics) == 0 {
		return nil, errors.New("没有数据或数据尚未准备好")
	}

	_, ok = statistics[a_addr]
	if !ok {
		return nil, errors.New("A链地址不存在")
	}


	_ = contract.Transfer(contract.Caller(), 100)
	return nil, nil
}

func ReceiveData(args map[string]string) (interface{}, error) {
	recordArgs := map[string]string{
		"state": "success",
	}
	log.Infof("ReceiveData 方法收到参数：%+v", args)
	data, ok := args["data"]
	if !ok {
		recordArgs["state"] = "fail"
		_, err := contract.Call("oracle", "RecordEvent", recordArgs)
		return meta.ContractUpdateData{}, err
	}
	log.Infof("ReceiveData 收到跨链数据：%s", data)
	statistics = util.JsonToMap(data)
	amount = totalMoney(statistics)
	log.Infof("amount: %v\n", amount)
	ready = true
	_, _ = contract.Call("oracle", "RecordEvent", recordArgs)
	return nil, nil
}

func totalMoney(args map[string]interface{}) int {
	d := args["Money"].(map[string]interface{})

	var tmp float64
	for _, v := range d {
		tmp += v.(float64)
	}
	return int(tmp)
}
