package main
import (
	"reflect"
	_"fmt"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/net"
	"gitee.com/jntse/minehero/pbmsg"
	"gitee.com/jntse/minehero/server/tbl"
	pb "github.com/gogo/protobuf/proto"
)


//func init() {
//	log.Info("rs2gw_msghandler.init")
//	NewRS2GWMsgHandler()
//}

type RS2GWMsgHandler struct {
	msgparser *network.ProtoParser
}

func NewRS2GWMsgHandler() *RS2GWMsgHandler {
	handler := &RS2GWMsgHandler{}
	handler.Init()
	return handler
}

func (this* RS2GWMsgHandler) Init() {

	this.msgparser = network.NewProtoParser("RS2GW_MsgParser", tbl.ProtoMsgIndexGenerator)
	if this.msgparser == nil {
		return
	}


	// 收
	this.msgparser.RegistProtoMsg(msg.RS2GW_ReqRegist{}, on_RS2GW_ReqRegist)
	this.msgparser.RegistProtoMsg(msg.RS2GW_RetUserDisconnect{}, on_RS2GW_RetUserDisconnect)
	this.msgparser.RegistProtoMsg(msg.RS2GW_MsgTransfer{}, on_RS2GW_MsgTransfer)
	this.msgparser.RegistProtoMsg(msg.GW2C_MsgNotify{}, on_GW2C_MsgNotify)
	//this.msgparser.RegistProtoMsg(msg.BT_GameInit{}, on_BT_GameInit)
	//this.msgparser.RegistProtoMsg(msg.BT_SendBattleUser{}, on_BT_SendBattleUser)
	//this.msgparser.RegistProtoMsg(msg.BT_GameStart{}, on_BT_GameStart)
	this.msgparser.RegistProtoMsg(msg.BT_GameEnd{}, on_BT_GameEnd)
	//this.msgparser.RegistProtoMsg(msg.BT_PickItem{}, on_BT_PickItem)

	// 发
	this.msgparser.RegistSendProto(msg.GW2RS_RetRegist{})
	this.msgparser.RegistSendProto(msg.GW2RS_UserDisconnect{})
	this.msgparser.RegistSendProto(msg.GW2RS_MsgTransfer{})
	this.msgparser.RegistSendProto(msg.BT_UploadGameUser{})
	this.msgparser.RegistSendProto(msg.BT_ReqEnterRoom{})
	this.msgparser.RegistSendProto(msg.BT_ReqQuitGameRoom{})
	this.msgparser.RegistSendProto(msg.BT_UpdateMoney{})
	this.msgparser.RegistSendProto(msg.C2GW_StartLuckyDraw{})
}

func on_RS2GW_ReqRegist(session network.IBaseNetSession, message interface{}) {
	//tmsg := message.(*msg.RS2GW_ReqRegist)
	//log.Info(reflect.TypeOf(tmsg).String())

	// TODO: 重复注册服务器为了第一时间发现问题使用了panic
	//name := tmsg.GetAgentname()
	//if RoomSvrMgr().IsRegisted(name) == true {
	//	log.Fatal(fmt.Sprintf("重复注册房间服务器 %s", name))
	//	session.SendCmd(&msg.GW2RS_RetRegist{Errcode: pb.String("重复注册房间服务器"), Agentname:pb.String(GateSvr().Name())})
	//	return
	//}

	//RoomSvrMgr().AddNew(session, name)
	//session.SendCmd(&msg.GW2RS_RetRegist{Agentname:pb.String(GateSvr().Name())})
}

func on_GW2C_MsgNotify(session network.IBaseNetSession, message interface{}) {
	tmsg := message.(*msg.GW2C_MsgNotify)
	if user := UserMgr().FindById(tmsg.GetUserid()); user != nil {
		user.SendMsg(tmsg)
	}
}

