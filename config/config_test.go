package config

import (
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/global"
	"github.com/ssbcV2/util"
	"io/ioutil"
	"testing"
)

func TestConfigGet(t *testing.T) {
	global.RootDir = "/Users/wsc/Go/src/ssbcV2"
	if Get("env.language") != "golang" {
		t.Errorf("err")
	}
}

func TestContractParse(t *testing.T) {
	buf, err := ioutil.ReadFile("./test.go")
	if err != nil {
		log.Info("ReadFile Error")
	}
	util.ParseContract(string(buf))
}
