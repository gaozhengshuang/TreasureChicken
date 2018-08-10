'use strict';
var fs = require('fs');
var node_xj = require('xls-to-json');
var process = require('child_process');

module.exports = {
    load() {
        // execute when package loaded
    },

    unload() {
        // execute when package unloaded
    },

    // register your ipc messages here
    messages: {
        'genJson'() {
            // open entry panel registered in package.json
            Editor.log('开始生成JSON');
            let sourcePath = __dirname + '\\..\\..\\..\\docs\\tbl\\proto_index.json';
            let targetFile = __dirname + '\\..\\..\\assets\\resources\\Json\\';
            let cmd = 'copy ' + sourcePath + ' ' + targetFile;
            process.exec(cmd, function (err, stdout, stderr) {
                if (err) {
                    Editor.log('++++++ Json 错误 ++++++' + stderr);
                } else {
                    sourcePath = __dirname + '\\..\\..\\..\\docs\\json\\*.json';
                    let cmd = 'copy ' + sourcePath + ' ' + targetFile;
                    process.exec(cmd, function (err, stdout, stderr) {
                        if (err) {
                            Editor.log('++++++ Json 错误 ++++++' + stderr);
                        } else {
                            Editor.log('生成成功');
                        }
                    });
                }
            });
        },
        'genProto'() {
            Editor.log('开始生成ProtoMsg');
            let sourcePath = __dirname + '\\..\\..\\..\\protocol\\*.proto';
            let targetFile = __dirname + '\\..\\..\\assets\\Script\\Util\\ProtoMsg.js';
            let cmd = 'pbjs -t static-module -w commonjs -o ' + targetFile + '  ' + sourcePath;
            process.exec(cmd, function (err, stdout, stderr) {
                if (err) {
                    Editor.log('++++++ 生成ProtoMsg 错误 ++++++' + err);
                } else {
                    Editor.log('生成成功');
                }
            });
        }
    },
};