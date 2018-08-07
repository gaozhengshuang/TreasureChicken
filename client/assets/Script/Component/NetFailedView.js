let Game = require('../Game');

cc.Class({
    extends: cc.Component,

    properties: {
    },
    onLoad() {
        Game.NotificationController.On(Game.Define.EVENT_KEY.CONNECT_TO_GATESERVER, this, this.onClosePanel);
    },
    start() {
    },
    update(dt) {
    },
    onDestroy() {
        Game.NotificationController.Off(Game.Define.EVENT_KEY.CONNECT_TO_GATESERVER, this, this.onClosePanel);
    },
    onClickResetNet() {
        if (Game.Platform.PLATFORM == 'Normal') {
            Game.LoginController.ConnectToLoginServer(function () {
                Game.NetWorkController.Send('msg.C2L_ReqLogin', Game.UserModel.loginInfo);
            }.bind(this));
        } else {
            Game.UserModel.GetUser(function (usr) {
                let loginInfo = { token: usr.token, account: usr.tvmid, face: usr.avatar, nickname: usr.nickname  }
                Game.UserModel.loginInfo = loginInfo;
                Game.LoginController.ConnectToLoginServer(function () {
                    Game.NetWorkController.Send('msg.C2L_ReqLogin', loginInfo);
                }.bind(this));
            });
        }
    },
    onClosePanel() {
        this.node.destroy();
    }
});
