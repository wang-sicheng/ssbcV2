package contract

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/account"
	"github.com/ssbcV2/global"
	"github.com/ssbcV2/meta"
	"github.com/ssbcV2/util"
	"os/exec"
	"plugin"
)

// 将智能合约编译成动态库
func GoBuildPlugin(contractName string) error {
	var output bytes.Buffer

	// 执行编译命令
	cmd := exec.Command("go", "build", "-buildmode=plugin", contractName+".go")
	log.Infof("node id: %s", global.NodeID)
	cmd.Dir = "./contract/contract/" + global.NodeID + "/" + contractName
	cmd.Stderr = &output
	err := cmd.Run()
	if err != nil {
		log.Info("合约部署错误: " + err.Error())
		return err
	}
	return nil
}

func execute(name, method string, args map[string]interface{}) (interface{}, error) {
	defer func() {
		if err := recover(); err != nil {
			Info("合约执行异常，请检查代码和参数\n", err)
		}
	}()

	// 参数校验
	if name == "" || method == "" {
		return nil, errors.New("invalid call params")
	}

	dir := "./contract/contract/" + global.NodeID + "/" + name + "/"

	p, err := plugin.Open(dir + name + ".so")
	if err != nil {
		return nil, err
	}
	f, err := p.Lookup(method)
	if err != nil {
		Infof("找不到目标方法：%v，执行FallBack方法", method)
		f, err := p.Lookup("Fallback")
		if err != nil {
			Info("没有提供Fallback方法")
			return nil, err
		}
		f, ok := f.(func(map[string]interface{}) (interface{}, error))
		if !ok {
			Info("调用失败，方法参数类型应为 map[string]interface{}，返回值应为 (interface{}, error)")
			return nil, err
		}
		a, err := f.(func(map[string]interface{}) (interface{}, error))(args)
		if err != nil {
			return nil, err
		}
		log.Infof("执行结果：%v\n", a)
		return a, nil
	}
	Infof("调用 %v 方法", method)
	f, ok := f.(func(map[string]interface{}) (interface{}, error))
	if !ok {
		Info("调用失败，方法参数类型应为 map[string]interface{}，返回值应为 (interface{}, error)")
		return nil, err
	}
	a, err := f.(func(map[string]interface{}) (interface{}, error))(args)
	if err != nil {
		return nil, err
	}
	log.Infof("执行结果：%v\n", a)
	return a, nil
}


// 获取目标智能合约中的指定数据
func Get(name string, targets []string) (map[string]interface{}, error) {
	// 参数校验
	if name == "" || targets == nil || len(targets) == 0 {
		return nil, errors.New("invalid get params")
	}

	dir := "./contract/contract/" + global.NodeID + "/" + name + "/"

	p, err := plugin.Open(dir + name + ".so")
	if err != nil {
		return nil, err
	}
	res := map[string]interface{}{}
	for _, target := range targets {
		f, err := p.Lookup(target)
		if err != nil {
			log.Infof("找不到数据：%v\n", target)
			continue
		}
		res[target] = f
	}
	return res, nil
}

// 第一次调用合约前加载合约信息
func SetContext(task meta.ContractTask) {
	contractAccount := account.GetContractByName(task.Name)

	curContext = context{
		Name:    task.Name,
		Address: contractAccount.Address,
		Method:  task.Method,
		Args:    task.Args,
		Balance: contractAccount.Balance,
		Caller:  task.Caller,
		Origin:  task.Caller,
		Value:   task.Value,
	}
}

// 合约调用合约时设置合约信息
func SetRecurContext(name string, method string, args map[string]interface{}, value int) {
	if len(stack.contexts) == 0 { // 用户调用合约时（第一次调用）context已经设置好
		stack.Push(curContext) // context设置完毕，入栈
		return
	}
	currContract := account.GetContractByName(curContext.Name) // 当前合约账户信息
	nextContract := account.GetContractByName(name)            // 即将被调合约账户信息

	curContext = context{
		Name:    name,
		Address: nextContract.Address,
		Method:  method,
		Args:    args,
		Balance: nextContract.Balance,
		Caller:  currContract.Address, // 调用者为当前合约
		Origin:  curContext.Origin,
		Value:   value,
	}

	stack.Push(curContext) // context设置完毕，入栈
}

func PrintContext() {
	bs, _ := json.Marshal(curContext)
	var out bytes.Buffer
	_ = json.Indent(&out, bs, "", "\t")
	log.Infof("当前合约调用的context: %v\n", out.String())
}

// 生成合约地址（虽然合约地址不应该由公私钥生成）
func GenerateContractAddress() string {
	//首先生成公私钥
	_, pubKey := util.GetKeyPair()
	//账户地址
	//将公钥进行hash
	pubHash, _ := util.CalculateHash(pubKey)
	//将公钥hash作为账户地址,256位
	address := hex.EncodeToString(pubHash)
	log.Infof("contract account address len: %d", len(address))
	return address
}
