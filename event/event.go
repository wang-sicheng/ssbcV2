package event

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/prometheus/common/log"
	"github.com/ssbcV2/account"
	"github.com/ssbcV2/common"
	"github.com/ssbcV2/levelDB"
	"github.com/ssbcV2/meta"
	"github.com/ssbcV2/smart_contract"
	"github.com/ssbcV2/util"
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

func UpdateToLevelDB(data map[string]meta.JFTreeData)  {
	dataBytes, _ := json.Marshal(data)
	levelDB.DBPut(common.EventAllDataKey, dataBytes)
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

// 执行智能合约，并将事件触发的智能合约放入队列
// todo：将智能合约的执行结果更新到accounts
func HandleContractTask(taskList *[]ContractTask) error {
	task := (*taskList)[0]
	*taskList = (*taskList)[1:]

	res, err := smart_contract.CallContract(task.Name, task.Method, task.Args)
	if err != nil {
		return err
	}
	data, ok := res.(meta.ContractUpdateData)
	if !ok {
		log.Error("contract update data decode error")
	}
	for _, msg := range data.Messages {
		e, ok := EventData[msg.EventID]
		if !ok {
			log.Errorf("event is not exist: %+v", msg)
			continue
		}
		eData, ok := e.(meta.Event)
		if !ok {
			log.Error("event data decode error")
			continue
		}
		subs := eData.Subscriptions
		for _, sub := range subs {
			s, ok := EventData[sub]
			if !ok {
				log.Errorf("sub is not exist: %+v", s)
			}
			sData, ok := s.(meta.EventSub)
			if !ok {
				log.Errorf("sub data decode error")
				continue
			}
			*taskList = append(*taskList, ContractTask{
				Name:   sData.Callback.Contract,
				Method: sData.Callback.Method,
				Args:   sData.Callback.Args,
			})
		}
	}
	return nil
}

// 生成事件和订阅数据（暂时不考虑更新），在部署合约时使用
// address:合约地址，from:部署合约的外部账户地址
// 暂时不考虑订阅当前智能合约
func UpdateEventData(name string, address string, from string) ([]meta.JFTreeData, error) {
	var args map[string]string
	var treeDataList []meta.JFTreeData
	res, err := smart_contract.CallContract(name, "initEvent", args) // 事件数据在智能合约中的initEvent函数中定义
	if err != nil {
		log.Errorf("initEvent run error: %s", err)
		return nil, err
	}
	data, ok := res.(meta.ContractUpdateData)
	if !ok {
		return nil, errors.New("contract update data decode error")
	}
	events := data.Events
	subs := data.EventSubs
	if !account.ContainsAddress(from) {
		return nil, errors.New("from address is not exist in db: " + from)
	}
	ac := account.GetAccount(from)
	curSeq := ac.Seq
	// 生成新的event
	for index, _ := range events {
		curSeq ++
		eventHash, _ := util.CalculateHash([]byte(from+string(curSeq))) // 外部账户地址和seq唯一决定一个事件
		events[index].EventID = hex.EncodeToString(eventHash)
		events[index].FromAddress = from
		EventData[events[index].EventID] = events[index] // 先更新到内存中，最后统一落库
		treeDataList = append(treeDataList, events[index])
	}
	// 生成新的订阅
	for index, s := range subs {
		eid := s.EventID
		if !IsContainsKey(eid) { // 要订阅的事件不存在
			log.Errorf("the event to sub is not exist: %s", eid)
			continue
		}
		curSeq ++
		subHash, _ := util.CalculateHash([]byte(from+string(curSeq)))
		subs[index].SubID = hex.EncodeToString(subHash)
		subs[index].FromAddress = from
		edata, _ := EventData[eid].(meta.Event)
		edata.Subscriptions = append(edata.Subscriptions, subs[index].SubID) // 更新事件的订阅信息

		EventData[subs[index].SubID] = subs[index]
		treeDataList = append(treeDataList, subs[index])
	}
	UpdateToLevelDB(EventData)
	return treeDataList, nil
}

