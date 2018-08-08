package main
import (
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/util"
	_"gitee.com/jntse/gotoolkit/eventqueue"
	"gitee.com/jntse/gotoolkit/net"
	"gitee.com/jntse/minehero/server/tbl"
	"gitee.com/jntse/minehero/server/def"
	"gitee.com/jntse/minehero/pbmsg"
	pb "github.com/golang/protobuf/proto"
	"encoding/json"
	"net/http"
	"strconv"
	"crypto/md5"
	"fmt"
	"time"
	"strings"
)

// 添加道具
func (this *GateUser) AddItem(item uint32, num uint32, reason string) {

	if item == uint32(msg.ItemId_YuanBao) {
		this.AddYuanbao(num, reason)
	}else if item == uint32(msg.ItemId_Gold) {
		this.AddMoney(num, reason)
	}else if item == uint32(msg.ItemId_Coupon) {
		this.AddCoupon(num, reason)
	}else if item == uint32(msg.ItemId_FreeStep) {
		//this.AddFreeStep(int32(num), reason)
	}else {
		base, ok := tbl.ItemBase.ItemBaseDataById[item]
		if ok && this.bag.AddItem(item, num, reason) != nil {
			this.PlatformPushLootMoney(float32(base.RealPrice) * float32(num))
		}
	}
	//CountMgr().AddGet(item, num)
	RCounter().IncrByDate("item_add", uint32(item), num)
}

// 扣除道具
func (this *GateUser) RemoveItem(item uint32, num uint32, reason string) bool{
	base, ok := tbl.ItemBase.ItemBaseDataById[item]
	if ok && this.bag.RemoveItem(item, num, reason) {
		this.PlatformPushConsumeMoney(float32(base.RealPrice) * float32(num))
		return true
	}
	return false
}

// 金币
func (this *GateUser) GetMoney() uint32   { return this.money }
func (this *GateUser) AddMoney(gold uint32, reason string) {
	this.money = this.GetMoney() + gold
	log.Info("玩家[%d] 添加金币[%d] 库存[%d] 原因[%s]", this.Id(), gold, this.GetMoney(), reason)
}

// 添加元宝
func (this *GateUser) GetYuanbao() uint32 { return this.yuanbao }
func (this *GateUser) AddYuanbao(yuanbao uint32, reason string) {
	this.yuanbao = this.GetYuanbao() + yuanbao
	send := &msg.GW2C_UpdateYuanbao{Num:pb.Uint32(this.GetYuanbao())}
	this.SendMsg(send)
	RCounter().IncrByDate("item_add", uint32(msg.ItemId_YuanBao), yuanbao)
	log.Info("玩家[%d] 添加元宝[%d] 库存[%d] 原因[%s]", this.Id(), yuanbao, this.GetYuanbao(), reason)
	this.PlatformPushLootMoney(float32(yuanbao))
}
func (this *GateUser) RemoveYuanbao(yuanbao uint32, reason string) bool {
	if this.GetYuanbao() >= yuanbao {
		this.yuanbao = this.GetYuanbao() - yuanbao
		send := &msg.GW2C_UpdateYuanbao{Num:pb.Uint32(this.GetYuanbao())}
		this.SendMsg(send)
		log.Info("玩家[%d] 扣除元宝[%d] 库存[%d] 原因[%s]", this.Id(), yuanbao, this.GetYuanbao(), reason)
		RCounter().IncrByDate("item_remove", uint32(msg.ItemId_YuanBao), yuanbao)
		this.PlatformPushConsumeMoney(float32(yuanbao))
		return true
	}
	log.Info("玩家[%d] 扣除元宝[%d]失败 库存[%d] 原因[%s]", this.Id(), yuanbao, this.GetYuanbao(), reason)
	return false
}


