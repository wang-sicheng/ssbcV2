package main

import (
	"encoding/json"
	"errors"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/meta"
	"github.com/ssbcV2/smart_contract" // 调用其他智能合约时引入
)

// 定义事件函数
func NewEvent(args map[string]string) (interface{}, error) {
	var res meta.ContractUpdateData
	var event meta.Event

	event.FromAddress = smart_contract.Caller
	event.Args = args
	eventType, ok := args["event_type"]
	if ok {
		event.Type = eventType
	}
	res.Events = append(res.Events, event)
	return res, nil
}

// 定义订阅函数
func NewSub(args map[string]string) (interface{}, error) {
	var res meta.ContractUpdateData
	var sub meta.EventSub
	sub.FromAddress = smart_contract.Caller
	eid, ok := args["event_id"]
	if !ok {
		return nil, errors.New("miss event_id field")
	}
	cbStr, ok := args["callback"]
	if !ok {
		return nil, errors.New("miss callback field")
	}
	var cb meta.Callback
	err := json.Unmarshal([]byte(cbStr), &cb)
	if err != nil {
		log.Errorf("callback unmarshal error: %s", err)
		return nil, err
	}
	sub.EventID = eid
	sub.Callback = cb
	res.EventSubs = append(res.EventSubs, sub)
	return res, nil
}

/*
type: 请求类型（api,chain）
url: 数据url
callback: 回调函数
*/
func QueryData(args map[string]string) (interface{}, error) {
	var res meta.ContractUpdateData
	qType, ok := args["type"] // 查询类型，对应生成不同的事件类型
	eventArgs := make(map[string]string)
	subArgs := make(map[string]string)
	if !ok {
		return nil, errors.New("miss type args")
	}
	url, ok := args["url"]
	if !ok {
		return nil, errors.New("miss url args")
	}
	eventArgs["url"] = url

	switch qType {
	case "api":
		eventArgs["event_type"] = "1"
	case "chain":
		eventArgs["event_type"] = "2"
	}

	callback, ok := args["callback"]
	if !ok {
		return nil, errors.New("miss callback args")
	}
	subArgs["callback"] = callback

	res1, err := smart_contract.CallContract("oracle", "NewEvent", eventArgs)
	if err != nil {
		log.Errorf("call newEvent contract error: %s", err)
		return nil, err
	}
	eventRes, _ := res1.(meta.ContractUpdateData)
	res2, err := smart_contract.CallContract("oracle", "NewSub", subArgs)
	if err != nil {
		log.Errorf("call newSub contract error: %s", err)
		return nil, err
	}
	subRes, _ := res2.(meta.ContractUpdateData)
	res.Events = eventRes.Events
	res.EventSubs = subRes.EventSubs
	return res, nil
}
