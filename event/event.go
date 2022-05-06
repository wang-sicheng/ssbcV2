package event

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/account"
	"github.com/ssbcV2/common"
	"github.com/ssbcV2/contract"
	"github.com/ssbcV2/global"
	"github.com/ssbcV2/levelDB"
	"github.com/ssbcV2/meta"
	"github.com/ssbcV2/redis"
	"github.com/ssbcV2/util"
)

// 事件树状态数据
var EventData map[string]meta.JFTreeData
// 事件数据
var EventDbInfo map[string]meta.Event
// 订阅数据
var SubDbInfo map[string]meta.EventSub
// 预言机合约账户
var OracleAccounts map[string]meta.Account
// 链下报告
var OracleReports []meta.UnderChainReport

func init() {
	EventData = map[string]meta.JFTreeData{}
	EventDbInfo = map[string]meta.Event{}
	SubDbInfo = map[string]meta.EventSub{}
	OracleAccounts = map[string]meta.Account{}
	OracleReports = []meta.UnderChainReport{}
}

func IsContainsKey(key string) bool {
	_, ok := EventData[key]
	return ok
}

func IsContainsEventKey(key string) bool {
	_, ok := EventDbInfo[key]
	return ok
}

func IsContainsSubKey(key string) bool {
	_, ok := SubDbInfo[key]
	return ok
}

func InitEventData() {
	eventBytes := levelDB.DBGet(common.EventKey)
	if len(eventBytes) != 0 {
		err := json.Unmarshal(eventBytes, &EventDbInfo)
		if err != nil {
			log.Errorf("事件数据解析失败：%s", err)
			return
		}
	}
	subBytes := levelDB.DBGet(common.EventSubKey)
	if len(subBytes) != 0 {
		err := json.Unmarshal(subBytes, &SubDbInfo)
		if err != nil {
			log.Errorf("订阅数据解析失败：%s", err)
			return
		}
	}
	for id, e := range EventDbInfo {
		EventData[id] = e
	}
	for id, s := range SubDbInfo {
		EventData[id] = s
	}
}

func UpdateToLevelDB(data map[string]meta.JFTreeData) {
	for id, info := range data {
		event, ok := info.(meta.Event)
		if ok {
			EventDbInfo[id] = event
		}
		sub, ok := info.(meta.EventSub)
		if ok {
			SubDbInfo[id] = sub
		}
	}
	eventBytes, _ := json.Marshal(EventDbInfo)
	subBytes, _ := json.Marshal(SubDbInfo)
	levelDB.DBPut(common.EventKey, eventBytes)
	levelDB.DBPut(common.EventSubKey, subBytes)
	log.Infof("事件数据已更新至leveldb: %+v", EventDbInfo)
	log.Infof("订阅数据已更新至leveldb: %+v", SubDbInfo)
}

// 将事件消息转换成需要上链的交易
func EventToTransaction(message meta.EventMessage) ([]meta.Transaction, error) {
	log.Infof("eventToTransaction: current eventData: %+v", EventData)
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
		if len(contractArgs) == 0 { // 初始化
			contractArgs = make(map[string]interface{})
		}
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
		log.Infof("事件消息转换交易列表: %+v", trans)
		// 更新订阅信息
		sub.Total ++
		EventData[subKey] = sub
	}
	UpdateToLevelDB(EventData)
	return trans, nil
}

// 执行智能合约，并将事件触发的智能合约放入队列，更新事件
// todo：将智能合约的执行结果更新到accounts
func HandleContractTask() error {
	task := global.TaskList[0]
	global.TaskList = global.TaskList[1:]

	contract.SetContext(task) // 加载合约的相关信息，供合约内部使用
	res, err := contract.Call(task.Name, task.Method, task.Args)
	if err != nil {
		log.Infof("合约执行错误：%v", err)
		return err
	}
	if res == nil {
		log.Infof("合约执行结果为空：%s, %s", task.Name, task.Method)
		return nil
	}
	data, ok := res.(meta.ContractUpdateData)
	if !ok {
		log.Error("contract update data decode error")
	}
	// 处理事件消息
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
			global.TaskList = append(global.TaskList, meta.ContractTask{
				Caller: sData.Callback.Caller,
				Value:  sData.Callback.Value,
				Name:   sData.Callback.Contract,
				Method: sData.Callback.Method,
				Args:   sData.Callback.Args,
			})
		}
	}

	// 处理事件和订阅信息
	eList, err := UpdateEventData(data, contract.Caller())
	if err != nil {
		log.Error(err)
		return err
	}
	global.TreeData = append(global.TreeData, eList...)
	return nil
}

