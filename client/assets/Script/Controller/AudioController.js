import _ from 'lodash';

import Define from '../Util/Define';
import Tools from '../Util/Tools';
import NotificationController from './NotificationController';

let AudioController = function () {
    this.audioClips = {};
    this.audio = null;
    this.effectIds = [];
    this.effectCallbacks = {}
    this.disableMusic = cc.sys.localStorage.getItem(Define.DATA_KEY.DISABLE_MUSIC) == 'true';
    this.disableEffect = cc.sys.localStorage.getItem(Define.DATA_KEY.DISABLE_EFFECT) == 'true';
    this.audioName = '';
};

AudioController.prototype.Init = function (cb) {
    NotificationController.On(Define.EVENT_KEY.MUSIC_CHANGE, this, this.onChangeMusic);
    NotificationController.On(Define.EVENT_KEY.EFFECT_CHANGE, this, this.onChangeEffect.bind(this));
    cc.loader.loadResDir('Audio/', cc.AudioClip, function (err, ress, urls) {
        if (err) {
            console.log('[严重错误] 奖励资源加载错误 ' + err);
        } else {
            for (let i = 0; i < ress.length; i++) {
                this.audioClips[urls[i]] = ress[i];
            }
        }
        Tools.InvokeCallback(cb, err);
    }.bind(this));
}

AudioController.prototype.PlayMusic = function (name, loop = true) {
    this.audioName = name;
    if (this.disableMusic) {
        return;
    }
    if (this.audio != null) {
        cc.audioEngine.stop(this.audio);
        this.audio = null;
    }
    this.audio = cc.audioEngine.play(this.audioClips[name], loop, 1);
    cc.audioEngine.setFinishCallback(this.audio, this._onMusicFinish.bind(this));
}

AudioController.prototype.PlayEffect = function (name, cb) {
    if (this.disableEffect) {
        return;
    }
    let id = cc.audioEngine.play(this.audioClips[name], false, 1);
    this.effectIds.push(id);
    this.effectCallbacks[id] = cb;
    cc.audioEngine.setFinishCallback(id, this._onEffectFinish.bind(this, id));
};

AudioController.prototype.StopAllEffect = function () {
    for (let i = 0; i < this.effectIds.length; i++) {
        cc.audioEngine.stop(this.effectIds[i]);
    }
    this.effectIds = [];
    for (let key in this.effectCallbacks) {
        let cb = this.effectCallbacks[key];
        Tools.InvokeCallback(cb);
    }
    this.effectCallbacks = {};
}

AudioController.prototype.SetMusicVolume = function (val) {
    if (this.audio != null) {
        cc.audioEngine.setVolume(this.audio, val);
    }
}

AudioController.prototype.onChangeMusic = function (disable) {
    this.disableMusic = disable;
    if (disable) {
        if (this.audio != null) {
            cc.audioEngine.stop(this.audio);
            this.audio = null;
        }
    } else {
        if (this.audioName != '') {
            this.playAudio(this.audioName);
        }
    }
}
AudioController.prototype.onChangeEffect = function (disable) {
    this.disableEffect = disable;
    if (disable) {
        for (let i = 0; i < this.effectIds.length; i++) {
            cc.audioEngine.stop(this.effectIds[i]);
        }
        this.effectIds = [];
        for (let key in this.effectCallbacks) {
            let cb = this.effectCallbacks[key];
            Tools.InvokeCallback(cb);
        }
        this.effectCallbacks = {};
    }
}

AudioController.prototype._onMusicFinish = function () {
    this.audio = null;
    this.audioName = '';
}

AudioController.prototype._onEffectFinish = function (id) {
    _.remove(this.effectIds, function (n) {
        return n == id;
    });
    let cb = this.effectCallbacks[id];
    Tools.InvokeCallback(cb);
    delete this.effectCallbacks[id];
}

module.exports = new AudioController();