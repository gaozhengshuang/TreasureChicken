let Game = require('../../Game');
cc.Class({
    extends: cc.Component,

    properties: {
        playerNode: { default: null, type: cc.Node },
        playerInfo: { default: null },
        playerSkeleton: { default: null, type: sp.Skeleton },
    },

    onLoad() {
    },

    start() {
        this._playRunAction();
    },

    update(dt) {
    },

    onDestroy() {
    },

    UpdateInfo(info) {
        this.playerInfo = info;
    },
    _playRunAction() {
        this.playerSkeleton.setAnimation(0, 'run1', true);
    },
    _playDieAction(callback) {
        this.playerSkeleton.setAnimation(0, 'dead1', false);
        if (Game._.isFunction(callback)) {
            this.node.runAction(cc.sequence([
                cc.delayTime(2),
                cc.callFunc(callback)
            ]))
        }
    }
});
