package main

import (
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/meta"
)

func ContractTest(args map[string]interface{}) (interface{}, error) {
	log.Infof("ContractTest: args from client: %+v", args)
	return meta.ContractUpdateData{}, nil
}

// 回退函数，当没有方法匹配时执行此方法
func Fallback(args map[string]interface{}) (interface{}, error) {
	return meta.ContractUpdateData{}, nil
}
