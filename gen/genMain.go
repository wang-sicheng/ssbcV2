package gen

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

func Gen() (string, error) {
	ok := checkSourcePath(bpmnFilepath, "bpmn")
	if !ok {
		return "", errors.New("找不到bpmn文件")
	}

	ok = checkSourcePath(owlFilepath, "owl")
	if !ok {
		return "", errors.New("找不到本体文件")
	}

	tasks := parseAndGen(bpmnFilepath)
	// 本体->数据结构
	str, _ := genOWL(owlFilepath)
	res := genTasks(tasks, str)
	return res, nil
}

func checkSourcePath(filepath string, fileType string) bool {
	_, err := os.OpenFile(filepath, os.O_RDONLY, 0755)
	if errors.Is(err, os.ErrNotExist) {
		fmt.Println(fileType + "地址输入错误")
		return false
	}
	return true
}

func parseAndGen(bpmn_filePath string) []*Task {
	// tree
	var doc = parseElementTree(bpmn_filePath)
	// sequence
	var flows = getFlows(doc)

	// task name and type
	var nameMap = make(map[string]string) // make(map[key_type]value_type)
	var typeMap = make(map[string]string)
	var functionNames = NewSet()
	var err = getNameAndTypeMappings(doc, typeMap, nameMap, functionNames)
	if err != "" {
		log.Fatal(err)
	}

	// access control
	var orgs []string
	var laneMap = make(map[string]string)
	err = getOrgsAndAccess(doc, orgs, laneMap)
	if err != "" {
		log.Fatal(err)
	}

	//task flow
	var incomingMap = make(map[string][]string)
	var outgoingMap = make(map[string][]string)
	err = getDependencies(flows, incomingMap, outgoingMap, typeMap, nameMap, laneMap, functionNames)

	taskObjArray := formArray(incomingMap, outgoingMap, typeMap, nameMap, laneMap)
	return taskObjArray
}

// 生成本体->数据结构的代码
func genOWL(owlPath string) (string, error) {
	classes := parseOWL(owlPath)
	if len(classes) == 0 {
		return "", errors.New("本体模型中未描述类或解析失败")
	}

	var writeIn string
	//自动生成标识
	writeIn += "/**以下代码由程序自动生成，注释标识的方法需要手动填充内容*/\n"
	writeIn += "package main \n\n"
	writeIn += "import (\n\t\"github.com/ssbcV2/contract\"\n\t\"time\"\n\t\"github.com/cloudflare/cfssl/log\"\n)\n\n"
	for _, class := range classes {
		classObjectName := strings.ToLower(class.name[1:2]) + class.name[2:]
		writeIn += "type " + class.name[1:] + " struct{\n"
		if class.subClassOf != "" {
			writeIn += "\t" + class.subClassOf[1:] + "\n"
		}
		for _, dp := range class.dps {
			writeIn += "\t" + dp.name[1:] + " " + dp.dataType + "\n"
		}
		writeIn += "}\n\n"

		for _, op := range class.ops {
			if op.hasFlag {
				writeIn += "func " + op.name[1:] + " () " + op.dataType + "{\n"
				writeIn += "\treturn " + strings.ToLower(class.name[1:2]) + class.name[2:] + "." + op.name[4:] + " \n}\n"
			}
		}

		writeIn += "var " + classObjectName + " " + class.name[1:] + "\n\n"
	}
	writeIn = genFuncs(owlPath, writeIn)
	//加入Fallback函数
	writeIn += "\n\nfunc Fallback(args map[string]string) (interface{}, error) {\n\tcontract.Transfer(contract.Caller(), contract.Value()) \n\treturn nil, nil\n}"
	//err := ioutil.WriteFile("./out/"+unique_id+"/struct.go", []byte(writeIn), os.ModePerm) //写入文件
	//if err != nil {
	//	return errors.New("本体写入失败")
	//}

	return writeIn, nil
}

