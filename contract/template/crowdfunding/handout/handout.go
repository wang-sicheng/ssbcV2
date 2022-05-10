package main

import (
	"errors"
	"github.com/ssbcV2/contract"
	"github.com/ssbcV2/util"
)

var Statistics map[string]int         // 接收来自A链的统计数据
var Publisher string				  // 合约发布人
var Ready bool                        // 数据是否已经准备好

func init() {
	Statistics = map[string]int{}
	Publisher = contract.Caller()
	Ready = false
}

func GetData(args map[string]interface{}) (interface{}, error) {
	if !Ready {
		res, _ := contract.GetCrossContractData(
			"ssbc2",
			"deposit",
			"Money",
			"ReceiveData")
		return res, nil
	}
	return nil, nil
}

func GetCoin(args map[string]interface{}) (interface{}, error) {
	a_addr, ok := args["a_addr"].(string)
	if !ok {
		return nil, errors.New("缺少a_addr参数")
	}

	if len(Statistics) == 0 {
		return nil, errors.New("没有数据或数据尚未准备好")
	}

	amount, ok := Statistics[a_addr]
	if !ok {
		return nil, errors.New("A链地址不存在")
	}

	err := contract.TransferFrom(Publisher, contract.Caller(), amount)
	if err != nil {
		return nil, err
	}

	Statistics[a_addr] = 0
	return nil, nil
}

func ReceiveData(args map[string]interface{}) (interface{}, error) {
	contract.Infof("ReceiveData 方法收到参数：%+v", args)

	data, ok := args["data"].(string)
	if !ok {
		contract.Info("接收数据失败")
		return nil, errors.New("接收数据失败")
	}
	contract.Info("ReceiveData 收到跨链数据：%s", data)
	Statistics = getStatistics(util.JsonToMap(data))
	Ready = true
	return nil, nil
}

func getStatistics(args map[string]interface{}) map[string]int {
	d := args["Money"].(map[string]interface{})
	res := map[string]int{}

	for k, v := range d {
		tmp := v.(float64)
		res[k] = int(tmp)
	}
	return res
}
