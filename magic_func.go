package httpwrapper

import (
	"github.com/spf13/cast"
	"math/rand"
	"time"
)

var TemplateFunc = map[string]interface{}{
	"getId":     getId,
	"getSid":    getSid,
	"toFloat64": cast.ToFloat64,
}

func getSid() int64 {
	return time.Now().Unix()
}

func getId(id int) int {
	return rand.Intn(id)
}
