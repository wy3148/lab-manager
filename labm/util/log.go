package util

import (
	"github.com/astaxie/beego/logs"
)

var Log *logs.BeeLogger

func init() {
	Log = logs.NewLogger()
	Log.SetLogger("console")
	Log.SetLogger(logs.AdapterFile, `{"filename":"./labm.log","level":7,"maxlines":0,"maxsize":0,"daily":true,"maxdays":10}`)
}
