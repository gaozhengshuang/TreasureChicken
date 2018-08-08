package def
import (
	"net/http"
	"strings"
	"strconv"
	"crypto/md5"
	"encoding/json"
	"sort"
	"fmt"
	"time"
	"reflect"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/net"
	"gitee.com/jntse/gotoolkit/util"
	"gitee.com/jntse/minehero/server/tbl"
	"github.com/go-redis/redis"
)


func GetUserToken(redis *redis.Client, charid uint64) string {
	if redis == nil {
		log.Error("redis is nil")
		return ""
	}

	tokenkey := fmt.Sprintf("charid_%d_token", charid)
	token, err := redis.Get(tokenkey).Result()
	if err != nil {
		log.Error("redis get[%s] is err[%s]", tokenkey, err)
		return ""
	}

	return token
}

func HttpPostDataStatistics(charid uint64, mapset map[string]interface{}, sortVal, tvmid, url, secret, mode string) (errcode string, resp *network.HttpResponse) {
	if tbl.Global.IntranetFlag {
		log.Warn("本版本推送暂不可用")
		return "本版本推送暂不可用", nil
	}

	// 签名
	strtimestamp := strconv.FormatInt(mapset["timestamp"].(int64), 10)
	signtemp := mapset["key"].(string) + sortVal + mapset["nonce"].(string) + strtimestamp + secret 
	signbytes := []byte(signtemp)
	md5array := md5.Sum(signbytes)
	md5bytes := []byte(md5array[:])
	md5string := fmt.Sprintf("%x", md5bytes)
	mapset["sign"] = strings.ToLower(md5string)

	// 序列化
	postbody, jsonerr := json.Marshal(mapset)
	if jsonerr != nil {
		log.Error("[%s] 玩家[%d] json.Marshal err[%s]", mode, charid, jsonerr)
		return "json.Marshal Fail", nil
	}
	log.Trace("[%s] 玩家[%d] url[%s] postbody[%s]", mode, charid, url, postbody)

	// post
	resp, posterr := network.HttpPost(url, util.BytesToString(postbody))
	if posterr != nil {
		log.Error("[%s] 玩家[%d] 推送失败 error[%s] resp[%#v]", mode, charid, posterr, resp)
		return "HttpPost失败", nil
	}

	// response
	if resp.Code != http.StatusOK { 
		log.Info("[%s] 玩家[%d] 推送失败 errcode[%d] status[%s]", mode, charid, resp.Code, resp.Status)
		return "HttpPost ErrorCode", nil
	}
	log.Trace("[%s] 玩家[%d] 推送完成 [%s]", mode, charid, util.BytesToString(resp.Body))
	return "", resp
}


// 新建角色
func HttpRequestNewUserArglist(arglist []interface{}) []interface{} {
	if len(arglist) != 4 {
		return nil
	}
	tvmid, ok := arglist[0].(string)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[0]).String()); return nil }

	token, ok := arglist[1].(string)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[1]).String()); return nil }

	username, ok := arglist[2].(string)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[2]).String()); return nil }

	charid, ok := arglist[3].(uint64)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[3]).String()); return nil }

	HttpRequestNewUser(tvmid,token,username,charid)
	return nil
}

