package main

import (
	"encoding/json"
	"errors"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/contract" // 调用其他智能合约时引入
	"github.com/ssbcV2/global"
	"github.com/ssbcV2/meta"
)

// 定义事件函数
func NewEvent(args map[string]interface{}) (interface{}, error) {
	var res meta.ContractUpdateData
	var event meta.Event

	event.FromAddress = contract.Caller()
	event.Args = args
	eventType, ok := args["eventType"].(string)
	if ok {
		event.Type = eventType
	}
	event.ChainId = global.ChainID
	res.Events = append(res.Events, event)
	return res, nil
}

// 定义订阅函数
func NewSub(args map[string]interface{}) (interface{}, error) {
	var res meta.ContractUpdateData
	var sub meta.EventSub
	sub.FromAddress = contract.Caller()
	eid, ok := args["event_id"].(string)
	if ok {
		sub.EventID = eid // 不是必须，也可以自定义targetEvent
	}
	cbStr, ok := args["callback"].(string)
	if !ok {
		return nil, errors.New("miss callback field")
	}
	var cb meta.Callback
	err := json.Unmarshal([]byte(cbStr), &cb)
	if err != nil {
		log.Errorf("callback unmarshal error: %s", err)
		return meta.ContractUpdateData{}, err
	}
	log.Infof("newSub callback info: %+v", cb)
	sub.EventID = eid
	sub.Callback = cb
	res.EventSubs = append(res.EventSubs, sub)
	return res, nil
}

/*
type: api
url: 数据url
callback: 回调函数

type: chain
name: 数据链链名
dataType: 跨链数据类型
params: 查询参数
callback: 回调函数
*/
func QueryData(args map[string]interface{}) (interface{}, error) {
	var res meta.ContractUpdateData
	qType, ok := args["type"] // 查询类型，对应生成不同的事件类型
	eventArgs := make(map[string]interface{})
	subArgs := make(map[string]interface{})
	if !ok {
		return nil, errors.New("miss type args")
	}
	switch qType {
	case "api":
		eventArgs["eventType"] = "1"
		url, ok := args["url"].(string)
		if !ok {
			return nil, errors.New("miss url args")
		}
		eventArgs["url"] = url
	case "chain":
		eventArgs["eventType"] = "2"
		name, ok := args["name"].(string)
		if !ok {
			return nil, errors.New("miss name args")
		}
		eventArgs["name"] = name
		dataType, ok := args["dataType"].(string)
		if !ok {
			return nil, errors.New("miss dataType args")
		}
		eventArgs["dataType"] = dataType
		params, ok := args["params"].(string)
		if !ok {
			return nil, errors.New("miss params args")
		}
		eventArgs["params"] = params
	}

	callback, ok := args["callback"].(string)
	if !ok {
		return nil, errors.New("miss callback args")
	}
	subArgs["callback"] = callback

	res1, err := contract.Call("oracle", "NewEvent", eventArgs)
	if err != nil {
		log.Errorf("call newEvent contract error: %s", err)
		return nil, err
	}
	eventRes, _ := res1.(meta.ContractUpdateData)
	res2, err := contract.Call("oracle", "NewSub", subArgs)
	if err != nil {
		log.Errorf("call newSub contract error: %s", err)
		return nil, err
	}
	subRes, _ := res2.(meta.ContractUpdateData)
	// 自动订阅获取数据的事件
	subRes.EventSubs[0].TargetEvent = eventRes.Events[0]
	//res.Events = eventRes.Events
	res.EventSubs = subRes.EventSubs
	log.Infof("queryData res: %+v", res)
	return res, nil
}

/*
type: "chain"
chainName: 链名
address: 合约地址
contract: 合约名
function: 方法名
tData: 传输的数据
*/
func TransferData(args map[string]interface{}) (interface{}, error) {
	var res meta.ContractUpdateData
	eType, ok := args["type"]
	eventArgs := make(map[string]interface{})
	if !ok {
		return nil, errors.New("miss type args")
	}
	switch eType {
	case "chain":
		eventArgs["eventType"] = "3"
		name, ok := args["chainName"]
		if !ok {
			return nil, errors.New("miss chainName args")
		}
		eventArgs["chainName"] = name
		cont, ok := args["contract"]
		if !ok {
			return nil, errors.New("miss contract args")
		}
		eventArgs["contract"] = cont
		function, ok := args["function"]
		if !ok {
			return nil, errors.New("miss function args")
		}
		eventArgs["function"] = function
		data, ok := args["tData"]
		if !ok {
			return nil, errors.New("miss tData args")
		}
		eventArgs["tData"] = data
		address, ok := args["address"]
		if !ok {
			return nil, errors.New("miss address args")
		}
		eventArgs["address"] = address
	}
	res1, err := contract.Call("oracle", "NewEvent", eventArgs)
	if err != nil {
		log.Errorf("call newEvent contract error: %s", err)
		return nil, err
	}
	res, _ = res1.(meta.ContractUpdateData)
	return res, nil
}


// 回退函数，当没有方法匹配时执行此方法
func Fallback(args map[string]interface{}) (interface{}, error) {
	return meta.ContractUpdateData{}, nil
}
