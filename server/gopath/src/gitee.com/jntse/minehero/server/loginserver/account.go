package main

import (
	"fmt"
	"github.com/go-redis/redis"
	pb "github.com/gogo/protobuf/proto"
	"strings"
	"time"

	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/net"
	"gitee.com/jntse/gotoolkit/redis"
	"gitee.com/jntse/gotoolkit/util"

	"gitee.com/jntse/minehero/pbmsg"
	"gitee.com/jntse/minehero/server/def"
	"gitee.com/jntse/minehero/server/tbl"
)

type ClientAccount struct {
	session  network.IBaseNetSession
	account  string
	tm_login int64
}

// 查找账户绑定Gate
func FindAccountGateWay(account string) (*msg.AccountGateInfo, error) {
	info := &msg.AccountGateInfo{}
	key := fmt.Sprintf("%s_%s", def.RedisKeyAccountGate, account)
	err := utredis.GetProtoBin(Redis(), key, info)
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	ip, port := info.GetIp(), int(info.GetPort())
	if GateMgr().IsRegisted(ip, port) == false {
		log.Error("账户%s 存储的网关host %s:%d 不可用", account, ip, port)
		return nil, nil
	}

	return info, nil
}

// 绑定账户一个Gate
func BindingAccountGateWay(account string, ip string, port int, vkey string) error {
	key := fmt.Sprintf("%s_%s", def.RedisKeyAccountGate, account)
	info := &msg.AccountGateInfo{Ip: pb.String(ip), Port: pb.Int(port), Verifykey: pb.String(vkey)}
	if err := utredis.SetProtoBin(Redis(), key, info); err != nil {
		return err
	}
	err := InsertAccountToGate(account, ip, port)
	return err
}

// 移除账户和Gate的绑定关系
func UnBindingAccountGateWay(account string) error {
	key := fmt.Sprintf("%s_%s", def.RedisKeyAccountGate, account)
	err := Redis().Del(key).Err()
	return err
}

// 插入账户到Gate
func InsertAccountToGate(account string, ip string, port int) error {
	key := fmt.Sprintf("%s_%s:%d", def.RedisKeyGateAccounts, ip, port)
	err := Redis().SAdd(key, account).Err()
	return err
}

// 从Gate查找账户
func IsFindAccountFromGate(account string, ip string, port int) (bool, error) {
	key := fmt.Sprintf("%s_%s:%d", def.RedisKeyGateAccounts, ip, port)
	ok, err := Redis().SIsMember(key, account).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}

func AmountGateAccount(ip string, port int) int64 {
	key := fmt.Sprintf("%s_%s:%d", def.RedisKeyGateAccounts, ip, port)
	num, err := Redis().SCard(key).Result()
	if err != nil {
		return 0
	}
	return num
}

func QuickLogin(session network.IBaseNetSession, account string) bool {
	host, err := FindAccountGateWay(account)
	if err != nil && err != redis.Nil {
		log.Error("账户%s Find Account GateWay 报错err: %v", account, err)
		return false
	}

	if host == nil {
		return false
	}

	// 检查Gate是否存在
	ip, port, vkey := host.GetIp(), int(host.GetPort()), host.GetVerifykey()
	if GateMgr().IsRegisted(ip, port) == false {
		return false
	}

	// 检查Gate上是否清除了玩家信息，例如Gate重启过
	isfind, err := IsFindAccountFromGate(account, ip, port)
	if err != nil {
		log.Error("账户%s Is Find Account FromGate 报错err: %s", account, err)
		return false
	}

	// 解除绑定Account上的Gateway信息，使用普通登陆
	if isfind == false {
		UnBindingAccountGateWay(account)
		//RemoveAccountsFromGate(account, ip, port)
		return false
	}

	log.Info("账户[%s] 快速登陆Gate[ip:%s port:%d]", account, ip, port)
	session.SendCmd(newL2C_RetLogin("", ip, port, vkey))
	Login().AddAuthenAccount(account, session) // 避免同时登陆
	return true
}

