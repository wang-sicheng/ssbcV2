package meta

type Event struct {
	EventID string
	Args map[string]string
	FromAddress string // 事件定义方
}

type EventSub struct {
	SubID string
	EventID string
	Callback Callback // 回调智能合约，处理逻辑
	Publisher []string // 支持对部分发布者产生响应
	EventRate int // 触发频率
	Useful bool // 是否生效
	FromAddress string // 事件订阅方
	Total int // 触发数量
}

type EventMessage struct {
	EventID string
	Data []byte
	Sign []byte
	PublicKey string
	TimeStamp string
}

type Callback struct {
	Contract ContractPost
}
