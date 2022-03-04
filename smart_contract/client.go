package smart_contract

import (
	"bytes"
	"errors"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/global"
	"os/exec"
	"plugin"
)

/* 智能合约模板（golang plugin）
`
package main	// 包名必须为main

import (
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/smart_contract"		// 调用其他智能合约时引入
)

// 参数必须为 map[string]string, 返回结果必须为 (interface{}, error)
func Multiply(args map[string]string) (interface{}, error) {
	// 调用其他合约，自行封装参数
	num, err := smart_contract.CallContract("random", "GetRandom", map[string]string{})
	if err != nil {
		log.Infof("[Multiply] 调用random失败")
	}
	a := num.(int)
	log.Infof("[Multiply] 调用 random.GetRandom 成功，结果：%v\n", a)
	ans := a * a
	log.Infof("[Multiply] 结果：%v\n", ans)
	return ans, nil
}
`
*/

// 将智能合约编译成动态库
func GoBuildPlugin(contractName string) (err error, errStr string) {
	var output bytes.Buffer

	// 执行编译命令
	cmd := exec.Command("go", "build", "-buildmode=plugin", contractName + ".go")
	cmd.Dir = "./smart_contract/contract/" + global.NodeID + "/" + contractName
	cmd.Stderr = &output
	err = cmd.Run()
	if err != nil {
		log.Info(output.String())
		log.Error(err)
		return err, output.String()
	} else {
		log.Info(output.String())
	}
	return nil, ""
}

// 调用智能合约
func CallContract(name string, method string, args map[string]string) (interface{}, error) {
	// 参数校验
	if name == "" || method == "" {
		return nil, errors.New("invalid call params")
	}

	dir := "./smart_contract/contract/" + global.NodeID + "/" + name + "/"
	log.Info("call contract: " + dir)
	p, err := plugin.Open(dir + name + ".so")
	if err != nil {
		return nil, err
	}
	f, err := p.Lookup(method)
	if err != nil {
		log.Infof("找不到目标方法：%v，执行FallBack方法", method)
		f, err := p.Lookup("Fallback")
		if err != nil {
			log.Info("没有提供Fallback方法")
			return nil, err
		}
		a, _ := f.(func(map[string]string) (interface{}, error))(args)
		log.Infof("执行结果：%v\n", a)
		return a, nil
	}

	a, _ := f.(func(map[string]string) (interface{}, error))(args)
	log.Infof("执行结果：%v\n", a)
	return a, nil
}

