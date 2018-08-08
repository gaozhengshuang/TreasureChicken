let Game = require('../../Game');
cc.Class({
    extends: cc.Component,

    properties: {
    },

    // LIFE-CYCLE CALLBACKS:

    onLoad() {

    },

    start() {

    },

    update(dt) {

    },

    onDestroy() {
        Game.NotificationController.Off(Game.Define.EVENT_KEY.ROOMINFO_UPDATEINFO, this, this.onUpdateRoomInfo);
    },

    onStartGame() {
        Game.RoomModel.RestartGame();
        Game.GameController.RestartGame();
        Game.NetWorkController.Send('msg.C2GW_JoinGame', { type: 1 });
        Game.NotificationController.On(Game.Define.EVENT_KEY.ROOMINFO_UPDATEINFO, this, this.onUpdateRoomInfo);
    },
    onUpdateRoomInfo(newList, updateList) {
        for (let i = 0; i < newList.length; i++) {
            let info = newList[i];
            if (info.uid == Game.UserModel.GetUserId()) {
                //收到自己的消息了 进入游戏吧
                cc.director.loadScene("GameScene");
            }
        }
    }
});
