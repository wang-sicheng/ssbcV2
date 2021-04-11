package chain

import (
	"encoding/json"
	"fmt"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/commoncon"
	"github.com/ssbcV2/merkle"
	"github.com/ssbcV2/meta"
	"github.com/ssbcV2/redis"
	"github.com/ssbcV2/util"
	"sync"
	"time"
)

var mutex = &sync.Mutex{}

var BlockChain []meta.Block //声明全局变量
func init() {
	BlockChain = make([]meta.Block, 0)
	gb := GenerateGenesisBlock()
	BlockChain = append(BlockChain, gb)
	StoreBlockChain(BlockChain)
}

//生成创世区块
func GenerateGenesisBlock() meta.Block {
	genesisBlock := meta.Block{}
	genesisBlock = meta.Block{
		Timestamp: time.Now().String(),
	}
	genesisBlock.Hash = util.CalculateBlockHash(genesisBlock)
	return genesisBlock
}

//将当前区块链存储
func StoreBlockChain(bc []meta.Block) {
	bcBytes, _ := json.Marshal(bc)
	redis.SetIntoRedis(commoncon.BlockChainKey, string(bcBytes))
}

//获取到当前区块链
func GetCurrentBlockChain() []meta.Block {
	bcStr, _ := redis.GetFromRedis(commoncon.BlockChainKey)
	bc := make([]meta.Block, 0)
	err := json.Unmarshal([]byte(bcStr), &bc)
	if err != nil {
		log.Errorf("[GetCurrentBlockChain] unmarshal failed:err=%v", err)
	}
	return bc
}

//生成新区块
func CreateNewBlock(txs []meta.Transaction) meta.Block {
	//首先获取到当前的区块链
	curBlockChain := GetCurrentBlockChain()
	length := len(curBlockChain)
	preBlock := curBlockChain[length-1]
	var newBlock = meta.Block{
		Height:    len(curBlockChain),
		Timestamp: time.Now().String(),
		PrevHash:  preBlock.Hash,
		TX:        txs,
	}
	//生成该区块的merkle root
	//区块中的每一笔交易生成对应的merkle path、index
	tree := merkle.GenerateMerkleTree(txs)
	newBlock.MerkleRoot = tree.MerkleRoot()
	newBlock.Hash = util.CalculateBlockHash(newBlock)
	return newBlock
}

//同步其他节点发来的区块链，进行本地更新
func UpdateChain(bc []meta.Block) {
	//首先获取本地的chain
	bcLocal := GetCurrentBlockChain()
	if len(bcLocal) < len(bc) {
		//更新
		StoreBlockChain(bc)
		log.Info("更新区块链成功")
		bytes, err := json.MarshalIndent(bc, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		// Green console color: 	\x1b[32m
		// Reset console color: 	\x1b[0m
		fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
		fmt.Println()
	}
}

//打印更新后的区块链
func PrintUpdatedBlockChain(bc []meta.Block) {
	bytes, err := json.MarshalIndent(bc, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	// Green console color: 	\x1b[32m
	// Reset console color: 	\x1b[0m
	fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
}
