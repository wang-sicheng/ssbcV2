package merkle

import (
	"encoding/json"
	"github.com/cloudflare/cfssl/log"
	"github.com/rjkris/go-jellyfish-merkletree/common"
	"github.com/rjkris/go-jellyfish-merkletree/jellyfish"
	"github.com/ssbcV2/meta"
)

var StatePath string

// 更新账户state数据到新的版本，生成新的root hash
func UpdateAccountState(accounts []meta.Account, version uint64) (common.HashValue, error) {
	db := jellyfish.NewTreeStore(StatePath)
	defer db.Db.Close()
	tree := jellyfish.JfMerkleTree{
		db,
		nil,
	}
	var kvs []jellyfish.ValueSetItem
	for _, account := range accounts {
		k := common.BytesToHash([]byte(account.Address))
		accountBytes, _ := json.Marshal(account)
		kvs = append(kvs, jellyfish.ValueSetItem{
			k,
			jellyfish.ValueT{accountBytes},
		})
	}
	rootHash, batch := tree.PutValueSet(kvs, jellyfish.Version(version))
	err := db.WriteTreeUpdateBatch(batch)
	if err != nil {
		log.Errorf("state update batch error: %s \n", err)
		return rootHash, err
	}
	return rootHash, nil
}

// 获取账户数据和proof
func getProofValue(address string, version uint64) (meta.Account, jellyfish.SparseMerkleProof, error) {
	db := jellyfish.NewTreeStore(StatePath)
	defer db.Db.Close()
	tree := jellyfish.JfMerkleTree{db, nil}
	k := common.BytesToHash([]byte(address))
	proofValue, proof := tree.GetWithProof(k, jellyfish.Version(version))
	var account meta.Account
	err := json.Unmarshal(proofValue.GetValue(), &account)
	if err != nil {
		log.Errorf("proofValue unmarshal error: %s\n", err)
		return account, proof, err
	}
	return account, proof, nil
}

// 存在性验证
func ProofVerify(rootHash common.HashValue, proof jellyfish.SparseMerkleProof, address string, value meta.Account) (bool, error) {
	k := common.BytesToHash([]byte(address))
	accountBytes, err := json.Marshal(value)
	if err != nil {
		log.Errorf("account marshal error: %s", err)
		return false, err
	}
	res := proof.Verify(rootHash, k, jellyfish.ValueT{accountBytes})
	return res, nil
}
