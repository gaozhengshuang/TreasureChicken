package main
import (
	"net/http"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/util"
	"encoding/json"
)

// TODO：回调函数有http底层单独的协程调用，想要单线程处理可以将数据push到主线程chan中
func HttpServerResponseCallBack(w http.ResponseWriter, urlpath string, rawquery string, body []byte) {
	log.Info("HttpServerResponseCallBack[rid=%d]", util.GetRoutineID())

	//
	log.Info("urlpath: %s", urlpath)
	log.Info("rawquery: %s", rawquery)
	log.Info("body: %s", string(body))


	switch {
	default:
		cmdmap := make(map[string]interface{})
		unerr := json.Unmarshal(body, &cmdmap)
		if unerr != nil {
			log.Error("HttpServerResponseCallBack json.Unmarshal err[%s]", unerr)
			break
		}

		// GM指令
		if _ , ok := cmdmap["gmcmd"]; ok {
			gmcommands:= make(map[string]string)
			for k ,v := range cmdmap { gmcommands[k] = v.(string) }
			DoGMCmd(gmcommands)
			break
		}
	}


	// header 属性设置
	head := w.Header()
	head.Set("Content-Type", "text/plain; charset=utf-8")       // default
	//head.Set("Content-Type", "application/json")
	//head.Set("Content-Type", "application/x-www-form-urlencoded")

	// ret code
	w.WriteHeader(http.StatusOK)
	//w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("hello golang http test"))
}


// GM指令处理
func DoGMCmd(cmd map[string]string) {
	value, ok := cmd["gmcmd"]
	if !ok {
		log.Error("找不到gmcmd字段")
		return
	}

	switch value {
	case "reload":
		DoReload(cmd)
		break
	}
}

func DoReload(cmd map[string]string) {
	Login().Reload()
}
