let _ = require('lodash');
let moment = require('moment');

let Define = require('../Util/Define');
let Tools = require('../Util/Tools');
let ConfigController = require('./ConfigController');
let NotificationController = require('./NotificationController');
let ProtoMsg = require('../Util/ProtoMsg');
let { stringify } = require('querystring');

var NetWorkController = function () {
    this.sock = null;
    this.listeners = {};
    this.protoIndexById = {};
    this.protoIndexByName = {};

    this.connectedCallback = null;
    this.closeedCallback = null;
}

NetWorkController.prototype.Init = function (cb) {
    let protoIndexs = ConfigController.GetConfig('ProtoId');
    if (protoIndexs == null) {
        Tools.InvokeCallback(cb, 'proto配置错误');
        return;
    }
    for (let i = 0; i < protoIndexs.length; i++) {
        let index = protoIndexs[i];
        let checks = Tools.CheckParams(index, ['Id', 'Name']);
        if (checks.length > 0) {
            console.log('proto Index 参数检查失败 ' + index);
            continue;
        }
        this.protoIndexById[index.Id] = index.Name;
        this.protoIndexByName[index.Name] = index.Id;
    }
    Tools.InvokeCallback(cb, null);
};

NetWorkController.prototype.Connect = function (url, cb) {
    this.connectedCallback = cb;
    this.sock = new WebSocket(url);
    this.sock.onmessage = this.onMessage.bind(this);
    this.sock.onerror = this.onError.bind(this);
    this.sock.onopen = this.onOpen.bind(this);
    this.sock.onclose = this.onClose.bind(this);
}

NetWorkController.prototype.Send = function (name, obj, cb) {
    if (this.sock == null || this.sock.readyState != WebSocket.OPEN) {
        return;
    }
    if (this.protoIndexByName[name] == null) {
        Tools.InvokeCallback(cb, '没找到消息 ' + name);
        return;
    }
    let proto = Tools.GetValueInObj(ProtoMsg, name);
    if (proto == null) {
        Tools.InvokeCallback(cb, '没有消息体 : ' + proto);
        return;
    }
    let message = proto.fromObject(obj);
    let result = proto.verify(message);
    if (result != null) {
        Tools.InvokeCallback(cb, '消息内容错误 : ' + result);
        return;
    }
    let msg = proto.encode(message).finish();
    let id = this.protoIndexByName[name];
    let head = new Uint8Array([(msg.length + 4) & 0xff, (msg.length + 4) >> 8, id & 0xff, id >> 8,]);
    let info = new Uint8Array(msg.length + 4);
    console.log('发送 ' + name + ' 消息');
    info.set(head, 0);
    info.set(msg, 4);
    this.sock.send(info);
    Tools.InvokeCallback(cb, null);
};

NetWorkController.prototype.Close = function (cb) {
    if (this.sock == null || this.sock.readyState != WebSocket.OPEN) {
        Tools.InvokeCallback(cb, '连接不可用');
        return;
    }
    this.closeedCallback = cb;
    this.sock.close();
}

NetWorkController.prototype.AddListener = function (name, caller, handler) {
    if (this.protoIndexByName[name] == null) {
        return '没找到消息 ' + name;
    }
    let id = this.protoIndexByName[name];
    let listenerlist = this.listeners[id];
    if (listenerlist == null) {
        listenerlist = [];
        this.listeners[id] = listenerlist;
    }
    listenerlist.push({ caller, handler });
    return null;
}

NetWorkController.prototype.RemoveListener = function (name, caller, handler) {
    if (this.protoIndexByName[name] == null) {
        return '没找到消息 ' + name;
    }
    let id = this.protoIndexByName[name];
    let listenerlist = this.listeners[id];
    if (listenerlist != null) {
        _.remove(listenerlist, function (h) {
            return h.caller == caller && h.handler == handler;
        })
    }
    return null;
}

//事件函数
NetWorkController.prototype.onMessage = function (obj) {
    let reader = new FileReader();
    reader.readAsArrayBuffer(obj.data);
    reader.onload = function () {
        var uint8View = new Uint8Array(reader.result);
        let length = (uint8View[0] & 0xff) + (uint8View[1] << 8);
        let msgid = (uint8View[2] & 0xff) + (uint8View[3] << 8);
        var msg = uint8View.subarray(4);
        if (length != uint8View.length) {
            console.log('长度不匹配 包体长度 : ' + uint8View.length + ' 消息长度 : ' + length);
        } else {
            let protoName = this.protoIndexById[msgid];
            if (protoName == null) {
                console.log('没有消息类型 : ' + msgid);
                return;
            }
            //解析proto数据
            let proto = Tools.GetValueInObj(ProtoMsg, protoName);
            if (proto == null) {
                console.log('没有消息体 : ' + protoName);
                return;
            }
            let message = proto.decode(msg);
            let obj = proto.toObject(message);
            console.log('收到 ' + protoName + ' 消息');
            //调用监听函数
            let listenerlist = this.listeners[msgid];
            if (listenerlist != null) {
                for (let i = 0; i < listenerlist.length; i++) {
                    let handler = listenerlist[i];
                    handler.handler.call(handler.caller, msgid, obj);
                }
            }
        }
    }.bind(this);
}

NetWorkController.prototype.onClose = function () {
    console.log(new Date() + '[网络消息] socket closed');
    this.sock = null;
    Tools.InvokeCallback(this.closeedCallback);
    this.closeedCallback = null;
    NotificationController.Emit(Define.EVENT_KEY.NET_CLOSE);
}

NetWorkController.prototype.onOpen = function (info) {
    console.log(new Date() + '[网络消息] socket opend ' + stringify(info));
    Tools.InvokeCallback(this.connectedCallback);
    this.connectedCallback = null;
    NotificationController.Emit(Define.EVENT_KEY.NET_OPEN);
}

NetWorkController.prototype.onError = function (err) {
    console.log(new Date() + '[网络消息] socket error ' + err);
    this.connectedCallback = null;
    NotificationController.Emit(Define.EVENT_KEY.NET_CLOSE);
}

module.exports = new NetWorkController();