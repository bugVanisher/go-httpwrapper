package httpwrapper

import (
	"bytes"
	"crypto/tls"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/myzhan/boomer"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"
)

var client *http.Client
var verbose = true

func init() {

	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 2000
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		MaxIdleConnsPerHost: 2000,
		DisableCompression:  false,
		DisableKeepAlives:   false,
	}
	client = &http.Client{
		Transport: tr,
		Timeout:   time.Duration(10) * time.Second,
	}
}

func genReqAction(fs FuncSet) func() {
	variables := fs.RScript.genVariables()
	initUrl := fs.getURL(variables.InitVariables)
	initBody := fs.getBody(variables.InitVariables)
	initHeaders := fs.getHeaders(variables.InitVariables)

	action := func() {
		var url string
		var body string
		var headers map[string]string
		runVariables := fs.RScript.genVariables()
		if !fs.RScript.WithInitVar && !fs.RScript.WithRunningVar {
			url = fs.Parsed.Url.ParsedValue
			body = fs.Parsed.Body.ParsedValue
		} else {
			if !fs.Parsed.Url.OriWithRunningVar {
				url = initUrl
			} else {
				url = fs.getURL(runVariables.MergedVariables)
			}

			if !fs.Parsed.Body.OriWithRunningVar {
				body = initBody
			} else {
				body = fs.getBody(runVariables.MergedVariables)
			}

			if !fs.Parsed.Header.OriWithRunningVar {
				headers = initHeaders
			} else {
				headers = fs.getHeaders(runVariables.MergedVariables)
			}

		}
		url = fmt.Sprintf("%s%s", fs.RScript.Domain, url)
		request, err := http.NewRequest(fs.Method, url, bytes.NewBuffer([]byte(body)))
		if err != nil {
			log.Fatalf("%v\n", err)
		}

		for k, v := range initHeaders {
			request.Header.Set(k, v)
		}

		for k, v := range headers {
			request.Header.Set(k, v)
		}

		if fs.RScript.Debug {
			log.Println(formatRequest(request))
		}

		startTime := time.Now()
		response, err := client.Do(request)
		elapsed := time.Since(startTime)

		if err != nil {
			if verbose {
				log.Printf("%v\n", err)
			}
			boomer.RecordFailure("http", "error", 0.0, err.Error())
		} else {

			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				log.Printf("%v\n", err)
			} else {
				var res map[string]interface{}
				_ = jsoniter.Unmarshal(body, &res)
				res["http_status_code"] = response.StatusCode
				merged := make(map[string]interface{})
				for k, v := range runVariables.MergedVariables {
					merged[k] = v
				}
				for k, v := range res {
					merged[k] = v
				}

				if fs.RScript.Debug {
					log.Printf("Status Code: %d\n", response.StatusCode)
					log.Println(string(body))

				} else {
					io.Copy(ioutil.Discard, response.Body)
				}

				if fs.assertTrue(merged) {
					fmt.Println("assert true", elapsed.Nanoseconds()/int64(time.Millisecond))
					boomer.RecordSuccess(fs.Key, strconv.Itoa(response.StatusCode),
						elapsed.Nanoseconds()/int64(time.Millisecond), response.ContentLength)
				} else {
					boomer.RecordFailure("assert failed", fs.Key, elapsed.Nanoseconds()/int64(time.Millisecond), "")
				}

			}

			response.Body.Close()
		}
	}
	return action
}

func formatRequest(r *http.Request) string {
	data, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Fatal("Error")
	}
	return string(data)
}
