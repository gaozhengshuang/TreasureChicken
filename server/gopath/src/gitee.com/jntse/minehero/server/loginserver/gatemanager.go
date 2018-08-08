package main
import (
	"sort"
	"fmt"
	"math"
	"time"
	"gitee.com/jntse/gotoolkit/net"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/util"
	"gitee.com/jntse/gotoolkit/eventqueue"
	pb "github.com/golang/protobuf/proto"
	"gitee.com/jntse/minehero/server/def"
)

// --------------------------------------------------------------------------
/// @brief 按人数排序GateAgent
// --------------------------------------------------------------------------
type SliceGate []*GateAgent
func (s SliceGate) Len() int {
	return len(s)
}
func (s SliceGate) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s SliceGate) Less(i, j int) bool {
	return s[i].usernum < s[j].usernum
}

// --------------------------------------------------------------------------
/// @brief GateServer
// --------------------------------------------------------------------------
type GateAgent struct {
	//sid	int
	session	network.IBaseNetSession
	ip 		string
	port 	int
	usernum	int64
	host	string
	asynev  eventque.AsynEventQueue
	ticker10ms  *util.GameTicker
}

func NewGateAgent(s network.IBaseNetSession, ip string ,port int) *GateAgent {
	gate := &GateAgent{session:s, ip:ip, port:port, usernum:0}
	gate.host = fmt.Sprintf("%s:%d", ip, port)
	return gate
}

func (this *GateAgent) Init() {
	this.ticker10ms = util.NewGameTicker(10 * time.Millisecond, this.Handler10msTick)
	this.ticker10ms.Start()
	this.asynev.Start(int64(this.Id()), 1000)
}

func (this *GateAgent) Id() int {
	return this.session.Id()
}

func (this *GateAgent) Name() string {
	return this.session.Name()
}

//func (this *GateAgent) AddUserNum(n int) {
//	this.usernum += n
//	if this.usernum <= 0 { this.usernum = 0 }
//}

func (this *GateAgent) SendMsg(msg pb.Message) bool {
	return this.session.SendCmd(msg)
}

func (this *GateAgent) Host() string {
	return this.host
}

func (this *GateAgent) Ip() string {
	return this.ip
}

func (this *GateAgent) Port() int {
	return this.port
}

func (this *GateAgent) OnEnd() {
	this.asynev.Shutdown()
}

func (this *GateAgent) Tick(now int64) {
	this.ticker10ms.Run(now)
}

func (this *GateAgent) Handler10msTick(now int64) {
	this.asynev.Dispatch()
}

func (this *GateAgent) DoCalcAccountAmount(argu []interface{}) []interface{} {
	key := fmt.Sprintf("%s_%s:%d", def.RedisKeyGateAccounts, this.ip, this.port)
	size, err := Redis().SCard(key).Result()
	if err != nil {
		log.Error("SCard[%s] 获取Gate账户失败 err: %s", key, err)
		this.usernum = math.MaxInt32
	}
	this.usernum = size
	return nil
}

func (this *GateAgent) AsynEventInsert(event eventque.IEvent) {
	this.asynev.Push(event)
}

func (this *GateAgent) AsynCalcAccountAmount() {
	if this.asynev.IsFull() { return }
	arglist := []interface{}{}
	event := eventque.NewCommonEvent(arglist, this.DoCalcAccountAmount, nil)
	this.AsynEventInsert(event)
}

// --------------------------------------------------------------------------
/// @brief 
// --------------------------------------------------------------------------
type GateManager struct {
	gates map[int]*GateAgent
	tm_lastcalculate int64
}

func (this *GateManager) Init() {
	this.gates = make(map[int]*GateAgent)
}

func (this *GateManager) Tick(now int64) {
	for _ ,v := range this.gates {
		v.Tick(now)
	}
}

func (this *GateManager) Num() int {
	return len(this.gates)
}

func (this *GateManager) AddGate(agent *GateAgent) {
	id := agent.Id()
	this.gates[id] = agent
}

func (this* GateManager) DelGate(id int) {
	//log.Info("反注册网关 gate=%v", agent)
	delete(this.gates, id)
}

func (this* GateManager) FindGate(id int) *GateAgent {
	agent, ok := this.gates[id]
	if ok == false {
		return nil
	}
	return agent
}

// TODO: 获取一个低负荷Gate
func (this *GateManager) FindLowLoadGate() *GateAgent {
	if len(this.gates) == 0 {
		return nil
	}

	now := util.CURTIMEMS()
	sortGates := make(SliceGate, 0)
	for _, v := range this.gates {
		if now >= this.tm_lastcalculate + 10 {	// 间隔计算
			v.AsynCalcAccountAmount()		// 多Gate下，异步计数Gate在线人数效率提升明显
		}
		sortGates = append(sortGates, v)
	}

	if len(sortGates) == 0 {
		return nil
	}

	this.tm_lastcalculate = now
	sort.Sort(sortGates)
	return sortGates[0]
}

func (this *GateManager) FindGateByHost(host string) *GateAgent {
	for _ ,v := range this.gates {
		if v.Host() == host {
			return v
		}
	}
	return nil
}

func (this *GateManager) IsRegisted(ip string, port int) bool {
	host := fmt.Sprintf("%s:%d", ip, port)
	for _, v := range this.gates {
		if v.Host() == host {
			return true
		}
	}
	return false
}

func (this *GateManager) IsRegistedByName(name string, ip string, port int) bool {
	host := fmt.Sprintf("%s:%d", ip, port)
	for _, v := range this.gates {
		if v.Host() == host || v.Name() == name { 
			return true 
		}
	}
	return false
}

func (this *GateManager) OnClose(sid int) {
	agent := this.FindGate(sid)
	if agent == nil { return }
	agent.OnEnd()
	this.DelGate(sid)
	log.Info("网关离线 id=%d [%v] 当前总数:%d", sid, agent.Host(), this.Num())
}

func (this *GateManager) AddNew(session network.IBaseNetSession, ip string ,port int) {
	agent := NewGateAgent(session, ip, port)
	agent.Init()
	this.AddGate(agent)
	log.Info("注册网关 id=%d [%s:%d] 当前总数:%d", agent.Id(), agent.ip, agent.port, this.Num())
}


