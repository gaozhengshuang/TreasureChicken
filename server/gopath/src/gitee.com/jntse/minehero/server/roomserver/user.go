package main
import (
	"gitee.com/jntse/minehero/pbmsg"
	"gitee.com/jntse/minehero/server/tbl"
	_"gitee.com/jntse/minehero/server/def"
	"gitee.com/jntse/gotoolkit/net"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/util"
	"gitee.com/jntse/gotoolkit/eventqueue"
	pb "github.com/gogo/protobuf/proto"
	"fmt"
	"strings"
	"strconv"
	"github.com/go-redis/redis"
	"time"
)

// --------------------------------------------------------------------------
/// @brief 玩家信息
// --------------------------------------------------------------------------
type RoomUser struct {
	roomkind  	int32
	roomid    	int64
	sid_gate  	int
	bin 		*msg.Serialize
	bag 		UserBag
	task       	UserTask
	token		string
	coins		uint32
	ticker1s  	*util.GameTicker
	ticker10ms  *util.GameTicker
	asynev      eventque.AsynEventQueue // 异步事件处理
	invitationcode string
	luckydraw	[]*msg.LuckyDrawItem
}

func NewRoomUser(rid int64, b *msg.Serialize, gate network.IBaseNetSession, roomkind int32) *RoomUser {
	user := &RoomUser{roomid:rid, bin:b, sid_gate:gate.Id(), roomkind:roomkind}
	user.ticker1s = util.NewGameTicker(1 * time.Second, user.Handler1sTick)
	user.ticker10ms = util.NewGameTicker(10 * time.Millisecond, user.Handler10msTick)
	user.ticker1s.Start()
	user.ticker10ms.Start()
	user.bag.Init(user)
	user.task.Init(user)
	user.bag.LoadBin(b)
	user.task.LoadBin(b)
	user.asynev.Start(int64(user.Id()), 10)
	for _, v := range user.UserBase().Luckydraw.Drawlist { 
		user.luckydraw = append(user.luckydraw, v) 
	}
	return user
}

func (this *RoomUser) Entity() *msg.EntityBase {
	return this.bin.GetEntity()
}

func (this *RoomUser) UserBase() *msg.UserBase {
	return this.bin.GetBase()
}

func (this *RoomUser) Id() uint64 {
	return this.Entity().GetId()
}

func (this *RoomUser) Name() string {
	return this.Entity().GetName()
}

func (this *RoomUser) Account() string {
	return this.Entity().GetAccount()
}

func (this *RoomUser) Face() string {
	return this.Entity().GetFace()
}

func (this *RoomUser) RoomId() int64 {
	return this.roomid
}

func (this *RoomUser) WechatOpenId() string {
	userbase := this.UserBase()
	return userbase.GetWechat().GetOpenid()
}

// 邀请人邀请码
func (this *RoomUser) InvitationCode() string {
	userbase := this.UserBase()
	return userbase.GetInvitationcode()
}

// 邀请人
func (this *RoomUser) Inviter() uint64 {
	if code := this.InvitationCode(); len(code) > 2 {
		inviter , _ := strconv.ParseUint(code[2:], 10, 64)
		return inviter
	}
	return 0
}


func (this *RoomUser) GetMoneyCost() int64 {
	userbase := this.UserBase()
	return userbase.GetScounter().GetMoneyCost()
}

func (this *RoomUser) GetMoneyCostReset() int64 {
	userbase := this.UserBase()
	return userbase.GetScounter().GetMoneyCostReset()
}

func (this *RoomUser) SetMoneyCost(cost int64) {
	userbase := this.UserBase()
	userbase.GetScounter().MoneyCost = pb.Int64(cost)
}

func (this *RoomUser) SetMoneyCostReset(reset int64) {
	userbase := this.UserBase()
	userbase.GetScounter().MoneyCostReset = pb.Int64(reset)
}
//
//func (this *RoomUser) DiamondRoomStep() int64 {
//	userbase := this.UserBase()
//	return userbase.GetScounter().GetDiamondRoomStep()
//}
//
//func (this *RoomUser) SetDiamondRoomCost(d int64) {
//	userbase := this.UserBase()
//	userbase.GetScounter().DiamondRoomCost = pb.Int64(d)
//}
//
//func (this *RoomUser) SetDiamondRoomIncome(d int64) {
//	userbase := this.UserBase()
//	userbase.GetScounter().DiamondRoomIncome = pb.Int64(d)
//}
//
//func (this *RoomUser) SetDiamondRoomStep(d int64) {
//	userbase := this.UserBase()
//	userbase.GetScounter().DiamondRoomStep = pb.Int64(d)
//}

