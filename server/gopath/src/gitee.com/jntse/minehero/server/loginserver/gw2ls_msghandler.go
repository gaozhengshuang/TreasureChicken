package main
import (
	"fmt"
	_"time"
	"reflect"
	_"strings"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/net"
	_"gitee.com/jntse/gotoolkit/util"
	pb"github.com/golang/protobuf/proto"
	"gitee.com/jntse/minehero/pbmsg"
	"gitee.com/jntse/minehero/server/tbl"
)


//func init() {
//	NewGW2LMsgHandler()
//}

type GW2LMsgHandler struct {
	msgparser *network.ProtoParser
}

func NewGW2LMsgHandler() *GW2LMsgHandler {
	handler := &GW2LMsgHandler{}
	handler.Init()
	return handler
}

func (this* GW2LMsgHandler) Init() {

	this.msgparser = network.NewProtoParser("GW2LS_MsgParser", tbl.ProtoMsgIndexGenerator)
	if this.msgparser == nil {
		return
	}

	// 收
	this.msgparser.RegistProtoMsg(msg.GW2L_ReqRegist{}, on_GW2L_ReqRegist)
	this.msgparser.RegistProtoMsg(msg.GW2L_HeartBeat{}, on_GW2L_HeartBeat)
	this.msgparser.RegistProtoMsg(msg.GW2L_RegistUserRet{}, on_GW2L_RegistUserRet)

	// 发
	this.msgparser.RegistSendProto(msg.L2GW_RetRegist{})
	this.msgparser.RegistSendProto(msg.L2GW_HeartBeat{})
	this.msgparser.RegistSendProto(msg.L2GW_ReqRegistUser{})

}

func on_GW2L_ReqRegist(session network.IBaseNetSession, message interface{}) {
	tmsg := message.(*msg.GW2L_ReqRegist)
	host := tmsg.GetHost()
	if host == nil {
		log.Error("Gate注册IP端口失败, Host is nil")
		return
	}

	ip, port := host.GetIp(), int(host.GetPort())
	if tmsg.GetAccount() != "gate_account_123" || tmsg.GetPasswd() != "gate_passwd_123" {
		log.Info("Gate验证失败，断开")
		session.Close()
		return
	}


	// TODO: 重复注册服务器为了第一时间发现问题使用了panic
	//if GateMgr().IsRegisted(ip, port) == true {
	if GateMgr().IsRegistedByName(tmsg.GetName(), ip, port) == true {
		log.Fatal(fmt.Sprintf("重复注册网关服务器 %s:%d", ip, port))
		session.SendCmd(&msg.L2GW_RetRegist{Errocde:pb.String("重复注册网关服务器"), Host:tmsg.Host})
		return
	}
	
	// 添加gate
	GateMgr().AddNew(session, ip, port)
	session.SendCmd(&msg.L2GW_RetRegist{Host:tmsg.Host})
}

func on_GW2L_HeartBeat(session network.IBaseNetSession, message interface{}) {
	tmsg := message.(*msg.GW2L_HeartBeat)
	log.Info(reflect.TypeOf(tmsg).String())
}

func on_GW2L_RegistUserRet(session network.IBaseNetSession, message interface{}) {
	tmsg := message.(*msg.GW2L_RegistUserRet)
	//log.Info(reflect.TypeOf(tmsg).String())

	account, ipport, errcode := tmsg.GetAccount(), tmsg.GetGatehost(), tmsg.GetErrcode()
	verifykey := tmsg.GetVerifykey()
	sid  := int(tmsg.GetSid())
	host := network.NewNetHost(ipport)

	// 获取玩家Session
	usession := Login().GetSession(sid)
	if usession == nil || reflect.ValueOf(usession).IsNil() {
		log.Error("注册玩家%s到Gate返回，玩家已经离开Login", account)
		return
	}

	if errcode != "" {
		log.Info("注册玩家%s到Gate失败 err=%s", account, errcode)
		Login().SendMsg(sid, newL2C_RetLogin("注册到Gate失败", "", 0, ""))
		usession.Close()
		return
	}

	// gate agent 实例
	gate := GateMgr().FindGateByHost(host.String())
	if gate == nil {
		log.Error("账户%s，获取Gate实例失败 host[%s]", account, host.String())
		Login().SendMsg(sid, newL2C_RetLogin("获取Gate实例失败", "", 0, ""))
		usession.Close()
		return
	}

	// 绑定账户Gateway信息，用于快速登陆
	if err := BindingAccountGateWay(account, host.Ip, host.Port, verifykey); err != nil {
		log.Error("账户%s,保存GateHost到Redis失败", account)
		Login().SendMsg(sid, newL2C_RetLogin("保存GateHost失败", "", 0, ""))
		usession.Close()
		return
	}

	Login().SendMsg(sid, newL2C_RetLogin("", host.Ip, host.Port, verifykey))
	log.Info("注册玩家到Gate成功 sid[%d] [%s] [%s] [%s]", sid, account, ipport, verifykey)
	return
}


