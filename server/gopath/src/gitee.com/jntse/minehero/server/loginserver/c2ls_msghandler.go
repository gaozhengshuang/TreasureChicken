package main

import (
	"crypto/md5"
	"fmt"
	_ "gitee.com/jntse/gotoolkit/eventqueue"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/net"
	"gitee.com/jntse/gotoolkit/util"
	"gitee.com/jntse/minehero/pbmsg"
	"gitee.com/jntse/minehero/server/tbl"
	pb "github.com/gogo/protobuf/proto"
	_ "math/rand"
	_ "reflect"
	"strings"
)

//func init() {
//	NewC2LSMsgHandler()
//}

type C2LSMsgHandler struct {
	msgparser *network.ProtoParser
}

func NewC2LSMsgHandler() *C2LSMsgHandler {
	handler := &C2LSMsgHandler{}
	handler.Init()
	return handler
}

func (this *C2LSMsgHandler) Init() {

	this.msgparser = network.NewProtoParser("C2LS_MsgParser", tbl.ProtoMsgIndexGenerator)
	if this.msgparser == nil {
		return
	}

	// 收
	this.msgparser.RegistProtoMsg(msg.C2L_ReqLogin{}, on_C2L_ReqLogin)
	this.msgparser.RegistProtoMsg(msg.C2L_ReqLoginWechat{}, on_C2L_ReqLoginWechat)
	this.msgparser.RegistProtoMsg(msg.C2L_ReqRegistAccount{}, on_C2L_ReqRegistAccount)
	this.msgparser.RegistProtoMsg(msg.C2L_ReqRegistAuthCode{}, on_C2L_ReqRegistAuthCode)

	// 发
	this.msgparser.RegistSendProto(msg.L2C_RetLogin{})
	this.msgparser.RegistSendProto(msg.L2C_RetRegistAccount{})
}

func newL2C_RetLogin(reason string, ip string, port int, key string) *msg.L2C_RetLogin {
	send := &msg.L2C_RetLogin{
		Result: pb.Int32(1),
		Reason: pb.String(reason),
		Gatehost: &msg.IpHost{
			Ip:   pb.String(ip),
			Port: pb.Int(port),
		},
		Verifykey: pb.String(key),
	}
	if reason != "" {
		send.Result = pb.Int32(0)
	}
	return send
}

// 获取手机验证码
func on_C2L_ReqRegistAuthCode(session network.IBaseNetSession, message interface{}) {
	//tmsg := message.(*msg.C2L_ReqRegistAuthCode)
	//GetRegistAuthCode(tmsg.GetPhone())
}

// 注册账户
func on_C2L_ReqRegistAccount(session network.IBaseNetSession, message interface{}) {
	tmsg := message.(*msg.C2L_ReqRegistAccount)
	errcode, phone, passwd, authcode, invitationcode := "", tmsg.GetPhone(), tmsg.GetPasswd(), tmsg.GetAuthcode(), tmsg.GetInvitationcode()
	nickname, account := tmsg.GetNickname(), phone

	switch {
	default:
		if errcode = RegistAccountCheck(account, passwd, invitationcode, authcode, nickname); errcode != "" {
			break
		}

		if errcode = RegistAccount(account, passwd, invitationcode, nickname, ""); errcode != "" {
			break
		}
	}

	// 回复
	send := &msg.L2C_RetRegistAccount{Errcode: pb.String(errcode)}
	session.SendCmd(send)
	if errcode != "" {
		log.Info("[注册] 账户[%s] 注册失败[%s]", phone, errcode)
	}
}

