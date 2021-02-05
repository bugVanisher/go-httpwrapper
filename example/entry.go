package main

import (
	"fmt"
	"github.com/myzhan/boomer"
	"httpwrapper"
	"time"
)

func main() {
	start := time.Now()
	templateJsonStr := `{
    "debug": true,
    "domain": "https://postman-echo.com",
    "declare": ["{{ $sessionId := getSid }}"],
    "init_variables": {
        "roomId": 1001,
        "sessionId": "{{ $sessionId }}",
        "ids": "{{ $sessionId }}"
    },
    "running_variables": {
    	"tid": "{{ getId 5000 }}"
    },
    "func_set": [
        {
            "key": "getTest",
            "method": "GET",
            "url": "/get?name=gannicus&roomId={{ .roomId }}&age=10&tid={{ .tid }}",
            "body": "{\"timeout\":10000}",
            "validator": "{{ and  (eq .http_status_code 200) (eq .args.age \"10\") }}"
        },
        {
            "key": "postTest",
            "method": "POST",
            "header":{
               "Cookie": "{{ .tid }}",
               "Content-Type": "application/json"
            },
            "url": "/post?name=gannicus",
            "body": "{\"timeout\":{{ .tid }}, \"retry\":true}",
            "validator": "{{ and  (eq .http_status_code 200) (eq .data.timeout (.tid | toFloat64 ) ) (eq .data.retry false) }}"
        }
    ]
}`
	tasks := httpwrapper.GetTaskList(templateJsonStr)
	boomer.Run(tasks...)
	fmt.Println("cost:", time.Since(start))
}
