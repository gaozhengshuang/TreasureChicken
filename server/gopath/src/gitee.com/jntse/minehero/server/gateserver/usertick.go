package main
import (
	"fmt"
	"time"
	"strings"
	"strconv"
	"gitee.com/jntse/gotoolkit/util"
	"gitee.com/jntse/gotoolkit/log"
	"github.com/go-redis/redis"
	"gitee.com/jntse/minehero/server/tbl"
)

// --------------------------------------------------------------------------
/// @brief 
// --------------------------------------------------------------------------
type UserTicker struct {
	ticker1s	*util.GameTicker
	ticker5s	*util.GameTicker
	ticker1m	*util.GameTicker
	ticker10ms 	*util.GameTicker
	ticker100ms *util.GameTicker
}

func (t *UserTicker) Init(handler10ms, handler100ms, handler1s, handler5s, handler1m util.TickerCallbackHandle) {
	t.ticker10ms = util.NewGameTicker(10 * time.Millisecond, handler10ms)
	t.ticker100ms = util.NewGameTicker(100 * time.Millisecond, handler100ms)
	t.ticker1s = util.NewGameTicker(1 * time.Second, handler1s)
	t.ticker5s = util.NewGameTicker(5 * time.Second, handler5s)
	t.ticker1m = util.NewGameTicker(1 * time.Minute, handler1m)
}


func (t *UserTicker) Start() {
	t.ticker10ms.Start()
	t.ticker100ms.Start()
	t.ticker1s.Start()
	t.ticker5s.Start()
	t.ticker1m.Start()
}

func (t *UserTicker) Stop() {
	t.ticker10ms.Stop()
	t.ticker100ms.Stop()
	t.ticker1s.Stop()
	t.ticker5s.Stop()
	t.ticker1m.Stop()
}

func (this *UserTicker) Run(now int64) {
	// 嵌套节省
	this.ticker10ms.Run(now)
	this.ticker100ms.Run(now)
	if this.ticker1s.Run(now) {
		if this.ticker5s.Run(now) { this.ticker1m.Run(now) }
	}
}

func (this *GateUser) OnTicker10ms(now int64) {
	this.asynev.Dispatch()
}

func (this *GateUser) OnTicker100ms(now int64) {
	this.CheckOffline(now)
	this.CheckDisconnectTimeOut(now)
}

func (this *GateUser) OnTicker1s(now int64) {
}

func (this *GateUser) OnTicker5s(now int64) {
	//this.CheckRechargeOrders()		// 不用我们的充值 2018年 05月 17日 星期四 19:34:03 CST
}

func (this *GateUser) OnTicker1m(now int64) {
}

func (this *GateUser) Tick(now int64) {
	this.tickers.Run(now)
}

// 处理充值订单
func (this *GateUser) CheckRechargeOrders() {
	if this.IsInRoom () == true {
		return
	}
	keyorder := fmt.Sprintf("%d_verified_recharge_orders", this.Id())
	order_amount, err := Redis().SPop(keyorder).Result()
	if err == redis.Nil {
		return
	} else if err != nil {
		log.Error("[充值] 从Redis Spop 验证订单失败 err:%s", err)
		return
	}

	// 字符串格式 recharge_order_userid_timestamp_amount_number
	orderparts := strings.Split(order_amount, "_")
	if len(orderparts) != 5 {
		log.Error("[充值] amount订单格式解析失败 [%s]", order_amount)
		return
	}
	
	amount, perr := strconv.ParseInt(orderparts[4], 10, 32)
	if perr != nil {
		log.Error("[充值] amount订单格式解析失败 [%s]", order_amount)
		return
	}

	this.AddYuanbao(uint32(amount), "充值获得")
}

// 心跳,毫秒
func (this *GateUser) SetHeartBeat(now int64) {
	tm_last := this.tm_heartbeat
	tm_delay := now - tm_last
	this.tm_heartbeat = now
	if tm_delay < 1000 {
		//log.Warn("玩家[%s %d] 心跳太过频繁[%d ms]", this.Name(), this.Id(), tm_delay)
	}else if tm_delay > 15000 {
		log.Warn("玩家[%s %d] 心跳延迟了[%d ms]，网络不好?", this.Name(), this.Id(), tm_delay)
	}
}

// 检查心跳，毫秒
func (this *GateUser) CheckOffline(now int64) {
	if this.online == false {
		return
	}
	tm_delay := now - this.tm_heartbeat
	if tm_delay > tbl.Global.Hearbeat.Timeout {
		log.Warn("玩家[%s %d] 心跳延迟达到[%d ms]，清理离线Session", this.Name(), this.Id(), tm_delay)
		this.KickOut("心跳超时")
	}
}