func HttpRequestNewUser(tvmid, token, username string, charid uint64) {

	//url := "https://open.yx.tvyouxuan.com/public/finance/MultiRecharge"
	//url := "http://open.std.tvmopt.com/public/finance/MultiRecharge"
	url := tbl.Global.HongBaoAPI.CharacterCreation
	secret := tbl.Global.HongBaoAPI.Secret
	key := tbl.Global.HongBaoAPI.Key
	mapset := make(map[string]interface{})
	mapset["token"] = token
	mapset["tvmid"] = tvmid
	mapset["timestamp"] = util.CURTIME()
	mapset["key"] = key
	mapset["nonce"] = strconv.FormatInt(util.CURTIMEUS(), 10)
	mapset["game_role_id"] = strconv.FormatInt(int64(charid), 10)
	mapset["role_name"] = username
	mapset["role_create_time"] = time.Now().Format("2006-01-02")


	// 除公共参数以外的内容，按键从小到大排序，将值拼接
	sortKeys := make(sort.StringSlice,0)
	sortKeys = append(sortKeys, "tvmid")
	sortKeys = append(sortKeys, "token")
	sortKeys = append(sortKeys, "game_role_id")
	sortKeys = append(sortKeys, "role_name")
	sortKeys = append(sortKeys, "role_create_time")
	sortKeys.Sort()
	sortVal := ""
	for _, v := range sortKeys { 
		if str, ok := mapset[v].(string); ok == true {
			sortVal += str
		}else if num, ok := mapset[v].(int64); ok == true {
			sortVal += strconv.FormatInt(num, 10)
		}else {
			log.Error("签名拼接参数[%s]的类型不是int64或者string", v )
			return
		}

	}

	HttpPostDataStatistics(charid, mapset, sortVal, tvmid, url, secret, "推送新用户")
}


// 推送在线时长，在线时间，单位分钟
func HttpRequestUserOnlineTimeArglist(arglist []interface{}) []interface{} {
	if len(arglist) != 4 {
		return nil
	}
	tvmid, ok := arglist[0].(string)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[0]).String()); return nil }

	token, ok := arglist[1].(string)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[1]).String()); return nil }

	charid, ok := arglist[2].(uint64)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[2]).String()); return nil }

	tm_online, ok := arglist[3].(int64)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[3]).String()); return nil }

	HttpRequestUserOnlineTime(tvmid, token, charid, tm_online)
	return nil
}
func HttpRequestUserOnlineTime(tvmid, token string, charid uint64, tm_online int64) {
	//token := GetUserToken(redis, charid)
	if token == "" {
		log.Error("玩家[%d] token is empty", charid)
		return
	}

	//url := "https://open.yx.tvyouxuan.com/public/finance/MultiRecharge"
	//url := "http://open.std.tvmopt.com/public/finance/MultiRecharge"
	url := tbl.Global.HongBaoAPI.Online
	secret := tbl.Global.HongBaoAPI.Secret
	key := tbl.Global.HongBaoAPI.Key
	mapset := make(map[string]interface{})
	mapset["token"] = token
	mapset["tvmid"] = tvmid
	mapset["timestamp"] = util.CURTIME()
	mapset["key"] = key
	mapset["nonce"] = strconv.FormatInt(util.CURTIMEUS(), 10)
	mapset["online_time"] = tm_online

	// 除公共参数以外的内容，按键从小到大排序，将值拼接
	sortKeys := make(sort.StringSlice,0)
	sortKeys = append(sortKeys, "tvmid")
	sortKeys = append(sortKeys, "token")
	sortKeys = append(sortKeys, "online_time")
	sortKeys.Sort()
	sortVal := ""
	for _, v := range sortKeys { 
		if str, ok := mapset[v].(string); ok == true {
			sortVal += str
		}else if num, ok := mapset[v].(int64); ok == true {
			sortVal += strconv.FormatInt(num, 10)
		}else {
			log.Error("签名拼接参数[%s]的类型不是int64或者string", v )
			return
		}

	}

	HttpPostDataStatistics(charid, mapset, sortVal, tvmid, url, secret, "推送在线时长")
}

