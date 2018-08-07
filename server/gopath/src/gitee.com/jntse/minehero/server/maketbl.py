#!/usr/bin/env python
#coding=utf-8

import os
import sys
import datetime
import string
import shutil
import subprocess

#获取系统变量
env_dict = os.environ


# 解析输入参数
g_GlobalSet = {}
for arg in sys.argv:
    if arg == '-v' or arg == '-verbose':    # 显示详细信息
        g_GlobalSet['verbose'] = True

def ExecShellCmd(cmd, check_call=False):
    if check_call or g_GlobalSet.has_key('verbose'):
        subprocess.check_call(cmd, shell=True)
    else:
        subprocess.check_output(cmd, shell=True)


#配置路径
TBL_EXCEL=env_dict['JUMPGAME_ROOT']+'/docs/excel'
TBL_JSON=env_dict['JUMPGAME_ROOT']+'/docs/json'
PROTOCOL_PATH=env_dict['JUMPGAME_ROOT']+'/protocol'


curdir = os.getcwd()
os.chdir(env_dict['JUMPGAME_ROOT']+'/protocol')
#print os.popen('./gen_protogo.sh','r').read()
ExecShellCmd('./gen_protogo.sh',True)
os.chdir(curdir)


# 工具生成protobuf协议index配置
os.chdir(env_dict['JUMPGAME_SERVER'])
pbindex_path = './tools/gen_pbindex'
if os.path.isfile(pbindex_path+'/gen_pbindex') == False:
    curdir = os.getcwd()
    os.chdir(pbindex_path)
    ExecShellCmd("gob", True)
    os.chdir(curdir)

cmd_genpb = '%s/gen_pbindex -root=%s -output=%s/proto_index.xlsx' % (pbindex_path, PROTOCOL_PATH, TBL_EXCEL)
ExecShellCmd(cmd_genpb,True)

# work path
os.chdir(env_dict['JUMPGAME_SERVER']+'/server')


# 输出路径
OUT_TBL_EXCEL_SERVER=env_dict['JUMPGAME_SERVER']+'/server/tbl/excel'
OUT_TBL_EXCEL_CLIENT=env_dict['JUMPGAME_ROOT']+'/docs/tbl'
OUT_TBL_JSON_SERVER=env_dict['JUMPGAME_SERVER']+'/server/tbl/json'
if os.path.exists(OUT_TBL_EXCEL_SERVER) == True:    shutil.rmtree(OUT_TBL_EXCEL_SERVER)
if os.path.exists(OUT_TBL_EXCEL_CLIENT) == True:    shutil.rmtree(OUT_TBL_EXCEL_CLIENT)
if os.path.exists(OUT_TBL_JSON_SERVER) == True:    shutil.rmtree(OUT_TBL_JSON_SERVER)
os.makedirs(OUT_TBL_EXCEL_SERVER, 0775)
os.makedirs(OUT_TBL_EXCEL_CLIENT, 0775)
os.makedirs(OUT_TBL_JSON_SERVER, 0775)


# protobuf版本 2.0/3.0
PROTOBUF_VER=2

