package main

import (
	"errors"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/smart_contract"
)

func Multiply(args map[string]string) (interface{}, error) {
	num, err := smart_contract.CallContract("random", "GetRandom", map[string]string{})
	if err != nil {
		errMsg := "[Multiply] 调用random失败"
		log.Infof(errMsg)
		return nil, errors.New(errMsg)
	}
	a := num.(int)
	log.Infof("[Multiply] 调用 random.GetRandom 成功，结果：%v\n", a)
	ans := a * a
	log.Infof("[Multiply] 结果：%v\n", ans)
	return ans, nil
}
