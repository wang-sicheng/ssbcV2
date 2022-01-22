package meta

// 交易类型
const (
	Transfer      int = iota // 0: 转账交易
	Register                 // 1: 注册账户
	Publish                  // 2: 发布合约
	Invoke                   // 3: 调用合约
	CrossTransfer            // 4: 跨链转账
)

type Transaction struct {
	From      string            `json:"from"`
	To        string            `json:"to"`
	Dest      string            `json:"dest"`
	Contract  string            `json:"contract"`
	Method    string            `json:"method"`
	Args      map[string]string `json:"args"`
	Data      TransactionData   `json:"data"`
	Value     int               `json:"value"`
	Id        []byte            `json:"id"`
	Timestamp string            `json:"timestamp"`
	Hash      []byte            `json:"hash"`
	PublicKey string            `json:"public_key"`
	Sign      []byte            `json:"sign"`
	Type      int               `json:"type"`

	// 以下参数只有在跨链时使用
	SourceChainId string         `json:"source_chain"`
	DestChainId   string         `json:"dest_chain"`
	TransId       []byte         `json:"tran_id"`
	Param         CrossTranParam `json:"param"`
	Proof         CrossTranProof `json:"proof"`
	Status        string         `json:"status"`
}

type TransactionData struct {
	Read map[string]string
	Set  map[string]string
	Code string
}

// 用户提交跨链交易的参数
type PostCrossTran struct {
	SourceChain string `json:"source_chain"`
	DestChain   string `json:"dest_chain"`
	From        string `json:"from"`
	To          string `json:"to"`
	Contract    string `json:"contract"`
	Method      string `json:"method"`
	Args        string `json:"args"`
	Value       int    `json:"value"`
	PrivateKey  string `json:"private_key"`
	PublicKey   string `json:"public_key"`
	Sign        string `json:"sign"`
	Type        int    `json:"type"`
}

type PostTran struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Dest       string `json:"dest"`
	Contract   string `json:"contract"`
	Method     string `json:"method"`
	Args       string `json:"args"`
	Value      int    `json:"value"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	Sign       string `json:"sign"`
	Type       int    `json:"type"`
}

type Block struct {
	Height     int
	Timestamp  string
	PrevHash   []byte
	MerkleRoot []byte
	Signature  []byte
	Hash       []byte
	TX         []Transaction
	StateRoot  []byte
	EventRoot  []byte
}
