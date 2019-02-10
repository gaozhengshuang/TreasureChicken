package main
import (
	"fmt"
	"time"
	"runtime/debug"
	"strings"
	"os"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/net"
	"gitee.com/jntse/gotoolkit/util"
	"gitee.com/jntse/minehero/pbmsg"
	"gitee.com/jntse/minehero/server/tbl"
	"gitee.com/jntse/minehero/server/tbl/excel"
	"gitee.com/jntse/minehero/server/def"
	pb "github.com/gogo/protobuf/proto"
	"github.com/go-redis/redis"
)

func SignalInt(signal os.Signal)    {
	log.Info("SignalInt");
	g_KeyBordInput.Insert("quit_graceful")
}

func SignalTerm(signal os.Signal)   {
	log.Info("SignalTerm");
	g_KeyBordInput.Insert("quit_graceful")
}

func SignalHup(signal os.Signal)    {
	log.Info("SignalHup");
	g_KeyBordInput.Insert("quit_graceful")
}

func SignalCoreDump(signal os.Signal)   {
	log.Info("Signal[%d] Received", signal);
	g_KeyBordInput.Insert("quit_graceful")
}


func init() {
	fmt.Println("roomserver.init()")
}

type RoomServer struct {
	netconf			*network.NetConf
	net				*network.NetWork
	matchsvr		network.IBaseNetSession
	hredis			*redis.Client
	gatemgr			GateManager
	roommgr			RoomManager
	usermgr			UserManager
	//sessions		map[int]network.IBaseNetSession     // 及时删除，没有任何地方引用golang才会GC
	msghandlers		[]network.IBaseMsgHandler
	clienthandler	*ClientMsgHandler
	tblloader		*tbl.TblLoader
	//countmgr		  CountManager
	rcounter		util.RedisCounter
	ticker1m		*util.GameTicker
	ticker5s		*util.GameTicker
	ticker100ms		*util.GameTicker
	runtimestamp 	int64
	quit_graceful 	bool
	noticerepeat 	[]*msg.RS2MS_MsgNotice
	noticepause 	int64
	itembase		[]*table.ItemBaseDataDefine
	namebase		[]*table.TNameDefine
}

var g_RoomServer *RoomServer = nil
func RoomServerIns() *RoomServer {
	if g_RoomServer == nil { g_RoomServer = &RoomServer{} }
	return g_RoomServer
}

func RoomSvr() *RoomServer {
	return g_RoomServer
}

func Match() network.IBaseNetSession {
	return RoomSvr().matchsvr
}

func GateMgr() *GateManager {
	return &RoomSvr().gatemgr
}

func RoomMgr() *RoomManager {
	return &RoomSvr().roommgr
}

func UserMgr() *UserManager {
	return &RoomSvr().usermgr
}

func Redis() *redis.Client {
	return RoomSvr().hredis
}

func CMHandler() *ClientMsgHandler {
	return RoomSvr().clienthandler
}

//func CountMgr() *CountManager {
//	return &RoomSvr().countmgr
//}

func RCounter() *util.RedisCounter {
	return &RoomSvr().rcounter
}

func (this *RoomServer) DoInputCmd(cmd string) {
	switch cmd {
	case "quit_graceful":
		this.QuitGraceful()
	case "gates":
		log.Info("show gates list")
	case "free":
		debug.FreeOSMemory()        // 谨慎使用
	case "reload":
		this.Reload()
	}
}

func (this *RoomServer) OnClose(session network.IBaseNetSession) {
	sid := session.Id()
	//delete(this.sessions, sid)
	subname := strings.Split(session.Name(), "_")
	switch subname[0] {
	case "GateConnector":
		log.Info("和GateServer连接断开 sid[%d]", sid)
		this.gatemgr.OnClose(session)
		this.roommgr.OnGateClose(sid)
	case "MatchConnector":
		this.matchsvr = nil
		log.Info("和MatchServer连接断开 sid[%d]", sid)
	default:
		log.Error("OnClose error not regist session:%+v", sid)
	}
}

func (this *RoomServer) OnConnect(session network.IBaseNetSession) {
	//this.sessions[session.Id()] = session
	//log.Trace("OnConnect session:%+v", session)
	subname := strings.Split(session.Name(), "_")
	switch subname[0] {
	case "GateConnector":
		this.RegistToGateServer(session)
		break
	case "MatchConnector":
		this.matchsvr = session
		this.RegistToMatchServer()
		break
	default:
		log.Error("OnConnect error not regist session:%+v", session)
		session.Close()
	}
}

