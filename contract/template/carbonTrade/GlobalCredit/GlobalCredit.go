package main

import (
	"errors"
	"github.com/ssbcV2/contract"
	"github.com/ssbcV2/meta"
)

/*
碳征信链
*/
var GlobalReputation float64 // 企业全局信誉

func init()  {
	GlobalReputation = 0.5 // 初始全局信誉为0.5
}

// 根据局部信誉更新全局信誉
func UpdateGlobalReputation(args map[string]interface{}) (interface{}, error){
	contract.Info("UpdateGlobalReputation收到数据：%+v", args)
	localRep, ok := args["localRep"].(float64)
	if !ok {
		contract.Info("localRep参数不存在")
		return meta.ContractUpdateData{}, errors.New("localRep参数不存在")
	}
	contract.Info("得到碳交易链的局部信誉值：%v", localRep)
	GlobalReputation += localRep
	contract.Info("全局信誉值更新完成：%v", GlobalReputation)
	return nil, nil
}

// 回退函数，当没有方法匹配时执行此方法
func Fallback(args map[string]interface{}) (interface{}, error) {
	return meta.ContractUpdateData{}, nil
}
