package main
import (
	"net/http"
	"strings"
	"strconv"
	"crypto/md5"
	"encoding/json"
	"sort"
	"fmt"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/net"
	"gitee.com/jntse/gotoolkit/util"
	"gitee.com/jntse/minehero/pbmsg"
	"gitee.com/jntse/minehero/server/tbl"
	pb "github.com/golang/protobuf/proto"
)

func HttpServerResponseCallBack(w http.ResponseWriter, urlpath string, rawquery string, body []byte) {
	log.Info("HttpServerResponseCallBack")

	//
	log.Info("urlpath: %s", urlpath)
	log.Info("rawquery: %s", rawquery)
	log.Info("body: %s", string(body))


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


// 获取玩家收货地址
func RequestUserHomeAddress(tvmid string, token string) ( receiver, phone, address string) {
	//url := "https://open.yx.tvyouxuan.com/public/user/GetDeliveryAddresses"
	//url := "http://open.std.tvmopt.com/public/user/GetDeliveryAddresses"
	url := tbl.Global.HongBaoAPI.Getaddress
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
	for _, v := range sortKeys { sortVal += mapset[v].(string)}

	
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
		log.Error("获取玩家账户[%s] 的收货地址失败 json.Marshal err[%s]", tvmid, jsonerr)
		return
	}
	log.Info("获取玩家账户[%s] 的收货地址 sign[%s]", tvmid, md5string)
	log.Info("获取玩家账户[%s] 的收货地址 postbody[%s]", tvmid, postbody)


	// post
	resp, posterr := network.HttpPost(url, util.BytesToString(postbody))
	if posterr != nil {
		log.Error("获取玩家账户[%s] 的收货地址失败 HttpPost err[%s] resp[%#v]", tvmid, posterr, resp)
		return
	}


	// response
	if resp.Code != http.StatusOK { 
		log.Info("获取玩家账户[%s] 的收货地址失败 errcode[%d] status[%s]", tvmid, resp.Code, resp.Status)
		return
	}
	log.Info("获取玩家账户[%s] 的收货地址返回[%s]", tvmid, util.BytesToString(resp.Body))


	// check status
	var respinfo map[string]interface{}
	unerr := json.Unmarshal(resp.Body, &respinfo)
	if unerr != nil {
		log.Info("获取玩家账户[%s] 的收货地址失败 json.Unmarshal 'status' Fail[%s] ", tvmid, unerr)
		return
	}

	var statuscode int32 = int32(respinfo["status"].(float64))
	if statuscode != 0 {
		log.Error("获取玩家账户[%s] 的收货地址失败 status[%d] msg[%s]", tvmid, statuscode, respinfo["msg"])
		return
	}

	//log.Info("获取玩家账户[%s] 的收货地址成功 [%#v]", tvmid, respinfo["data"])
	addrlist, convertok := respinfo["data"].([]interface{})
	if convertok == false {
		log.Error("获取玩家账户[%s] 的收货地址异常，respinfo[data]非期望类型")
		return
	}

	if len(addrlist) <= 0 {
		log.Error("获取玩家账户[%s] 的收货地址是空的 ", tvmid)
		return
	}

	addrinfo, convertok := addrlist[0].(map[string]interface{})
	if convertok == false {
		log.Error("获取玩家账户[%s] 的收货地址异常，respinfo[data]非期望类型")
		return
	}

	if addrinfo["receiver"] != nil && addrinfo["phone"] != nil && addrinfo["province"] != nil &&
	   addrinfo["city"] != nil && addrinfo["area"] != nil && addrinfo["street"] != nil {

		address = fmt.Sprintf("%s%s%s%s", addrinfo["province"].(string), addrinfo["city"].(string), addrinfo["area"].(string), addrinfo["street"].(string))
		receiver = addrinfo["receiver"].(string)
		phone = addrinfo["phone"].(string)
		log.Info("玩家账户[%s] 收件人[%s] 电话[%s] 收货地址[%s]", tvmid, receiver , phone, address)
	}

	return

	//type stUserHomeAddress struct {
	//	receiver 	string 	//收件人
	//	phone		string	//联系电话
	//	province	string	//省或直辖市
	//	city		string	//市
	//	area		string	//区
	//	town		string	//镇，可能没有
	//	street		string	//街道地址
	//	zipcode		string	//邮编，可能没有
	//}
	//addr := &stUserHomeAddress{}
	//addrerr := json.Unmarshal(addrinfo[""], addr)
	//if addrerr != nil {
	//	log.Error("获取玩家账户[%s] 的收货地址失败 json.Unmarshal err[%s]", tvmid, addrerr)
	//	return
	//}
	//log.Info("获取玩家账户[%s] 的收货地址成功[%v]", tvmid, addr)
}

