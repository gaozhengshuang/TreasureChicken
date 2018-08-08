package main
import (
	"fmt"
	"reflect"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/net"
	pb "github.com/golang/protobuf/proto"
	"gitee.com/jntse/minehero/pbmsg"
	"gitee.com/jntse/minehero/server/tbl"
	)


//func init() {
//	log.Info("ls2gw_msghandler.init")
//	NewLS2GMsgHandler()
//}

type LS2GMsgHandler struct {
	msgparser *network.ProtoParser
}

func NewLS2GMsgHandler() *LS2GMsgHandler {
	handler := &LS2GMsgHandler{}
	handler.Init()
	return handler
}

func (this* LS2GMsgHandler) Init() {

	this.msgparser = network.NewProtoParser("LS2GW_MsgParser", tbl.ProtoMsgIndexGenerator)
	if this.msgparser == nil {
		return
	}


	// 收
	this.msgparser.RegistProtoMsg(msg.L2GW_RetRegist{}, on_L2GW_RetRegist)
	this.msgparser.RegistProtoMsg(msg.L2GW_HeartBeat{}, on_L2GW_HeartBeat)
	this.msgparser.RegistProtoMsg(msg.L2GW_ReqRegistUser{}, on_L2GW_ReqRegistUser)

	// 发
	this.msgparser.RegistSendProto(msg.GW2L_ReqRegist{})
	this.msgparser.RegistSendProto(msg.GW2L_HeartBeat{})
	this.msgparser.RegistSendProto(msg.GW2L_RegistUserRet{})

}

func on_L2GW_RetRegist(session network.IBaseNetSession, message interface{}) {
	tmsg := message.(*msg.L2GW_RetRegist)

	// 启动了相同的网关
	if tmsg.GetErrocde() != "" {
		panic(fmt.Sprintf("Login服务器通知注册失败 原因：%s", tmsg.GetErrocde()))
		return
	}
	log.Info("注册网关[%v]到Login成功", tmsg.GetHost())
}

func on_L2GW_HeartBeat(session network.IBaseNetSession, message interface{}) {
	tmsg := message.(*msg.L2GW_HeartBeat)
	log.Info(reflect.TypeOf(tmsg).String())
}

func on_L2GW_ReqRegistUser(session network.IBaseNetSession, message interface{}) {
	tmsg := message.(*msg.L2GW_ReqRegistUser)
	//log.Info(reflect.TypeOf(tmsg).String())
	account, gate, tm_expire, vkey := tmsg.GetAccount(), tmsg.GetGatehost(), tmsg.GetExpire(), tmsg.GetVerifykey()
	errcode := ""

	switch {
	default:
		if UserMgr().FindByAccount(account) != nil {
			log.Warn("账户%s 注册到Gate失败: [%s]", account, "玩家还未完全下线")
			errcode = "玩家还未完全下线"
			break
		}

		if WaitPool().IsFind(account) == true {
			log.Warn("账户%s 注册到Gate失败: [%s]", account, "已经注册过了")
			errcode = "已经注册过了"
			break
		}

		WaitPool().Insert(account, vkey, tm_expire)
		log.Info("账户%s 注册到Gate成功: verify[%s]", account, vkey)
	}

	sendmsg := &msg.GW2L_RegistUserRet {
		Account : pb.String(account),
		Gatehost : pb.String(gate),
		Sid : tmsg.Sid,
		Verifykey : pb.String(vkey),
	}
	sendmsg.Errcode = pb.String(errcode)
	session.SendCmd(sendmsg)

}


