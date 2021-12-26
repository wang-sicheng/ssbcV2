package util

import (
	"crypto/rand"
	"github.com/cloudflare/cfssl/log"
	"math/big"
)

// 返回一个十位数的随机数，作为msgid
func GetRandom() int {
	x := big.NewInt(10000000000)
	for {
		result, err := rand.Int(rand.Reader, x)
		if err != nil {
			log.Error(err)
		}
		if result.Int64() > 1000000000 {
			return int(result.Int64())
		}
	}
}