// 推送战斗次数(开房间次数)
func HttpRequestUserBattleCountArglist(arglist []interface{}) []interface{} {
	if len(arglist) != 5 {
		return nil
	}
	tvmid, ok := arglist[0].(string)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[0]).String()); return nil }

	token, ok := arglist[1].(string)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[1]).String()); return nil }

	charid, ok := arglist[2].(uint64)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[2]).String()); return nil }

	playtype, ok := arglist[3].(string)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[3]).String()); return nil }

	palycount, ok := arglist[4].(int32)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[4]).String()); return nil }

	HttpRequestUserBattleCount(tvmid, token, charid, playtype, palycount)
	return nil
}
func HttpRequestUserBattleCount(tvmid, token string, charid uint64, playtype string, palycount int32) {

	url := tbl.Global.HongBaoAPI.Battles
	secret := tbl.Global.HongBaoAPI.Secret
	key := tbl.Global.HongBaoAPI.Key
	mapset := make(map[string]interface{})
	mapset["token"] = token
	mapset["tvmid"] = tvmid
	mapset["timestamp"] = util.CURTIME()
	mapset["key"] = key
	mapset["nonce"] = strconv.FormatInt(util.CURTIMEUS(), 10)
	mapset["play_type"] = playtype
	mapset["play_count"] = int64(palycount)

	// 除公共参数以外的内容，按键从小到大排序，将值拼接
	sortKeys := make(sort.StringSlice,0)
	sortKeys = append(sortKeys, "tvmid")
	sortKeys = append(sortKeys, "token")
	sortKeys = append(sortKeys, "play_type")
	sortKeys = append(sortKeys, "play_count")
	sortKeys.Sort()
	sortVal := ""
	for _, v := range sortKeys { 
		if str, ok := mapset[v].(string); ok == true {
			sortVal += str
		}else if num, ok := mapset[v].(int64); ok == true {
			sortVal += strconv.FormatInt(num, 10)
		}else {
			log.Error("签名拼接参数[%s]的类型不是int64或者string", v )
			return
		}

	}

	HttpPostDataStatistics(charid, mapset, sortVal, tvmid, url, secret, "推送战斗次数")
}


// 推送角色等级
func HttpRequestUserLevelArglist(arglist []interface{}) []interface{} {
	if len(arglist) != 4 {
		return nil
	}
	tvmid, ok := arglist[0].(string)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[0]).String()); return nil }

	token, ok := arglist[1].(string)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[1]).String()); return nil }

	charid, ok := arglist[2].(uint64)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[2]).String()); return nil }

	level, ok := arglist[3].(uint32)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[3]).String()); return nil }

	HttpRequestUserLevel(tvmid, token, charid, level)
	return nil
}

func HttpRequestUserLevel(tvmid, token string, charid uint64, level uint32) {

	//url := "https://open.yx.tvyouxuan.com/public/finance/MultiRecharge"
	//url := "http://open.std.tvmopt.com/public/finance/MultiRecharge"
	url := tbl.Global.HongBaoAPI.CharacterLevel
	secret := tbl.Global.HongBaoAPI.Secret
	key := tbl.Global.HongBaoAPI.Key
	mapset := make(map[string]interface{})
	mapset["token"] = token
	mapset["tvmid"] = tvmid
	mapset["timestamp"] = util.CURTIME()
	mapset["key"] = key
	mapset["nonce"] = strconv.FormatInt(util.CURTIMEUS(), 10)
	mapset["game_role_id"] = strconv.FormatInt(int64(charid), 10)
	mapset["level"] = int64(level)


	// 除公共参数以外的内容，按键从小到大排序，将值拼接
	sortKeys := make(sort.StringSlice,0)
	sortKeys = append(sortKeys, "tvmid")
	sortKeys = append(sortKeys, "token")
	sortKeys = append(sortKeys, "game_role_id")
	sortKeys = append(sortKeys, "level")
	sortKeys.Sort()
	sortVal := ""
	for _, v := range sortKeys { 
		if str, ok := mapset[v].(string); ok == true {
			sortVal += str
		}else if num, ok := mapset[v].(int64); ok == true {
			sortVal += strconv.FormatInt(num, 10)
		}else {
			log.Error("签名拼接参数[%s]的类型不是int64或者string", v )
			return
		}
	}

	HttpPostDataStatistics(charid, mapset, sortVal, tvmid, url, secret, "推送角色等级")
}


