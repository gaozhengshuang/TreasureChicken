package main
import (
	"fmt"
	"sync"
	"runtime/debug"
	"os"
	_"time"
	_"strconv"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/net"
	"gitee.com/jntse/gotoolkit/util"
	_"gitee.com/jntse/minehero/pbmsg"
	"gitee.com/jntse/minehero/server/def"
	"gitee.com/jntse/minehero/server/tbl"
	"github.com/go-redis/redis"
	pb "github.com/gogo/protobuf/proto"
	"gitee.com/jntse/minehero/pbmsg"
)


func SignalInt(signal os.Signal)    {
	log.Info("SignalInt");
	g_KeyBordInput.Insert("quit")
}

func SignalTerm(signal os.Signal)   {
	log.Info("SignalTerm");
	g_KeyBordInput.Insert("quit")
}

func SignalHup(signal os.Signal)    {
	log.Info("SignalHup");
	g_KeyBordInput.Insert("quit")
}

func SignalCoreDump(signal os.Signal)    {
	log.Info("Signal[%d] Received", signal);
	g_KeyBordInput.Insert("quit")
}

func init() {
	fmt.Println("loginserver.init()")
}

type MsgProcess struct {
}

type MatchServer struct {
	net				*network.NetWork
	netconf			*network.NetConf
	//sessions		map[int]network.IBaseNetSession		// 及时删除，没有任何地方引用golang才会GC
	hredis      	*redis.Client
	mutex			sync.Mutex
	gatemgr			GateManager
	//usermgr		UserManager
	roomsvrmgr		RoomSvrManager
	authens			map[string]int
	msghandlers     []network.IBaseMsgHandler
	tblloader       *tbl.TblLoader
	runtimestamp	int64
}

var g_MatchServer *MatchServer = nil
func NewMatchServer() *MatchServer {
	if g_MatchServer == nil {
		c := &MatchServer{}
		g_MatchServer = c
	}
	return g_MatchServer
}

func Match() *MatchServer {
	return g_MatchServer
}

//func UserMgr() *UserManager {
//	return &Match().usermgr
//}

func GateSvrMgr() *GateManager {
	return &Match().gatemgr
}

func RoomSvrMgr() *RoomSvrManager {
	return &Match().roomsvrmgr
}

func Redis() *redis.Client {
	return Match().hredis
}

func (this *MatchServer) DoInputCmd(cmd string) {
	switch cmd {
	case "gates":
		log.Info("show gates list")
	case "free":
		log.Info("Start FreeOSMemory...")
		debug.FreeOSMemory()        // 谨慎使用
	case "notice1":
		log.Info("Notice Suspension Test!")
		noticemsg := &msg.GW2C_MsgNotice{Userid:pb.Uint64(0), Name:pb.String("玩家名字"), Face:pb.String("")}
		text := def.MakeNoticeText("玩家名字", "#ffffff", 26) + def.MakeNoticeText("获得", "#fffc00", 30) + def.MakeNoticeText("道具名字", "#ffffff", 26)
		noticemsg.Text = pb.String(text)
		noticemsg.Type = pb.Int32(int32(msg.NoticeType_Suspension))
		send := &msg.MS2GW_MsgNotice{ Notice : noticemsg}
		this.BroadcastGateMsg(send)
	case "notice2":
		log.Info("Notice Marquee  Test!")
		noticemsg := &msg.GW2C_MsgNotice{Userid:pb.Uint64(0), Name:pb.String("玩家名字"), Face:pb.String(""), Type:pb.Int32(0)}
		text := def.MakeNoticeText("玩家名字", "#ffffff", 26) + def.MakeNoticeText("获得", "#fffc00", 30) + def.MakeNoticeText("道具名字", "#ffffff", 26)
		noticemsg.Text = pb.String(text)
		noticemsg.Type = pb.Int32(int32(msg.NoticeType_Marquee))
		send := &msg.MS2GW_MsgNotice{ Notice : noticemsg}
		this.BroadcastGateMsg(send)
	}
}


// --------------------------------------------------------------------------
// --------------------------------------------------------------------------
func (this *MatchServer) OnClose(session network.IBaseNetSession) {
	sid := session.Id()
	//delete(this.sessions, sid)
	switch session.Name() {
	case "TaskGate":
		log.Info("和Gate连接断开 sid[%d]", sid)
		this.gatemgr.OnClose(sid)
	case "TaskRoom":
		log.Info("和Room连接断开 sid[%d]", sid)
		this.roomsvrmgr.OnClose(sid)
	default:
		log.Error("OnClose error not regist session:%+v", sid)
	}
}


