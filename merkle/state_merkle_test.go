package merkle

import (
	"encoding/json"
	"github.com/rjkris/go-jellyfish-merkletree/common"
	"github.com/ssbcV2/meta"
	"gotest.tools/assert"
	"testing"
)

func TestOneUpdate(t *testing.T)  {
	var accounts []meta.Account
	accounts = append(accounts, meta.Account{
		Address:    "jklirogregerg",
		Balance:    0,
		Data:       meta.AccountData{},
		PublicKey:  "",
		PrivateKey: "",
		IsContract: false,
	})
	_, err := UpdateAccountState(accounts, uint64(0))
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateAndVerify(t *testing.T)  {
	var accounts []meta.Account
	nums := 10
	//var rootHashs []common.HashValue
	for i:=0; i < nums; i++ {
		key := common.HashValue{}.Random().Bytes() // 随机数确保每次生成的key不同
		//t.Errorf("key: %s", string(key))
		accounts = append(accounts, meta.Account{
			Address:    string(key),
			Balance:    100,
			Data:       meta.AccountData{},
			PublicKey:  "",
			PrivateKey: "",
			IsContract: false,
		})
	}
	StatePath = "../levelDB/db/path/statedb"
	rootHash, _ := UpdateAccountState(accounts, uint64(0))
	//for ver:=0; ver < 10; ver++ {
	//	rootHash, err := UpdateAccountState(accounts[ver*1000:(ver+1)*1000], uint64(ver))
	//	if err != nil {
	//		t.Errorf("update account state error: %+v", err)
	//	}
	//	rootHashs = append(rootHashs, rootHash)
	//}
	//for i:=0; i < nums; i++ {
	//	ver := i%1000
	//	account := accounts[i]
	//	actualAccount, proof, _ := getProofValue(account.Address, uint64(ver))
	//	assert.Equal(t, actualAccount, account)
	//	verifyRes, _ := ProofVerify(rootHashs[ver], proof, account.Address, account)
	//	assert.Equal(t, verifyRes, true)
	//}
	for i:=0; i < nums; i++ {
		account := accounts[i]
		actualAccount, proof, _ := getProofValue(account.Address, uint64(0))
		t.Errorf("actualAccount: %+v, address: %v", actualAccount, []byte(actualAccount.Address))
		t.Errorf("account: %+v, address: %v", account, []byte(account.Address))
		assert.Equal(t, actualAccount, account) // TODO:经json序列化后address发生变化，导致校验失败
		verifyRes, _ := ProofVerify(rootHash, proof, account.Address, account)
		assert.Equal(t, verifyRes, true)
	}
}

func TestEqual(t *testing.T)  {

	//account2 := meta.Account{
	//	Address:    "123",
	//	Balance:    0,
	//	Data:       meta.AccountData{},
	//	PublicKey:  "",
	//	PrivateKey: "",
	//	IsContract: false,
	//}
	//assert.Equal(t, account1, account2)
	key := common.HashValue{}.Random().Bytes()
	afterKey := []byte(string(key))
	t.Logf("key: %v, after: %v", key, afterKey)

	account1 := meta.Account{
		Address:    string(common.HashValue{}.Random().Bytes()),
		Balance:    0,
		Data:       meta.AccountData{},
		PublicKey:  "",
		PrivateKey: "",
		IsContract: false,
	}
	accountBytes, _ := json.Marshal(&account1)
	var newAccount1 meta.Account
	_ = json.Unmarshal(accountBytes, &newAccount1)
	t.Logf("account1: %v, new: %v", []byte(account1.Address), []byte(newAccount1.Address))
	assert.Equal(t, account1.Address, newAccount1.Address)
}