package meta

//账户

type Account struct {
	Address    string      `json:"address"` //账户地址
	Balance    int         `json:"balance"` //账户余额
	Data       AccountData `json:"data"`
	PublicKey  []byte      `json:"publickey"`  //账户公钥
	PrivateKey []byte      `json:"privatekey"` //账户私钥
	IsContract bool		   `json:"iscontract"` // 是否是智能合约账户
}

type AccountData struct {
	Code         string 	`json:"code"`
	ContractName string 	`json:"contractname"`
}