// --------------------------------------------------------------------------
// --------------------------------------------------------------------------
func (this* MatchServer) OnConnect(session network.IBaseNetSession) {

	//this.sessions[session.Id()] = session
	switch session.Name() {
	case "TaskClient":
		//log.Trace("OnConnect clientsession:%+v", session)
		//session.SetUserDefdata(this)
	case "TaskGate":
		//log.Trace("OnConnect gatesession:%+v", session)
		//session.SetUserDefdata(this)
	case "TaskRoom":
	default:
		log.Error("not regist client OnConnect session:%+v", session)
		session.Close()
	}
}

func (this *MatchServer) SendMsg(id int, msg pb.Message) bool {
	//session, ok := this.sessions[id]
	//if ok == true {
	//	return session.SendCmd(msg)
	//}
	//return false
	return this.net.SendMsg(id, msg)
}


func (this *MatchServer) GetSession(id int) network.IBaseNetSession {
	//session, _ := this.sessions[id]
	//return session
	return this.net.FindSession(id)
}

// 
func (this* MatchServer) Init(fileconf string) bool {
	// 服务器配置
	netconf := &network.NetConf{}
	jsonerr := util.JsonConfParser(fileconf, netconf)
	if jsonerr != nil || netconf == nil {
		log.Error("JsonParser Error or netconf is nil: '%s'", jsonerr)
		return false
	}
	this.netconf = netconf
	log.Info("加载服务器配置ok...")

	// 游戏配置
	this.tblloader = tbl.NewTblLoader(netconf.TblPath)

	// 消息handler
	this.InitMsgHandler()

	// 杂项
	this.gatemgr.Init()
	//this.usermgr.Init()
	this.roomsvrmgr.Init()
	//this.sessions = make(map[int]network.IBaseNetSession)
	this.authens = make(map[string]int)
	this.runtimestamp = 0

	return true
}

func (this *MatchServer) InitMsgHandler() {
	if this.tblloader == nil { panic("should init 'tblloader' first") }
	this.msghandlers = append(this.msghandlers, NewGW2MSMsgHandler())
	this.msghandlers = append(this.msghandlers, NewRS2MSMsgHandler())
}


// 启动redis
func (this *MatchServer) StartRedis() bool {
	this.hredis = redis.NewClient(&redis.Options {
		Addr:     this.netconf.Redis.Host.String(), // "ip:host"
		Password: this.netconf.Redis.Passwd,        // no passwd 
		DB:       this.netconf.Redis.DB,            // 0: use default DB
	})

	_, err := this.hredis.Ping().Result()
	if err != nil {
		panic(err)
		return false
	}

	log.Info("连接Redis[%s]成功", this.netconf.Redis.Host.String())
	return true
}


// 启动网络
func (this* MatchServer) StartNetWork() bool {
	this.net = network.NewNetWork()
	if this.net == nil {
		return false
	}
	this.net.Init(this.netconf, this)
	this.net.SetHttpResponseHandler(HttpServerResponseCallBack)	// Http监听,需要设置处理回调
	if this.net.Start() == false {
		log.Info("初始化网络error...")
		return false
	}
	log.Info("初始化网络ok...")
	return true
}

// 启动完成
func (this *MatchServer) OnStart() {
	log.Info("开始执行OnStart")
	this.runtimestamp = util.CURTIMEMS()
	log.Info("结束执行OnStart")
}

// 程序退出最后清理
func (this *MatchServer) OnStop() {
}


//  退出
func (this* MatchServer) Quit() {
	if this.net != nil {
		this.net.Shutdown()
	}
}

// 主循环
func (this* MatchServer) Run() {

	// TODO:每帧处理2000条
	now := util.CURTIMEMS()
	lastrun := now - this.runtimestamp
	this.net.Dispatch(network.KFrameDispatchNum * 2, 1000)
	tm_dispath := util.CURTIMEMS()

	//
	this.roomsvrmgr.Tick(now)
	tm_roomtick := util.CURTIMEMS()

	//
	delay := tm_roomtick - now
	if lastrun + delay > 20 {
		log.Warn("统计帧耗时 lastrun[%d] total[%d] dispatch[%d] roomtick[%d]", lastrun, delay, tm_dispath - now, tm_roomtick-tm_dispath)
	}

	//
	this.runtimestamp = util.CURTIMEMS()
}

// 公告
func (this *MatchServer) BroadcastGateMsg(msg pb.Message) {
	this.gatemgr.Broadcast(msg)
}

func (this *MatchServer) Reload() {
	if this.tblloader != nil { this.tblloader.Reload()  }
}


