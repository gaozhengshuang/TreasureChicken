package main
import (
	_"fmt"
	_"time"
	pb "github.com/golang/protobuf/proto"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/net"
	"gitee.com/jntse/minehero/pbmsg"
	_"gitee.com/jntse/gotoolkit/util"
)

// --------------------------------------------------------------------------
/// @brief 账户登陆等待池
// --------------------------------------------------------------------------
type stAccountLogin struct {
	account 	string
	tm_expire	int64
	verifykey	string
}

type LoginWaitPool struct {
	pool map[string]*stAccountLogin
}

func (this *LoginWaitPool) Init() {
	this.pool = make(map[string]*stAccountLogin)
}

func (this *LoginWaitPool) IsFind(acc string) bool {
	_, ok := this.pool[acc]
	return ok
}

func (this *LoginWaitPool) Find(acc string) *stAccountLogin {
	if a, ok := this.pool[acc]; ok {
		return a
	}
	return nil
}

func (this *LoginWaitPool) Insert(acc, key string, expire int64) {
	this.pool[acc] = &stAccountLogin{acc, expire, key}
}

func (this *LoginWaitPool) Remove(acc string) {
	delete(this.pool, acc)
}

func (this *LoginWaitPool) Clear() {
	this.pool= make(map[string]*stAccountLogin)
}

func (this *LoginWaitPool) Tick(now int64) {
	for _, v := range this.pool {
		if now >= v.tm_expire {
			UnBindingAccountGateWay(v.account)
			log.Info("账户:%s 超时还未登录Gate，删除注册信息", v.account)
			delete(this.pool, v.account)
		}
	}
}

// --------------------------------------------------------------------------
/// @brief 玩家管理器
// --------------------------------------------------------------------------
type UserManager struct {
	accounts	map[string]*GateUser
	ids			map[uint64]*GateUser
	names		map[string]*GateUser
}

func (this *UserManager) Init() {
	this.accounts = make(map[string]*GateUser)
	this.names = make(map[string]*GateUser)
	this.ids = make(map[uint64]*GateUser)
}

func (this *UserManager) CreateNewUser(session network.IBaseNetSession, account, key, token, face string) (*GateUser, string) {
	user := NewGateUser(account, key, token)
	if user.LoadDB() == false {
		return nil, "加载玩家DB数据失败"
	}
	user.SetFace(face)

	if user.Online(session) == false {
		return nil, "Online失败"
	}

	WaitPool().Remove(account)
	this.AddUser(user)
	log.Info("当前在线人数:%d", len(this.accounts))
	return user, ""
}

func (this *UserManager) Amount() int {
	return len(this.accounts)
}

func (this *UserManager) AmountOnline() int {
	count := 0
	for _, user := range this.accounts {
		if user.IsOnline() { count ++ }
	}
	return count
}

//func (this *UserManager) AddAccount(user *GateUser) {
//	this.accounts[user.Account()] = user
//}

func (this *UserManager) AddUser(user *GateUser) {
	this.accounts[user.Account()] = user
	this.ids[user.Id()] = user
	this.names[user.Name()] = user
}

func (this *UserManager) IsRegisted(acc string) bool {
	_, ok := this.accounts[acc]
	return ok
}

func (this *UserManager) FindByAccount(acc string) *GateUser {
	u, _ := this.accounts[acc]
	return u
}

func (this *UserManager) FindByName(name string) *GateUser {
	user, _ := this.names[name]
	return user
}

func (this *UserManager) FindById(id uint64) *GateUser {
	user, _ := this.ids[id]
	return user
}

func (this *UserManager) DelUser(user *GateUser) {
	delete(this.accounts, user.Account())
	delete(this.names, user.Name())
	delete(this.ids, user.Id())
	log.Info("当前在线人数:%d", len(this.accounts))
}

func (this *UserManager) Tick(now int64) {
	for _, user := range this.accounts {
		if this.IsRemove(user, now) {
			continue
		}
		user.Tick(now)
	}
}

func (this *UserManager) IsRemove(user *GateUser, now int64) bool {
	if user.IsCleanUp() {
		user.OnCleanUp()
		this.DelUser(user)
		return true
	}
	return false
}

//func (this *UserManager) LoadUserFromDB(user *GateUser) bool {
//	if user.LoadDB() == false { 
//		log.Error("账户[%s] 从DB加载玩家数据失败", user.Account())
//		return false 
//	}
//	this.ids[user.Id()] = user
//	this.names[user.Name()] = user
//	return true
//}

func (this *UserManager) OnMatchServerClose() {
}

// 房间服务器断开
func (this *UserManager) OnRoomServerClose(sid int) {
	for _, user := range this.accounts {
		if sid == user.RoomSid() {
			user.SendMsg(&msg.BT_GameOver{Roomid:pb.Int64(user.RoomId())})
			//user.GameEnd(nil , "房间服务器断开")
		}
	}
}

// 本服务器退出
func (this *UserManager) OnServerClose() {
	for _, user := range this.accounts {
		user.KickOut("服务器Shutdown")
	}
}

// 广播消息
func (this *UserManager) BroadcastMsg(msg pb.Message) {
	for _, user := range this.accounts {
		user.SendMsg(msg)
		//key := fmt.Sprintf("accounts_%s", user.Account())
		//Redis().Exists(key)
	}
}

// TODO:整点赠送免费步数，异步事件处理
func (this *UserManager) GiveFreeStep(now int64) {
	//for _, user := range this.accounts {
		//user.CheckGiveFreeStep(now, "整点在线获得")
		//event := NewGiveFreeStepEvent(now, "整点在线获得", user.CheckGiveFreeStep)
		//user.AsynEventInsert(event)
	//}
}

