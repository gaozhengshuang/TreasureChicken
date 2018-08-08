import moment from 'moment';

import Tools from '../util/Tools';
import Define from "../Util/Define";
import NetWorkController from './NetWorkController';
let TimeController = function () {
    this._recievedTime = Tools.GetMilliSecond();
    this._clientTime = Tools.GetMilliSecond();
};
/**
 * 
 * @param {Function} cb 
 */
TimeController.prototype.Init = function (cb) {
    NetWorkController.AddListener('msg.GW2C_HeartBeat', this, this.onGW2C_HeartBeat);
    Tools.InvokeCallback(cb, null);
};

TimeController.prototype.GetCurTime = function () {
    let cur = Tools.GetMilliSecond();
    return Math.floor((cur - this._clientTime + this._recievedTime) / 1000);
}

TimeController.prototype.GetCurTimeMs = function () {
    let cur = Tools.GetMilliSecond();
    return cur - this._clientTime + this._recievedTime;
}


TimeController.prototype.onGW2C_HeartBeat = function (msgid, data) {
    this._recievedTime = Math.floor(data.time / 1000.0);
    this._clientTime = Tools.GetMilliSecond();
}

module.exports = new TimeController();