package redis

import (
	"github.com/cloudflare/cfssl/log"
	"testing"
)

func TestExampleClient(t *testing.T) {
	ExampleClient()
}

func TestGetandSet(t *testing.T) {
	SetIntoRedis("ye", "depeng")
	v, _ := GetFromRedis("ye")
	log.Info(v)
	GetFromRedis("hu")
}
