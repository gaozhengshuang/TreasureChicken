let Game = require('../../Game');

const SkinName = [
    '37fen',
    'pingtou',
    'qiliuhai'
]
cc.Class({
    extends: cc.Component,

    properties: {
        playerNode: { default: null, type: cc.Node },
        playerInfo: { default: null },
        playerSkeleton: { default: null, type: sp.Skeleton },
        playerSprite: { default: null, type: cc.Sprite },
        playerHeadNode: { default: null, type: cc.Node },
        playerNameLabel: { default: null, type: cc.Label }
    },

    onLoad() {
    },

    start() {
        this.playerSkeleton.setSkin(SkinName[Game.Tools.GetRandomInt(0, 3)]);
        this._playRunAction();
    },

    update(dt) {
    },

    onDestroy() {
    },

    UpdateInfo(info, userHeadUrl) {
        this.playerInfo = info;
        this.playerNameLabel.string = info.name;
        this.playerHeadNode.active = (userHeadUrl != null);
        if (userHeadUrl != null) {
            cc.loader.load(userHeadUrl, function (err, res) {
                if (err) {
                    console.log('[严重错误] 加载资源失败 ' + err);
                } else {
                    this.playerSprite.spriteFrame = new cc.SpriteFrame(res);
                }
            }.bind(this));
        }
    },
    _playRunAction() {
        this.playerSkeleton.setAnimation(0, 'run1', true);
    },
    _playDieAction(callback) {
        this.playerSkeleton.setAnimation(0, 'dead3', false);
    }
});