// 添加金卷
func (this *GateUser) GetCoupon() uint32  { return this.coupon }
func (this *GateUser) AddCoupon(num uint32, reason string) {
	this.coupon = this.GetCoupon() + num
	send := &msg.GW2C_UpdateCoupon{Num:pb.Uint32(this.GetCoupon())}
	this.SendMsg(send)
	log.Info("玩家[%d] 添加金卷[%d] 库存[%d] 原因[%s]", this.Id(), num, this.GetCoupon(), reason)
}
func (this *GateUser) RemoveCoupon(num uint32, reason string) bool {
	if this.GetCoupon() >= num {
		this.coupon = this.GetCoupon() - num
		send := &msg.GW2C_UpdateCoupon{Num:pb.Uint32(this.GetCoupon())}
		this.SendMsg(send)
		log.Info("玩家[%d] 添加金卷[%d] 库存[%d] 原因[%s]", this.Id(), num, this.GetCoupon(), reason)
		//CountMgr().AddRemove(uint32(msg.ItemId_Coupon), num)
		RCounter().IncrByDate("item_remove", uint32(msg.ItemId_Coupon), num)
		return true
	}
	log.Info("玩家[%d] 添加金卷[%d]失败 库存[%d] 原因[%s]", this.Id(), num, this.GetCoupon(), reason)
	return false
}

// 购买道具
func (this* GateUser) BuyItem(productid uint32, num uint32) {
	product , ok := tbl.ShopBase.TShopById[productid]
	if ok == false {
		log.Error("[元宝商店] 玩家[%s %d] 购买商品[%d]，配置不存在", this.Name(), this.Id(), productid)
		return
	}

	itemid := uint32(product.Itemid)
	item, ok := tbl.ItemBase.ItemBaseDataById[itemid]
	if ok == false {
		log.Error("[元宝商店] 玩家[%s %d] 购买商品[%d]，无效的道具[%d]", this.Name(), this.Id(), productid, itemid)
		return
	}
	
	cost := uint32(product.Price) * num
	if this.GetYuanbao() < cost {
		log.Error("[元宝商店] 玩家[%s %d] 购买商品[%d]，元宝不足", this.Name(), this.Id(), itemid)
		this.SendNotify("元宝不足")		// 元宝要提示为积分，唉
		return;
	}
	this.RemoveYuanbao(cost, "商店购买道具")
	this.AddItem(itemid, num * uint32(product.Num), "商店购买道具")
	this.SendNotify("购买成功")
	GateSvr().SendNotice(	this.Face(), msg.NoticeType_Marquee, def.MakeNoticeText(this.Name(), "#00ff00", 26),
		def.MakeNoticeText("获得", "#00ff00", 26), def.MakeNoticeText(item.Name, "#00ff00", 26));
	;
}

// 道具换元宝
func (this *GateUser) SellBagItem(itemid, num uint32) {
	item, ok := tbl.ItemBase.ItemBaseDataById[itemid]
	if ok == false {
		log.Error("[道具换元宝] 玩家[%s %d] 无效的道具[%d]", this.Name(), this.Id(), itemid)
		return
	}
	
	have := this.bag.GetItemNum(itemid)
	if have < num {
		this.SendNotify("道具数量不足")
		return
	}
	this.RemoveItem(itemid, num, "出售道具");

	//
	yuanbao := uint32(item.Sold) * num		// 价值，元宝
	this.AddYuanbao(yuanbao, "出售道具")
	this.SendNotify("出售成功")
}


// 使用道具
func (this* GateUser) UseItem(itemid , num uint32) {
	have := this.bag.GetItemNum(itemid)
	if have < num {
		this.SendNotify("道具数量不足")
		return
	}

	this.bag.RemoveItem(itemid, num, "使用背包道具")
}