// 推送资源消耗，price 等值人民币金额，单位分
func HttpRequestUserResourceConsumeArglist(arglist []interface{}) []interface{} {
	if len(arglist) != 4 {
		return nil
	}
	tvmid, ok := arglist[0].(string)
	if ok == false { 
		log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[0]).String());
		return nil 
	}

	token, ok := arglist[1].(string)
	if ok == false { 
		log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[1]).String())
		return nil 
	}

	charid, ok := arglist[2].(uint64)
	if ok == false { 
		log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[2]).String())
		return nil 
	}

	price, ok := arglist[3].(uint32)
	if ok == false { 
		log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[3]).String())
		return nil 
	}

	HttpRequestUserResourceConsume(tvmid, token, charid, price)
	return nil
}

func HttpRequestUserResourceConsume(tvmid, token string, charid uint64, price uint32) {

	if price <= 0 {
		return
	}

	//url := "https://open.yx.tvyouxuan.com/public/finance/MultiRecharge"
	//url := "http://open.std.tvmopt.com/public/finance/MultiRecharge"
	url := tbl.Global.HongBaoAPI.ConsumeMoney
	secret := tbl.Global.HongBaoAPI.Secret
	key := tbl.Global.HongBaoAPI.Key
	mapset := make(map[string]interface{})
	mapset["token"] = token
	mapset["tvmid"] = tvmid
	mapset["timestamp"] = util.CURTIME()
	mapset["key"] = key
	mapset["nonce"] = strconv.FormatInt(util.CURTIMEUS(), 10)
	mapset["consume"] = int64(price)

	// 除公共参数以外的内容，按键从小到大排序，将值拼接
	sortKeys := make(sort.StringSlice,0)
	sortKeys = append(sortKeys, "tvmid")
	sortKeys = append(sortKeys, "token")
	sortKeys = append(sortKeys, "consume")
	sortKeys.Sort()
	sortVal := ""
	for _, v := range sortKeys { 
		if str, ok := mapset[v].(string); ok == true {
			sortVal += str
		}else if num, ok := mapset[v].(int64); ok == true {
			sortVal += strconv.FormatInt(num, 10)
		}else {
			log.Error("签名拼接参数[%s]的类型不是int64或者string", v )
			return
		}
	}

	HttpPostDataStatistics(charid, mapset, sortVal, tvmid, url, secret, "推送资源消耗")
}


// 推送资源获取，price 等值人民币金额，单位分
func HttpRequestUserResourceEarnArglist(arglist []interface{}) []interface{} {
	if len(arglist) != 4 {
		return nil
	}
	tvmid, ok := arglist[0].(string)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[0]).String()); return nil }

	token, ok := arglist[1].(string)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[1]).String()); return nil }

	charid, ok := arglist[2].(uint64)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[2]).String()); return nil }

	price, ok := arglist[3].(uint32)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[3]).String()); return nil }

	HttpRequestUserResourceEarn(tvmid, token, charid, price)
	return nil
}

func HttpRequestUserResourceEarn(tvmid, token string, charid uint64, price uint32) {

	if price <= 0 {
		return
	}
	//url := "https://open.yx.tvyouxuan.com/public/finance/MultiRecharge"
	//url := "http://open.std.tvmopt.com/public/finance/MultiRecharge"
	url := tbl.Global.HongBaoAPI.LootMoney
	secret := tbl.Global.HongBaoAPI.Secret
	key := tbl.Global.HongBaoAPI.Key
	mapset := make(map[string]interface{})
	mapset["token"] = token
	mapset["tvmid"] = tvmid
	mapset["timestamp"] = util.CURTIME()
	mapset["key"] = key
	mapset["nonce"] = strconv.FormatInt(util.CURTIMEUS(), 10)
	mapset["earm"] = int64(price)

	// 除公共参数以外的内容，按键从小到大排序，将值拼接
	sortKeys := make(sort.StringSlice,0)
	sortKeys = append(sortKeys, "tvmid")
	sortKeys = append(sortKeys, "token")
	sortKeys = append(sortKeys, "earm")
	sortKeys.Sort()
	sortVal := ""
	for _, v := range sortKeys { 
		if str, ok := mapset[v].(string); ok == true {
			sortVal += str
		}else if num, ok := mapset[v].(int64); ok == true {
			sortVal += strconv.FormatInt(num, 10)
		}else {
			log.Error("签名拼接参数[%s]的类型不是int64或者string", v )
			return
		}
	}

	HttpPostDataStatistics(charid, mapset, sortVal, tvmid, url, secret, "推送资源获取")
}


