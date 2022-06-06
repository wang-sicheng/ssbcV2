package gen

import (
	"github.com/beevik/etree"
	"regexp"
)

type Task struct {
	ID       string
	Name     string
	Type     string
	Lane     string
	Children []string
	Parents  []string
}

// mapping上id和incoming/outgoing id
func getDependencies(flows []*etree.Element, incomingMap map[string][]string, outgoingMap map[string][]string, typeMap map[string]string, nameMap map[string]string, laneMap map[string]string, functionNames *set) string {
	for iter := 0; iter < len(flows); iter++ {
		// XOR有条件注明的需要将条件转化为一个task 用于flow control
		if typeMap[flows[iter].SelectAttr("sourceRef").Value] == "XOR" && flows[iter].SelectAttr("id") != nil {
			// 在sequence id的基础上生成一个新的task id
			var oldId = flows[iter].SelectAttr("id").Value
			var newId = "Condition_" + oldId[13:] //13是'SequenceFlow_'的长度
			// 新方法需要加入typeMap functionSet nameMap laneMap
			typeMap[newId] = "task"
			if functionNames.Contains(flows[iter].SelectAttr("name")) {
				return "重复方法：" + flows[iter].SelectAttr("name").Value
			}
			functionNames.Add(flows[iter].SelectAttr("name"))
			nameMap[newId] = flows[iter].SelectAttr("name").Value
			laneMap[newId] = laneMap[flows[iter].SelectAttr("sourceRef").Value]

			insert(incomingMap, newId, flows[iter].SelectAttr("sourceRef").Value)
			insert(outgoingMap, flows[iter].SelectAttr("sourceRef").Value, newId)
			insert(incomingMap, flows[iter].SelectAttr("targetRef").Value, newId)
			insert(outgoingMap, newId, flows[iter].SelectAttr("targetRef").Value)

		} else {
			insert(incomingMap, flows[iter].SelectAttr("targetRef").Value, flows[iter].SelectAttr("sourceRef").Value)
			insert(incomingMap, flows[iter].SelectAttr("sourceRef").Value, flows[iter].SelectAttr("targetRef").Value)
		}
	}

	for source := range outgoingMap {
		if typeMap[source] == "XOR" && len(outgoingMap[source]) > 1 {
			for iter := 0; iter < len(outgoingMap[source]); iter++ {
				target := outgoingMap[source][iter]
				if typeMap[target] == "XOR" || typeMap[target] == "AND" {
					return "不支持嵌套gateway"
				}
			}
		}
	}

	return ""
}

//返回Task Array
func formArray(incomingMap map[string][]string, outgoingMap map[string][]string, typeMap map[string]string, nameMap map[string]string, laneMap map[string]string) []*Task {
	var arr []*Task
	for id := range typeMap {
		task := new(Task)
		task.ID = id
		task.Type = typeMap[id]
		task.Lane = laneMap[id]
		task.Name = nameMap[id]
		task.Children = outgoingMap[id]
		task.Parents = incomingMap[id]
		arr = append(arr, task)
	}
	return arr
}

func insert(incomingMap map[string][]string, id string, value string) {
	//if len(incomingMap[id])==0{
	//	incomingMap[id] := []string
	//}
	incomingMap[id] = append(incomingMap[id], value)
}

// laneMap: id:laneName
func getOrgsAndAccess(doc *etree.Document, orgs []string, laneMap map[string]string) string {
	// laneMap:
	var lanes = doc.FindElements("//laneSet/lane")
	var laneNames = NewSet()

	if len(lanes) == 0 {
		return "流程图至少需要一条泳道lane"
	}

	for iter := 0; iter < len(lanes); iter++ {
		var err = processLane(lanes[iter], orgs, laneNames, laneMap)
		if err != "" {
			return err
		}
	}

	return ""
}

