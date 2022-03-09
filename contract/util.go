package contract

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/account"
	"github.com/ssbcV2/global"
	"github.com/ssbcV2/meta"
	"os/exec"
	"plugin"
)

// 将智能合约编译成动态库
func GoBuildPlugin(contractName string) (err error, errStr string) {
	var output bytes.Buffer

	// 执行编译命令
	cmd := exec.Command("go", "build", "-buildmode=plugin", contractName+".go")
	log.Infof("node id: %s", global.NodeID)
	cmd.Dir = "./contract/contract/" + global.NodeID + "/" + contractName
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

func execute(name, method string, args map[string]string) (interface{}, error) {
	// 参数校验
	if name == "" || method == "" {
		return nil, errors.New("invalid call params")
	}

	dir := "./contract/contract/" + global.NodeID + "/" + name + "/"
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

// 第一次调用合约前加载合约信息
func SetContext(task meta.ContractTask) {
	contractAccount := account.GetAccount(task.Name)
	curContext.Name = task.Name
	curContext.Balance = contractAccount.Balance
	curContext.Caller = task.Caller
	curContext.Origin = task.Caller
	curContext.Value = task.Value
	curContext.Method = task.Method
}

// 合约调用合约时设置合约信息
func SetRecurContext(name string, method string, args map[string]string, value int) {
	if len(stack.contexts) == 0 { // 用户调用合约时（第一次调用）不执行该函数
		stack.Push(curContext) // context设置完毕，入栈
		return
	}
	curContext.Caller = curContext.Name // 调用者为上一个合约

	curContext.Name = name
	curContext.Method = method
	curContext.Args = args
	curContext.Value = value

	contract := account.GetAccount(name)
	curContext.Balance = contract.Balance

	stack.Push(curContext) // context设置完毕，入栈
}

func PrintContext() {
	bs, _ := json.Marshal(curContext)
	var out bytes.Buffer
	_ = json.Indent(&out, bs, "", "\t")
	log.Infof("当前合约调用的context: %v\n", out.String())
}