// --------------------------------------------------------------------------
/// @brief 账户校验
/// @return
// --------------------------------------------------------------------------
func Authenticate(session network.IBaseNetSession, account string, passwd string) string {

	// 校验账户密码
	key := fmt.Sprintf("accounts_passwd_%s", account)
	svrpasswd, err := Redis().Get(key).Result()
	if err == redis.Nil {
		return "账户未注册"
	}

	if svrpasswd != passwd {
		return "密码错误"
	}

	return ""
}

func ProcessInvitationUser(charid uint64, invitationcode string) {

	// 保存邀请人信息
	if len(invitationcode) < 2 {
		return
	}

	invitation_user := invitationcode[2:]
	invitkey := fmt.Sprintf("user_%d_inviter", charid)
	Redis().Set(invitkey, invitation_user, 0)

	// 转账给邀请人
	invitation_openid, geterror := Redis().Get(fmt.Sprintf("user_%d_wechat_openid", invitation_user)).Result()
	if geterror != nil {
		return
	}
	def.HttpWechatCompanyPay(invitation_openid, 1)

}

// 微信小游戏，直接注册账户
func RegistAccountFromWechatMiniGame(account, passwd, invitationcode, name, face string) string {
	// 获取账户信息
	key := "accounts_" + account
	exist, err := Redis().Exists(key).Result()
	if err != nil {
		return "检查账户存在 Redis报错"
	}

	if exist == 1 {
		return ""
	}

	if errcode := RegistAccount(account, passwd, invitationcode, name, face); errcode != "" {
		return fmt.Sprintf("注册账户失败 账户[%s] 错误[%s]", account, errcode)
	}

	return ""
}

// 获取注册短信验证码
func GetRegistAuthCode(phone string) string {
	errcode, keyauthcode := "", fmt.Sprintf("regist_phone_%s", phone)
	switch {
	default:
		// 手机是否已经注册过
		accountkey := fmt.Sprintf("accounts_%s", phone)
		accountexist, _ := Redis().Exists(accountkey).Result()
		if accountexist == 1 {
			errcode = "该手机号已经注册过了"
			break
		}

		// 检查redis是否获取过验证码(自动过期)
		exist, _ := Redis().Exists(keyauthcode).Result()
		if exist == 1 {
			errcode = "稍后再试"
			break
		}

		authcode := def.SendSms(phone)
		if authcode == "" {
			errcode = "发送验证码失败"
			break
		}

		// 缓存验证码
		Redis().Set(keyauthcode, authcode, time.Second*60).Result()
	}

	if errcode != "" {
		log.Error("获取注册验证码失败 %s [%s]", keyauthcode, errcode)
	}
	return errcode
}

func RegistAccountCheck(phone, passwd, invitationcode, authcode, nickname string) (errcode string) {
	if phone == "" {
		errcode = "手机号不能为空"
		return
	}

	if authcode == "" {
		errcode = "请填写验证码"
		return
	}

	if passwd == "" {
		errcode = "密码不能为空"
		return
	}

	if nickname == "" {
		errcode = "昵称不能为空"
		return
	}

	if strings.Count(nickname, "")-1 > 8 {
		errcode = "昵称长度不能大于8个字符"
		return
	}

	if issp, _ := util.ContainsSpecialCharacter(nickname); issp == true {
		errcode = "昵称不能含有标点和特殊字符"
		return
	}

	// 账户检查重复
	keyaccount := fmt.Sprintf("accounts_%s", phone)
	bexist, _ := Redis().Exists(keyaccount).Result()
	if bexist == 1 {
		errcode = "账户已经存在"
		return
	}

	// 是否是机器人注册
	if authcode == "robot@free@regist" {
		freeregist, _ := Redis().Get(authcode).Int64() // Robot自由注册redis标记
		if freeregist == 0 {
			errcode = "使用了机器人自由注册码，但服务器没有Robot自由注册标记"
			return
		}
	} else {
		key := fmt.Sprintf("regist_phone_%s", phone)
		svrauthcode, err := Redis().Get(key).Result()
		if err == redis.Nil {
			errcode = "请先获取验证码"
			return
		} else if err != nil {
			errcode = "redis暂时不可用"
			log.Error("检查账户是否存在 Redis错误:%s", err)
			return
		}

		if svrauthcode == "" {
			errcode = "验证码已过期"
			return
		}

		if svrauthcode != authcode {
			errcode = "验证码错误"
			return
		}
	}

	// 昵称是否重复
	keynickname := fmt.Sprintf("accounts_nickname")
	keyvalue, err := Redis().SIsMember(keynickname, nickname).Result()
	if err != nil && err != redis.Nil {
		errcode = "redis暂时不可用"
		log.Error("检查昵称是否重复 Redis错误:%s", err)
		return
	}

	if keyvalue == true {
		errcode = "昵称重复"
		return
	}

	return ""
}