func (this *RoomUser) Token() string {
	return this.token
}

func (this *RoomUser) UpdateToken(t string) {
	this.token = t
}

func (this *RoomUser) Level() uint32 {
	return this.UserBase().GetLevel()
}

func (this *RoomUser) AddLevel(num uint32) {
	this.UserBase().Level = pb.Uint32(this.Level() + num)
	//send := &msg.BT_LevelUp{Userid:pb.Uint64(this.Id()), Level:pb.Uint32(this.Level())}
	//this.SendMsg(send)
}

func (this *RoomUser) Exp() uint32 {
	return this.UserBase().GetExp()
}

func (this *RoomUser) SetExp(num uint32) {
	this.UserBase().Exp = pb.Uint32(num)
}

// 添加经验
func (this *RoomUser) AddExp(num uint32, reason string, syn bool) {
	old, exp := this.Level(), this.Exp() + num
	for {
		lvlbase, ok := tbl.LevelBasee.TLevelById[this.Level()]
		if ok == false {
			break
		}

		// 下一级需要经验
		if exp < uint32(lvlbase.ExpNums) || lvlbase.ExpNums == 0 {
			break
		}

		exp = exp - uint32(lvlbase.ExpNums)
		this.OnLevelUp()
	}
	this.SetExp(exp)
	if syn == true { this.SendBattleUser() }

	log.Info("玩家[%d] 添加经验[%d] 老等级[%d] 新等级[%d] 经验[%d] 原因[%s]", this.Id(), num, old, this.Level(), this.Exp(), reason)
}

// 升级
func (this *RoomUser) OnLevelUp() {
	this.AddLevel(1)
	
	//升级拿元宝
	lvlbase, ok := tbl.LevelBasee.TLevelById[this.Level()-1]
	if ok == true { this.AddYuanbao(uint32(lvlbase.Reward), "升级奖元宝") }

	// 临时
	//arglist := []interface{}{this.Account(), this.Token(), uint64(this.Id()), uint32(this.Level())}
	//event := eventque.NewCommonEvent(arglist, def.HttpRequestUserLevelArglist, nil)
	//this.AsynEventInsert(event)
}

// 打包二进制数据
func (this *RoomUser) PackBin() *msg.Serialize {
	bin := &msg.Serialize{}

	// 基础信息
	bin.Entity = pb.Clone(this.bin.GetEntity()).(*msg.EntityBase)

	// 玩家信息
	bin.Base = pb.Clone(this.bin.GetBase()).(*msg.UserBase)
	bin.Base.Luckydraw.Drawlist = make([]*msg.LuckyDrawItem,0)
	for _, v := range this.luckydraw {
		bin.Base.Luckydraw.Drawlist = append(bin.Base.Luckydraw.Drawlist, v)
	}

	// 背包
	this.bag.PackBin(bin)
	this.task.PackBin(bin)


	return bin
}

// 游戏结束，将数据回传Gate
func (this *RoomUser) OnEnd(now int64) {
	this.ticker1s.Stop()
	this.ticker10ms.Stop()
	this.asynev.Shutdown()
	UserMgr().DelUser(this)
	//this.bin = this.PackBin()
}

func (this *RoomUser) SendMsg(m pb.Message) bool {
	return RoomSvr().SendMsg(this.sid_gate , m)
}

// 转发消息到gate
func (this *RoomUser) SendClientMsg(m pb.Message) bool {
	name := pb.MessageName(m)
	if name == "" { 
		log.Fatal("SendClientMsg 获取proto名字失败[%s]", m)
		return false 
	}
	msgbuf, err := pb.Marshal(m)
	if err != nil { 
		log.Fatal("SendClientMsg 序列化proto失败[%s][%s]", name, err)
		return false 
	}

	send := &msg.RS2GW_MsgTransfer{ Uid:pb.Uint64(this.Id()), Name:pb.String(name), Buf:msgbuf }
	return this.SendMsg(send)
}

func (this *RoomUser) SidGate() int {
	return this.sid_gate
}

func (this *RoomUser) GetCoins() uint32 {
	return this.coins
}

