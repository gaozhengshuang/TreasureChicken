let _ = require('lodash');
let moment = require('moment');

let Tools = {
    GetRandomInt: function (min, max) {
        min = Math.ceil(min);
        max = Math.floor(max);
        return Math.floor(Math.random() * (max - min)) + min;
    },
    GetRandomResult: function () {
        return Math.random() < 0.5;
    },
    CheckParams: function (obj, params) {
        var lostParams = [];
        for (var i = 0; i < params.length; i++) {
            if (obj[params[i]] == null) {
                lostParams.push(params[i]);
            }
        }
        return lostParams;
    },
    GetValueInObj: function (obj, key) {
        let keys = key.split('.');
        let source = obj;
        for (let i = 0; i < keys.length; i++) {
            if (source == null) {
                return null;
            }
            source = source[keys[i]];
        }
        return source;
    },
    GetValuesInArray: function (arr, key) {
        if (_.isArray(arr)) {
            let ret = [];
            for (let i = 0; i < arr.length; i++) {
                let obj = arr[i];
                ret.push(Tools.GetValueInObj(obj, key));
            }
            return ret;
        }
        return [];
    },
    GetRandomString: function (len) {
        len = len || 32;
        var $chars = 'ABCDEFGHJKMNPQRSTWXYZabcdefhijkmnprstwxyz2345678';    /****默认去掉了容易混淆的字符oOLl,9gq,Vv,Uu,I1****/
        var maxPos = $chars.length;
        var pwd = '';
        for (let i = 0; i < len; i++) {
            pwd += $chars.charAt(Math.floor(Math.random() * maxPos));
        }
        return pwd;
    },
    InvokeCallback: function (cb) {
        if (_.isFunction(cb)) {
            cb.apply(null, Array.prototype.slice.call(arguments, 1));
            return true;
        }
        return false;
    },
    GetMilliSecond: function () {
        return moment().unix() * 1000 + moment().milliseconds();
    }
}

module.exports = Tools;