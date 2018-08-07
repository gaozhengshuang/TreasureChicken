let _ = require('lodash');

let Tools = require('../util/Tools');
var NotificationController = function () {
    this._eventHandlesList = {};
}

NotificationController.prototype.Init = function (cb) {
    Tools.InvokeCallback(cb, null);
};
/**
 * 注册事件
 * @param {Number} type 
 * @param {Object} caller 
 * @param {Function} handler 
 */
NotificationController.prototype.On = function (type, caller, handler) {
    let handlers = this._eventHandlesList[type];
    if (handlers == null) {
        handlers = [];
        this._eventHandlesList[type] = handlers;
    }
    handlers.push({ caller, handler });
};
/**
 * 移除注册
 * @param {Number} type 
 * @param {Object} caller 
 * @param {Function} handler 
 */
NotificationController.prototype.Off = function (type, caller, handler) {
    let handlers = this._eventHandlesList[type];
    if (handlers != null) {
        let result = _.remove(handlers, function (h) {
            return h.caller == caller && h.handler == handler;
        });
    }
}

NotificationController.prototype.Emit = function (type, ...args) {
    let handlers = this._eventHandlesList[type];
    if (handlers == null) {
        return;
    }
    for (let i = 0; i < handlers.length; i++) {
        let handler = handlers[i];
        handler.handler.apply(handler.caller, args);
    }
}
module.exports = new NotificationController();