// 请求充值
func RequestRecharge(tvmid string, token string, user *GateUser, amount uint32) {
	if user == nil {
		return
	}
	//url := "https://open.yx.tvyouxuan.com/public/finance/MultiRecharge"
	//url := "http://open.std.tvmopt.com/public/finance/MultiRecharge"
	url := tbl.Global.HongBaoAPI.Recharge
	secret := tbl.Global.HongBaoAPI.Secret
	key := tbl.Global.HongBaoAPI.Key
	order_id := strconv.FormatInt(int64(user.Id()),10) + "_recharge_" + strconv.FormatInt(util.CURTIMEUS(), 10)
	mapset := make(map[string]interface{})
	mapset["token"] = token
	mapset["tvmid"] = tvmid
	mapset["timestamp"] = util.CURTIME()
	mapset["key"] = key
	mapset["nonce"] = strconv.FormatInt(util.CURTIMEUS(), 10)
	mapset["amount"] = int64(amount)	// 单位分
	mapset["order_id"] = order_id
	//mapset["callback"] = "http://210.73.214.67:19000"
	mapset["callback"] = tbl.Global.RechargeCallback
	mapset["redirect"] = "http://jump.giantfun.cn/rs.html"
	mapset["description"] = "充值"


	// 除公共参数以外的内容，按键从小到大排序，将值拼接
	sortKeys := make(sort.StringSlice,0)
	sortKeys = append(sortKeys, "tvmid")
	sortKeys = append(sortKeys, "token")
	sortKeys = append(sortKeys, "amount")
	sortKeys = append(sortKeys, "order_id")
	sortKeys = append(sortKeys, "callback")
	sortKeys = append(sortKeys, "redirect")
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
			return
		}
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
		log.Error("[充值] 玩家账户[%s] 充值失败 json.Marshal err[%s]", tvmid, jsonerr)
		return
	}
	//log.Info("[充值] 玩家账户[%s] 获取充值页面地址 sign[%s]", tvmid, md5string)
	log.Info("[充值] 玩家账户[%s] 获取充值页面地址 postbody[%s]", tvmid, postbody)


	// post
	resp, posterr := network.HttpPost(url, util.BytesToString(postbody))
	if posterr != nil {
		log.Error("[充值] 玩家账户[%s] 获取充值页面地址失败 HttpPost err[%s] resp[%#v]", tvmid, posterr, resp)
		return
	}

	// response
	if resp.Code != http.StatusOK { 
		log.Info("[充值] 玩家账户[%s] 获取充值页面地址失败 errcode[%d] status[%s]", tvmid, resp.Code, resp.Status)
		return
	}
	log.Info("[充值] 玩家账户[%s] 获取充值页面地址返回[%s]", tvmid, util.BytesToString(resp.Body))


	// check status
	var respinfo map[string]interface{}
	unerr := json.Unmarshal(resp.Body, &respinfo)
	if unerr != nil {
		log.Info("[充值] 玩家账户[%s] 获取充值页面地址失败 json.Unmarshal 'status' Fail[%s] ", tvmid, unerr)
		return
	}

	// 
	var statuscode int32 = int32(respinfo["status"].(float64))
	if statuscode != 0 {
		log.Error("[充值] 玩家账户[%s] 获取充值页面地址失败 status[%d] msg[%s]", tvmid, statuscode, respinfo["msg"])
		return
	}

	//log.Info("获取玩家账户[%s] 的收货地址成功 [%#v]", tvmid, respinfo["data"])
	data, convertok := respinfo["data"].(map[string]interface{})
	if convertok == false {
		log.Info("[充值] 玩家账户[%s] 获取充值页面地址失败 convertok=false", tvmid)
		return
	}
	log.Trace("[充值] 玩家账户[%s] 获取充值页面地址data=%#v", tvmid, data)

	url_checkout , findit := data["checkout"]
	if findit == true {
		send := &msg.GW2C_RetRechargeMoney{Urlcheckout:pb.String(url_checkout.(string))}
		user.SendMsg(send)
		log.Info("[充值] 玩家账户[%s] 获得充值页面成功 地址:%s 订单id:%s 金额:%d", tvmid, url_checkout, order_id, amount)

		// 订单详情
		Redis().Set(order_id, util.BytesToString(postbody), 0)
		
		// 绑定订单到玩家身上
		key_unfinish_order := fmt.Sprintf("%d_submit_recharge_orders", user.Id())
		Redis().SAdd(key_unfinish_order, order_id)
	}
}

