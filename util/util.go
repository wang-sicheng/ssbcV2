package util

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/meta"
	"go/ast"
	"go/parser"
	"go/token"
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

// 从合约代码中解析出包名、方法、全局变量
func ParseContract(code string) meta.ContractInfo {
	set := token.NewFileSet()
	f, err := parser.ParseFile(set, "", code, 0)
	if err != nil {
		fmt.Println("Failed to parse code:", err)
		return meta.ContractInfo{}
	}

	decls := f.Decls
	variables := []string{}
	for _, decl := range decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		ast.Inspect(genDecl, func(n ast.Node) bool {
			var s string
			switch x := n.(type) {
			case *ast.ValueSpec:
				for _, vs := range x.Names {
					if vs.Obj != nil && vs.Obj.Kind == ast.Var {
						s = vs.Name
					}
					if s != "" {
						variables = append(variables, s)
					}
				}
			}
			return true
		})
	}

	methods := []string{}

	for _, d := range f.Decls {
		if fn, isFn := d.(*ast.FuncDecl); isFn {
			methods = append(methods, fn.Name.String())
		}
	}

	log.Infof("包名: %v \n", f.Name)
	log.Infof("合约的全部变量：%+v\n", variables)
	log.Infof("合约的全部方法: %+v\n", methods)

	return meta.ContractInfo{
		Package:   f.Name.Name,
		Variables: variables,
		Methods:   methods,
	}
}

func MapToJson(param map[string]string) string {
	dataType , _ := json.Marshal(param)
	dataString := string(dataType)
	return dataString
}

func JsonToMap(str string) map[string]interface{} {
	var tempMap map[string]interface{}
	err := json.Unmarshal([]byte(str), &tempMap)
	if err != nil {
		panic(err)
	}
	return tempMap
}
