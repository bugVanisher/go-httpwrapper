package main

import (
	"bytes"
	"fmt"
	"github.com/ghodss/yaml"
	jsoniter "github.com/json-iterator/go"
	"httpwrapper"
	"log"
	"os"
	"reflect"
	"strings"
	"text/template"
)

func Funcs() {
	const templateText = `
    Nrd   Friend : {{last .Friends}}
`

	templateFunc := map[string]interface{}{
		"last": func(s []string) string { return s[len(s)-1] },
	}
	type Recipient struct {
		Name    string
		Friends []string
	}

	recipient := Recipient{
		Name:    "Jack",
		Friends: []string{"Bob", "Json", "Tom"},
	}
	t := template.Must(template.New("").Funcs(templateFunc).Parse(templateText))
	err := t.Execute(os.Stdout, recipient)
	if err != nil {
		fmt.Println("Executing template:", err)
	}
}
func Yaml() {
	yamlStr := `declare: '{{ $sessionId := getSid }}'
init_variables:
  roomId: 1001
  sessionId: {{ $sessionId }}
  ids: {{ getIds $sessionId }}
running_variables: ~
func_set:
  - key: getTest
    method: GET
    url: /gameabr/api/v1/1001/settings/
    body: {"timeout": 10000}
    validator: {{ and  (eq .roomId.Int64 1001) }}
    `
	getIds := func(sid int) string {
		return "3000"
	}
	getSid := func() int {
		return 100100
	}
	templateFunc := map[string]interface{}{
		"getIds": getIds,
		"getSid": getSid,
	}
	t := template.Must(template.New("Variables").Funcs(templateFunc).Parse(yamlStr))
	var tmpBytes bytes.Buffer
	_ = t.Execute(&tmpBytes, nil)
	fmt.Println(tmpBytes.String())
	bout, err := yaml.YAMLToJSON(tmpBytes.Bytes())
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bout))
	var variable httpwrapper.Variables
	decoder := jsoniter.NewDecoder(strings.NewReader(string(bout)))
	decoder.UseNumber()
	err = decoder.Decode(&variable)
	if err != nil {
		fmt.Println(err)
		log.Fatal()
	}
	validator(variable)
}

func validator(v httpwrapper.Variables) {
	validator := `{{ and  (eq .roomId.Int64 1001) }}`
	t := template.Must(template.New("Validator").Parse(validator))
	var bs bytes.Buffer
	fmt.Println(reflect.TypeOf(v.InitVariables["roomId"]), reflect.TypeOf(v.InitVariables["sessionId"]))
	err := t.Execute(&bs, v.InitVariables)
	if err != nil {
		log.Fatal()
	}
	fmt.Println(bs.String())
}
