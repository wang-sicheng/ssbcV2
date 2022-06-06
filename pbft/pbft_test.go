package pbft

import (
	"encoding/hex"
	"fmt"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/util"
	"regexp"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	pri, pub := util.GetKeyPair()
	log.Info("pri=", string(pri))
	log.Info("pub=", string(pub))
	//将公钥进行hash
	pubHash, _ := util.CalculateHash(pub)
	//将公钥的前20位作为账户地址
	account := hex.EncodeToString(pubHash[:20])
	log.Info(account)
}

func TestRe(t *testing.T) {
	r, _ := regexp.Compile("(.*).go")
	res := r.FindStringSubmatch("oracle.go")
	fmt.Println(res)
}
