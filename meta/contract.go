package meta

type ContractInfo struct {
	Package   string   `json:"package"`		// 合约的包名
	Variables []string `json:"variables"`	// 合约的所有变量
	Methods   []string `json:"methods"`		// 合约的所有方法
}

type ContractPost struct {
	Account    string `json:"account"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	Code       string `json:"code"`
	Name       string `json:"name"`
}

type ContractUpdateData struct {
	Events    []Event
	EventSubs []EventSub
	Messages  []EventMessage
	StateData interface{}
}

type ContractTask struct {
	Caller string            // 合约调用者（外部账户、合约账户、事件...
	Value  int               // 外部账户调用合约交易的转账金额
	Name   string            // 合约名称
	Method string            // 方法
	Args   map[string]string // 参数
}
