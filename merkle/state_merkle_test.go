package merkle

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/rjkris/go-jellyfish-merkletree/common"
	"github.com/ssbcV2/meta"
	"gotest.tools/assert"
	"math/rand"
	"testing"
	"time"
)

func RandString(len int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		b := r.Intn(26) + 65
		bytes[i] = byte(b)
	}
	return string(bytes)
}

func TestOneUpdate(t *testing.T) {
	var accounts []meta.JFTreeData
	path := "../levelDB/db/path/statedb"
	accounts = append(accounts, meta.Account{
		Address:    "jklirogregerg",
		Balance:    0,
		Data:       meta.AccountData{},
		PublicKey:  "",
		PrivateKey: "",
		IsContract: false,
	})
	_, err := UpdateStateTree(accounts, uint64(0), path)
	_, err = UpdateStateTree(accounts, uint64(3), path)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateAndVerify(t *testing.T) {
	var accounts []meta.JFTreeData
	nums := 100
	//var rootHashs []common.HashValue
	for i := 0; i < nums; i++ {
		//key := RandString(32)// 随机生成address
		key := common.HashValue{}.Random().Bytes()
		keyStr := hex.EncodeToString(key)
		//t.Errorf("key: %s", string(key))
		accounts = append(accounts, meta.Account{
			Address:    keyStr,
			Balance:    100,
			Data:       meta.AccountData{},
			PublicKey:  "",
			PrivateKey: "",
			IsContract: false,
		})
	}
	path := "../levelDB/db/path/statedb"
	rootHash, _ := UpdateStateTree(accounts, uint64(0), path)
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
	for i := 0; i < nums; i++ {
		account, _ := accounts[i].(meta.Account)
		actualAccountBytes, proof, _ := getProofValue(account.Address, uint64(0), path)
		var actualAccount meta.Account
		_ = json.Unmarshal(actualAccountBytes, &actualAccount)
		t.Logf("actualAccount: %+v, address: %v", actualAccount, []byte(actualAccount.Address))
		t.Logf("account: %+v, address: %v", account, []byte(account.Address))
		assert.Equal(t, actualAccount, account)
		verifyRes, _ := ProofVerify(rootHash, proof, account.Address, account)
		assert.Equal(t, verifyRes, true)
	}
}

func TestEqual(t *testing.T) {

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

func TestByteList(t *testing.T) {
	key := common.HashValue{}.Random().Bytes()
	var byteArray [32]byte
	for i, v := range key {
		fmt.Println(i)
		byteArray[i] = v
	}
}