// 提钻石
func (this *GateUser) DeliveryDiamond(list []*msg.DeliveryGoods, token string, phone string) {
	defer this.SetDeliveryState(false)
	var diamondNum, partsNum uint32 = 0, 0
	for _, item := range list {

		base, ok := tbl.ItemBase.ItemBaseDataById[item.GetItemid()]
		if ok == false { 
			this.SendNotify("存在无效的道具")
			return
		}

		if item.GetNum() == 0 {
			this.SendNotify("提取数量不能是0")
			return
		}

		if base.Id == uint32(msg.ItemId_Diamond) {
			diamondNum += item.GetNum()
		}else if base.Id == uint32(msg.ItemId_DiamondParts) {
			partsNum += item.GetNum()
		}else {
			this.SendNotify("只能提取钻石道具")
			return
		}

		if this.bag.GetItemNum(item.GetItemid()) < item.GetNum()  {
			this.SendNotify("提钻石失败，道具数量不足")
			log.Error("[提钻石] 玩家[%s %d] 提钻石[%d] 数量[%d]，道具数量不足", this.Name(), this.Id(), item.GetItemid(), item.GetNum())
			return
		}
	}

	// 扣除钻石
	//if diamondNum > 0 { this.RemoveItem(uint32(msg.ItemId_Diamond), diamondNum, "提取钻石") }

	// 碎片转换钻石
	convertDiamond, remainParts := uint32(0), partsNum
	for remainParts >= uint32(tbl.Room.DiamondpartsToDiamond) {
		convertDiamond += 1
		remainParts -= uint32(tbl.Room.DiamondpartsToDiamond)
	}
	//if partsNum - remainParts > 0 { this.RemoveItem(uint32(msg.ItemId_DiamondParts), partsNum - remainParts, "提取碎片") }


	totalNum := diamondNum + convertDiamond
	if totalNum <= 0 {
		this.SendNotify(fmt.Sprintf("钻石券不足%d个，无法提取", tbl.Room.DiamondpartsToDiamond))
		return
	}

	// Http请求
	if def.HttpRequestIncrDiamonds(this.Id(), token, this.Account(), int32(totalNum), "提取钻石") == true {
		// 扣除钻石和碎片
		if diamondNum > 0 { this.RemoveItem(uint32(msg.ItemId_Diamond), diamondNum, "提取钻石") }
		if partsNum - remainParts > 0 { this.RemoveItem(uint32(msg.ItemId_DiamondParts), partsNum - remainParts, "提取碎片") }

		log.Info("玩家[%d] 添加钻石[%d] 库存[%d] 原因[%s]", this.Id(), totalNum, 0, "提取钻石")
		send := &msg.GW2C_RetDeliveryDiamond{}
		send.Diamond = pb.Int32(int32(diamondNum))
		send.Diamondparts = pb.Int32(int32(partsNum - remainParts))
		send.Total = pb.Int32(int32(totalNum))
		this.SendMsg(send)
	}else {
		log.Info("玩家[%s %d] 提取钻石异常，消耗钻石%d个，卷%d个", this.Name(), this.Id(), diamondNum, partsNum-remainParts)
		this.SendNotify("今日可兑换钻石额度已达上线，明日再来兑换")
	}

}