// 胜利次数,拿大奖次数
func HttpRequestUserVictoryArglist(arglist []interface{}) []interface{} {
	if len(arglist) != 5 {
		return nil
	}
	tvmid, ok := arglist[0].(string)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[0]).String()); return nil }

	token, ok := arglist[1].(string)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[1]).String()); return nil }

	charid, ok := arglist[2].(uint64)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[2]).String()); return nil }

	victype, ok := arglist[3].(string)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[3]).String()); return nil }

	count, ok := arglist[4].(int32)
	if ok == false { log.Fatal("类型转换错误 argu 真实类型是：%s", reflect.TypeOf(arglist[4]).String()); return nil }

	HttpRequestUserVictory(tvmid, token, charid, victype, count)
	return nil
}

func HttpRequestUserVictory(tvmid, token string, charid uint64, victype string, count int32) {

	//url := "https://open.yx.tvyouxuan.com/public/finance/MultiRecharge"
	//url := "http://open.std.tvmopt.com/public/finance/MultiRecharge"
	url := tbl.Global.HongBaoAPI.Victory
	secret := tbl.Global.HongBaoAPI.Secret
	key := tbl.Global.HongBaoAPI.Key
	mapset := make(map[string]interface{})
	mapset["token"] = token
	mapset["tvmid"] = tvmid
	mapset["timestamp"] = util.CURTIME()
	mapset["key"] = key
	mapset["nonce"] = strconv.FormatInt(util.CURTIMEUS(), 10)
	mapset["victory_type"] = victype 
	mapset["victory_count"] = int64(count)

	// 除公共参数以外的内容，按键从小到大排序，将值拼接
	sortKeys := make(sort.StringSlice,0)
	sortKeys = append(sortKeys, "tvmid")
	sortKeys = append(sortKeys, "token")
	sortKeys = append(sortKeys, "victory_type")
	sortKeys = append(sortKeys, "victory_count")
	sortKeys.Sort()
	sortVal := ""
	for _, v := range sortKeys {
		if str, ok := mapset[v].(string); ok == true {
			sortVal += str
		}else if num, ok := mapset[v].(int64); ok == true {
			sortVal += strconv.FormatInt(num, 10)
		}else {
			log.Error("签名拼接参数[%s]的类型不是int64或者string", v )
			return
		}
	}

	HttpPostDataStatistics(charid, mapset, sortVal, tvmid, url, secret, "推送胜利次数")
}