func (this *RoomUser) UpdateCoins(amount uint32) {
	this.coins = amount
}

// 获取平台金币
//func (this *RoomUser) QueryPlatformCoins() {
//	event := NewQueryPlatformCoinsEvent(this, this.SyncPlatformCoins)
//	this.AsynEventInsert(event)
//}

// 同步平台金币
//func (this *RoomUser) SyncPlatformCoins() {
//	//
//	tvmid := this.Account()
//	errcode, coins, _ := def.HttpRequestFinanceQuery(this.Id(), this.Token(), tvmid)
//	if errcode != "" {
//		return
//	}
//	this.UpdateCoins(uint32(coins))
//	this.SendBattleUser()	// 同步玩家数据
//}

func (this *RoomUser) GetMoney() uint32 {
	return this.UserBase().GetMoney()
}

func (this *RoomUser) RemoveMoney(gold uint32, reason string, syn bool) bool {
	if this.GetMoney() > gold {
		userbase := this.UserBase()
		userbase.Money = pb.Uint32(this.GetMoney() - gold)
		if syn {
			send := &msg.GW2C_UpdateGold{Num:pb.Uint32(this.GetMoney())}
			this.SendClientMsg(send)
		}
		log.Info("玩家[%d] 扣除金币[%d] 剩余[%d] 原因[%s]", this.Id(), gold, this.GetMoney(), reason)

		RCounter().IncrByDate("item_remove", int32(msg.ItemId_Gold), int32(gold))
		return true
	}
	log.Info("玩家[%d] 扣除金币失败[%d] 原因[%s]", this.Id(), gold, reason)
	return false
}

func (this *RoomUser) AddMoney(gold uint32, reason string, syn bool) {
	userbase := this.UserBase()
	userbase.Money = pb.Uint32(this.GetMoney() + gold)
	if syn {
		send := &msg.GW2C_UpdateGold{Num:pb.Uint32(this.GetMoney())}
		this.SendClientMsg(send)
	}
	log.Info("玩家[%d] 添加金币[%d] 剩余[%d] 原因[%s]", this.Id(), gold, this.GetMoney(), reason)
}

func (this *RoomUser) SetMoney(gold uint32, reason string, syn bool) {
	userbase := this.UserBase()
	userbase.Money = pb.Uint32(gold)
	if syn {
		send := &msg.GW2C_UpdateGold{Num:pb.Uint32(this.GetMoney())}
		this.SendClientMsg(send)
	}
	log.Info("玩家[%d] 设置金币[%d] 剩余[%d] 原因[%s]", this.Id(), gold, this.GetMoney(), reason)
}


// 元宝
func (this *RoomUser) GetYuanbao() uint32 {
	return this.UserBase().GetYuanbao()
}

func (this *RoomUser) AddYuanbao(yuanbao uint32, reason string) {
	userbase := this.bin.GetBase()
	userbase.Yuanbao = pb.Uint32(userbase.GetYuanbao() + yuanbao)
	RCounter().IncrByDate("room_output", int32(this.roomkind), int32(yuanbao))
	//this.PlatformPushLootMoney(float32(yuanbao))
	log.Info("玩家[%d] 添加元宝[%d] 剩余[%d] 原因[%s]", this.Id(), yuanbao, userbase.GetYuanbao(), reason) 
}

func (this *RoomUser) RemoveYuanbao(yuanbao uint32, reason string) bool {
	if this.GetYuanbao() >= yuanbao {
		userbase := this.bin.GetBase()
		userbase.Yuanbao = pb.Uint32(this.GetYuanbao() - yuanbao)
		RCounter().IncrByDate("item_remove", int32(msg.ItemId_YuanBao), int32(yuanbao))
		RCounter().IncrByDate("room_income", int32(this.roomkind), int32(yuanbao))
		//this.PlatformPushConsumeMoney(float32(yuanbao))
		log.Info("玩家[%d] 扣除元宝[%d] 库存[%d] 原因[%s]", this.Id(), yuanbao, this.GetYuanbao(), reason)
		return true
	}
	log.Info("玩家[%d] 扣除元宝[%d]失败 库存[%d] 原因[%s]", this.Id(), yuanbao, this.GetYuanbao(), reason)
	return false
}

func (this *RoomUser) GetCoupon() uint32 {
	return this.UserBase().GetCoupon()
}

