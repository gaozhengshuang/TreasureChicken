let Game = require('../../Game');

cc.Class({
    extends: cc.Component,

    properties: {
        tipParentNode: { default: null, type: cc.Node },
        notifyPrefab: { default: null, type: cc.Prefab },
        resultPrefab: { default: null, type: cc.Prefab },
        barragePrefab: { default: null, type: cc.Prefab },
        barragePositionIndex: { default: 0, type: cc.Integer },
    },

    onLoad() {
        cc.game.addPersistRootNode(this.node);
    },

    start() {
        Game.NotificationController.On(Game.Define.EVENT_KEY.TIP_TIPS, this, this.onShowTips);
        Game.NotificationController.On(Game.Define.EVENT_KEY.TIP_RESULT, this, this.onShowResult);
        Game.NotificationController.On(Game.Define.EVENT_KEY.TIP_BARRAGE, this, this.onBarrage);
    },

    update(dt) {
    },

    //漂浮提示代码--------------------------------------start-----------------------------------
    onShowTips(data) {
    },
    //漂浮提示代码--------------------------------------end-------------------------------------

    //抓取结果提示代码--------------------------------------start-----------------------------------
    onShowResult(data) {
    },
    //抓取结果提示代码--------------------------------------end-----------------------------------

    //弹幕代码--------------------------------------start-----------------------------------
    onBarrage(data) {
    },
    //弹幕代码--------------------------------------end------------------------------------
});
