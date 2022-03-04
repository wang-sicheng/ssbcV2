package config

import (
	"github.com/ssbcV2/global"
	"testing"
)

func TestConfigGet(t *testing.T) {
	global.RootDir = "/Users/wsc/Go/src/ssbcV2"
	if Get("env.language") != "golang" {
		t.Errorf("err")
	}
}
