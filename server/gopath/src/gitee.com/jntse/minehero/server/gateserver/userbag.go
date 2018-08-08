package main
import "gitee.com/jntse/minehero/pbmsg"
import "gitee.com/jntse/gotoolkit/log"
import pb "github.com/golang/protobuf/proto"
import "gitee.com/jntse/minehero/server/tbl"
import "gitee.com/jntse/minehero/server/tbl/excel"


// --------------------------------------------------------------------------
/// @brief 道具
// --------------------------------------------------------------------------
type Item struct {
	bin *msg.ItemData
	base *table.ItemBaseDataDefine
}

func (t *Item) Id() uint32 {
	return t.bin.GetId()
}

func (t *Item) BaseId() uint32 {
	return t.base.Id
}

func (t *Item) Name() string {
	return t.base.Name
}

func (t *Item) Pos() int32 {
	return t.bin.GetPos()
}

func (t *Item) Num() uint32 {
	return t.bin.GetNum()
}

func (t *Item) Inc(num uint32) {
	t.bin.Num = pb.Uint32(t.Num() + num)
}

func (t *Item) Dec(num uint32) {
	if t.Num() < num { num = t.Num() }
	t.bin.Num = pb.Uint32(t.Num() - num)
}

func NewItem(data *msg.ItemData) *Item {
	base := FindItemBase(data.GetId())
	if base == nil { 
		log.Error("创建道具[%d]，找不到配置", data.GetId())
		return nil
	}

	item := new(Item)
	item.bin = data
	item.base = base
	return item
}

func NewItemData(id uint32, num uint32, pos int32) *msg.ItemData {
	data := &msg.ItemData { Id:pb.Uint32(id), Num:pb.Uint32(num), Pos:pb.Int32(pos) }
	return data
}

func FindItemBase(id uint32) *table.ItemBaseDataDefine {
	return tbl.ItemBase.ItemBaseDataById[id]
}

// --------------------------------------------------------------------------
/// @brief 背包
// --------------------------------------------------------------------------
type UserBag struct {
	items	map[uint32]*Item
	names	map[string]*Item
	owner   *GateUser
}

func (this *UserBag) Init(user *GateUser) {
	this.items = make(map[uint32]*Item)
	this.names = make(map[string]*Item)
	this.owner = user
}

func (this *UserBag) LoadBin(bin *msg.Serialize) {
	itembin := bin.GetItem();
	if itembin == nil { return }
	for _, data := range itembin.GetItems() {
		if data.GetPos() != int32(msg.ItemPos_Bag) { continue }
		item := NewItem(data)
		if item != nil {  this.LoadItemObj(item) }
	}
}

func (this *UserBag) PackBin(bin *msg.Serialize) {
	bin.Item = &msg.ItemBin{Items:make([]*msg.ItemData,0)}
	itembin := bin.GetItem();
	for _, item := range this.items {
		itembin.Items = append(itembin.Items, item.bin)
	}
}

func (this *UserBag) FindById(id uint32) *Item {
	return this.items[id]
}

func (this *UserBag) FindByName(name string) *Item {
	return this.names[name]
}

func (this *UserBag) LoadItemObj(item *Item) {
	this.items[item.Id()] = item
	this.names[item.Name()] = item
	log.Trace("加载背包道具%+v", item.bin)
	log.Trace("加载背包道具%+v", item.base)
}

func (this *UserBag) AddItem(id uint32, num uint32, reason string) *Item {
	item := this.FindById(id)
	if item != nil {
		item.Inc(num)
	}else {
		item = NewItem(NewItemData(id, num, int32(msg.ItemPos_Bag)))
		if item == nil { 
			log.Error("玩家[%d] 添加道具[%d]失败", this.owner.Id(), id)
			return nil
		}
		this.items[item.Id()] = item
		this.names[item.Name()] = item
	}

	send := &msg.GW2C_AddPackageItem{Itemid:pb.Uint32(id), Num:pb.Uint32(num) }
	this.owner.SendMsg(send)
	log.Info("玩家[%d] 添加道具[%d] 数量[%d] 库存[%d] 原因[%s]", this.owner.Id(), id, num, item.Num(), reason)
	return item
}

func (this *UserBag) RemoveItem(id uint32, num uint32, reason string) bool {
	item, leftnum := this.FindById(id), uint32(0)
	if item == nil { 
		log.Error("[道具] 道具[%d] 数量为0", id)
		return  false
	}

	if item.Num() < num {
		log.Info("玩家[%d] 扣除道具[%d]失败 数量[%d] 库存[%d] 原因[%s]", this.owner.Id(), id, num, leftnum, reason)
		return false
	}

	if item.Num() > num {
		item.Dec(num)
		leftnum = item.Num()
	} else if item.Num() == num{
		delete(this.items, id)
		delete(this.names, item.Name())
	}
	send := &msg.GW2C_RemovePackageItem{Itemid:pb.Uint32(id), Num:pb.Uint32(num) }
	this.owner.SendMsg(send)
	log.Info("玩家[%d] 扣除道具[%d] 数量[%d] 库存[%d] 原因[%s]", this.owner.Id(), id, num, leftnum, reason)
	return true
}

func (this *UserBag) GetItemNum(id uint32) uint32 {
	item, _ := this.FindById(id), uint32(0)
	if item == nil { 
		return  0
	}
	return item.Num()
}

func (this *UserBag) Clean() {
	this.items , this.names = nil, nil
	this.items = make(map[uint32]*Item)
	this.names = make(map[string]*Item)
}
