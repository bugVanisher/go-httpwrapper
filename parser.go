package httpwrapper

import (
	"bytes"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/myzhan/boomer"
	"github.com/rs/zerolog/log"
	"math/rand"
	"strings"
	"text/template"
	"time"
)

const (
	NoValue = "<no value>"
)

type Variable map[string]interface{}

type Variables struct {
	Declare          []string               `json:"declare"`
	InitVariables    map[string]interface{} `json:"init_variables"`
	RunningVariables map[string]interface{} `json:"running_variables"`
	MergedVariables  map[string]interface{}
}

type RunScript struct {
	Debug  bool              `json:"debug"`
	Domain string            `json:"domain"`
	Header map[string]string `json:"header"`
	Variables
	FuncSet        []FuncSet `json:"func_set"`
	WithInitVar    bool
	WithRunningVar bool
}

type FuncSet struct {
	Key         string            `json:"key"`
	Method      string            `json:"method"`
	Body        string            `json:"body"`
	Url         string            `json:"url"`
	Header      map[string]string `json:"header"`
	Probability int               `json:"probability"`
	Validator   string            `json:"validator"`
	Parsed      struct {
		Body   StrComponent
		Url    StrComponent
		Header SMapComponent
	}
	RScript *RunScript
}

type Component struct {
	OriWithInitVar    bool
	OriWithRunningVar bool
}

type StrComponent struct {
	Component
	ParsedValue string
}

type SMapComponent struct {
	Component
	ParsedValue map[string]string
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (rs *RunScript) genVariables() Variables {
	varsBytes, _ := jsoniter.Marshal(rs.Variables)
	getId := func(sid int) int {
		return rand.Intn(sid)
	}
	getSid := func() int64 {
		return time.Now().Unix()
	}
	templateFunc := map[string]interface{}{
		"getId":  getId,
		"getSid": getSid,
	}
	t := template.Must(template.New("Variables").Funcs(templateFunc).Parse(string(varsBytes)))
	var tmpBytes bytes.Buffer
	_ = t.Execute(&tmpBytes, nil)
	var variables Variables
	decoder := jsoniter.NewDecoder(strings.NewReader(tmpBytes.String()))
	decoder.UseNumber()
	err := decoder.Decode(&variables)
	if err != nil {
		log.Fatal()
	}
	merged := make(map[string]interface{})
	for k, v := range variables.InitVariables {
		merged[k] = v
	}
	for k, v := range variables.RunningVariables {
		merged[k] = v
	}

	variables.MergedVariables = merged

	return variables
}

func (rs *RunScript) init() {
	if nil != rs.RunningVariables && len(rs.RunningVariables) > 0 {
		rs.WithRunningVar = true
	}
	if nil != rs.InitVariables && len(rs.InitVariables) > 0 {
		rs.WithInitVar = true
	}
}

func (fs *FuncSet) parseVars(rs RunScript) {
	fs.RScript = &rs
	// no variables
	if !fs.RScript.WithInitVar && !fs.RScript.WithRunningVar {
		fs.Parsed.Url.ParsedValue = fs.Url
		fs.Parsed.Body.ParsedValue = fs.Body
		fs.Parsed.Header.ParsedValue = fs.Header
		return
	}

	parsedUrl := fs.getURL(rs.InitVariables)
	if strings.Contains(parsedUrl, NoValue) {
		fs.Parsed.Url.OriWithRunningVar = true
	}
	parsedUrl = fs.getURL(rs.RunningVariables)
	if strings.Contains(parsedUrl, NoValue) {
		fs.Parsed.Url.OriWithInitVar = true
	}

	parsedBody := fs.getBody(rs.InitVariables)
	if strings.Contains(parsedBody, NoValue) {
		fs.Parsed.Body.OriWithRunningVar = true
	}
	parsedBody = fs.getBody(rs.RunningVariables)
	if strings.Contains(parsedBody, NoValue) {
		fs.Parsed.Body.OriWithInitVar = true
	}
	parsedHeader := fs.getHeaders(rs.InitVariables)
	for _, v := range parsedHeader {
		if strings.Contains(v, NoValue) {
			fs.Parsed.Header.OriWithRunningVar = true
		}
	}
	parsedHeader = fs.getHeaders(rs.RunningVariables)
	for _, v := range parsedHeader {
		if strings.Contains(v, NoValue) {
			fs.Parsed.Header.OriWithInitVar = true
		}
	}

}

func (fs *FuncSet) getURL(v Variable) string {
	tmpl, err := template.New("URL").Parse(fs.Url)
	if err != nil {
		panic(err)
	}
	var tmplBytes bytes.Buffer
	err = tmpl.Execute(&tmplBytes, v)
	if err != nil {
		panic(err)
	}
	return tmplBytes.String()
}

func (fs *FuncSet) getBody(v Variable) string {
	tmpl, err := template.New("Body").Parse(fs.Body)
	if err != nil {
		panic(err)
	}
	var tmplBytes bytes.Buffer
	err = tmpl.Execute(&tmplBytes, v)
	if err != nil {
		panic(err)
	}
	return tmplBytes.String()
}

func (fs *FuncSet) getHeaders(v Variable) (hmap map[string]string) {
	headerBytes, err := jsoniter.Marshal(fs.Header)
	tmpl, err := template.New("Header").Parse(string(headerBytes))
	if err != nil {
		panic(err)
	}
	var tmplBytes bytes.Buffer
	err = tmpl.Execute(&tmplBytes, v)
	if err != nil {
		panic(err)
	}
	err = jsoniter.Unmarshal(tmplBytes.Bytes(), &hmap)
	if err != nil {
		panic(err)
	}
	fmt.Println(hmap)
	return hmap
}

func (fs *FuncSet) assertTrue(mapping map[string]interface{}) bool {
	t := template.Must(template.New("Validator").Parse(fs.Validator))
	var bs bytes.Buffer
	//for _, v := range mapping {
	//	fmt.Println(v, reflect.TypeOf(v))
	//}
	err := t.Execute(&bs, mapping)
	if err != nil {
		panic(err)
	}
	return "true" == bs.String()
}

func GetTaskList(baseJson string) []*boomer.Task {
	rs := RunScript{}
	err := jsoniter.Unmarshal([]byte(baseJson), &rs)
	if err != nil {
		panic(err)
	}
	rs.init()
	var tasks []*boomer.Task
	for _, req := range rs.FuncSet {
		req.parseVars(rs)
		action := genReqAction(req)
		task := boomer.Task{
			Name:   req.Key,
			Weight: req.Probability,
			Fn:     action,
		}
		tasks = append(tasks, &task)
	}
	return tasks
}