// 获得平台金币和钻石
func HttpRequestFinanceQuery(charid uint64, token, tvmid string) (errocde string, coin, diamond int32) {

	//url := "https://open.yx.tvyouxuan.com/public/finance/Query"
	//url := "http://open.std.tvmopt.com/public/finance/Query"
	url := tbl.Global.HongBaoAPI.FinanceQuery
	secret := tbl.Global.HongBaoAPI.Secret
	key := tbl.Global.HongBaoAPI.Key
	mapset := make(map[string]interface{})
	mapset["token"] = token
	mapset["tvmid"] = tvmid
	mapset["timestamp"] = util.CURTIME()
	mapset["key"] = key
	mapset["nonce"] = strconv.FormatInt(util.CURTIMEUS(), 10)

	// 除公共参数以外的内容，按键从小到大排序，将值拼接
	sortKeys := make(sort.StringSlice,0)
	sortKeys = append(sortKeys, "tvmid")
	sortKeys = append(sortKeys, "token")
	sortKeys.Sort()
	sortVal := ""
	for _, v := range sortKeys {
		if str, ok := mapset[v].(string); ok == true {
			sortVal += str
		}else if num, ok := mapset[v].(int64); ok == true {
			sortVal += strconv.FormatInt(num, 10)
		}else {
			log.Error("签名拼接参数[%s]的类型不是int64或者string", v )
			return "签名拼接参数错误", 0, 0
		}
	}

	errcode, resp := HttpPostDataStatistics(charid, mapset, sortVal, tvmid, url, secret, "查询平台货币")
	if errcode != "" || resp == nil {
		return errcode, 0, 0
	}

	// check status
	type stFinance struct {
		Coins int32
		Diamonds int32
	}
	type stFinanceResp struct {
		Status 	int32
		Data	stFinance
		Msg		string
	}
	
	var fResp stFinanceResp
	unerr := json.Unmarshal(resp.Body, &fResp)
	if unerr != nil {
		log.Info("玩家[%d] 查询平台金币 json.Unmarshal 'status' Fail[%s] ", charid, unerr)
		return "json.Unmarshal resp.Body 失败", 0, 0
	}

	if fResp.Status != 0 {
		log.Error("玩家[%d] 查询平台金币失败 status[%d] msg[%s]", charid, fResp.Status, fResp.Msg)
		return fResp.Msg, 0, 0
	}

	return "", fResp.Data.Coins, fResp.Data.Diamonds

	//var respinfo map[string]interface{}
	//unerr := json.Unmarshal(resp.Body, &respinfo)
	//if unerr != nil {
	//	log.Info("玩家[%d] 扣平台金币 json.Unmarshal 'status' Fail[%s] ", charid, unerr)
	//	return 0, 0
	//}

	//var statuscode int32 = int32(respinfo["status"].(float64))
	//if statuscode != 0 {
	//	log.Error("玩家[%d] 扣平台金币 status[%d] msg[%s]", charid, statuscode, respinfo["msg"])
	//	return 0, 0
	//}

	//moneypair, convertok := respinfo["data"].(map[string]interface{})
	//if convertok == false {
	//	panic(convertok)
	//}

	//return int32(moneypair["coins"].(float64)), int32(moneypair["diamonds"].(float64))
}

// 扣除平台金币
func HttpRequestDecrCoins(charid uint64, token, tvmid string, count int32, desc string) bool {

	//url := "https://open.yx.tvyouxuan.com/public/finance/DecrCoins"
	//url := "http://open.std.tvmopt.com/public/finance/DecrCoins"
	url := tbl.Global.HongBaoAPI.DecrCoins
	secret := tbl.Global.HongBaoAPI.Secret
	key := tbl.Global.HongBaoAPI.Key
	order_id := strconv.FormatInt(int64(charid),10) + "_decrcoins_" + strconv.FormatInt(util.CURTIMEUS(), 10)
	mapset := make(map[string]interface{})
	mapset["token"] = token
	mapset["tvmid"] = tvmid
	mapset["timestamp"] = util.CURTIME()
	mapset["key"] = key
	mapset["nonce"] = strconv.FormatInt(util.CURTIMEUS(), 10)
	mapset["order_id"] = order_id
	mapset["amount"] = int64(count)
	mapset["description"] = desc

	// 除公共参数以外的内容，按键从小到大排序，将值拼接
	sortKeys := make(sort.StringSlice,0)
	sortKeys = append(sortKeys, "tvmid")
	sortKeys = append(sortKeys, "token")
	sortKeys = append(sortKeys, "order_id")
	sortKeys = append(sortKeys, "amount")
	sortKeys = append(sortKeys, "description")
	sortKeys.Sort()
	sortVal := ""
	for _, v := range sortKeys {
		if str, ok := mapset[v].(string); ok == true {
			sortVal += str
		}else if num, ok := mapset[v].(int64); ok == true {
			sortVal += strconv.FormatInt(num, 10)
		}else {
			log.Error("签名拼接参数[%s]的类型不是int64或者string", v )
			return false
		}
	}

	errcode, resp := HttpPostDataStatistics(charid, mapset, sortVal, tvmid, url, secret, "扣平台金币")
	if errcode != "" || resp == nil {
		return false
	}

	// check status
	var respinfo map[string]interface{}
	unerr := json.Unmarshal(resp.Body, &respinfo)
	if unerr != nil {
		log.Info("玩家[%d] 扣平台金币 json.Unmarshal 'status' Fail[%s] ", charid, unerr)
		return false
	}

	// statuscode 非0 不一定是余额不足
	var statuscode int32 = int32(respinfo["status"].(float64))
	if statuscode != 0 {
		log.Error("玩家[%d] 扣平台金币 status[%d] msg[%s]", charid, statuscode, respinfo["msg"])
		return false
	}

	return true
}

