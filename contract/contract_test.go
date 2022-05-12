package contract

import (
	"encoding/json"
	"fmt"
	"testing"
)

type Student struct {
	Name   string   `json:"name"`   // 姓名
	Age    int      `json:"age"`    // 年龄
	Gender string   `json:"gender"` // 性别
	Score  float64  `json:"score"`  // 分数
	Course []string `json:"course"` // 课程
}

func Test(t *testing.T) {
	serialize1()
	serialize2()
	serialize3()
	deserialize1()
	deserialize2()
	deserialize3()
}

// slice序列化
func serialize1() {
	m1 := map[string]interface{}{
		"name":    "张三",
		"address": "广东省深圳市",
	}
	m2 := map[string]interface{}{
		"name":    "李四",
		"address": "广东省广州市",
	}
	var slice []map[string]interface{}
	slice = append(slice, m1, m2)

	data, err := json.Marshal(slice)
	if err != nil {
		fmt.Println("序列化失败", err)
	} else {
		fmt.Printf("slice  序列化后 str = %v\n", string(data))
	}
}

// map序列化
func serialize2() {
	m := map[string]interface{}{ // 使用一个空接口表示可以是任意类型
		"name":     "张三",
		"province": "广东省",
		"city":     "深圳市",
		"salary":   1000,
		"hobby":    []string{"看书", "旅游", "学习"},
	}

	data, err := json.Marshal(m)
	if err != nil {
		fmt.Println("序列化失败", err)
	} else {
		fmt.Printf("map    序列化后 str = %v\n", string(data))
	}
}

// struct序列化
func serialize3() {
	stu := Student{
		"张三",
		20,
		"男",
		78.6,
		[]string{"语文", "数学", "音乐"},
	}
	data, err := json.Marshal(&stu)
	if err != nil {
		fmt.Println("序列化失败", err)
	} else {
		fmt.Printf("struct 序列化后 str = %v\n", string(data))
	}
}

// 反序列化为slice
func deserialize1() {
	var slice []map[string]interface{}
	str := `[{"address":"广东省深圳市","name":"张三"},{"address":"广东省广州市","name":"李四"}]`
	err := json.Unmarshal([]byte(str), &slice)
	if err != nil {
		fmt.Println("反序列化失败", err)
	}
	fmt.Printf("反序列化后 slice  = %v\n", slice)

}

// 反序列化为map
func deserialize2() {
	var m map[string]interface{} // 使用一个空接口表示可以是任意类型
	str := `{"city":"深圳市","hobby":["看书","旅游","学习"],"name":"张三","province":"广东省","salary":1000}`

	err := json.Unmarshal([]byte(str), &m)
	if err != nil {
		fmt.Println("反序列化失败", err)
	}
	fmt.Printf("反序列化后 map    = %v\n", m)
}

// 反序列化为struct
func deserialize3() {
	var stu Student
	str := `{"name":"张三","age":20,"gender":"男","score":78.6,"course":["语文","数学","音乐"]}`

	err := json.Unmarshal([]byte(str), &stu)
	if err != nil {
		fmt.Println("反序列化失败", err)
	}
	fmt.Printf("反序列化后 struct = %v\n", stu)
}
