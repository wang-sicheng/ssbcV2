package main

import (
	"github.com/ssbcV2/contract"
)

var Res float64

func Add(args map[string]interface{}) (interface{}, error) {
	a, ok := args["A"].(float64)
	if !ok {
		contract.Info("缺少参数A")
	}

	b, ok := args["B"].(float64)
	if !ok {
		contract.Info("缺少参数B")
	}

	Res = a + b
	return Res, nil
}