// 增加钻石，需要绑定微信号
func HttpRequestIncrDiamonds(charid uint64, token, tvmid string, count int32, desc string) bool {

	//url := "https://open.yx.tvyouxuan.com/public/finance/IncrDiamonds"
	//url := "http://open.std.tvmopt.com/public/finance/IncrDiamonds"
	url := tbl.Global.HongBaoAPI.IncrDiamonds
	secret := tbl.Global.HongBaoAPI.Secret
	key := tbl.Global.HongBaoAPI.Key
	order_id := strconv.FormatInt(int64(charid),10) + "_incdian_" + strconv.FormatInt(util.CURTIME(), 10)
	mapset := make(map[string]interface{})
	mapset["token"] = token
	mapset["tvmid"] = tvmid
	mapset["timestamp"] = util.CURTIME()
	mapset["key"] = key
	mapset["nonce"] = strconv.FormatInt(util.CURTIMEUS(), 10)
	mapset["order_id"] = order_id
	mapset["amount"] = int64(count)
	mapset["description"] = desc

	// 除公共参数以外的内容，按键从小到大排序，将值拼接
	sortKeys := make(sort.StringSlice,0)
	sortKeys = append(sortKeys, "tvmid")
	sortKeys = append(sortKeys, "token")
	sortKeys = append(sortKeys, "order_id")
	sortKeys = append(sortKeys, "amount")
	sortKeys = append(sortKeys, "description")
	sortKeys.Sort()
	sortVal := ""
	for _, v := range sortKeys {
		if str, ok := mapset[v].(string); ok == true {
			sortVal += str
		}else if num, ok := mapset[v].(int64); ok == true {
			sortVal += strconv.FormatInt(num, 10)
		}else {
			log.Error("签名拼接参数[%s]的类型不是int64或者string", v )
			return false
		}
	}

	errcode, resp := HttpPostDataStatistics(charid, mapset, sortVal, tvmid, url, secret, "增加平台钻石")
	if errcode != "" || resp == nil {
		return false
	}

	// check status
	var respinfo map[string]interface{}
	unerr := json.Unmarshal(resp.Body, &respinfo)
	if unerr != nil {
		log.Info("玩家[%d] 增加平台钻石 json.Unmarshal 'status' Fail[%s] ", charid, unerr)
		return false
	}

	var statuscode int32 = int32(respinfo["status"].(float64))
	if statuscode != 0 {
		log.Error("玩家[%d] 增加平台钻石 status[%d] msg[%s]", charid, statuscode, respinfo["msg"])
		return false
	}

	return true
}

