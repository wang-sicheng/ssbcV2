package util

import (
	"crypto/rand"
	"encoding/json"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/meta"
	"math/big"
	"net"
	"os"
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

// 判断文件或文件夹是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		log.Info(err)
		return false
	}
	return true
}

// 判断数组是否包含该元素
func Contains(arr []string, target string) bool {
	for _, a := range arr {
		if a == target {
			return true
		}
	}
	return false
}

// 使用tcp发送消息
func TCPSend(msg meta.TCPMessage, addr string) {
	conn, err := net.Dial("tcp", addr)
	defer conn.Close()
	if err != nil {
		log.Error("[TCPSend]connect error,err:", err, "msg:", msg, "addr:", addr)
		return
	}
	context, _ := json.Marshal(msg)
	_, err = conn.Write(context)
	if err != nil {
		log.Error(err)
	}
}
