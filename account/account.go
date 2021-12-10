package account

import (
	"encoding/json"
	"github.com/cloudflare/cfssl/log"
	commonconst "github.com/ssbcV2/common"
	"github.com/ssbcV2/levelDB"
	"github.com/ssbcV2/meta"
)

/* 这里封装了所有的对账户的操作
 * 每个节点默认包含一个全局的State，所以这里直接将State设置为私有，
 * 不用每个节点显式地创建，直接在init()创建
 * 通过调用函数进行操作
 */

var state State	// 私有，通过函数进行操作

type State struct {
	Accounts map[string]meta.Account	// Accounts 存储了所有账户（普通账户和合约账户），key: 账户地址(合约账户用的name) - val: 账户信息
}

func init() {
	state.Accounts = map[string]meta.Account{}
}

// 创建普通账户
func CreateAccount(address, publicKey string, balance int) meta.Account{
	account := meta.Account{
		Address:    address,
		Balance:    balance,
		Data:       meta.AccountData{},
		PrivateKey: "",
		PublicKey:  publicKey,
	}
	state.Accounts[address] = account

	PutIntoDisk(state.Accounts)
	return account
}

// 创建智能合约账户
func CreateContract(address, publicKey, code, name string) meta.Account{
	contract := meta.Account{
		Address: address,
		Balance: 0,
		Data:  meta.AccountData{
			Code:         code,
			ContractName: name,
		},
		PublicKey: publicKey,
		IsContract: true,
	}
	// 用智能合约的名称作为key，合约地址暂时没有使用
	state.Accounts[name] = contract

	PutIntoDisk(state.Accounts)
	return contract
}

func SubBalance(sender string, amount int) meta.Account{
	senderAccount := state.Accounts[sender]
	if senderAccount.Balance < amount {		// 调用SubBalance前会先调用CanTransfer，理论上不会出现余额不足的情况
		log.Infof("[SubBalance]: Insufficient balance.")
	}
	senderAccount.Balance -= amount
	state.Accounts[sender] = senderAccount

	PutIntoDisk(state.Accounts)
	return senderAccount
}

func AddBalance(receiver string, amount int) meta.Account{
	receiverAccount := state.Accounts[receiver]
	receiverAccount.Balance += amount
	state.Accounts[receiver] = receiverAccount

	PutIntoDisk(state.Accounts)
	return receiverAccount
}

// 判断交易发起方是否有足够余额
func CanTransfer(sender string, amount int) bool {
	senderAccount := state.Accounts[sender]
	if senderAccount.Balance < amount {
		log.Infof("[CanTransfer]: Insufficient balance.")
		return false
	}
	return true
}

// 持久化（每次对账户信息的更改都需要持久化到磁盘）
// 目前也还没有考虑事务和回滚
func PutIntoDisk(accounts map[string]meta.Account) {
	bytes, _ := json.Marshal(accounts)
	levelDB.DBPut(commonconst.AccountsKey, bytes)
}

// 从磁盘获取已有的账户信息（在节点启动时执行）
func GetFromDisk() {
	accountBytes := levelDB.DBGet(commonconst.AccountsKey)
	_ = json.Unmarshal(accountBytes, &state.Accounts)
}

// 账户地址是否存在
func ContainsAddress(address string) bool {
	_, ok := state.Accounts[address]
	return ok
}

// 获取账户信息
func GetAccount(address string) meta.Account {
	return state.Accounts[address]
}

// 获取所有的账户地址
func GetTotalAddress() []string {
	var totalAddress []string
	for address := range state.Accounts {
		totalAddress = append(totalAddress, address)
	}
	return totalAddress
}

// 是否为普通账户地址
func IsOrdinaryAccount(address string) bool {
	return !state.Accounts[address].IsContract
}

// 是否为智能合约账户账户地址
func IsContractAccount(address string) bool {
	return state.Accounts[address].IsContract
}