// 增加金币，需要绑定微信号
func HttpRequestIncrCoins(charid uint64, token, tvmid string, count int32, desc string) bool {

	//url := "https://open.yx.tvyouxuan.com/public/finance/IncrDiamonds"
	//url := "http://open.std.tvmopt.com/public/finance/IncrDiamonds"
	url := tbl.Global.HongBaoAPI.IncrGolds
	secret := tbl.Global.HongBaoAPI.Secret
	key := tbl.Global.HongBaoAPI.Key
	order_id := strconv.FormatInt(int64(charid),10) + "_incrcoin_" + strconv.FormatInt(util.CURTIMEMS(), 10)
	mapset := make(map[string]interface{})
	mapset["token"] = token
	mapset["tvmid"] = tvmid
	mapset["timestamp"] = util.CURTIME()
	mapset["key"] = key
	mapset["nonce"] = strconv.FormatInt(util.CURTIMEUS(), 10)
	mapset["order_id"] = order_id
	mapset["amount"] = int64(count)
	mapset["description"] = desc

	// 除公共参数以外的内容，按键从小到大排序，将值拼接
	sortKeys := make(sort.StringSlice,0)
	sortKeys = append(sortKeys, "tvmid")
	sortKeys = append(sortKeys, "token")
	sortKeys = append(sortKeys, "order_id")
	sortKeys = append(sortKeys, "amount")
	sortKeys = append(sortKeys, "description")
	sortKeys.Sort()
	sortVal := ""
	for _, v := range sortKeys {
		if str, ok := mapset[v].(string); ok == true {
			sortVal += str
		}else if num, ok := mapset[v].(int64); ok == true {
			sortVal += strconv.FormatInt(num, 10)
		}else {
			log.Error("签名拼接参数[%s]的类型不是int64或者string", v )
			return false
		}
	}

	errcode, resp := HttpPostDataStatistics(charid, mapset, sortVal, tvmid, url, secret, "增加平台金币")
	if errcode != "" || resp == nil {
		return false
	}

	// check status
	var respinfo map[string]interface{}
	unerr := json.Unmarshal(resp.Body, &respinfo)
	if unerr != nil {
		log.Info("玩家[%d] 增加平台金币 json.Unmarshal 'status' Fail[%s] ", charid, unerr)
		return false
	}

	var statuscode int32 = int32(respinfo["status"].(float64))
	if statuscode != 0 {
		log.Error("玩家[%d] 增加平台金币 status[%d] msg[%s]", charid, statuscode, respinfo["msg"])
		return false
	}

	return true
}

// 检查是否绑定微信
func HttpRequestCheckWechatBound(charid uint64, token, tvmid string) {

	//url := "https://open.yx.tvyouxuan.com/public/user/CheckWechatBound"
	//url := "http://open.std.tvmopt.com/public/user/CheckWechatBound"
	url := tbl.Global.HongBaoAPI.CheckWechatBound
	secret := tbl.Global.HongBaoAPI.Secret
	key := tbl.Global.HongBaoAPI.Key
	mapset := make(map[string]interface{})
	mapset["token"] = token
	mapset["phone"] = tvmid
	mapset["timestamp"] = util.CURTIME()
	mapset["key"] = key
	mapset["nonce"] = strconv.FormatInt(util.CURTIMEUS(), 10)

	// 除公共参数以外的内容，按键从小到大排序，将值拼接
	sortKeys := make(sort.StringSlice,0)
	sortKeys = append(sortKeys, "token")
	sortKeys = append(sortKeys, "phone")
	sortKeys.Sort()
	sortVal := ""
	for _, v := range sortKeys {
		if str, ok := mapset[v].(string); ok == true {
			sortVal += str
		}else if num, ok := mapset[v].(int64); ok == true {
			sortVal += strconv.FormatInt(num, 10)
		}else {
			log.Error("签名拼接参数[%s]的类型不是int64或者string", v )
			return
		}
	}

	HttpPostDataStatistics(charid, mapset, sortVal, tvmid, url, secret, "检查绑定微信")
}

