// Generated by github.com/davyxu/tabtoy
// Version: 2.8.7
// DO NOT EDIT!!
package table

import (
	"fmt"
	"encoding/json"
	"io/ioutil"
)

// Defined in table: ProtoMsgIndex
type ProtoMsgIndex struct {

	//ProtoId
	ProtoId []*ProtoIdDefine
}

// Defined in table: ProtoId
type ProtoIdDefine struct {

	//唯一id
	Id int32 `json:"id"`

	Name string `json:"name"`
}

// 添加init初始实例 "Add By Tse"
var InsProtoMsgIndexTable *ProtoMsgIndexTable

func init() {
	InsProtoMsgIndexTable = NewProtoMsgIndexTable()
}

// 添加Reload方法 "Add By Tse"
func (self *ProtoMsgIndexTable) Reload() error {
	return self.Load(self.configfile)
}

// ProtoMsgIndex 访问接口
type ProtoMsgIndexTable struct {

	// 表格原始数据
	ProtoMsgIndex

	// 索引函数表
	indexFuncByName map[string][]func(*ProtoMsgIndexTable) error

	// 清空函数表
	clearFuncByName map[string][]func(*ProtoMsgIndexTable) error

	// 加载前回调
	preFuncList []func(*ProtoMsgIndexTable) error

	// 加载后回调
	postFuncList []func(*ProtoMsgIndexTable) error

	ProtoIdById map[int32]*ProtoIdDefine

	ProtoIdByName map[string]*ProtoIdDefine

	// 配置文件
	configfile string
}

// 从json文件加载
func (self *ProtoMsgIndexTable) Load(filename string) error {
	self.configfile = filename
	data, err := ioutil.ReadFile(filename)

	if err != nil {
		return err
	}

	var newTab ProtoMsgIndex

	// 读取
	err = json.Unmarshal(data, &newTab)
	if err != nil {
		return err
	}

	// 所有加载前的回调
	for _, v := range self.preFuncList {
		if err = v(self); err != nil {
			return err
		}
	}

	// 清除前通知
	for _, list := range self.clearFuncByName {
		for _, v := range list {
			if err = v(self); err != nil {
				return err
			}
		}
	}

	// 复制数据
	self.ProtoMsgIndex = newTab

	// 生成索引
	for _, list := range self.indexFuncByName {
		for _, v := range list {
			if err = v(self); err != nil {
				return err
			}
		}
	}

	// 所有完成时的回调
	for _, v := range self.postFuncList {
		if err = v(self); err != nil {
			return err
		}
	}

	return nil
}

// 注册外部索引入口, 索引回调, 清空回调
func (self *ProtoMsgIndexTable) RegisterIndexEntry(name string, indexCallback func(*ProtoMsgIndexTable) error, clearCallback func(*ProtoMsgIndexTable) error) {

	indexList, _ := self.indexFuncByName[name]
	clearList, _ := self.clearFuncByName[name]

	if indexCallback != nil {
		indexList = append(indexList, indexCallback)
	}

	if clearCallback != nil {
		clearList = append(clearList, clearCallback)
	}

	self.indexFuncByName[name] = indexList
	self.clearFuncByName[name] = clearList
}

// 注册加载前回调
func (self *ProtoMsgIndexTable) RegisterPreEntry(callback func(*ProtoMsgIndexTable) error) {

	self.preFuncList = append(self.preFuncList, callback)
}

// 注册所有完成时回调
func (self *ProtoMsgIndexTable) RegisterPostEntry(callback func(*ProtoMsgIndexTable) error) {

	self.postFuncList = append(self.postFuncList, callback)
}

// 创建一个ProtoMsgIndex表读取实例
func NewProtoMsgIndexTable() *ProtoMsgIndexTable {
	return &ProtoMsgIndexTable{

		indexFuncByName: map[string][]func(*ProtoMsgIndexTable) error{

			"ProtoId": {func(tab *ProtoMsgIndexTable) error {

				// ProtoId
				for _, def := range tab.ProtoId {

					if _, ok := tab.ProtoIdById[def.Id]; ok {
						panic(fmt.Sprintf("duplicate index in ProtoIdById: %v", def.Id))
					}

					if _, ok := tab.ProtoIdByName[def.Name]; ok {
						panic(fmt.Sprintf("duplicate index in ProtoIdByName: %v", def.Name))
					}

					tab.ProtoIdById[def.Id] = def
					tab.ProtoIdByName[def.Name] = def

				}

				return nil
			}},
		},

		clearFuncByName: map[string][]func(*ProtoMsgIndexTable) error{

			"ProtoId": {func(tab *ProtoMsgIndexTable) error {

				// ProtoId

				tab.ProtoIdById = make(map[int32]*ProtoIdDefine)
				tab.ProtoIdByName = make(map[string]*ProtoIdDefine)

				return nil
			}},
		},

		ProtoIdById: make(map[int32]*ProtoIdDefine),

		ProtoIdByName: make(map[string]*ProtoIdDefine),
	}
}
