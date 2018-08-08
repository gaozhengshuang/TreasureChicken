package main
import (
	"gitee.com/jntse/minehero/pbmsg"
	"gitee.com/jntse/gotoolkit/eventqueue"
	"gitee.com/jntse/gotoolkit/util"
	"gitee.com/jntse/gotoolkit/log"
)
// --------------------------------------------------------------------------
/// @brief 玩家异步事件
// --------------------------------------------------------------------------


// 获取充值页面
type RechargeEventHandle func(tvmid string, token string, user *GateUser, amount uint32)
type UserRechargeEvent struct {
	user *GateUser
	tvmid string
	token string
	amount uint32
	handler RechargeEventHandle
}

func NewUserRechargeEvent(user *GateUser, tvmid, token string, amount uint32, handler RechargeEventHandle) *UserRechargeEvent {
	return &UserRechargeEvent{tvmid:tvmid,token:token,user:user,amount:amount,handler:handler}
}

func (this *UserRechargeEvent) Process(ch_fback chan eventque.IEvent) {
	this.handler(this.tvmid, this.token, this.user, this.amount)
}

func (this *UserRechargeEvent) Feedback() {
}


// 玩家存盘
type UserSaveEventHandle func()
type UserSaveEvent struct {
	handler UserSaveEventHandle
}

func NewUserSaveEvent(handler UserSaveEventHandle) *UserSaveEvent {
	return &UserSaveEvent{handler:handler}
}

func (this *UserSaveEvent) Process(ch_fback chan eventque.IEvent) {
	this.handler()
}

func (this *UserSaveEvent) Feedback() {
}


// 玩家发货
type DeliveryGoodsHandle func(list []*msg.DeliveryGoods, token string, phone string)
type DeliveryGoodsEvent struct {
	list []*msg.DeliveryGoods
	token string
	handler DeliveryGoodsHandle
	phone string
}

func NewDeliveryGoodsEvent(list []*msg.DeliveryGoods, token string, phone string, handler DeliveryGoodsHandle) *DeliveryGoodsEvent {
	return &DeliveryGoodsEvent{list:list, token:token, phone:phone, handler:handler}
}

func (this *DeliveryGoodsEvent) Process(ch_fback chan eventque.IEvent) {
	tm1 := util.CURTIMEMS()
	this.handler(this.list, this.token, this.phone)
	log.Trace("[异步事件] DeliveryGoodsEvent 本次消耗 %dms", util.CURTIMEMS() - tm1)
}

func (this *DeliveryGoodsEvent) Feedback() {
}


// 整点给玩家免费步数
type GiveFreeStepHandle func(now int64, reason string)
type GiveFreeStepEvent struct {
	now int64
	reason string
	handler GiveFreeStepHandle
}

func NewGiveFreeStepEvent(now int64, reason string, handler GiveFreeStepHandle) *GiveFreeStepEvent { 
	return &GiveFreeStepEvent{now:now,reason:reason,handler:handler}
}

func (this *GiveFreeStepEvent) Process(ch_fback chan eventque.IEvent) {
	this.handler(this.now, this.reason)
}

func (this *GiveFreeStepEvent) Feedback() {
}

// 同步平台金币
type QueryPlatformCoinsEventHandle func()
type QueryPlatformCoinsFeedback func()
type QueryPlatformCoinsEvent struct {
	handler QueryPlatformCoinsEventHandle
	feedback QueryPlatformCoinsFeedback
	//coins int32
	//diamonds int32

	//tm_done int64
	//tm_start int64
	//tm_processed int64
}

func NewQueryPlatformCoinsEvent(handler QueryPlatformCoinsEventHandle) *QueryPlatformCoinsEvent {
	return &QueryPlatformCoinsEvent{handler:handler, feedback:nil}
}

func (this *QueryPlatformCoinsEvent) Process(ch_fback chan eventque.IEvent) {
	//tm1 := util.CURTIMEMS()
	this.handler()
	//log.Trace("[异步事件] QueryPlatformCoinsEvent 本次消耗 %dms", util.CURTIMEMS() - tm1)
}

func (this *QueryPlatformCoinsEvent) Feedback() {
}

// 增加平台金币
type AddPlatformCoinsEventHandle func(amount int32, desc string) bool
type AddPlatformCoinsFeedback func()
type AddPlatformCoinsEvent struct {
	handler AddPlatformCoinsEventHandle
	feedback AddPlatformCoinsFeedback
	coins int32
	desc string
	//diamonds int32

	//tm_done int64
	//tm_start int64
	//tm_processed int64
}

func NewAddPlatformCoinsEvent(amount int32, desc string, handler AddPlatformCoinsEventHandle) *AddPlatformCoinsEvent {
	return &AddPlatformCoinsEvent{coins:amount, desc:desc, handler:handler, feedback:nil}
}

func (this *AddPlatformCoinsEvent) Process(ch_fback chan eventque.IEvent) {
	//tm1 := util.CURTIMEMS()
	this.handler(this.coins, this.desc)
	//log.Trace("[异步事件] QueryPlatformCoinsEvent 本次消耗 %dms", util.CURTIMEMS() - tm1)
}

func (this *AddPlatformCoinsEvent) Feedback() {
}


// 扣除平台金币
type RemovePlatformCoinsEventHandle func(amount int32, desc string) bool
type RemovePlatformCoinsFeedback func(removeok bool, amount int32, bid int32)
type RemovePlatformCoinsEvent struct {
	amount int32
	boxid int32
	desc string
	handler RemovePlatformCoinsEventHandle

	removeok bool
	feedback RemovePlatformCoinsFeedback
}

func NewRemovePlatformCoinsEvent(amount int32, bid int32, desc string, handler RemovePlatformCoinsEventHandle, feedback RemovePlatformCoinsFeedback) *RemovePlatformCoinsEvent {
	return &RemovePlatformCoinsEvent{amount, bid, desc, handler, false, feedback}
}

func (this *RemovePlatformCoinsEvent) Process(ch_fback chan eventque.IEvent) {
	//tm1 := util.CURTIMEMS()
	this.removeok = this.handler(this.amount, this.desc)
	//log.Trace("[异步事件] 玩家[%d] QueryPlatformCoinsEvent 本次消耗 %dms", this.room.Id(), this.room.ownerid, util.CURTIMEMS() - tm1)
	ch_fback <- this
}

func (this *RemovePlatformCoinsEvent) Feedback() {
	this.feedback(this.removeok, this.amount, this.boxid)
}