func on_BT_GameInit(session network.IBaseNetSession, message interface{}) {
	tmsg := message.(*msg.BT_GameInit)
	user := UserMgr().FindById(tmsg.GetOwnerid())
	if user == nil {
		log.Error("房间[%d] BT_GameInit 找不到玩家[%d]", tmsg.GetRoomid(), tmsg.GetOwnerid())
		return
	}
	user.SendMsg(tmsg)
}

func on_BT_SendBattleUser(session network.IBaseNetSession, message interface{}) {
	tmsg := message.(*msg.BT_SendBattleUser)
	user := UserMgr().FindById(tmsg.GetOwnerid())
	if user == nil {
		log.Error("BT_SendBattleUser 找不到玩家[%d]", tmsg.GetOwnerid())
		return
	}
	user.SendMsg(tmsg)
}

func on_BT_GameStart(session network.IBaseNetSession, message interface{}) {
	tmsg := message.(*msg.BT_GameStart)
	user := UserMgr().FindById(tmsg.GetOwnerid())
	if user == nil {
		log.Error("房间[%d] BT_GameStart 找不到玩家[%d]", tmsg.GetRoomid(), tmsg.GetOwnerid())
		return
	}
	user.SendMsg(tmsg)
}

func on_BT_GameEnd(session network.IBaseNetSession, message interface{}) {
	tmsg := message.(*msg.BT_GameEnd)
	user := UserMgr().FindById(tmsg.GetOwnerid())
	if user == nil {
		log.Error("房间[%d] BT_GameEnd 找不到玩家[%d]", tmsg.GetRoomid(), tmsg.GetOwnerid())
		return
	}
	if user.IsOnline() { user.SendMsg(&msg.BT_GameOver{Roomid:tmsg.Roomid}) }
	//user.GameEnd(tmsg.GetBin(), tmsg.GetReason())
	log.Info("房间[%d] BT_GameEnd 游戏结束，Owner[%d]", tmsg.GetRoomid(), tmsg.GetOwnerid())
}

func on_BT_PickItem(session network.IBaseNetSession, message interface{}) {
	tmsg := message.(*msg.BT_PickItem)
	//log.Info(reflect.TypeOf(tmsg).String())
	user := UserMgr().FindById(tmsg.GetUserid())
	if user == nil {
		log.Error("BT_PickItem 找不到玩家[%d]", tmsg.GetUserid())
		return
	}
	user.SendMsg(tmsg)
}

func on_RS2GW_RetUserDisconnect(session network.IBaseNetSession, message interface{}) {
	tmsg := message.(*msg.RS2GW_RetUserDisconnect)
	//log.Info(reflect.TypeOf(tmsg).String())
	roomid, userid := tmsg.GetRoomid(), tmsg.GetUserid()
	user := UserMgr().FindById(userid)
	if user == nil {
		log.Error(" RS2GW_RetUserDisconnect 找不到玩家[%d]", userid)
		return
	}

	// RS关闭房间失败，直接下线玩家
	if tmsg.GetErrcode() != "" {
		log.Info("房间[%d] 玩家[%d] GW通知RS玩家断开连接，处理报错，原因[%s]", roomid, userid, tmsg.GetErrcode())
		user.Logout()
	}
}

func on_RS2GW_MsgTransfer(session network.IBaseNetSession, message interface{}) {
	tmsg := message.(*msg.RS2GW_MsgTransfer)
	msg_type := pb.MessageType(tmsg.GetName())
	if msg_type == nil {
		log.Fatal("消息转发解析失败，找不到proto msg=%s" , tmsg.GetName())
		return
	}

	protomsg := reflect.New(msg_type.Elem()).Interface()
	err := pb.Unmarshal(tmsg.GetBuf(), protomsg.(pb.Message))
	if err != nil {
		log.Fatal("消息转发解析失败，Unmarshal失败 msg=%s" , tmsg.GetName())
		return
	}

	user := UserMgr().FindById(tmsg.GetUid())
	if user == nil { return }
	user.SendMsg(protomsg.(pb.Message))
}


