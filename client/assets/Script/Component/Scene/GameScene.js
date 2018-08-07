let Game = require('../../Game');
let PlayerView = require('../View/PlayerView');
const bgHeight = 1100;
const moveSpeed = 300;
cc.Class({
    extends: cc.Component,

    properties: {
        bg1: { default: null, type: cc.Node },
        bg2: { default: null, type: cc.Node },
        resultButton: { default: [], type: [cc.Button] },
        result: { default: 1, type: cc.Integer },
        playerViewPrefab: { default: null, type: cc.Prefab },
        playerViews: { default: [], type: [PlayerView] },
        selfPlayerView: { default: null, type: PlayerView },
        playerViewParent: { default: null, type: cc.Node },
        leftNode: { default: null, type: cc.Node },
        rightNode: { default: null, type: cc.Node },
        countDownLabel: { default: null, type: cc.Label },
        questionLabel: { default: null, type: cc.Label },
        lastLabel: { default: null, type: cc.Label },
        redpackNode: { default: null, type: cc.Node },
    },

    onLoad() {
    },

    start() {
        //轮播背景
        this.bg1.y = 0;
        this.bg2.y = -bgHeight;
        this.bg1.runAction(cc.repeatForever(cc.sequence([
            cc.moveTo((bgHeight - this.bg1.y) / moveSpeed, 0, bgHeight),
            cc.callFunc(function () {
                this.bg1.y = -bgHeight;
            }, this),
            cc.moveTo((this.bg1.y + bgHeight) / moveSpeed, 0, 0),
        ])));
        this.bg2.runAction(cc.repeatForever(cc.sequence([
            cc.moveTo((bgHeight - this.bg2.y) / moveSpeed, 0, bgHeight),
            cc.callFunc(function () {
                this.bg2.y = -bgHeight;
            }, this),
        ])));
        this.redpackNode.rotation = -10;
        this.redpackNode.runAction(cc.repeatForever(
            cc.sequence([
                cc.rotateTo(0.18, 10),
                cc.rotateTo(0.18, -10),
            ])
        ));

        this._updateResultButton();
        let node = cc.instantiate(this.playerViewPrefab);
        node.position = this._randomLeftPos();
        this.playerViewParent.addChild(node);
        this.selfPlayerView = node.getComponent(PlayerView);
        this.selfPlayerView.UpdateAnswer(this.result);
    },

    update(dt) {
    },

    onDestroy() {
    },

    onResultClick(event, customData) {
        customData = parseInt(customData);
        this.result = customData;
        this._updateResultButton();
        let targetPos = cc.Vec2.ZERO;
        if (customData == Game.ChickenDefine.GAME_RESULT.RESULT_RIGHT) {
            targetPos = this._randomLeftPos();
        } else {
            targetPos = this._randomRightPos();
        }
        this.selfPlayerView.UpdateAnswer(customData);
        this.selfPlayerView.node.stopAllActions();
        this.selfPlayerView.node.runAction(cc.moveTo(0.5, targetPos));

    },
    _updateResultButton() {
        for (let i = 0; i < this.resultButton.length; i++) {
            let btn = this.resultButton[i];
            btn.interactable = !((i + 1) == this.result);
        }
    },
    _randomLeftPos() {
        let x = Game.Tools.GetRandomInt(0, this.leftNode.width) + this.leftNode.x;
        let y = Game.Tools.GetRandomInt(0, this.leftNode.height) + this.leftNode.y;
        return cc.v2(x, y);
    },
    _randomRightPos() {
        let x = Game.Tools.GetRandomInt(0, this.rightNode.width) + this.rightNode.x;
        let y = Game.Tools.GetRandomInt(0, this.rightNode.height) + this.rightNode.y;
        return cc.v2(x, y);
    }
});