// 提货
func (this *GateUser) DeliveryGoods(list []*msg.DeliveryGoods, token string, tel string) {
	defer this.SetDeliveryState(false)

	// 虚拟道具不扣运费，小商品不能单独邮寄
	IsFreeDelivercost := true
	for _, item := range list {
		base, ok := tbl.ItemBase.ItemBaseDataById[item.GetItemid()]
		if ok == false { 
			this.SendNotify("存在无效的道具")
			return
		}
		if base.Type != int32(msg.ItemType_MobileCard) && base.Type != int32(msg.ItemType_ShoppingCard) {
			IsFreeDelivercost = false
		}
		if base.Type != int32(msg.ItemType_Smallware) {
			//IsAllSmallware = false
		}
	}

	// 小商品不能单独邮寄
	//if IsAllSmallware == true {
	//	this.SendNotify("小商品不能单独邮寄")
	//	return
	//}

	// 虚拟道具不扣运费，实物道具邮寄少于2个道具扣运费
	amount, delivercost := uint32(0), uint32(0)
	if IsFreeDelivercost == false {
		for _, item := range list {
			amount += item.GetNum()
		}
		if int64(amount) < tbl.Global.Delivery.Freelimit  {
			delivercost = uint32(tbl.Global.Delivery.Cost)
			//if this.GetYuanbao() < delivercost {
			//	this.SendNotify("运费不足")
			//	return
			//}
		}
	}


	// 玩家收货地址为空，从红包获取一次
	if this.GetAddressSize() == 0 && tel == ""{
		receiver, phone, address := RequestUserHomeAddress(this.Account(), token)
		if address == "" || phone == "" || receiver == "" {
			this.SendNotify("提货失败，没有收货地址")
			log.Error("[提货] 玩家[%s %d] 没有收货地址", this.Name(), this.Id())
			return
		}
		this.AddAddress(receiver, phone, address)
		//this.AddAddress("测试测试", "13681626939", "上海市闵行区1333弄28号31栋901")
	}

	// 检查
	for _, item := range list {
		if this.bag.GetItemNum(item.GetItemid()) < item.GetNum()  {
			this.SendNotify("提货失败，道具数量不足")
			log.Error("[提货] 玩家[%s %d] 提货[%d] 数量[%d]，道具数量不足", this.Name(), this.Id(), item.GetItemid(), item.GetNum())
			return
		}
	}


	// 扣钱
	if delivercost > 0 { 
		if this.RemovePlatformCoins(int32(delivercost), "提货扣运费") == false {
			this.SendNotify(fmt.Sprintf("运费不足, 需要%d金币", delivercost))
			return
		}
	}

	// 扣道具
	for _, item := range list {
		this.RemoveItem(item.GetItemid(), item.GetNum(), "玩家提货")
	}


	// 发送到订单到平台
	type OrderItem struct {
		Id		uint32		`json:"itemid"`
		Desc 	string		`json:"itemname"`
		Count	uint32		`json:"count"`
	}
	type DeliveryOrder struct {
		Id			uint64		`json:"uid"`
		Way			string		`json:"way"`
		Name		string		`json:"playerName"`
		Phone		string		`json:"playerMobile"`
		Address		string		`json:"playerAddr"`
		Nonce		string		`json:"nonce"`
		Sign		string		`json:"sign"`
		GameId		string		`json:"gameid"`
		Items		[]OrderItem	`json:"itemlist"`
	}

	// 发货平台提供的Key，不是红包key
	key1 := "top5737e7d6cb76cc84cb74bd9df946c"
	key2 := "01ca55b6371a5e5023130a0c6ea5dd11883af1c7bbbede3dea52d01a25877019"

	var order DeliveryOrder
	if tel == "" {
		address := this.GetDefaultAddress();
		order = DeliveryOrder{ Id:this.Id(), Way: "0", Name:address.GetReceiver(), Phone:address.GetPhone(), Address:address.GetAddress(), GameId:tbl.Global.Delivery.GameID}
	}else{
		order = DeliveryOrder{ Id:this.Id(), Way: "0", Name:"", Phone:tel, Address:"", GameId:tbl.Global.Delivery.GameID}
	}
	for _, item := range list {
		base := FindItemBase(item.GetItemid())
		if base == nil { log.Error("[提货] 玩家[%s %d] 道具[%d]配置无效", this.Name(), this.Id(), item.GetItemid()); continue }
		orderitem := OrderItem{Id:base.Id, Desc:base.Name, Count:item.GetNum()}
		order.Items = append(order.Items, orderitem)
	}

	order.Nonce = strconv.FormatInt(int64(this.Id()), 10) + "_" + strconv.FormatInt(util.CURTIMEUS(), 10)
	signbytes := []byte(key1 + order.Nonce + key2)
	md5array  := md5.Sum(signbytes)
	md5bytes  := []byte(md5array[:])
	md5string := fmt.Sprintf("%x", md5bytes)
	order.Sign = md5string

	orderbytes, err := json.Marshal(order)
	if err != nil {
		this.SendNotify("提货异常，稍后再试")
		log.Error("[提货] 玩家[%s %d] 序列化json订单失败[%s]", this.Name(), this.Id(), err)
		return
	}


	orderstr := util.BytesToString(orderbytes)
	url_platform := tbl.Global.Delivery.URLAPI;
	log.Info("[提货] 玩家[%s %d] 向平台[%s] 推送订单[%s]\n", this.Name(), this.Id(), url_platform, orderstr)
	resp, posterr := network.HttpPost(url_platform, orderstr)
	if posterr != nil {
		log.Error("[提货] 玩家[%s %d] HttpPost失败 err[%v] resp[%v]", this.Name(), this.Id(), posterr, resp)
		return
	}

	if resp.Code != http.StatusOK {
		log.Error("[提货] 玩家[%s %d] HttpPost返回错误码 resp[%s]", this.Name(), this.Id(), resp)
		return
	}

	type stRespBody struct {
		Msg		string
		Code 	int32
		Data	string
	}
	var jresp stRespBody
	unerr := json.Unmarshal(resp.Body, &jresp)
	if unerr != nil {
		log.Info("[提货] 玩家[%s %d] json.Unmarshal解析失败[%s] ", this.Name(), this.Id(), unerr)
		return
	}

	if jresp.Code != 0 {
		log.Error("[提货] 玩家[%s %d] Http返回失败 status[%d] msg[%s]", this.Name(), this.Id(), jresp.Code, jresp.Msg)
		return
	}

	log.Info("[提货] 玩家[%s %d] 提货成功 Data=%s", this.Name(), this.Id(), jresp.Data)
	this.SendNotify("提货成功,请耐心等待系统发货！")
}

