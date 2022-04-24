package main


import (
	"errors"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/meta"
	"strconv"
)

/*
碳征信链
*/
var GlobalReputation string // 企业全局信誉

func init()  {
	GlobalReputation = "0.5" // 初始全局信誉为0.5
}

// 根据局部信誉更新全局信誉
func UpdateGlobalReputation(args map[string]string) (interface{}, error){
	log.Infof("UpdateGlobalReputation收到数据：%+v", args)
	localRep, ok := args["localRep"]
	if !ok {
		log.Error("localRep参数不存在")
		return meta.ContractUpdateData{}, errors.New("localRep参数不存在")
	}
	log.Infof("得到碳交易链的局部信誉值：%v", localRep)
	GlobalRepNum, _ := strconv.ParseFloat(GlobalReputation, 32)
	localRepNum, _ := strconv.ParseFloat(localRep, 32)
	GlobalRepNum += localRepNum
	NewGlobalReputation := strconv.FormatFloat(GlobalRepNum, 'E', -1, 32)
	GlobalReputation = NewGlobalReputation
	log.Infof("全局信誉值更新完成：%v", GlobalReputation)
	return nil, nil
}

// 回退函数，当没有方法匹配时执行此方法
func Fallback(args map[string]string) (interface{}, error) {
	return meta.ContractUpdateData{}, nil
}