// 生成事件和订阅数据（暂时不考虑更新），在部署合约时使用
// address:合约地址，from:部署合约的外部账户地址
// 暂时不考虑订阅当前智能合约
func ExecuteInitEvent(name string, address string, from string) ([]meta.JFTreeData, error) {
	var args map[string]interface{}

	contract.SetContext(meta.ContractTask{
		Caller: from,
		Name:   name,
		Method: "initEvent",
		Args:   args,
	})
	res, err := contract.Call(name, "initEvent", args) // 事件数据在智能合约中的initEvent函数中定义
	if err != nil {
		log.Errorf("initEvent run error: %s", err)
		return nil, err
	}
	data, ok := res.(meta.ContractUpdateData)
	if !ok {
		return nil, errors.New("contract update data decode error")
	}
	dataList, err := UpdateEventData(data, from)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return dataList, err
}

// 生成事件和订阅数据
func UpdateEventData(data meta.ContractUpdateData, from string) ([]meta.JFTreeData, error) {
	var treeDataList []meta.JFTreeData
	events := data.Events
	subs := data.EventSubs
	if !account.ContainsAddress(from) {
		return nil, errors.New("from address is not exist in db: " + from)
	}
	ac := account.GetAccount(from)

	curSeq := ac.Seq
	// 生成新的event
	for index, _ := range events {
		curSeq++
		eventHash, _ := util.CalculateHash([]byte(from + string(curSeq))) // 外部账户地址和seq唯一决定一个事件
		events[index].EventID = hex.EncodeToString(eventHash)
		events[index].FromAddress = from
		EventData[events[index].EventID] = events[index] // 先更新到内存中，最后统一落库
		treeDataList = append(treeDataList, events[index])
		log.Infof("注册事件成功: %+v", events[index])
		if global.NodeID == global.Master {
			_ = pushEventToRedis(events[index])
		}
	}
	// 生成新的订阅
	for index, s := range subs {
		eid := s.EventID
		tarEvent := s.TargetEvent
		if eid == "" { // eid不存在，自动生成event
			curSeq++
			eventHash, _ := util.CalculateHash([]byte(from + string(curSeq))) // 外部账户地址和seq唯一决定一个事件
			tarEvent.EventID = hex.EncodeToString(eventHash)
			tarEvent.FromAddress = from
			EventData[tarEvent.EventID] = tarEvent // 先更新到内存中，最后统一落库
			eid = tarEvent.EventID
			subs[index].EventID = eid
			treeDataList = append(treeDataList, tarEvent)
			if global.NodeID == global.Master {
				_ = pushEventToRedis(tarEvent)
			}
			log.Infof("注册事件成功: %+v", tarEvent)
		} else {
			if !IsContainsEventKey(eid){ // 要订阅的事件不存在
				log.Errorf("the event to sub is not exist: %s", eid)
				continue
			}
		}
		curSeq++
		subHash, _ := util.CalculateHash([]byte(from + string(curSeq)))
		subs[index].SubID = hex.EncodeToString(subHash)
		subs[index].FromAddress = from
		subs[index].Useful = true
		edata, _ := EventData[eid].(meta.Event)
		edata.Subscriptions = append(edata.Subscriptions, subs[index].SubID) // 更新事件的订阅信息
		log.Infof("订阅信息注册成功: %+v", subs[index])

		EventData[subs[index].SubID] = subs[index]
		EventData[eid] = edata
		treeDataList = append(treeDataList, subs[index]) // 更新到状态树
		treeDataList = append(treeDataList, edata)
	}
	account.SeqIncrease(curSeq, from) // 更新from账户seq
	UpdateToLevelDB(EventData)

	return treeDataList, nil
}

