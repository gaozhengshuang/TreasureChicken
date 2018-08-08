let async = require('async');

let NetWorkController = require('./NetWorkController');
let NotificationController = require('./NotificationController');
let Tools = require("../Util/Tools");
let moment = require("moment/moment");
let Define = require("../Util/Define");
let Platform = require('../Util/Platform');

var LoginController = function () {
    this.sendHeartBeatTime = 0;
    this.revHeartBeat = 0;
    this.loginServerUrl = '';
    this.loginedToGate = false;

    this.userid = null;
    //断线重连标志位
    this.reconnecting = false;
}

LoginController.prototype.Init = function (cb) {
    NetWorkController.AddListener('msg.GW2C_HeartBeat', this, this.onGW2C_HeartBeat);
    NetWorkController.AddListener('msg.L2C_RetLogin', this, this.onL2C_RetLogin);
    NotificationController.On(Define.EVENT_KEY.CONNECT_TO_GATESERVER, this, this.onLoginedToGate);
    NotificationController.On(Define.EVENT_KEY.NET_OPEN, this, this.onNetOpen);
    NotificationController.On(Define.EVENT_KEY.NET_CLOSE, this, this.onNetClose);
    Tools.InvokeCallback(cb, null);
}

LoginController.prototype.ConnectToLoginServer = function (cb) {
    let url = 'ws://' + Platform.LoginHost + ':' + Platform.LoginPort + '/' + Platform.LoginSuffix;
    this.loginServerUrl = url;
    NetWorkController.Connect(url, cb);
}

//回调函数
LoginController.prototype.onGW2C_HeartBeat = function (msgid, data) {
    console.log(new Date() + '[网络消息] 接收心跳包');
    this.revHeartBeat = moment().unix();
}

LoginController.prototype.onL2C_RetLogin = function (msgid, data) {
    //连接gate server
    console.log(data);
    let UserModel = require('../Model/User');
    let url = 'ws://' + data.gatehost.ip + ':' + data.gatehost.port + '/ws_handler';
    async.waterfall([
        function (anext) {
            //第一步断开websocket
            NetWorkController.Close(anext);
        },
        function (anext) {
            //第二步 连接gateserver
            NetWorkController.Connect(url, anext);
        },
        function (anext) {
            //第三步 发送登录消息
            NetWorkController.Send('msg.C2GW_ReqLogin', {
                account: UserModel.GetAccount(),
                verifykey: data.verifykey,
                token: UserModel.loginInfo.token
            }, function (err) {
                anext(err);
            })
        }.bind(this)
    ], function (err) {
        if (err) {
            console.log('[严重错误] ' + err);
        }
    }.bind(this));
}

LoginController.prototype.onLoginedToGate = function (event) {
    let UserModel = require('../Model/User');
    this.userid = UserModel.GetUserId();
    this.loginedToGate = true;
    this.reconnecting = false;
    this.revHeartBeat = moment().unix();
}

LoginController.prototype.onNetOpen = function (event) {
    this.sendHeartBeatTime = moment().unix();
}

LoginController.prototype.onNetClose = function (event) {
    if (this.loginedToGate) {
        this.loginedToGate = false;
        this._showNetFailed('网路断开');
    }
}

LoginController.prototype.Update = function (dt) {
    let curTime = moment().unix();
    let UserModel = require('../Model/User');
    if (this.loginedToGate) {
        //这个状态下需要检测心跳
        if ((NetWorkController.sock != null && NetWorkController.sock.readyState == WebSocket.OPEN)) {
            //发送心跳包了
            if (curTime - this.sendHeartBeatTime > Define.HEART_BEAT.INTERVAL) {
                if (this.userid == null) {
                    this.userid = UserModel.getUserId();
                }
                if (this.userid != null) {
                    console.log(new Date() + '[网络消息] 发送心跳包');
                    NetWorkController.Send('msg.C2GW_HeartBeat', {
                        uid: this.userid,
                        time: moment().unix()
                    })
                    this.sendHeartBeatTime = curTime;
                }
            }
            //检查心跳包
            if (this.userid == null) {
                this.userid = UserModel.getUserId();
            }
            if (this.userid != null) {
                //说明登录gate成功了
                if (this.revHeartBeat > 0 && curTime - this.revHeartBeat > Define.HEART_BEAT.TIMEOUT) {
                    console.log('超时了');
                    this._showNetFailed('心跳超时');
                    this.loginedToGate = false;
                }
            }
        } else {
            //说明有问题了，启动自动登录流程
            this._showNetFailed('socket 状态有问题');
            this.loginedToGate = false;
        }
    }
}

LoginController.prototype._showNetFailed = function (reason) {
    console.log(new Date() + ' [断线重连] 原因 ' + reason);
    cc.loader.loadRes("Prefab/NetFailedView", function (err, prefab) {
        if (err) {
            console.log('[严重错误] 奖励资源加载错误 ' + err);
        } else {
            var newNode = cc.instantiate(prefab);
            cc.director.getScene().getChildByName('Canvas').addChild(newNode);
        }
    });
}

module.exports = new LoginController();