package main
import (
	"fmt"
	"time"
	"runtime"
	"runtime/debug"
	"strings"
	"os"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/net"
	"gitee.com/jntse/gotoolkit/util"
	"gitee.com/jntse/minehero/pbmsg"
	"gitee.com/jntse/minehero/server/tbl"
	"gitee.com/jntse/minehero/server/def"
	pb "github.com/gogo/protobuf/proto"
	"github.com/go-redis/redis"
)

func SignalInt(signal os.Signal)    { 
	log.Info("SignalInt")
	g_KeyBordInput.Insert("quit_graceful")
}

func SignalTerm(signal os.Signal)   { 
	log.Info("SignalTerm")
	g_KeyBordInput.Insert("quit_graceful")
}

func SignalHup(signal os.Signal)    { 
	log.Info("SignalHup")
	g_KeyBordInput.Insert("quit_graceful")
}

func SignalCoreDump(signal os.Signal)   { 
	log.Info("Signal[%d] Received", signal)
	g_KeyBordInput.Insert("quit_graceful")
}

func init() {
	fmt.Println("gateserver.init()")
}


type GateServer struct {
	netconf			*network.NetConf
	net				*network.NetWork
	loginsvr		network.IBaseNetSession
	matchsvr		network.IBaseNetSession
	hredis			*redis.Client
	usermgr			UserManager
	waitpool		LoginWaitPool
	roomsvrmgr		RoomSvrManager
	msghandlers		[]network.IBaseMsgHandler
	tblloader		*tbl.TblLoader
	//countmgr		CountManager
	rcounter		util.RedisCounter
	ticker1m		*util.GameTicker
	ticker1s		*util.GameTicker
	ticker100ms		*util.GameTicker
	quit_graceful 	bool
	runtimestamp	int64
	hourmonitor		*util.IntHourMonitorPool
}

var g_GateServer *GateServer = nil
func NewGateServer() *GateServer {
	if g_GateServer == nil { g_GateServer = &GateServer{} }
	return g_GateServer
}

func GateSvr() *GateServer {
	return g_GateServer
}

func Match() network.IBaseNetSession {
	return GateSvr().matchsvr
}

func UserMgr() *UserManager {
	return &GateSvr().usermgr
}

func WaitPool() *LoginWaitPool {
	return &GateSvr().waitpool
}

func RoomSvrMgr() *RoomSvrManager {
	return &GateSvr().roomsvrmgr
}

//func CountMgr() *CountManager {
//	return &GateSvr().countmgr
//}

func RCounter() *util.RedisCounter {
	return &GateSvr().rcounter
}

func Redis() *redis.Client {
	return GateSvr().hredis
}

func (this *GateServer) Name() string {
	return this.netconf.Name
}

func (this *GateServer) DoInputCmd(cmd string) {
	subcmd := strings.Split(cmd, " ")
	switch subcmd[0] {
	case "quit_graceful":
		this.QuitGraceful()
	case "gates":
		log.Info("show gates list")
	case "reload":
		this.tblloader.Reload()
	case "num":
		log.Info("user registed in gate num=%d", this.usermgr.Amount())
	case "online":
		log.Info("user online num=%d", this.usermgr.AmountOnline())
	case "post":
		TestHttpPostRedPacketPlatform(subcmd)
	case "gc":
		log.Info("Start Force GC...")
		runtime.GC()
	case "free":
		log.Info("Start FreeOSMemory...")
		debug.FreeOSMemory()		// 谨慎使用
	default:
		log.Error("Input Cmd Invalid cmd[%s]", cmd)
	}
}

func (this *GateServer) OnClose(session network.IBaseNetSession) {
	sid := session.Id()
	switch session.Name() {
	case "LoginConnector":
		this.loginsvr = nil
		log.Info("sid[%d] 和LoginServer连接断开", sid)
	case "MatchConnector":
		this.matchsvr = nil
		log.Info("sid[%d] 和MatchServer连接断开", sid)
		this.usermgr.OnMatchServerClose()
	case "TaskRoom":
		log.Info("sid[%d] 和RoomServer连接断开", sid)
		//this.roomsvrmgr.OnClose(sid)
		break
	case "TaskClient":
		log.Info("sid[%d] 和客户端连接断开", sid)
		user := ExtractSessionUser(session)
		if user != nil { user.OnDisconnect() }
	default:
		log.Error("OnClose error not regist session:%+v", sid)
	}
}

func (this *GateServer) OnConnect(session network.IBaseNetSession) {
	//log.Trace("OnConnect session:%+v", session)
	switch session.Name() {
	case "TaskClient":
	case "TaskRoom":
	case "LoginConnector":
		this.loginsvr = session.(network.IBaseNetSession)
		this.RegistToLoginServer()
	case "MatchConnector":
		this.matchsvr = session.(network.IBaseNetSession)
		this.RegistToMatchServer()
		break
	default:
		log.Error("OnConnect error not regist session:%+v", session)
		session.Close()
	}
}