func (this *RoomServer) SendMsg(id int, msg pb.Message) bool {
	//session, ok := this.sessions[id]
	//if ok == true {
	//	return session.SendCmd(msg)
	//}
	//return false
	return this.net.SendMsg(id, msg)
}


func (this *RoomServer) GetSession(id int) network.IBaseNetSession {
	//session, _ := this.sessions[id]
	//return session
	return this.net.FindSession(id)
}

// 
func (this *RoomServer) Init(fileconf string) bool {

	//加载服务器配置
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

	//
	//this.sessions = make(map[int]network.IBaseNetSession)
	this.gatemgr.Init()
	this.roommgr.Init()
	this.usermgr.Init()

	//this.countmgr.Init()
	this.ticker1m = util.NewGameTicker(60 * time.Second, this.Handler1mTick)
	this.ticker5s = util.NewGameTicker(05 * time.Second, this.Handler5sTick)
	this.ticker100ms = util.NewGameTicker(100 * time.Millisecond, this.Handler100msTick)

	this.ticker1m.Start()
	this.ticker5s.Start()
	this.ticker100ms.Start()

	//初始道具和名字slice
	this.itembase = make([]*table.ItemBaseDataDefine,0)
	for _, v := range tbl.ItemBase.ItemBaseDataById { if v.Type != 1 && v.Type != 10 && v.Type != 6 { this.itembase = append(this.itembase, v) } }
	this.namebase = make([]*table.TNameDefine,0)
	for _, v := range tbl.NameBase.TNameById { this.namebase = append(this.namebase, v) }

	//
	this.runtimestamp = 0
	return true
}

func (this *RoomServer) Handler1mTick(now int64) {
	//log.Trace("开始批量统计")
	this.rcounter.BatchSave(10)
}

func (this *RoomServer) Handler5sTick(now int64) {
	//this.TickCacheNotice(now)
}

func (this *RoomServer) Handler100msTick(now int64) {
	if this.quit_graceful && this.roommgr.Num() == 0 {
		g_KeyBordInput.Insert("quit")
	}
}

func (this *RoomServer) InitMsgHandler() {
	if this.tblloader == nil { panic("should init 'tblloader' first") }
	this.msghandlers = append(this.msghandlers, NewC2GWMsgHandler())
	this.msghandlers = append(this.msghandlers, NewMS2RSMsgHandler())
	this.clienthandler = NewClientMsgHandler()
}