// 移除金卷
func (this *RoomUser) RemoveCoupon(num uint32, reason string) bool {
	userbase := this.bin.GetBase()
	if ( userbase.GetCoupon() >= num ) {
		userbase.Coupon = pb.Uint32(userbase.GetCoupon() - num)
		log.Info("玩家[%d] 扣除金卷[%d] 剩余[%d] 原因[%s]", this.Id(), num, userbase.GetCoupon(), reason)
		RCounter().IncrByDate("item_remove", int32(msg.ItemId_Coupon), int32(num))
		return true
	}
	log.Info("玩家[%d] 扣除金卷[%d]失败 剩余[%d] 原因[%s]", this.Id(), num, userbase.GetCoupon(), reason)
	return false
}

// 添加金卷
func (this *RoomUser) AddCoupon(num uint32, reason string) {
	userbase := this.bin.GetBase()
	userbase.Coupon = pb.Uint32(userbase.GetCoupon() + num)
	log.Info("玩家[%d] 添加金卷[%d] 剩余[%d] 原因[%s]", this.Id(), num, userbase.GetCoupon(), reason)
}

// 添加道具
func (this *RoomUser) AddItem(item uint32, num uint32, reason string) {

    if item == uint32(msg.ItemId_YuanBao) {
        this.AddYuanbao(num, reason)
    }else if item == uint32(msg.ItemId_Gold) {
        this.AddMoney(num, reason, true)
    }else if item == uint32(msg.ItemId_Coupon) {
		this.AddCoupon(num, reason)
	}else if item == uint32(msg.ItemId_FreeStep) {
		this.AddFreeStep(num, reason)
	}else{
		this.bag.AddItem(item, num, reason)
	}
	RCounter().IncrByDate("item_add", int32(item), int32(num))

}

// 扣除道具
func (this *RoomUser) RemoveItem(item uint32, num uint32, reason string) bool{
	return this.bag.RemoveItem(item, num, reason)
}

func (this *RoomUser) GetFreeStep() int32 {
	userbase := this.UserBase()
	return userbase.GetScounter().GetFreestep()
}

func (this *RoomUser) RemoveFreeStep(num int32, reason string) bool {
	freestep := this.GetFreeStep()
	if freestep >= num {
		this.bin.GetBase().GetScounter().Freestep = pb.Int32(freestep - num)
		//log.Trace("玩家[%d] 添加免费步数[%d] 库存[%d] 原因[%s]", this.Id(), num, this.GetFreeStep(), reason)
		return true
	}
	return false
}

func (this *RoomUser) AddFreeStep(num uint32, reason string) {
	this.bin.GetBase().GetScounter().Freestep = pb.Int32(this.GetFreeStep() + int32(num))
	//log.Trace("玩家[%d] 添加免费步数[%d] 库存[%d] 原因[%s]", this.Id(), num, this.GetFreeStep(), reason)
}

func (this *RoomUser) SendBattleUser() {
	send := &msg.BT_SendBattleUser	{ 
		Ownerid:pb.Uint64(this.Id()),
		Yuanbao:pb.Uint32(this.GetYuanbao()),
		Level:pb.Uint32(this.Level()),
		Freestep:pb.Int32(this.GetFreeStep()),
		Gold:pb.Uint32(this.GetMoney()),
	}
	this.SendClientMsg(send)
}


func (this *RoomUser) SendNotify(text string) {
	send := &msg.GW2C_MsgNotify{Userid:pb.Uint64(this.Id()), Text:pb.String(text)}
	this.SendMsg(send)
}

func (this *RoomUser) Tick(now int64) {
	if this.ticker10ms.Run(now) {
		this.ticker1s.Run(now)
	}
}

func (this *RoomUser) Handler10msTick(now int64) {
	this.asynev.Dispatch()
}

func (this *RoomUser) Handler1sTick(now int64) {

	// 不用我们的充值 2018年 05月 17日 星期四 19:34:03 CST
	//this.CheckRechargeOrders()

	// 充值，异步事件队列
	//event := NewRechargeCheckEvent(this, this.HaveRechargeOrders, this.CheckRechargeOrders)
	//this.AsynEventInsert(event)
}

// 处理充值订单
func (this *RoomUser) CheckRechargeOrders() {
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
	this.SendBattleUser()
}