class TblGenerator:

    def __init__(self, in_excel, in_json, out_tbl_excel, out_tbl_client, out_tbl_json):
        self.in_excel = in_excel
        self.in_json = in_json
        self.out_tbl_excel = out_tbl_excel
        self.out_tbl_client = out_tbl_client
        self.out_tbl_json = out_tbl_json
        self.excels = {}
        self.jsons = {}

    #// --------------------------------------------------------------------------
    #/// @param 表格归属的结构体所在的结构的名称
    #/// @param 生成文件名
    #/// @param 表格文件名
    #// --------------------------------------------------------------------------
    def maketbl_excel(self, combname, outfile, infile):
        cmd = ""
        cmd += "tabtoy -mode=exportorv2 -protover=%d --combinename=%s " % (PROTOBUF_VER, combname)
        cmd += "-go_out=%s/%s.go -json_out=%s/%s.json " % (self.out_tbl_excel, outfile, self.out_tbl_excel, outfile)
        cmd += "-lua_out=%s/%s.lua -proto_out=%s/%s.proto " % (self.out_tbl_client, outfile, self.out_tbl_client, outfile)
        cmd += "%s/%s" % (self.in_excel, infile)
        ExecShellCmd(cmd)
        self.excels[combname] = outfile + ".json"
        print "make excel config ", infile
        self.maketbl_excel_client(combname, outfile, infile)

    def maketbl_excel_client(self, combname, outfile, infile):
        cmd = ""
        cmd += "tabtoy -mode=exportorv2 -protover=%d --combinename=%s " % (PROTOBUF_VER, combname)
        cmd += "-json_out=%s/%s.json " % (self.out_tbl_client, outfile)
        cmd += "%s/%s" % (self.in_excel, infile)
        ExecShellCmd(cmd)
        self.excels[combname] = outfile + ".json"
        print "make client config ", infile

    #// --------------------------------------------------------------------------
    #/// @brief 参数1 包名
    #/// @brief 参数2 主结构名称
    #/// @brief 参数3 文件名
    #// --------------------------------------------------------------------------
    def maketbl_json(self, pkgName, tableIns, infile):
        jsonFile = '%s/%s' % (self.in_json, infile)
        print "copy %s -> %s" % (jsonFile, self.out_tbl_json)
        shutil.copy(jsonFile, self.out_tbl_json)

        cmd = "gojson -fmt=json -pkg=%s -name=%s -subStruct=true " % (pkgName, tableIns)
        cmd += "-input=%s/%s -o=%s/%s.go" % (self.in_json, infile, self.out_tbl_json, infile)
        ExecShellCmd(cmd)
        self.jsons[tableIns] = infile
        print "\bmake json config ", infile 
        print ""

    def gen_excelcode(self, dirName, outFile, fileTempate):
        if os.path.exists(dirName) == False:
            os.makedirs(dirName, 0775)
        filePath  = '%s/%s' % (dirName, outFile)
        class_file = open(filePath,'w')

        #模版文件
        template_file = open(dirName + '/' + fileTempate,'r')
        tmpl = string.Template(template_file.read())

        ##
        impl_line = ""
        for combname , v in self.excels.items():
            impl_line += "var %s = table.Ins%sTable\n" % (combname , combname)

        #模版替换
        lines, tm_now = [], datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')
        lines.append(tmpl.substitute(GENE_DATE = tm_now, IMPL_INSTANCE = impl_line))

            # 0.将生成的代码写入文件
        class_file.writelines(lines)
        class_file.close()

    def gen_jsoncode(self, dirName, outFile, fileTempate):
        if os.path.exists(dirName) == False:
            os.makedirs(dirName, 0775)
        filePath  = '%s/%s' % (dirName, outFile)
        class_file = open(filePath,'w')

        #模版文件
        template_file = open(dirName + '/' + fileTempate,'r')
        tmpl = string.Template(template_file.read())

        ##
        impl_line = ""
        for tableIns , v in self.jsons.items():
            impl_line += "var {0} *{1}.{0} = new({1}.{0})\n".format(tableIns, 'table')

        #模版替换
        lines, tm_now = [], datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')
        lines.append(tmpl.substitute(GENE_DATE = tm_now, IMPL_INSTANCE = impl_line))

            # 0.将生成的代码写入文件
        class_file.writelines(lines)
        class_file.close()

    def gen_tblcode(self, dirName, outFile, fileTempate):
        if os.path.exists(dirName) == False:
            os.makedirs(dirName, 0775)
        filePath  = '%s/%s' % (dirName, outFile)
        class_file = open(filePath,'w')

        #模版文件
        template_file = open(dirName + '/' + fileTempate,'r')
        tmpl = string.Template(template_file.read())

        ##
        loadexcel = ""
        for combname , jsonfile in self.excels.items():
            loadexcel += "this.LoadExcel(%s, \"%s\")\n\t" % (combname, jsonfile)

        loadjson = ""
        for tableIns, jsonfile in self.jsons.items():
            loadjson += "this.LoadJson(%s, \"%s\")\n\t" % (tableIns, jsonfile)

        #模版替换
        lines, tm_now = [], datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')
        lines.append(tmpl.substitute(GENE_DATE = tm_now, LOAD_INS_EXCEL = loadexcel, LOAD_INS_JSON = loadjson))

            # 0.将生成的代码写入文件
        class_file.writelines(lines)
        class_file.close()


#// --------------------------------------------------------------------------
#/// @brief 打表
#// --------------------------------------------------------------------------
gen = TblGenerator(TBL_EXCEL, TBL_JSON, OUT_TBL_EXCEL_SERVER, OUT_TBL_EXCEL_CLIENT, OUT_TBL_JSON_SERVER)
gen.maketbl_excel('ProtoMsgIndex', 'proto_index', 'proto_index.xlsx')
gen.maketbl_excel('ItemBase', 'itembase', 'TItem.xlsx')
#gen.maketbl_excel('DungeonsBase', 'dungeons', 'TDungeons.xlsx')
gen.maketbl_excel('ShopBase', 'shopbase', 'TShop.xlsx')
#gen.maketbl_excel('RandomNameBase', 'randomname', 'TRandomName.xlsx')
#gen.maketbl_excel('TurntableBase', 'turntable', 'TTurntableNew.xlsx')
gen.maketbl_excel('SignBase', 'signbase', 'TSign.xlsx')
gen.maketbl_excel('RechargeBase', 'recharge', 'TRecharge.xlsx')
gen.maketbl_excel('NoticeBase', 'noticebase', 'TNotice.xlsx')
gen.maketbl_excel('MusicBase', 'music', 'TMusic.xlsx')
gen.maketbl_excel('LevelBasee', 'levelbase', 'TLevel.xlsx')
gen.maketbl_excel('NameBase', 'namebase', 'TName.xlsx')

gen.maketbl_excel('TBallBase', 'ballbase', 'TBall.xlsx')
gen.maketbl_excel('TBirckBase', 'birckbase', 'TBirck.xlsx')
gen.maketbl_excel('TBirckItembase', 'birckitem', 'TBirckItem.xlsx')
gen.maketbl_excel('TbirckRefreshbase', 'birckrefreshbase', 'TBirckRefresh.xlsx')
gen.maketbl_excel('TbirckInfobase', 'birckinfobase', 'TBirckInfo.xlsx')
gen.maketbl_excel('TBallGiftbase', 'ballgiftbase', 'TBallGift.xlsx')
gen.maketbl_excel('GiftProBase', 'giftprobase', 'TGiftPro.xlsx')
gen.maketbl_excel('TaskBase', 'taskbase', 'TTask.xlsx')
gen.maketbl_excel('Question', 'question', 'TQuestion.xlsx')


gen.maketbl_json('table', 'Global', 'global.json')
gen.maketbl_json('table', 'Room', 'room.json')
gen.maketbl_json('table', 'Game', 'game.json')

#// --------------------------------------------------------------------------
#/// @brief 生成自动加载代码
#// --------------------------------------------------------------------------
gen.gen_excelcode('./tbl', 'excel.go', 'excel.go.template')
gen.gen_jsoncode('./tbl', 'json.go', 'json.go.template')
gen.gen_tblcode('./tbl', 'tbl.go', 'tbl.go.template')

#
print "make all res finish!"


