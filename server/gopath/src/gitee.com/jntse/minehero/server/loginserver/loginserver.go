package main
import (
	"fmt"
	"sync"
	"time"
	_"strconv"
	"runtime"
	"runtime/debug"
	"os"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/net"
	"gitee.com/jntse/gotoolkit/util"
	"gitee.com/jntse/gotoolkit/eventqueue"
	_"gitee.com/jntse/minehero/pbmsg"
	"gitee.com/jntse/minehero/server/tbl"
	"github.com/go-redis/redis"
	pb "github.com/golang/protobuf/proto"
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

func SignalCoreDump(signal os.Signal)   {
	log.Info("Signal[%d] Received", signal);
	g_KeyBordInput.Insert("quit")
}


func init() {
	fmt.Println("loginserver.init()")
}

type MsgProcess struct {
}

type LoginServer struct {
	net				*network.NetWork
	netconf			*network.NetConf
	//sessions		map[int]network.IBaseNetSession		// 及时删除，没有任何地方引用golang才会GC
	hredis      	*redis.Client
	mutex			sync.Mutex
	gatemgr			GateManager
	//usermgr		UserManager
	login_now		map[string]*ClientAccount			// 正在登陆中的账户(不是在线)
	msghandlers		[]network.IBaseMsgHandler
	tblloader		*tbl.TblLoader
	runtimestamp    int64
	asynev      	eventque.AsynEventQueue 			// 异步事件处理
	ticker1ms    	*util.GameTicker
}

var g_LoginServer *LoginServer = nil
func NewLoginServer() *LoginServer {
	if g_LoginServer == nil {
		c := &LoginServer{}
		g_LoginServer = c
	}
	return g_LoginServer
}

func Login() *LoginServer {
	return g_LoginServer
}

//func UserMgr() *UserManager {
//	return &Login().usermgr
//}

func GateMgr() *GateManager {
	return &Login().gatemgr
}

func Redis() *redis.Client {
	return Login().hredis
}

func (this *LoginServer) DoInputCmd(cmd string) {
	switch cmd {
	case "gates":
		log.Info("show gates list")
	case "reload":
		this.tblloader.Reload()
	case "num":
		log.Info("num sessions=%d", this.net.SessionSize())
	case "gc":
		log.Info("Start Force GC...")
		runtime.GC()
	case "free":
		log.Info("Start FreeOSMemory...")
		debug.FreeOSMemory()        // 谨慎使用
	}
}


// --------------------------------------------------------------------------
// --------------------------------------------------------------------------
func (this *LoginServer) OnClose(session network.IBaseNetSession) {
	sid := session.Id()
	//delete(this.sessions, sid)
	switch session.Name() {
	case "TaskClient":
		log.Info("和客户端连接断开 sid[%d]", sid)
		this.DelAuthenAccount(sid)
	case "TaskGate":
		log.Info("和GateServer连接断开 sid[%d]", sid)
		this.gatemgr.OnClose(sid)
	default:
		log.Error("OnClose error not regist session:%+v", sid)
	}
}


// --------------------------------------------------------------------------
// --------------------------------------------------------------------------
func (this* LoginServer) OnConnect(session network.IBaseNetSession) {

	//_, findid := this.sessions[session.Id()]
	//if findid == true {
	//	panic(fmt.Sprintf("sid[%d] OnConnect 重复", session.Id()))
	//	return
	//}
	//this.sessions[session.Id()] = session

	switch session.Name() {
	case "TaskClient":
		//log.Info("OnConnect clientsession:%+v", session)
		//session.SetUserDefdata(this)
	case "TaskGate":
		//log.Trace("OnConnect gatesession:%+v", session)
		//session.SetUserDefdata(this)
	default:
		log.Error("not regist client OnConnect session:%+v", session)
		session.Close()
	}
}

func (this *LoginServer) SendMsg(id int, msg pb.Message) bool {
	//session, ok := this.sessions[id]
	//if ok == true {
	//	return session.SendCmd(msg)
	//}
	//return false
	return this.net.SendMsg(id, msg)
}


func (this *LoginServer) GetSession(id int) network.IBaseNetSession {
	//session, _ := this.sessions[id]
	//return session
	return this.net.FindSession(id)
}

func (this *LoginServer) InitMsgHandler() {
	if this.tblloader == nil { panic("should init 'tblloader' first") }
	this.msghandlers = append(this.msghandlers, NewGW2LMsgHandler())
	this.msghandlers = append(this.msghandlers, NewC2LSMsgHandler())
}

// 重新加载配置
func (this *LoginServer) Reload() {
	if this.tblloader != nil { this.tblloader.Reload() }
}

// 
func (this *LoginServer) Init(fileconf string) bool {

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
	this.ticker1ms = util.NewGameTicker(10 * time.Millisecond, this.Handler10msTick)
	this.ticker1ms.Start()
	this.gatemgr.Init()
	//this.sessions = make(map[int]network.IBaseNetSession)
	this.login_now = make(map[string]*ClientAccount)
	this.runtimestamp = 0
	this.asynev.Start(1,10000)

	return true
}


// 启动redis
func (this *LoginServer) StartRedis() bool {
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

	log.Info("连接Redis成功")
	return true
}


// 启动网络
func (this* LoginServer) StartNetWork() bool {
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
func (this *LoginServer) OnStart() {
	log.Info("开始执行OnStart")
	this.runtimestamp = util.CURTIMEMS()
	log.Info("结束执行OnStart")
}

// 程序退出最后清理
func (this *LoginServer) OnStop() {
}

//  退出
func (this* LoginServer) Quit() {
	if this.net != nil {
		this.asynev.Shutdown()
		this.net.Shutdown()
	}
}


// 主循环
func (this* LoginServer) Run() {

	// TODO:每帧处理1000条
	now := util.CURTIMEMS()
	lastrun := now - this.runtimestamp
	this.net.Dispatch(network.KFrameDispatchNum * 2)
	tm_dispath := util.CURTIMEMS()

	// 测试日志
	doEventStatistics(this)
	this.ticker1ms.Run(now)
	this.gatemgr.Tick(now)
	tm_tick := util.CURTIMEMS()


	// 每帧统计耗时
	delay := tm_tick - now
	if lastrun + delay > 20 {
		log.Warn("统计帧耗时 lastrun[%d] total[%d] dispatch[%d] tick[%d]", lastrun, delay, tm_dispath-now, tm_tick-tm_dispath)
	}

	//
	this.runtimestamp = util.CURTIMEMS()
}

func (this *LoginServer) Handler10msTick(now int64) {
	this.TickAuthenAccount(now)
	this.asynev.Dispatch()
}

// 添加正在登陆的账户
func (this *LoginServer) AddAuthenAccount(acc string, session network.IBaseNetSession)	{
	client := &ClientAccount{session:session, account:acc, tm_login:util.CURTIMEMS()}
	this.login_now[acc] = client
}

// 删除正在登陆的账户
func (this *LoginServer) DelAuthenAccount(sid int)	{
	for k, v := range this.login_now {
		if v.session.Id() != sid { continue }
		delete(this.login_now, k)
		return
	}
}

// 查找正在登陆的账户
func (this *LoginServer) FindAuthenAccount(acc string) bool {
	_, ok := this.login_now[acc];
	return ok
}

// 检查超时，客户端session长时间不断开会话或服务器没有收到断开
func (this *LoginServer) TickAuthenAccount(now int64)	{
	var timeout int64 = 10000
	for k, v := range this.login_now {
		if now > v.tm_login + timeout {		// 超过30秒
			log.Error("账户[%s] sid[%d] 长时间[%dms]与loginserver未断开，服务器主动断开", v.account, v.session.Id(), timeout)
			v.session.Close()
			delete(this.login_now, k)
		}
	}
}

// 插入新异步事件
func (this *LoginServer) AsynEventInsert(event eventque.IEvent) {
	this.asynev.Push(event)
}

// 生成唯一userid
func GenerateUserId() (userid uint64, errcode string ) {
	key := "genuserid"
	id, err := Redis().Incr(key).Result()
	var idstart uint64 = 1000000
	if err != nil {
		log.Error("生成userid redis报错, err: %s", err)
		return 0, "redis不可用"
	}

	return idstart + uint64(id), ""
}


