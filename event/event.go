package event

import (
	"encoding/json"
	"errors"
	"github.com/prometheus/common/log"
	"github.com/ssbcV2/common"
	"github.com/ssbcV2/levelDB"
	"github.com/ssbcV2/meta"
)

var EventData map[string]meta.JFTreeData

type ContractTask struct {
	Name string
	Method string
	Args map[string]string
}


func init() {
	EventData = map[string]meta.JFTreeData{}
}

func IsContainsKey(key string) bool {
	_, ok := EventData[key]
	return ok
}

func InitEventData() {
	dataBytes := levelDB.DBGet(common.EventAllDataKey)
	_ = json.Unmarshal(dataBytes, &EventData)
}

// 将事件消息转换成需要上链的交易
func EventToTransaction(message meta.EventMessage) ([]meta.Transaction, error) {
	value := EventData[message.EventID]
	event, ok := value.(meta.Event)
	if !ok {
		return nil, errors.New("event type convert error")
	}
	subs := event.Subscriptions // 事件的订阅者id
	var trans []meta.Transaction
	for _, subKey := range subs {
		subValue, ok := EventData[subKey]
		if !ok {
			log.Infof("sub key not exit: %s", subKey)
			continue
		}
		sub, ok := subValue.(meta.EventSub)
		contractArgs := sub.Callback.Args
		for k, v := range message.Data { // 增加来自event message的参数
			contractArgs[k] = v
		}
		trans = append(trans, meta.Transaction{
			From:      message.From, // 来自外部账户
			To:        sub.Callback.Address,
			Dest:      "",
			Contract:  sub.Callback.Contract,
			Method:    sub.Callback.Method,
			Args:      contractArgs,
			Data:      meta.TransactionData{},
			Value:     0,
			Id:        nil,
			Timestamp: "",
			Hash:      nil,
			PublicKey: message.PublicKey,
			Sign:      message.Sign,
			Type:      meta.Invoke,
		})
	}
	return trans, nil
}