// 启动redis
func (this *RoomServer) StartRedis() bool {
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
func (this *RoomServer) StartNetWork() bool {
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

// 启动完成
func (this *RoomServer) OnStart() {
	log.Info("开始执行OnStart")

	this.runtimestamp = util.CURTIMEMS()
	this.cleanRoom()	// 删除房间
	this.rcounter.Init(Redis())	// 计数器

	log.Info("结束执行OnStart")
}

// 程序退出最后清理
func (this *RoomServer) OnStop() {
	this.ticker1m.Stop()
	this.ticker5s.Stop()
	this.ticker100ms.Stop()
}

// 优雅退出
func (this *RoomServer) QuitGraceful() {
	this.quit_graceful = true
	this.roommgr.Shutdown()
	RCounter().Save()
	log.Info("服务器 QuitGraceful")
}       


//  强制退出
func (this *RoomServer) Quit() {
	log.Info("服务器 QuitForce")
	if this.net != nil {
		this.net.Shutdown()
	}       
}       


// 主循环
func (this *RoomServer) Run() {

	// TODO:每帧处理2000条
	now := util.CURTIMEMS()
	lastrun := now - this.runtimestamp
	this.net.Dispatch(network.KFrameDispatchNum * 2, 1000)
	tm_dispath := util.CURTIMEMS()

	// 测试日志
	doEventStatistics(this)

	//
	this.roommgr.Tick(now)
	this.usermgr.Tick(now)
	tm_roomticker := util.CURTIMEMS()

	//
	this.ticker1m.Run(now)
	this.ticker5s.Run(now)
	this.ticker100ms.Run(now)
	tm_svrticker := util.CURTIMEMS()

	// 每帧统计耗时
	delay := tm_svrticker - now
	if lastrun + delay > 20 {
		log.Warn("统计帧耗时 lastrun[%d] total[%d] dispatch[%d] roomticker[%d] svrticker[%d]",
			lastrun, delay, tm_dispath - now, tm_roomticker - tm_dispath, tm_svrticker - tm_roomticker)
	}

	//
	this.runtimestamp = util.CURTIMEMS()
}

func (this *RoomServer) Net() *network.NetWork {
	return this.net
}

func (this *RoomServer) Name() string {
	return this.netconf.Name
}

func (this *RoomServer) RegistToMatchServer() {
	if this.matchsvr == nil {
		return
	}

	send := &msg.RS2MS_ReqRegist{
		Account : pb.String("room_account_123"),
		Passwd : pb.String("room_passwd_123"),
		Name : pb.String(this.Name()),
	}
	this.matchsvr.SendCmd(send)
	log.Info("请求注册Room[%s]到Match",this.Name())
}

func (this *RoomServer) RegistToGateServer(session network.IBaseNetSession) {
	send := &msg.RS2GW_ReqRegist {
		Account : pb.String("room_account_123"),
		Passwd : pb.String("room_passwd_123"),
		Agentname : pb.String(this.Name()),
	}
	session.SendCmd(send)
	log.Info("请求注册Room[%s]到Gate[%s]",this.Name(), session.Name())
}

func RoomSizeKey() string {
	key := fmt.Sprintf("RS_%s_RoomSize", RoomSvr().Name())
	return key
}

func (this *RoomServer) cleanRoom() {
	key := RoomSizeKey()
	_, err := Redis().Del(key).Result()
	if err != nil {
		log.Error("def key:%s err: %s", key, err)
		return
	}
	log.Info("del key:%s ok", key)
}

// 通用公告
func (this *RoomServer) SendNotice(face string, ty msg.NoticeType, subtext ...string) {
	noticemsg := &msg.GW2C_MsgNotice{Userid:pb.Uint64(0), Name:pb.String(""), Face:pb.String(face), Type:pb.Int32(int32(ty))}
	noticemsg.Text = pb.String(strings.Join(subtext, ""))
	send := &msg.RS2MS_MsgNotice{ Notice: noticemsg}
	Match().SendCmd(send)

	this.noticepause = util.CURTIMEMS() + 10000	// 临时暂停重复公告
	this.CacheNotice(send)
}

func (this *RoomServer) SendNoticeByMsg(notice *msg.RS2MS_MsgNotice) {
	Match().SendCmd(notice)
}

// 重新加载配置
func (this *RoomServer) Reload() {
	if this.tblloader != nil { this.tblloader.Reload() }
}

func (this *RoomServer) CacheNotice(notice *msg.RS2MS_MsgNotice) {
	this.noticerepeat = append(this.noticerepeat, notice)
}

// 随机重复公告
func (this *RoomServer) TickCacheNotice(now int64) {
	if now < this.noticepause || Match() == nil { 
		return 
	}

	amount := int32(len(this.noticerepeat))
	if amount < 100 || util.SelectPercent(50) == true {
		// 头像
		noticemsg := &msg.GW2C_MsgNotice{Userid:pb.Uint64(0), Name:pb.String(""), Face:pb.String(""), Type:pb.Int32(int32(msg.NoticeType_Suspension))}
		imageindex := util.RandBetween(0, 1200)
		faceurl := fmt.Sprintf("http://jump.cdn.giantfun.cn/cdn/jumphead/tx (%d).jpg",imageindex)

		itemname, username := "钻石", this.GetRandNickName()
		if util.SelectPercent(50) == true { itemname = this.GetRandItemName() }
		if itemname == "" || username == "" { return }
		subtext := []string	{
			def.MakeNoticeText("恭喜","#ffffff", 26), def.MakeNoticeText(username,"#ffffff", 26),
			def.MakeNoticeText("获得","#fffc00", 26), def.MakeNoticeText(itemname,"#ffffff", 26),
		}

		noticemsg.Face, noticemsg.Text = pb.String(faceurl), pb.String(strings.Join(subtext, ""))
		send := &msg.RS2MS_MsgNotice{ Notice:noticemsg }
		this.SendNoticeByMsg(send)
		return
	}

	randnotice := util.RandBetween(0, amount-1)
	if randnotice >= 0 && randnotice < amount {
		this.SendNoticeByMsg(this.noticerepeat[randnotice])
	}
}

func (this *RoomServer) GetRandItemName() string {
	lenlist := int32(len(this.itembase))
	if lenlist <= 0 {
		return ""
	}

	rnd := util.RandBetween(0, lenlist-1)
	if rnd >= 0 && rnd < lenlist {
		return this.itembase[rnd].Name
	}

	return ""
}

func (this *RoomServer) GetRandNickName() string {
	lenlist := int32(len(this.namebase))
	if lenlist <= 0 {
		return ""
	}

	rnd := util.RandBetween(0, lenlist-1)
	if rnd >= 0 && rnd < lenlist {
		return this.namebase[rnd].Name
	}

	return ""
}