// 检查是否有充值订单
func (this *RoomUser) HaveRechargeOrders() bool {
	//log.Info("HaveRechargeOrders")
	keyorder := fmt.Sprintf("%d_verified_recharge_orders", this.Id())
	amount, err := Redis().SCard(keyorder).Result()
	if err == redis.Nil || amount == 0 {
		return false
	} else if err != nil {
		log.Error("[充值] 从Redis SCard 订单失败 err:%s", err)
		return false
	}

	return true
}

// 插入新异步事件
func (this *RoomUser) AsynEventInsert(event eventque.IEvent) {
	this.asynev.Push(event)
}

// 推送资源消耗
//func (this *RoomUser) PlatformPushConsumeMoney(yuanbao float32) {
//	rmbcent := 100.0 * yuanbao / float32(tbl.Room.RmbToYuanbao)
//	arglist := []interface{}{this.Account(), this.Token(), uint64(this.Id()), uint32(rmbcent)}
//	event := eventque.NewCommonEvent(arglist, def.HttpRequestUserResourceConsumeArglist, nil)
//	this.AsynEventInsert(event)
//}
//
//// 推送资源获取
//func (this *RoomUser) PlatformPushLootMoney(yuanbao float32) {
//	rmbcent := 100.0 * yuanbao / float32(tbl.Room.RmbToYuanbao)
//	arglist := []interface{}{this.Account(), this.Token(), uint64(this.Id()), uint32(rmbcent)}
//	event := eventque.NewCommonEvent(arglist, def.HttpRequestUserResourceEarnArglist, nil)
//	this.AsynEventInsert(event)
//}

func (this *RoomUser) LuckyDraw() {
	// 检查消耗
	cost := uint32(tbl.Game.LuckDrawPrice)
	if this.GetMoney() < cost {
		this.SendNotify("金币不足")
		return
	}
	this.RemoveMoney(cost, "幸运抽奖", true)

	// 每周一重置
	curtime := util.CURTIME()
	if util.IsSameWeek(this.GetMoneyCostReset(), curtime) != false {
		this.SetMoneyCost(0)
		this.SetMoneyCostReset(util.CURTIME())
	}

	// 解析概率配置
	ParseProString := func (sliceweight* []util.WeightOdds, Pro []string) (bool) {
		for _ , strpro := range Pro {
			slicepro := strings.Split(strpro, "-")
			if len(slicepro) != 2 {
				log.Error("[%d %s] 抽奖异常，解析概率配置异常 strpro=%s", this.Id(), this.Name(), strpro)
				return false
			}
			id    , _ := strconv.ParseInt(slicepro[0], 10, 32)
			weight, _ := strconv.ParseInt(slicepro[1], 10, 32)
			*sliceweight = append(*sliceweight, util.WeightOdds{Weight:int32(weight), Uid:int64(id)})
		}
		return true
	}

	giftweight := make([]util.WeightOdds, 0)
	for _, v := range tbl.GiftProBase.TGiftProById {
		if this.GetMoneyCost() <= int64(v.Limitmin) {
			if ParseProString(&giftweight, v.Pro) == false { return }
			break
		}
	}

	//giftweight := make([]util.WeightOdds, 0)
	//for k ,v := range tbl.TBallGiftbase.TBallGiftById {
	//	giftweight = append(giftweight, util.WeightOdds{Weight:v.Pro, Uid:int64(k)})
	//}
	index := util.SelectByWeightOdds(giftweight)
	if index < 0 || index >= int32(len(giftweight)) {
		log.Error("[%d %s] 抽奖异常，无法获取抽奖id", this.Id(), this.Name())
		return
	}

	uid := giftweight[index].Uid
	gift, find := tbl.TBallGiftbase.TBallGiftById[uint32(uid)]
	if find == false {
		log.Error("[%d %s] 无效的奖励id[%d]", this.Id(), this.Name(), uid)
		return
	}

	this.AddItem(uint32(gift.ItemId), uint32(gift.Num), "幸运抽奖")
	drawitem := &msg.LuckyDrawItem{Time:pb.Int64(curtime), Item:pb.Int32(gift.ItemId), Num:pb.Int32(gift.Num), Worth:pb.Int32(gift.Cost)}
	this.luckydraw = append(this.luckydraw, drawitem)

	// feedback
	send := &msg.GW2C_LuckyDrawHit{Id:pb.Int32(int32(uid))}
	this.SendClientMsg(send)
}

