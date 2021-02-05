package httpwrapper

import (
	"github.com/spf13/cast"
	"math/rand"
	"time"
)

var TemplateFunc = map[string]interface{}{
	"getRandomId": getRandomId,
	"getSid":      getSid,
	"toFloat64":   cast.ToFloat64,
	"toString":    cast.ToString,
}

func getSid() int64 {
	return time.Now().Unix()
}

func getRandomId(id int) int {
	return rand.Intn(id)
}