func (this *GateServer) Reload() {
	if this.tblloader != nil { this.tblloader.Reload() }
}

// 
func (this *GateServer) Init(fileconf string) bool {
	// 服务器配置
	netconf := &network.NetConf{}
	jsonerr := util.JsonConfParser(fileconf, netconf)
	if jsonerr != nil || netconf == nil {
		log.Error("JsonParser Error or netconf is nil: '%s'", jsonerr)
		return false
	}
	this.netconf = netconf
	log.Info("加载[%s]服务器配置ok...", netconf.Name)

	// 游戏配置
	this.tblloader = tbl.NewTblLoader(netconf.TblPath)

	// 消息handler
	this.InitMsgHandler()

	// 整点回调
	hourmonitor := util.NewIntHourMonitorPool()
	hourmonitor.Init()
	hourmonitor.Regist(0, ZeroHourClockCallback)
	for i:= 1; i < 24; i++ { hourmonitor.Regist(int64(i), IntHourClockCallback) }
	this.hourmonitor = hourmonitor

	//
	this.usermgr.Init()
	this.waitpool.Init()
	this.roomsvrmgr.Init()
	//this.countmgr.Init()
	//this.gamemgr.Init()
	this.ticker1m = util.NewGameTicker(60 * time.Second, this.Handler1mTick)
	this.ticker1s = util.NewGameTicker(01 * time.Second, this.Handler1sTick)
	this.ticker100ms = util.NewGameTicker(100 * time.Millisecond, this.Handler100msTick)
	this.ticker1m.Start()
	this.ticker1s.Start()
	this.ticker100ms.Start()
	this.runtimestamp = 0
	return true
}

func (this *GateServer) Handler1mTick(now int64) {
	this.rcounter.BatchSave(20)
}

func (this *GateServer) Handler1sTick(now int64) {
}

func (this *GateServer) Handler100msTick(now int64) {
	this.waitpool.Tick(now)

	// 
	if this.quit_graceful && this.usermgr.Amount() == 0 {
		g_KeyBordInput.Insert("quit")
	}

    this.roomsvrmgr.Tick(now)
}

func (this *GateServer) InitMsgHandler() {
	if this.tblloader == nil { panic("should init 'tblloader' first") }
	this.msghandlers = append(this.msghandlers, NewC2GWMsgHandler())
	this.msghandlers = append(this.msghandlers, NewLS2GMsgHandler())
	this.msghandlers = append(this.msghandlers, NewMS2GWMsgHandler())
	this.msghandlers = append(this.msghandlers, NewRS2GWMsgHandler())

}