func processLane(lane *etree.Element, orgs []string, laneNames *set, laneMap map[string]string) string {
	laneName := lane.SelectAttr("name")
	if laneName == nil {
		return "所有泳道lane必须被命名" + lane.SelectAttr("id").Value
	}
	if match, _ := regexp.Match("^[0-9a-zA-Z_]+$", []byte(laneName.Value)); !match {
		return "所有泳道lane的命名只能包括a-z 0-9和下划线_" + laneName.Value
	}
	if lane.SelectElement("childLaneSet") != nil {
		return "不支持嵌套lane"
	}

	if laneNames.Contains(laneName) {
		return "泳道命名重复" + laneName.Value
	}
	laneNames.Add(laneName.Value)
	orgs = append(orgs, laneName.Value)
	var allTasks = lane.SelectElements("flowNodeRef")
	var numTasks = len(allTasks)
	for iter := 0; iter < numTasks; iter++ {
		laneMap[allTasks[iter].Text()] = laneName.Value
	}
	return ""
}

//mapping包括typeMap: id string: type string;nameMap: id string:name string
func getNameAndTypeMappings(doc *etree.Document, typeMap map[string]string, nameMap map[string]string, functionNames *set) string {
	// 对所有task标签做mapping
	var sendTasks = doc.FindElements("//sendTask")
	var receiveTasks = doc.FindElements("//receiveTask")
	var userTasks = doc.FindElements("//userTask")
	var manualTasks = doc.FindElements("//manualTask")
	var businessRuleTasks = doc.FindElements("//businessRuleTask")
	var scriptTasks = doc.FindElements("//scriptTask")
	var serviceTasks = doc.FindElements("//serviceTask")

	var tasks = doc.FindElements("//task")
	tasks = append(tasks, sendTasks...)
	tasks = append(tasks, receiveTasks...)
	tasks = append(tasks, userTasks...)
	tasks = append(tasks, manualTasks...)
	tasks = append(tasks, businessRuleTasks...)
	tasks = append(tasks, scriptTasks...)
	tasks = append(tasks, serviceTasks...)

	for iter := 0; iter < len(tasks); iter++ {
		typeMap[tasks[iter].SelectAttr("id").Value] = "task"
		if functionNames.Contains(tasks[iter].SelectAttr("name").Value) {
			return "有重复的task名称： " + tasks[iter].SelectAttr("name").Value
		} else {
			nameMap[tasks[iter].SelectAttr("id").Value] = tasks[iter].SelectAttr("name").Value
			functionNames.Add(tasks[iter].SelectAttr("name").Value)
		}
	}

	// 构建所有start的map
	var starts = doc.FindElements("//startEvent")
	for iter := 0; iter < len(starts); iter++ {
		curId := starts[iter].SelectAttr("id").Value
		typeMap[curId] = "START"
		nameMap[curId] = starts[iter].SelectAttr("name").Value
	}

	// 构建所有end的map
	var ends = doc.FindElements("//endEvent")
	for iter := 0; iter < len(ends); iter++ {
		curId := ends[iter].SelectAttr("id").Value
		typeMap[curId] = "END"
		nameMap[curId] = ends[iter].SelectAttr("name").Value
	}

	// 事件
	var events = doc.FindElements("//intermediateThrowEvent")
	for iter := 0; iter < len(events); iter++ {
		curId := events[iter].SelectAttr("id").Value
		typeMap[curId] = "event"
		nameMap[curId] = ends[iter].SelectAttr("name").Value
	}

	// 异或
	var xors = doc.FindElements("//exclusiveGateway")
	for iter := 0; iter < len(xors); iter++ {
		curId := xors[iter].SelectAttr("id").Value
		typeMap[curId] = "XOR"
		nameMap[curId] = ends[iter].SelectAttr("name").Value
	}

	// 与
	var ands = doc.FindElements("//parallelGateway")
	for iter := 0; iter < len(ands); iter++ {
		curId := ands[iter].SelectAttr("id").Value
		typeMap[curId] = "AND"
		nameMap[curId] = ends[iter].SelectAttr("name").Value
	}

	//不支持or
	var ors = doc.FindElements("//inclusiveGateway")
	if len(ors) > 0 {
		return "不支持或"
	}
	return ""
}

func getFlows(doc *etree.Document) []*etree.Element {
	return doc.FindElements("//sequenceFlow")
}
func parseElementTree(filePath string) *etree.Document {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(filePath); err != nil {
		panic(err)
	}
	return doc
}