// --------------------------------------------------------------------------
/// @brief 注册账户
///
/// @param account 账户名
/// @param passwd 密码
/// @param invitationcode 邀请码
/// @param
///
/// @return
// --------------------------------------------------------------------------
func RegistAccount(account, passwd, invitationcode, nickname, face string) (errcode string) {
	errcode = ""
	switch {
	default:

		// 保存密码
		passwdkey := fmt.Sprintf("accounts_passwd_%s", account)
		if _, errpasswd := Redis().Set(passwdkey, passwd, 0).Result(); errpasswd != nil {
			errcode = "缓存账户密码失败"
			return
		}

		// 保存昵称
		keynickname := fmt.Sprintf("accounts_nickname")
		_, errnick := Redis().SAdd(keynickname, nickname).Result()
		if errnick != nil {
			errcode = "redis暂时不可用"
			log.Error("保存全局昵称 Redis错误:%s", errnick)
			return
		}

		// 实名认证
		// 生成唯一userid
		userid, errstr := GenerateUserId()
		if errstr != "" {
			errcode = errstr
			break
		}

		// 新建账户
		info := &msg.AccountInfo{
			Account: &account,
			Passwd:  &passwd,
			Userid:  pb.Uint64(userid),
		}

		keyaccount := fmt.Sprintf("accounts_%s", account)
		if errsetbin := utredis.SetProtoBin(Redis(), keyaccount, info); errsetbin != nil {
			errcode = "插入账户数据失败"
			log.Error("新建账户%s失败，err: %s", account, errsetbin)
			break
		}

		// 初始元宝和金卷
		Yuanbao := uint32(tbl.Global.Newuser.Yuanbao)
		userinfo := &msg.Serialize{
			Entity: &msg.EntityBase{Id: pb.Uint64(userid), Name: pb.String(nickname), Face: pb.String(""), Account: pb.String(account)},
			Base:   &msg.UserBase{Money: pb.Uint32(1000), Invitationcode: pb.String(invitationcode), Yuanbao: pb.Uint32(Yuanbao), Level: pb.Uint32(1)},
			Item:   &msg.ItemBin{},
		}
		userkey := fmt.Sprintf("userbin_%d", userid)
		log.Info("userinfo=%v", userinfo)
		if err := utredis.SetProtoBin(Redis(), userkey, userinfo); err != nil {
			errcode = "插入玩家数据失败"
			log.Error("新建账户%s插入玩家数据失败，err: %s", account, err)
			break
		}

		log.Info("账户[%s] UserId[%d] 创建新用户成功", account, userid)
		ProcessInvitationUser(userid, invitationcode)
	}

	if errcode != "" {
		log.Info("账户[%s] 创建新用户失败 err[%s]", account, errcode)
	}

	return errcode
}

// 检查新账户
func CheckNewAccount(session network.IBaseNetSession, account, name, face, passwd string) string {
	// 获取账户信息
	key := "accounts_" + account
	exist, err := Redis().Exists(key).Result()
	if err != nil {
		return "登陆检查账户是否存在"
	}

	if exist == 1 {
		return ""
	}
	log.Info("注册新账户[%s]", account) // 不存在的账户
	name = fmt.Sprintf("player_%s", account)
	if errcode := RegistAccount(account, passwd, "", name, ""); errcode != "" {
		return fmt.Sprintf("注册账户失败 账户[%s] 错误[%s]", account, errcode)
	}
	return ""
}