// 输出到redis消息队列用于预言机监听
func pushEventToRedis(event meta.Event) error {
	eventBytes, _ := json.Marshal(event)
	err := redis.PushToList(common.RedisEventKey, string(eventBytes))
	if err != nil {
		return err
	}
	log.Infof("事件输出到队列: %+v", event)
	return nil
}

func GetAllEventData() ([]meta.EventInfo, error) {
	EventBytes := levelDB.DBGet(common.EventKey)
	SubBytes := levelDB.DBGet(common.EventSubKey)
	var eventMap map[string]meta.Event
	var subMap map[string]meta.EventSub
	res := []meta.EventInfo{}	// 初始化为空数组，而非nil，以便前端展示
	if len(EventBytes) != 0 {
		err := json.Unmarshal(EventBytes, &eventMap)
		if err != nil {
			log.Errorf("事件数据解析失败：%s", err)
		}
	}
	if len(SubBytes) != 0 {
		err := json.Unmarshal(SubBytes, &subMap)
		if err != nil {
			log.Errorf("订阅数据解析失败：%s", err)
		}
	}
	for _, e := range eventMap {
		res = append(res, meta.EventInfo{
			Type:          e.Type,
			EventID:       e.EventID,
			Args:          e.Args,
			FromAddress:   e.FromAddress,
			Subscriptions: e.Subscriptions,
			ChainId:       e.ChainId,
		})
	}
	for _, s := range subMap {
		res = append(res, meta.EventInfo{
			Type:           "4",
			EventID:        s.EventID,
			FromAddress:    s.FromAddress,
			SubID:          s.SubID,
			ContractName:   s.Callback.Contract,
			ContractMethod: s.Callback.Method,
			Total:          s.Total,
		})
	}
	return res, nil
}

// 新建预言机智能合约账户
func CreateOracleAccount(name string, code string)  {
	contractInfo := util.ParseContract(code)
	newAccount := account.CreateContract(name, contract.GenerateContractAddress(), code, "", contractInfo)
	OracleAccounts[name] = newAccount
}

//func UpdateEventData(name string, address string, from string) ([]meta.JFTreeData, error) {
//	var args map[string]interface{}
//	var treeDataList []meta.JFTreeData
//	res, err := contract.Call(name, "initEvent", args) // 事件数据在智能合约中的initEvent函数中定义
//	if err != nil {
//		log.Errorf("initEvent run error: %s", err)
//		return nil, err
//	}
//	data, ok := res.(meta.ContractUpdateData)
//	if !ok {
//		return nil, errors.New("contract update data decode error")
//	}
//	events := data.Events
//	subs := data.EventSubs
//	if !account.ContainsAddress(from) {
//		return nil, errors.New("from address is not exist in db: " + from)
//	}
//	ac := account.GetAccount(from)
//	curSeq := ac.Seq
//	// 生成新的event
//	for index, _ := range events {
//		curSeq ++
//		eventHash, _ := util.CalculateHash([]byte(from+string(curSeq))) // 外部账户地址和seq唯一决定一个事件
//		events[index].EventID = hex.EncodeToString(eventHash)
//		events[index].FromAddress = from
//		EventData[events[index].EventID] = events[index] // 先更新到内存中，最后统一落库
//		treeDataList = append(treeDataList, events[index])
//	}
//	// 生成新的订阅
//	for index, s := range subs {
//		eid := s.EventID
//		if !IsContainsKey(eid) { // 要订阅的事件不存在
//			log.Errorf("the event to sub is not exist: %s", eid)
//			continue
//		}
//		curSeq ++
//		subHash, _ := util.CalculateHash([]byte(from+string(curSeq)))
//		subs[index].SubID = hex.EncodeToString(subHash)
//		subs[index].FromAddress = from
//		edata, _ := EventData[eid].(meta.Event)
//		edata.Subscriptions = append(edata.Subscriptions, subs[index].SubID) // 更新事件的订阅信息
//
//		EventData[subs[index].SubID] = subs[index]
//		treeDataList = append(treeDataList, subs[index])
//	}
//	UpdateToLevelDB(EventData)
//	return treeDataList, nil
//}
