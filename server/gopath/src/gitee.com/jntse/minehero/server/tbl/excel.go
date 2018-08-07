/// [注：本文件为自动生成，不需要人为编辑，若有修改，请通过配置py脚本来重新生成.]
/// @author xiejian
/// @generate date: GENE_DATE

package tbl
import "gitee.com/jntse/minehero/server/tbl/excel"

type IBaseExcel interface {
	Load(filename string) error
	Reload() error
}

// --------------------------------------------------------------------------
/// @brief 为excel config 实例取一个别名
// --------------------------------------------------------------------------
var TaskBase = table.InsTaskBaseTable
var MusicBase = table.InsMusicBaseTable
var RechargeBase = table.InsRechargeBaseTable
var LevelBasee = table.InsLevelBaseeTable
var TBirckItembase = table.InsTBirckItembaseTable
var TbirckInfobase = table.InsTbirckInfobaseTable
var SignBase = table.InsSignBaseTable
var ProtoMsgIndex = table.InsProtoMsgIndexTable
var GiftProBase = table.InsGiftProBaseTable
var NameBase = table.InsNameBaseTable
var TBirckBase = table.InsTBirckBaseTable
var Question = table.InsQuestionTable
var TBallGiftbase = table.InsTBallGiftbaseTable
var NoticeBase = table.InsNoticeBaseTable
var ShopBase = table.InsShopBaseTable
var TbirckRefreshbase = table.InsTbirckRefreshbaseTable
var TBallBase = table.InsTBallBaseTable
var ItemBase = table.InsItemBaseTable

