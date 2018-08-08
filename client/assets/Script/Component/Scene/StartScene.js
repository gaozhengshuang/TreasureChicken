let Game = require('../../Game');
cc.Class({
    extends: cc.Component,

    properties: {
    },

    // LIFE-CYCLE CALLBACKS:

    onLoad() {
        Game.NetWorkController.AddListener('msg.GW2C_JoinOk', this, this.onGW2C_JoinOk);
    },

    start() {

    },

    update(dt) {

    },

    onDestroy() {
        Game.NetWorkController.RemoveListener('msg.GW2C_JoinOk', this, this.onGW2C_JoinOk);
    },

    onStartGame(event, customData) {
        Game.RoomModel.RestartGame();
        Game.GameController.RestartGame();
        Game.NetWorkController.Send('msg.C2GW_JoinGame', { type: parseInt(customData) });
    },
    onGW2C_JoinOk() {
        //收到自己的消息了 进入游戏吧
        cc.director.loadScene("GameScene");
    }
});
