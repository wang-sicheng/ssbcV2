package meta

import (
	"encoding/hex"
	"github.com/rjkris/go-jellyfish-merkletree/common"
)

// 账户结构（外部账户和合约账户共用，通过IsContract字段区分）
type Account struct {
	Address    string      `json:"address"`    // 账户地址
	Balance    int         `json:"balance"`    // 账户余额
	Data       AccountData `json:"data"`       // 智能合约数据
	PublicKey  string      `json:"publickey"`  // 账户公钥
	PrivateKey string      `json:"privatekey"` // 账户私钥（用户的私钥不应该出现在这里，后续删除）
	IsContract bool        `json:"iscontract"` // 是否是智能合约账户
	Seq        int         `json:"seq"`        // 该账户下定义的事件序列号
}

type AccountData struct {
	Code         string `json:"code"`         // 合约代码
	ContractName string `json:"contractname"` // 合约名称
	Publisher    string `json:"publisher"`    // 部署合约的外部账户地址
}

func (ac Account) GetKey() common.HashValue {
	keyBytes, _ := hex.DecodeString(ac.Address)
	return common.BytesToHash(keyBytes)
}
