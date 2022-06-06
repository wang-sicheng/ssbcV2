package gen

import (
	"regexp"
)

type metaDataProperty struct {
	name     string
	dataType string
}

type metaObjectProperty struct {
	hasFlag  bool
	name     string
	dataType string
}

type metaClass struct {
	subClassOf string
	name       string
	dps        []metaDataProperty
	ops        []metaObjectProperty
}

// 解析OWL文件中的本体
func parseOWL(owlPath string) []metaClass {
	doc := parseElementTree(owlPath)

	//构建map存储新构建的数据结构
	var classMap = make(map[string]metaClass)
	var dpMap = make(map[string]metaDataProperty)
	var opMap = make(map[string]metaObjectProperty)

	//1.从Declaration标签中将所有数据结构都解析出来，存在name-object map中
	for _, object := range doc.FindElements("//Declaration/Class") {
		var newClass = metaClass{}
		newClass.name = object.SelectAttr("IRI").Value
		classMap[newClass.name] = newClass
	}

	for _, object := range doc.FindElements("//Declaration/DataProperty") {
		var newDP = metaDataProperty{}
		newDP.name = object.SelectAttr("IRI").Value
		dpMap[newDP.name] = newDP
	}

	for _, object := range doc.FindElements("//Declaration/ObjectProperty") {
		var newOP = metaObjectProperty{}
		var attrStr = object.SelectAttr("IRI").Value
		//对于object property还需要解析是否描述有某个数据域
		// 如果是，那么将flag置为true，并把对应的数据域名称解析出来 数据域名称要求首字母为大写，其余字符均为字母
		if boolean, _ := regexp.MatchString("#has[A-Z][A-Za-z]+$", attrStr); boolean == true {
			newOP.hasFlag = true
		} else {
			newOP.hasFlag = false
		}
		newOP.name = attrStr
		opMap[newOP.name] = newOP
	}

	//2. 解析所有SubClassOf的标签，填充metaClass
	for _, subClasses := range doc.FindElements("//SubClassOf") {
		classes := subClasses.SelectElements("Class")
		child := classes[0].SelectAttr("IRI").Value
		parent := classes[1].SelectAttr("IRI").Value
		cur := classMap[child]
		cur.subClassOf = parent
		classMap[child] = cur
	}

	//3. 解析所有DataPropertyRange,填充metaDataProperty
	for _, dprs := range doc.FindElements("//DataPropertyRange") {
		dataTypeName := dprs.SelectElement("Datatype").SelectAttr("abbreviatedIRI").Value
		dpName := dprs.SelectElement("DataProperty").SelectAttr("IRI").Value
		//需要将dataType转化为go语言中的数据类型
		dataTypeName = getType(dataTypeName)
		cur := dpMap[dpName]
		cur.dataType = dataTypeName
		dpMap[dpName] = cur
	}

	//4. 解析所有ObjectPropertyDomain,填充metaClass
	for _, opds := range doc.FindElements("//ObjectPropertyDomain") {
		className := opds.SelectElement("Class").SelectAttr("IRI").Value
		opName := opds.SelectElement("ObjectProperty").SelectAttr("IRI").Value
		cur := classMap[className]
		curOp := opMap[opName]
		//// 非has开头的函数跳过这一步
		//if !curOp.hasFlag {
		//	continue
		//}
		curOp.dataType = dpMap["#"+opName[4:]].dataType
		opMap[opName] = curOp
		cur.ops = append(cur.ops, opMap[opName])
		classMap[className] = cur
	}

	//5. 解析所有DataPropertyDomain，将DataProperty填入Class中
	for _, dpds := range doc.FindElements("//DataPropertyDomain") {
		dpName := dpds.SelectElement("DataProperty").SelectAttr("IRI").Value
		className := dpds.SelectElement("Class").SelectAttr("IRI").Value
		cur := classMap[className]
		cur.dps = append(cur.dps, dpMap[dpName])
		classMap[className] = cur
	}

	return getAllValues(classMap)
}

func getType(owlType string) string {
	typeDict := make(map[string]string)
	typeDict["xsd:int"] = "int"
	typeDict["xsd:integer"] = "int"
	typeDict["xsd:double"] = "float64"
	typeDict["xsd:float"] = "float32"
	typeDict["xsd:string"] = "string"
	typeDict["xsd:boolean"] = "bool"
	typeDict["xsd:dateTime"] = "time.Time"

	if res, ok := typeDict[owlType]; ok {
		return res
	} else {
		//没有标注的数据类型都做string处理
		return "string"
	}
}

func getAllValues(m map[string]metaClass) []metaClass {
	values := make([]metaClass, 0)
	for _, value := range m {
		values = append(values, value)
	}
	return values
}

func parseSWRL(filePath string) ([]metaSWRL, []string) {
	doc := parseElementTree(filePath)

	var swrls []metaSWRL
	biaSet := make(map[string]bool)

	//遍历所有的DLSafeRule标签
	for _, rule := range doc.FindElements("//DLSafeRule") {
		var cur metaSWRL
		var classMap = make(map[string]string) // 记录var对应的class类型

		//1. 找出SWRL的name
		cur.name = rule.FindElement("Head/ObjectPropertyAtom/ObjectProperty").SelectAttr("IRI").Value[1:]

		//2. 找出所有class
		for _, class := range rule.FindElements("Body/ClassAtom") {
			className := class.SelectElement("Class").SelectAttr("IRI").Value[1:]
			v := class.SelectElement("Variable").SelectAttr("IRI").Value
			classMap[className] = v
		}
		cur.classMap = classMap

		//3. 构建OPA Map op->[]string 两个var,前一个是input var，后一个是output var
		for _, opa := range rule.FindElements("Body/ObjectPropertyAtom") {
			var newFunc metaFunc
			newFunc.funcName = opa.SelectElement("ObjectProperty").SelectAttr("IRI").Value[1:]
			var vars []string
			for _, variable := range opa.SelectElements("Variable") {
				vars = append(vars, variable.SelectAttr("IRI").Value)
			}
			newFunc.funcVar = vars
			cur.opas = append(cur.opas, newFunc)
		}

		//4. 解析所有BuiltInAtom，解析为bia数组
		for _, bia := range rule.FindElements("Body/BuiltInAtom") {
			var curBia metaFunc
			curBia.funcName = bia.SelectAttr("IRI").Value[32:] //http://www.w3.org/2003/11/swrlb# 32字符
			var vars []string
			for _, variable := range bia.SelectElements("Variable") {
				vars = append(vars, variable.SelectAttr("IRI").Value)
			}
			curBia.funcVar = vars
			cur.bia = append(cur.bia, curBia)

			biaSet[curBia.funcName] = true
		}

		swrls = append(swrls, cur)
	}
	//获取set中的funcNames
	j := 0
	var temp []string
	for k, _ := range biaSet {
		temp = append(temp, k)
		j++
	}
	return swrls, temp
}

type metaSWRL struct {
	name     string
	classMap map[string]string
	opas     []metaFunc
	bia      []metaFunc
}

type metaFunc struct {
	funcName string
	funcVar  []string
}
