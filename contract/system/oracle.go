package main

import (
	"encoding/json"
	"errors"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/contract" // 调用其他智能合约时引入
	"github.com/ssbcV2/meta"
)

// 定义事件函数
func NewEvent(args map[string]string) (interface{}, error) {
	var res meta.ContractUpdateData
	var event meta.Event

	event.FromAddress = contract.Caller()
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
	sub.FromAddress = contract.Caller()
	eid, ok := args["event_id"]
	if ok {
		sub.EventID = eid // 不是必须，也可以自定义targetEvent
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
	log.Infof("newSub callback info: %+v", cb)
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

// 回退函数，当没有方法匹配时执行此方法
func Fallback(args map[string]string) (interface{}, error) {
	return nil, nil
}
