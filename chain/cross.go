package chain

import (
	"github.com/ssbcV2/meta"
)

//获取到本链的抽象区块头集合
func GetLocalAbstractBlockChainHeaders(chainId string) []meta.AbstractBlockHeader {
	//首先获取到本链区块链
	bc := GetCurrentBlockChain()
	bcHeaders := make([]meta.AbstractBlockHeader, 0)
	for h, block := range bc {
		var bcHeader meta.AbstractBlockHeader
		bcHeader = meta.AbstractBlockHeader{
			ChainId:    chainId,
			Height:     h,
			Hash:       block.Hash,
			PreHash:    block.PrevHash,
			MerkleRoot: block.MerkleRoot,
		}
		bcHeaders = append(bcHeaders, bcHeader)
	}
	return bcHeaders
}
