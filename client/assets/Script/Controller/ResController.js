let _ = require('lodash');
let Tools = require('../Util/Tools');

var ResController = function () {
    this.spriteFrames = {};
}

ResController.prototype.Init = function (cb) {
    Tools.InvokeCallback(cb);
};

ResController.prototype.GetSpriteFrameByName = function (name, cb) {
    let spriteFrame = this.spriteFrames[name];
    if (spriteFrame == null) {
        cc.loader.loadRes(name, cc.SpriteFrame, function (err, res) {
            if (err) {
                Tools.InvokeCallback(cb, err);
            } else {
                this.spriteFrames[name] = res;
                Tools.InvokeCallback(cb, null, res);
            }
        }.bind(this));
    } else {
        Tools.InvokeCallback(cb, null, spriteFrame);
    }
};

ResController.prototype.DestoryAllChildren = function (node) {
    if (node && node.children) {
        let children = node.children;
        for (let i = 0; i < children.length; i++) {
            let child = children[i];
            if (child) {
                child.destroy();
            }
        }
        node.removeAllChildren(true);
    }
}


module.exports = new ResController();