// 启动redis
func (this *GateServer) StartRedis() bool {
	this.hredis = redis.NewClient(&redis.Options {
		Addr:     this.netconf.Redis.Host.String(),	// "ip:host"
		Password: this.netconf.Redis.Passwd,		// no passwd 
		DB:       this.netconf.Redis.DB,  			// 0: use default DB
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
func (this *GateServer) StartNetWork() bool {
	
	// tcp network
	this.net = network.NewNetWork()
	if this.net == nil {
		return false
	}
	this.net.Init(this.netconf, this)
	this.net.SetHttpResponseHandler(HttpServerResponseCallBack) // Http监听,需要设置处理回调
	if this.net.Start() == false {
		log.Info("初始化网络error...")
		return false
	}
	log.Info("初始化网络ok...")
	return true
}


//  退出
func (this *GateServer) Quit() {
	log.Info("服务器 QuitForce")
	if this.net != nil {
		this.net.Shutdown()
	}
}

// 优雅退出
func (this *GateServer) QuitGraceful() {
	this.quit_graceful = true
	UserMgr().OnServerClose()
	log.Info("服务器 QuitGraceful")
}

// 启动完成
func (this *GateServer) OnStart() {
	log.Info("开始执行OnStart")

	// 清除Gate上的账户信息
	clientconf := this.ClientListenerConf()
	ClearGateAccounts(clientconf.Host.Ip, clientconf.Host.Port)

	//
	this.rcounter.Init(Redis())

	//
	this.runtimestamp = util.CURTIMEMS()
	log.Info("结束执行OnStart")
}

// 程序退出最后清理
func (this *GateServer) OnStop() {
	this.ticker1s.Stop()
	this.ticker1m.Stop()
	this.ticker100ms.Stop()
}

func (this *GateServer) ClientListenerConf() *network.WsListenConf {
	conf , findok := this.netconf.FindWsListenConf("ClientListener")
	if findok == false { panic("not found ClientListener") }
	return &conf
}



// 主循环
func (this *GateServer) Run() {

	// TODO:每帧处理1000条
	now := util.CURTIMEMS()		// 毫秒
	lastrun := now - this.runtimestamp
	this.net.Dispatch(network.KFrameDispatchNum)
	tm_dispath := util.CURTIMEMS()

	// 测试日志
	doEventStatistics(this)

	//
	this.usermgr.Tick(now)
	tm_usrticker := util.CURTIMEMS()
	
	//
	this.hourmonitor.Run(now/1000)	// 秒
	this.ticker1m.Run(now)			// 毫秒
	this.ticker1s.Run(now)			// 毫秒
	this.ticker100ms.Run(now)		// 毫秒
	tm_svrticker := util.CURTIMEMS()

	// 每帧统计耗时
	delay := tm_svrticker - now
	if lastrun + delay > 20 {	// 20毫秒
		log.Warn("统计帧耗时 lastrun[%d] total[%d] dispatch[%d] userticker[%d] svrticker[%d]", 
			lastrun, delay, tm_dispath - now, tm_usrticker - tm_dispath, tm_svrticker - tm_usrticker)
	}
	this.runtimestamp = util.CURTIMEMS()
}


// 注册到Login
func (this *GateServer) RegistToLoginServer() {
	if this.loginsvr == nil { return }
	conf := this.ClientListenerConf()
	send := &msg.GW2L_ReqRegist {
		Account : pb.String("gate_account_123"),
		Passwd : pb.String("gate_passwd_123"),
		Host : &msg.IpHost{
		Ip: pb.String(conf.Host.Ip), 
		Port: pb.Int(conf.Host.Port)},
		Name: pb.String(this.Name()),
	}
	this.loginsvr.SendCmd(send)
	log.Info("请求注册网关'%s'到Login", conf.Host.String())
}

func (this *GateServer) RegistToMatchServer() {
	if this.matchsvr == nil {
		return
	}

	conf , findok := this.netconf.FindTcpListenConf("RoomListener")
	if findok == false {
		log.Error("not found conf 'ClientListener'")
		return
	}

	send := &msg.GW2MS_ReqRegist {
		Account : pb.String("gate_account_123"),
		Passwd : pb.String("gate_passwd_123"),
		Agentname : pb.String(this.Name()),
		Host : &msg.IpHost{
		Ip: pb.String(conf.Host.Ip), 
		Port: pb.Int(conf.Host.Port),
		},
	}
	this.matchsvr.SendCmd(send)
	log.Info("请求注册网关'%s'到Match", conf.Host.String())
}

func (this *GateServer) RegistRoomServer(agent *RoomAgent) {
	this.roomsvrmgr.AddRoom(agent)
	log.Info("注册房间服 id=%d 当前总数:%d", agent.Id(), this.roomsvrmgr.Num())
}

//func (this *GateServer) GetUserBySession(session network.IBaseNetSession) *GateUser {
//	defdata := session.UserDefData()
//	if defdata == nil {
//		log.Error("客户端Session没有绑定账户数据 sid[%d]", session.Id())
//		return nil
//	}
//
//	account, ok := defdata.(string)
//	if ok == false {
//		log.Error("客户端Session绑定了错误的数据")
//		return nil
//	}
//
//	return UserMgr().FindByAccount(account)
//}

// 通用公告
func (this *GateServer) SendNotice(face string, ty msg.NoticeType, subtext ...string) {
	noticemsg := &msg.GW2C_MsgNotice{Userid:pb.Uint64(0), Name:pb.String(""), Face:pb.String(face), Type:pb.Int32(int32(ty))}
	noticemsg.Text = pb.String(strings.Join(subtext, ""))

	send := &msg.GW2MS_MsgNotice{ Notice: noticemsg}
	Match().SendCmd(send)
}

// 重启服务器清理服务器上账户信息
func ClearGateAccounts(ip string, port int) error {
	key := fmt.Sprintf("%s_%s:%d", def.RedisKeyGateAccounts, ip, port)
	err := Redis().Del(key).Err();
	if err == nil { log.Info("Del Redis Key[%s] Ok", key) }
	return err
}

// Redis移除账户网关记录
func UnBindingAccountGateWay(account string) {

	// 解除绑定账户的GateWay信息
	key := fmt.Sprintf("%s_%s", def.RedisKeyAccountGate, account)
	Redis().Del(key)

	// GateWay移除账户信息
	clientconf := GateSvr().ClientListenerConf()
	ip, port := clientconf.Host.Ip, clientconf.Host.Port
	key = fmt.Sprintf("%s_%s:%d", def.RedisKeyGateAccounts, ip, port)
	Redis().SRem(key, account)
}

func ZeroHourClockCallback(now int64) {
	log.Info("==========零点回调开始===========")
	//UserMgr().GiveFreeStep(now)
	log.Info("==========零点回调结束===========")
}

func IntHourClockCallback(now int64) {
	log.Info("==========整点回调开始===========")
	//UserMgr().GiveFreeStep(now)
	log.Info("==========整点点回调结束===========")
}


