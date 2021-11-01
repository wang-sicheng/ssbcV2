package meta

//账户

type Account struct {
	Address    string      `json:"address"` //账户地址
	Balance    int         `json:"balance"` //账户余额
	Data       AccountData `json:"data"`
	PublicKey  []byte      `json:"publickey"`  //账户公钥
	PrivateKey []byte      `json:"privatekey"` //账户私钥
}

type AccountData struct {
	Code         string //合约代码
	ContractName string //合约名称
}
