package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PostTran struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Dest     string `json:"dest"`
	Contract string `json:"contract"`
	Method   string `json:"method"`
	Args     string `json:"args"`
	Value      int    `json:"value"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	Sign       string `json:"sign"`
}

const (
	SCISSORS	= 1		// 剪刀
	HAMMER   	= 2		// 石头
	CLOTH	 	= 3		// 布
)

const (
	A	= "c511c24e01a6c9d287a8bb011b2f4bb9b798d7a7"
	B	= "5ca4b2a0d7bcf0340442ebec25e8a436c7783e57"
	C	= "a4eef1dea05ed28c51dc23e3438c236f95924ee4"
)
var states = map[string]int {
	A: 0,
	B: 0,
	C: 0,
}

func setAction(sender string, action int) bool {
	if (sender == A || sender == B || sender == C) && (action == SCISSORS || action == HAMMER || action == CLOTH)  {
		states[sender] = action
		return true
	}
	return false
}

func whoIsWinner() (string, bool) {
	if states[A] == 0 || states[B] == 0 || states[C] == 0 {	// 还有人没有出结果
		return "", false
	}

	if states[A] != states[B] && states[A] != states[C] && states[B] != states[C] {
		reset()
		return "", false
	}
	if states[A] == states[B] {
		if check(states[C], states[B]) {
			return C, true
		} else {
			return "", false
		}
	}
	if states[A] == states[C] {
		if check(states[B], states[A]) {
			return B, true
		} else {
			return "", false
		}
	}
	if states[B] == states[C] {
		if check(states[A], states[B]) {
			return A, true
		} else {
			return "", false
		}
	}
	return "", false
}

func check(a, b int) bool {
	if a == 1 && b == 3 {
		return true
	} else if a == 2 && b == 1 {
		return true
	} else if a == 3 && b == 2 {
		return true
	}
	return false
}


func reset() {
	states[A] = 0
	states[B] = 0
	states[C] = 0
}

func postTran(from, to string, value int) {
	url := "http://host.docker.internal:9999/postTran"
	//url := "http://localhost:9999/postTran"
	info := PostTran{
		From: from,
		To: to,
		PrivateKey: "MIICXQIBAAKBgQDimHHmdmFHmfqYIMV+xsvG6nFKd8PqU6ljlaV10eAdc8mTSsW9 keZ+uQFLYUlh45APPAiyRcInHhn2FzFtEO7XIxK+/0CqKEUzexGiXzeISVwuFLTo CNlfzZwnRHnfr/jRY3MDZ1EZrQgRZGlqyTyxKPghqinzZEmGPNVX/uzn0QIDAQAB AoGBAIyLc3I3kMUBe44qHXUxxqj9NwGyYUERXSoYYoU+hNyfubJzGU0olqeZBnWD xSlDJVJdsSMp42+x2vZpkk2MyCZbWjnWcBVx6p1q4s4yM6umofwrJHfD4e+Y2Nl5 2DKr5NFv2u4rSbTQxUtB1hCAFL4UK56AQ7iMETE+MFSmKQPxAkEA5o+hIHiPBiBP ccHCsKRj30d+G9bqbonCtSKeE9ewpVuy02jZ1mQ6MtMFWHO/eOSIr4OyBmoTq1lD pnS21tTpJQJBAPuYzh+C2YD57MmFHwzvkegFQ+nS7qfls0JNUrwTWJMynwcQ7q+6 i+PFQreeA5iCcM4wBMds3RKG0FR60rrd0j0CQGji8lQJRFdvH3UKxn0BbAXJSk9z 59Y9iXxJsUwplUzEeIfAbUkg83DnmsjwbyyaGqxt5vEQFL6gryvscLkuxpkCQQCe O3b/KGski4pZHyjtGMqZsp4Is4k2/Oalfz3WXRBq2v5bElIbIOaT5F7WXkGCrB7H /jkzNws+eJ0TVH+t2I49AkBYk7gmoegdWqukH3NIOOWt1m4zwncrYxLNisWrONoh NI1qD0cJGwp3DP7AYQTekQChv9RAyabCqxC0ri/Ukg6v",
		PublicKey: "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDimHHmdmFHmfqYIMV+xsvG6nFK d8PqU6ljlaV10eAdc8mTSsW9keZ+uQFLYUlh45APPAiyRcInHhn2FzFtEO7XIxK+ /0CqKEUzexGiXzeISVwuFLToCNlfzZwnRHnfr/jRY3MDZ1EZrQgRZGlqyTyxKPgh qinzZEmGPNVX/uzn0QIDAQAB",
		Value: value,
		Contract: "",
		Method: "method",
		Dest: "dest",
		Sign: "",
		Args: "{}",
	}
	jsons,_:=json.Marshal(info)
	result :=string(jsons)
	jsoninfo :=strings.NewReader(result)

	req, _ := http.NewRequest("POST", url, jsoninfo)
	res, err := http.DefaultClient.Do(req)
	if  err !=nil{
		log.Printf("调用接口异常%v",err.Error())
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if  err !=nil{
		log.Printf("调用接口异常%v",err.Error())
	}
	fmt.Println(string(body))
}

func handler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	sender := query.Get("sender")
	actionStr := query.Get("action")
	action, _ := strconv.Atoi(actionStr)
	setAction(sender, action)
	if winner, ok := whoIsWinner(); ok {
		_, _ = w.Write([]byte(fmt.Sprintf("winner: %s\n", winner)))
		var losers []string
		if winner == A {
			losers = append(losers, B, C)
		} else if winner == B {
			losers = append(losers, A, C)
		} else {
			losers = append(losers, A, B)
		}
		for _, loser := range losers {
			_, _ = w.Write([]byte(fmt.Sprintf("loser: %s\n", loser)))
			postTran(loser, winner, 888)
			time.Sleep(5 * time.Second)
		}
	}
	_, _ = w.Write([]byte(fmt.Sprintf("Hello, %s, your action: %d!\n", sender, action)))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/setAction", handler)
	server := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Println("Starting Server.")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
