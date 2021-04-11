package chain

import (
	"github.com/davecgh/go-spew/spew"
	"testing"
)

func TestSetAndStore(t *testing.T) {
	gb := GenerateGenesisBlock()
	BlockChain = append(BlockChain, gb)
	StoreBlockChain(BlockChain)
	bc := GetCurrentBlockChain()
	spew.Dump(bc)
}

func TestGetLocalAbstractBlockChainHeaders(t *testing.T) {
	abs := GetLocalAbstractBlockChainHeaders("ssbc")
	spew.Dump(abs)
}