func genFuncs(filePath string, writeIn string) string {
	metaSWRLs, biaSet := parseSWRL(filePath)
	// 初始化用于比较的两个参数
	writeIn += "var var1 interface{}\nvar var2 interface{}"
	for _, metaSWRL := range metaSWRLs {
		// 暂时先设定为可调用的函数
		writeIn += "\n\n func " + strings.ToUpper(metaSWRL.name[:1]) + metaSWRL.name[1:] + "(args map[string]string)(interface{}, error){\n"

		for k, v := range metaSWRL.classMap {
			if strings.HasPrefix(v, "#") {
				//应对 IRI为"#a"标记方式
				v = v[1:]
			} else {
				//应对IRI为"urn:swrl:var#h"标记方式的
				v = v[13:]
			}
			writeIn += "\t" + v + " := " + strings.ToLower(k[:1]) + k[1:] + "\n"
			// 将类信息打印出来
			writeIn += "\tlog.Info(" + v + ")\n"
		}

		for _, opa := range metaSWRL.opas {
			//input := opa.funcVar[0]
			output := opa.funcVar[1]
			if strings.HasPrefix(output, "#") {
				//应对 IRI为"#a"标记方式
				output = output[1:]
			} else {
				//应对IRI为"urn:swrl:var#h"标记方式的
				output = output[13:]
			}
			writeIn += "\t" + output + " := " + opa.funcName + "()\n"
			// 将取出的参数都用log打印出来
			writeIn += "\tlog.Info(" + output + ")\n"
		}

		writeIn += "\treturn true"
		for _, bia := range metaSWRL.bia {
			//var1 := bia.funcVar[0]
			//var2 := bia.funcVar[1]
			writeIn += "&&" + bia.funcName + "()"
		}
		writeIn += ",nil\n}"
	}
	// 把内置函数放进去 支持greaterThanOrEqual,greaterThan
	for _, name := range biaSet {
		writeIn += "\n" + addBIA(name)
	}

	return writeIn
}

type Event struct {
	Token           int
	Type            string
	ID              string
	Name            string
	AndToken        string
	Children        string
	Lane            string
	FunctionControl string
	StartControl    string
}

func addBIA(biaName string) string {
	if biaName == "greaterThanOrEqual" {
		return "\nfunc greaterThanOrEqual()bool{\n\tif v1,ok1 := var1.(int);ok1{\n\t\tv2,_:=var2.(int)\n\t\treturn v1>=v2\n\t} else if v1,ok1 := var1.(float32);ok1{\n\t\tv2,_:=var2.(float32)\n\t\treturn v1>=v2\n\t} else if v1,ok1 := var1.(time.Time);ok1{\n\t\tv2,_:=var2.(time.Time)\n\t\treturn v1.Before(v2)\n\t}else{\n\t\treturn false\n\t}\n}"
	} else if biaName == "greaterThan" {
		return "\nfunc greaterThan()bool{\n\tif v1,ok1 := var1.(int);ok1{\n\t\tv2,_:=var2.(int)\n\t\treturn v1>v2\n\t} else if v1,ok1 := var1.(float32);ok1{\n\t\tv2,_:=var2.(float32)\n\t\treturn v1>v2\n\t} else if v1,ok1 := var1.(time.Time);ok1{\n\t\tv2,_:=var2.(time.Time)\n\t\treturn v1.Before(v2)\n\t}else{\n\t\treturn false\n\t}\n}"
	} else {
		return ""
	}
}

func genTasks(tasks []*Task, writeIn string) string {
	for _, task := range tasks {
		// 非task类型函数不用生成方法
		if task.Type != "task" {
			continue
		}

		if task.Name == "" {
			continue
		}

		// task name必须以大写字母开头
		task.Name = strings.ToUpper(task.Name[:1]) + task.Name[1:]

		//生成空方法
		writeIn += "\n\n /**此方法需要手动填充内容*/\nfunc " + task.Name + "(args map[string]string)(interface{}, error){\n\treturn nil,nil\n}"
	}
	return writeIn
}
