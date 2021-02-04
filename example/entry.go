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
    "debug": false,
    "domain": "https://postman-echo.com",
    "declare": ["{{ $sessionId := getSid }}"],
    "init_variables": {
        "roomId": 1001,
        "sessionId": "{{ $sessionId }}",
        "ids": "{{ $sessionId }}"
    },
    "running_variables": {
    	"tid" : "{{ getId $sessionId }}"
    },
    "func_set": [
        {
            "key": "getTest",
            "method": "GET",
            "header":{
               "Cookies": "{{ .tid }}"
            },
            "url": "/get?name=gannicus&roomId={{ .roomId }}&age=10&tid={{ .tid }}",
            "body": "{\"timeout\":10000}",
            "validator": "{{ and  (eq .http_status_code 200) (eq .age \"10\") }}"
        },
        {
            "key": "postTest",
            "method": "POST",
            "header":{
               "Cookie": "{{ .tid }}",
               "Content-Type": "application/json"
            },
            "url": "/post?name=gannicus",
            "body": "{\"timeout\":{{ .tid }}}",
            "validator": "{{ and  (eq .http_status_code 200) (eq .data.timeout \"100\") }}"
        }
    ]
}`
	tasks := httpwrapper.GetTaskList(templateJsonStr)
	boomer.Run(tasks...)
	fmt.Println("cost:", time.Since(start))
}