// 同步大奖被拿走次数
func (this *GateUser) SyncBigRewardPickNum() {
	send := &msg.Sync_BigRewardPickNum{}
	picklist , err := Redis().ZRangeWithScores("bigreward_picknum", 0, -1).Result()
	if err != nil {
		this.SendMsg(send)
		log.Error("玩家[%s %d] 同步奖被拿走计数失败 err[%s]", this.Name(), this.Id(), err)
		return
	}
	
	for _, v := range picklist {
		stritemid := v.Member.(string)
		itemid, _ := strconv.ParseInt(stritemid, 10, 32)
		send.List = append(send.List, &msg.BigRewardItem{Id:pb.Uint32(uint32(itemid)), Num:pb.Uint32(uint32(v.Score))})
	}
	this.SendMsg(send)
}

// 签到
func (this *GateUser) GetSignReward(){
	if util.IsSameDay(int64(this.signtime), util.CURTIME()) {
		return
	}
	sconfig , ok := tbl.SignBase.TSignById[uint32(this.signreward + 1)]
	if !ok {
		return
	}
	this.SendNotify("领取成功")
	this.signtime = uint32(util.CURTIME())
	this.AddItem(uint32(sconfig.CostId), uint32(sconfig.Num), "签到奖励")
	this.signreward += 1
	if this.signreward >= 7 {
		this.signreward = 0
	}

	item, itemok := tbl.ItemBase.ItemBaseDataById[uint32(sconfig.CostId)]
	if itemok {
		GateSvr().SendNotice(this.Face(), msg.NoticeType_Marquee, 
			def.MakeNoticeText(this.Name(), "#00ff00", 26),
			def.MakeNoticeText("领取每日签到奖励", "#00ff00", 26), 
			def.MakeNoticeText(strconv.Itoa(int(sconfig.Num)), "#00ff00", 26),
			def.MakeNoticeText(item.Name, "#00ff00", 26),
		);
	}
}

func (this *GateUser) SendSign(){
	if util.IsSameDay(int64(this.signtime), util.CURTIME()) {
		return
	}
    send := &msg.GW2C_Ret7DayReward{Day: pb.Uint32(this.signreward + 1)}
	this.SendMsg(send)
}

// TODO: 获得补偿
func (this *GateUser) CheckHaveCompensation() {
	strkey := fmt.Sprintf("compen_%d", this.Id())
	members := Redis().SMembers(strkey).Val()
	//log.Info("key %s", strkey)
	for _, value := range members {
		para := strings.Split(value, ":")
		if len(para) != 2 {
			continue
		}

		intid, iderr := strconv.Atoi(para[0])
		intnum, numerr := strconv.Atoi(para[1])
		if iderr != nil || numerr != nil {
			continue
		}

		_, ok := tbl.ItemBase.ItemBaseDataById[uint32(intid)]
		if ok == false {
			log.Error("玩家[%s %d] 获得无效的补偿道具[%d]", this.Name(), this.Id(), intid)
			continue
		}

		this.AddItem(uint32(intid), uint32(intnum), "系统补偿")
		Redis().Del(strkey)
		//log.Info("玩家%s获得系统补偿 id:%d, 数量:%d", this.account, intid, intnum)
	}
}

// 统计登陆
func (this *GateUser) LoginStatistics() {
	datetime := time.Now().Format("2006-01-02")
	if this.tm_login == 0 {
		key := fmt.Sprintf("%s_create", datetime)
		Redis().Incr(key)
		key = fmt.Sprintf("%s_loginsum", datetime)
		Redis().Incr(key)
		this.continuelogin = 1
		return
	}
	if util.IsNextDay(this.tm_login, util.CURTIME()) {
		this.continuelogin += 1
		if this.nocountlogin == 0 {
			key := fmt.Sprintf("%s_login_%d", datetime, this.continuelogin)
			Redis().Incr(key)
		}
		key2 := fmt.Sprintf("%s_loginsum", datetime)
		Redis().Incr(key2)
	} else {
		if !util.IsSameDay(this.tm_login, util.CURTIME()) {
			this.continuelogin = 1
			this.nocountlogin = 1;
			key := fmt.Sprintf("%s_loginsum", datetime)
			Redis().Incr(key)
		}
	}
}