// 请求登陆验证
func on_C2L_ReqLogin(session network.IBaseNetSession, message interface{}) {
	tmsg := message.(*msg.C2L_ReqLogin)
	tm1 := util.CURTIMEUS()
	errcode, account, passwd := "", tmsg.GetAccount(), tmsg.GetPasswd()
	switch {
	default:
		// 1. 多Gate时避免多客户端同时登陆
		// 2. 已经登陆上的账户，在对应的Gate中要验证是否重复登陆
		if account == "" || strings.Contains(account, " ") {
			errcode = "账户空字符串或包含空格"
			break
		}

		// 登陆验证
		//if errcode = Authenticate(session, account, passwd); errcode != "" {
		//	break
		//}
		if CheckNewAccount(session, account, "", "", passwd) != "" {
			errcode = "检查新账户报错"
			break
		}

		if Login().FindAuthenAccount(account) == true {
			errcode = "同时登陆多个账户"
			break
		}

		// TODO: 从Redis获取账户缓存Gate信息，实现快速登陆
		log.Info("账户[%s]登陆Login ", account)
		if ok := QuickLogin(session, account); ok == true {
			return
		}

		// 挑选一个负载较低的agent
		agent := GateMgr().FindLowLoadGate()
		if agent == nil {
			errcode = "没有可用Gate"
			break
		}

		// 生成校验key，玩家简单信息发送到对应Gate,用于验证玩家登陆Gate合法
		now := util.CURTIMEMS()
		signbytes := []byte(fmt.Sprintf("<%d-%s>", now, account))
		md5array := md5.Sum(signbytes)
		md5bytes := []byte(md5array[:])
		md5string := fmt.Sprintf("%s_%x", account, md5bytes)

		sendmsg := &msg.L2GW_ReqRegistUser{
			Account:   pb.String(account),
			Expire:    pb.Int64(now + 10000), // 10秒
			Gatehost:  pb.String(agent.Host()),
			Sid:       pb.Int(session.Id()),
			Timestamp: pb.Int64(now),
			Verifykey: pb.String(md5string),
		}
		agent.SendMsg(sendmsg)
		Login().AddAuthenAccount(account, session) // 避免同时登陆
		tm5 := util.CURTIMEUS()
		log.Info("登陆验证通过，请求注册玩家到Gate sid[%d] account[%s] host[%s] 登陆耗时%dus", session.Id(), account, agent.Host(), tm5-tm1)
		return
	}

	if errcode != "" {
		log.Info("账户:[%s] sid[%d] 登陆失败[%s]", account, session.Id(), errcode)
		session.SendCmd(newL2C_RetLogin(errcode, "", 0, ""))
		session.Close()
	}
}

// 微信小游戏登陆
func on_C2L_ReqLoginWechat(session network.IBaseNetSession, message interface{}) {
	tmsg := message.(*msg.C2L_ReqLoginWechat)
	tm1 := util.CURTIMEUS()
	errcode, openid, face, nickname := "", tmsg.GetOpenid(), tmsg.GetFace(), tmsg.GetNickname()
	account, passwd, invitationcode := openid, "", ""
	switch {
	default:
		// 1. 多Gate时避免多客户端同时登陆
		// 2. 已经登陆上的账户，在对应的Gate中要验证是否重复登陆
		if account == "" || strings.Contains(account, " ") {
			errcode = "账户空字符串或包含空格"
			break
		}

		//// 登陆验证
		//if errcode = Authenticate(session, account, passwd); errcode != "" {
		//	break
		//}
		if errcode = RegistAccountFromWechatMiniGame(account, passwd, invitationcode, nickname, face); errcode != "" {
			break
		}

		if Login().FindAuthenAccount(account) == true {
			errcode = "同时登陆多个账户"
			break
		}

		// TODO: 从Redis获取账户缓存Gate信息，实现快速登陆
		log.Info("账户[%s]登陆Login ", account)
		if ok := QuickLogin(session, account); ok == true {
			return
		}

		// 挑选一个负载较低的agent
		agent := GateMgr().FindLowLoadGate()
		if agent == nil {
			errcode = "没有可用Gate"
			break
		}

		// 生成校验key，玩家简单信息发送到对应Gate,用于验证玩家登陆Gate合法
		now := util.CURTIMEMS()
		signbytes := []byte(fmt.Sprintf("<%d-%s>", now, account))
		md5array := md5.Sum(signbytes)
		md5bytes := []byte(md5array[:])
		md5string := fmt.Sprintf("%s_%x", account, md5bytes)

		sendmsg := &msg.L2GW_ReqRegistUser{
			Account:   pb.String(account),
			Expire:    pb.Int64(now + 10000), // 10秒
			Gatehost:  pb.String(agent.Host()),
			Sid:       pb.Int(session.Id()),
			Timestamp: pb.Int64(now),
			Verifykey: pb.String(md5string),
		}
		agent.SendMsg(sendmsg)
		Login().AddAuthenAccount(account, session) // 避免同时登陆
		tm5 := util.CURTIMEUS()
		log.Info("登陆验证通过，请求注册玩家到Gate sid[%d] account[%s] host[%s] 登陆耗时%dus", session.Id(), account, agent.Host(), tm5-tm1)
		return
	}

	if errcode != "" {
		log.Info("账户:[%s] sid[%d] 登陆失败[%s]", account, session.Id(), errcode)
		session.SendCmd(newL2C_RetLogin(errcode, "", 0, ""))
		session.Close()
	}
}
