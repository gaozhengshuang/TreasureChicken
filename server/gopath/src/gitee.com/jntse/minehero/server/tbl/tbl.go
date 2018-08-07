/// [注：本文件为自动生成，不需要人为编辑，若有修改，请通过配置py脚本来重新生成.]
/// @author xiejian
/// @generate date: GENE_DATE

package tbl
import "os"
import "path/filepath"
import "encoding/json"
import "gitee.com/jntse/gotoolkit/log"
import "gitee.com/jntse/gotoolkit/net"

// --------------------------------------------------------------------------
/// @brief 配置管理器
// --------------------------------------------------------------------------
type TblLoader struct {
	confpath network.TablePathConf
	tbls map[string]IBaseExcel
	jsons map[string]IBaseJson
}

func NewTblLoader(conf network.TablePathConf) *TblLoader {
	loader := new(TblLoader)
	loader.confpath = conf
	loader.tbls = make(map[string]IBaseExcel)
	loader.jsons= make(map[string]IBaseJson)
	loader.Init()
	return loader
}
 
func (this *TblLoader) Reload() {
	log.Info("重新加载tbl配置")
	this.ReloadExcel()
	this.ReloadJson()
}

func (this *TblLoader) LoadExcel(conf IBaseExcel, file string) {
	fullpath := this.confpath.Excel + file
	err := conf.Load(fullpath)
	if err != nil { panic(err) }
	log.Info("加载配置文件[%s]", fullpath)
	this.tbls[fullpath] = conf
}

func (this *TblLoader) ReloadExcel() {
	for k, v := range this.tbls {
		if err := v.Reload(); err != nil {
			panic(err)
		}
		log.Info("加载配置文件[%s]", k)
	}
}

func (this *TblLoader) LoadJson(conf IBaseJson, file string) {
	fullpath := this.confpath.Json + file
	ifile, err := os.Open(fullpath)
	if err != nil  { panic(err) }
	defer ifile.Close()

	fileinfo, _:= ifile.Stat()
	buf := make([]byte, fileinfo.Size())
	ifile.Read(buf)
	//buflen, _:= ifile.Read(buf)
	//log.Info("file=%s buflen=%d buf=\n%v\n", file, buflen, string(buf))
	if err := json.Unmarshal(buf, conf); err != nil {
		panic(err)
	}
	this.jsons[fullpath] = conf
	log.Info("加载配置文件[%s]", fullpath)
}

func (this *TblLoader) ReloadJson() {
	for f, v := range this.jsons {
		this.LoadJson(v, filepath.Base(f))
	}
}

func (this *TblLoader) Init() {

	// load excels
    this.LoadExcel(TaskBase, "taskbase.json")
	this.LoadExcel(MusicBase, "music.json")
	this.LoadExcel(RechargeBase, "recharge.json")
	this.LoadExcel(LevelBasee, "levelbase.json")
	this.LoadExcel(TBirckItembase, "birckitem.json")
	this.LoadExcel(TbirckInfobase, "birckinfobase.json")
	this.LoadExcel(SignBase, "signbase.json")
	this.LoadExcel(ProtoMsgIndex, "proto_index.json")
	this.LoadExcel(GiftProBase, "giftprobase.json")
	this.LoadExcel(NameBase, "namebase.json")
	this.LoadExcel(TBirckBase, "birckbase.json")
	this.LoadExcel(Question, "question.json")
	this.LoadExcel(TBallGiftbase, "ballgiftbase.json")
	this.LoadExcel(NoticeBase, "noticebase.json")
	this.LoadExcel(ShopBase, "shopbase.json")
	this.LoadExcel(TbirckRefreshbase, "birckrefreshbase.json")
	this.LoadExcel(TBallBase, "ballbase.json")
	this.LoadExcel(ItemBase, "itembase.json")
	

	// load jsons
    this.LoadJson(Game, "game.json")
	this.LoadJson(Global, "global.json")
	this.LoadJson(Room, "room.json")
	
}

