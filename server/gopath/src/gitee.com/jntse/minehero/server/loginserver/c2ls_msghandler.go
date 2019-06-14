package main
import ( 
	_"time"
	_"reflect"
	"fmt"
	_"math/rand"
	"crypto/md5"
	"strings"
	pb "github.com/golang/protobuf/proto"
	_  "github.com/go-redis/redis"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/net"
	"gitee.com/jntse/gotoolkit/redis"
	"gitee.com/jntse/gotoolkit/util"
	"gitee.com/jntse/gotoolkit/eventqueue"
	"gitee.com/jntse/minehero/server/tbl"
	"gitee.com/jntse/minehero/pbmsg"
	"gitee.com/jntse/minehero/server/def"
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

func (this* C2LSMsgHandler) Init() {

	this.msgparser = network.NewProtoParser("C2LS_MsgParser", tbl.ProtoMsgIndexGenerator)
	if this.msgparser == nil {
		return
	}

	// 收
	this.msgparser.RegistProtoMsg(msg.C2L_ReqLogin{}, on_C2L_ReqLogin)
	this.msgparser.RegistProtoMsg(msg.C2L_ReqRegistAccount{}, on_C2L_ReqRegistAccount)


	// 发
	this.msgparser.RegistSendProto(msg.L2C_RetLogin{})
	this.msgparser.RegistSendProto(msg.L2C_RetRegistAccount{})
}

func newL2C_RetLogin(reason string, ip string, port int, key string) *msg.L2C_RetLogin {
	send := &msg.L2C_RetLogin {
		Result : pb.Int32(1),
		Reason : pb.String(reason),
		Gatehost : &msg.IpHost {
		Ip : pb.String(ip),
		Port : pb.Int(port),
		},
		Verifykey : pb.String(key),
	}
	if reason != "" {
		send.Result = pb.Int32(0)
	}
	return send
}


// 注册账户
func on_C2L_ReqRegistAccount(session network.IBaseNetSession, message interface{}) {
	tmsg := message.(*msg.C2L_ReqRegistAccount)

	account, passwd, name, face := tmsg.GetAccount(), tmsg.GetPasswd(), tmsg.GetName(), tmsg.GetFace()
	errcode := registAccount(account, passwd, name, face, "")

	// 回复
	send := &msg.L2C_RetRegistAccount {
		Account : pb.String(account),
		Errcode : pb.String(errcode),
	}
	session.SendCmd(send)
}

func registAccount(account, passwd, name, face, token string) (errcode string) {
	errcode = ""
	switch {
	default:

		// 账户检查重复
		key := fmt.Sprintf("accounts_%s", account)
		bexist , err := Redis().Exists(key).Result()
		if err != nil {
			errcode = "redis暂时不可用"
			log.Error("检查账户是否存在 Redis错误:%s", err)
			break
		}

		if bexist == 1 {
			errcode = "账户已经存在"
			break
		}

		// 实名认证
		// 生成唯一userid
		userid , errstr := GenerateUserId()
		if errstr !=  "" {
			errcode = errstr
			break
		}

		// 新建账户
		info := &msg.AccountInfo {
			Account: &account,
			Passwd: &passwd,
			Userid: pb.Uint64(userid),
		}

		if err := utredis.SetProtoBin(Redis(), key, info); err != nil {
			errcode = "插入账户数据失败"
			log.Error("新建账户%s失败，err: %s", account, err)
			break
		}
		
		// 初始元宝和金卷
		Yuanbao, Coupon := uint32(tbl.Global.Newuser.Yuanbao), uint32(tbl.Global.Newuser.Coupon)
		userinfo := &msg.Serialize {
			Entity : &msg.EntityBase{ Id:pb.Uint64(userid), Name:pb.String(name), Face:pb.String(face), Account:pb.String(account) },
			Base : &msg.UserBase{Money: pb.Uint32(0), Coupon:pb.Uint32(Coupon), Yuanbao:pb.Uint32(Yuanbao), Level:pb.Uint32(1)},
			Item : nil,
			//Item : &msg.ItemBin{Items:make([]*msg.ItemData,0)},
		}
		//Item , Pos := userinfo.GetItem(), int32(msg.ItemPos_Bag)
		//Item.Items = append(Item.Items, &msg.ItemData{Id:pb.Uint32(101),Num:pb.Uint32(10), Pos:pb.Int32(Pos)})
		//Item.Items = append(Item.Items, &msg.ItemData{Id:pb.Uint32(102),Num:pb.Uint32(20), Pos:pb.Int32(Pos)})
		userkey := fmt.Sprintf("userbin_%d", userid)
		//log.Info("userinfo=%v",userinfo)
		if err := utredis.SetProtoBin(Redis(), userkey, userinfo); err != nil {
			errcode = "插入玩家数据失败"
			log.Error("新建账户%s插入玩家数据失败，err: %s", account, err)
			break
		}

		arglist := []interface{}{account, token, name, uint64(userid)}
		event := eventque.NewCommonEvent(arglist, def.HttpRequestNewUserArglist, nil)
		Login().AsynEventInsert(event)
		log.Info("账户[%s] UserId[%d] 名字[%s] 创建新用户成功", account, userid, name)
	}

	if errcode != "" {
		log.Info("账户[%s] 创建新用户失败 err[%s]", account, errcode)
	}

	return errcode
}


// 请求登陆验证
func on_C2L_ReqLogin(session network.IBaseNetSession, message interface{}) {
	log.Info("开始验证登录")
	tmsg := message.(*msg.C2L_ReqLogin)
	tm1 := util.CURTIMEUS()
	errcode, account, name, face, token := "", tmsg.GetAccount(), tmsg.GetNickname(), tmsg.GetFace(), tmsg.GetToken()
	switch {
	default:
		// 1. 多Gate时避免多客户端同时登陆
		// 2. 已经登陆上的账户，在对应的Gate中要验证是否重复登陆
		if account == "" || strings.Contains(account, " ") {
			errcode = "账户空字符串或包含空格"
			break
		}

		if Login().FindAuthenAccount(account) == true {
			errcode = "同时登陆多个账户"
			break
		}

		// 校验，不需要密码了
		//errcode = Authenticate(session, account, "", name, face) 
		if CheckNewAccount(session, account, name, face, token) != "" {
			errcode = "检查新账户报错"
			break
		}

		// TODO: 从Redis获取账户缓存Gate信息，实现快速登陆
		log.Info("账户[%s]登陆Login token[%s]", account, token)
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
			Account : pb.String(account),
			Expire : pb.Int64(now+10000),	// 10秒
			Gatehost : pb.String(agent.Host()),
			Sid : pb.Int(session.Id()),
			Timestamp: pb.Int64(now),
			Verifykey : pb.String(md5string),
		}
		agent.SendMsg(sendmsg)
		Login().AddAuthenAccount(account, session)		// 避免同时登陆
		tm5 := util.CURTIMEUS()
		log.Info("登陆验证通过，请求注册玩家到Gate sid[%d] account[%s] host[%s] 登陆耗时%dus", session.Id(), account, agent.Host(), tm5-tm1)
		return
	}

	if errcode != "" {
		log.Info("账户:[%s] sid[%d] 登陆报错[%s]", account, session.Id(), errcode)
		session.SendCmd(newL2C_RetLogin(errcode, "", 0, ""))
		session.Close()
	}
}


