package merkle

import (
	"encoding/hex"
	"encoding/json"
	"github.com/cloudflare/cfssl/log"
	"github.com/rjkris/go-jellyfish-merkletree/common"
	"github.com/rjkris/go-jellyfish-merkletree/jellyfish"
	"github.com/ssbcV2/meta"
)

var AccountStatePath string
var EventStatePath string

var version uint64 = 0 // 只有在账户信息变动时，版本号才加一
var InitAccount meta.Account
var InitEvent meta.Event

func UpdateStateTree(data []meta.JFTreeData, version uint64, path string) (common.HashValue, error) {
	db := jellyfish.NewTreeStore(path)
	defer db.Db.Close()
	tree := jellyfish.JfMerkleTree{
		db,
		nil,
	}
	var kvs []jellyfish.ValueSetItem
	for _, item := range data {
		valueBytes, _ := json.Marshal(item)
		kvs = append(kvs, jellyfish.ValueSetItem{
			item.GetKey(),
			jellyfish.ValueT{valueBytes},
		})
	}
	rootHash, batch := tree.PutValueSet(kvs, jellyfish.Version(version))
	err := db.WriteTreeUpdateBatch(batch)
	if err != nil {
		log.Errorf("event state update error: %s", err)
		return rootHash, err
	}
	return rootHash, nil
}

// 获取账户数据和proof
func getProofValue(address string, version uint64, path string) ([]byte, jellyfish.SparseMerkleProof, error) {
	db := jellyfish.NewTreeStore(path)
	defer db.Db.Close()
	tree := jellyfish.JfMerkleTree{db, nil}
	addressBytes, _ := hex.DecodeString(address)
	k := common.BytesToHash(addressBytes)
	proofValue, proof := tree.GetWithProof(k, jellyfish.Version(version))
	return proofValue.GetValue(), proof, nil
}

// 存在性验证
func ProofVerify(rootHash common.HashValue, proof jellyfish.SparseMerkleProof, address string, value meta.JFTreeData) (bool, error) {
	addressBytes, _ := hex.DecodeString(address)
	k := common.BytesToHash(addressBytes)
	dataBytes, err := json.Marshal(value)
	if err != nil {
		log.Errorf("account marshal error: %s", err)
		return false, err
	}
	res := proof.Verify(rootHash, k, jellyfish.ValueT{dataBytes})
	return res, nil
}

func GetVersion() uint64 {
	curr := version
	version++
	return curr
}
