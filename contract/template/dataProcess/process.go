package main

import (
	"github.com/ssbcV2/contract"
)

var (
	A float64
	B string
	C []interface{}
	D bool
	E map[string]interface{}
)


func Add(args map[string]interface{}) (interface{}, error) {
	var ok bool
	A, ok = args["A"].(float64)
	if !ok {
		contract.Info("参数A格式错误或不存在")
	}

	B, ok = args["B"].(string)
	if !ok {
		contract.Info("参数B格式错误或不存在")
	}

	C, ok = args["C"].([]interface{})
	if !ok {
		contract.Info("参数C格式错误或不存在")
	}

	D, ok = args["D"].(bool)
	if !ok {
		contract.Info("参D格式错误或不存在")
	}

	E, ok = args["E"].(map[string]interface{})
	if !ok {
		contract.Info("参数E格式错误或不存在")
	}

	contract.Info("执行结果：", A, B, C, D, E)
	return nil, nil
}

/*
{
    "A":5.34,
    "B":"It's string",
    "C":[
        1,
        2,
        3,
        4
    ],
    "D":true,
    "E":{
        "Key":"Value"
    }
}
*/

