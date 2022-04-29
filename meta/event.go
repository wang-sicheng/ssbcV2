package meta

import (
	"encoding/hex"
	"github.com/rjkris/go-jellyfish-merkletree/common"
)

type JFTreeData interface {
	GetKey() common.HashValue
}

// 用于前端展示
type EventInfo struct {
	Type          string            `json:"type"`          // 0:内部事件;1:pull-api事件;2:pull-跨链事件;3:push事件;4:订阅
	EventID       string            `json:"event_id"`      // 事件ID
	Args          map[string]interface{} `json:"args"`          // 事件参数
	FromAddress   string            `json:"from_address"`  // 定义方
	Subscriptions []string          `json:"subscriptions"` // 订阅方
	ChainId       string            `json:"chain_id"`      // push目标链

	SubID          string `json:"sub_id"`          // 订阅ID
	ContractName   string `json:"contract_name"`   // 回调合约名
	ContractMethod string `json:"contract_method"` // 回调合约方法
	Total          int    `json:"total"`           // 触发数量
	Useful         bool   `json:"useful"`          // 是否生效
}

type Event struct {
	Type          string // 0:内部事件;1:pull-api事件;2:pull-跨链事件;3:push事件
	EventID       string
	Args          map[string]interface{}
	FromAddress   string   // 事件定义方
	Subscriptions []string // 订阅方
	ChainId string
}

type EventSub struct {
	SubID       string
	EventID     string
	TargetEvent Event    // 支持订阅自定义事件，此时不需要eventId
	Callback    Callback // 回调智能合约，处理逻辑
	Publisher   []string // 支持对部分发布者产生响应
	EventRate   int      // 触发频率
	Useful      bool     // 是否生效
	FromAddress string   // 事件订阅方
	Total       int      // 触发数量
}

type EventMessage struct {
	From      string
	EventID   string
	Data      map[string]interface{}
	Sign      []byte
	PublicKey string
	TimeStamp string
	Hash      []byte
}

type Callback struct {
	Caller   string // 调用者地址
	Value    int    // 调用合约的转账金额
	Contract string
	Method   string
	Args     map[string]interface{}
	Address  string // 合约地址
}

type EventMessageParams struct {
	From      string `json:"from"`
	EventKey  string `json:"event_key"`
	PublicKey string `json:"public_key"`
	Args      string `json:"args"`
}

func (e Event) GetKey() common.HashValue {
	keyBytes, _ := hex.DecodeString(e.EventID)
	return common.BytesToHash(keyBytes)
}

func (es EventSub) GetKey() common.HashValue {
	keyBytes, _ := hex.DecodeString(es.SubID)
	return common.BytesToHash(keyBytes)
}
