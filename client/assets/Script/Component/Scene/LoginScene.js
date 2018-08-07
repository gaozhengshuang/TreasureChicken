let Game = require('../../Game');

cc.Class({
    extends: cc.Component,

    properties: {
        accEditBox: { default: null, type: cc.EditBox, },
        targetCanvas: { default: null, type: cc.Canvas }
    },

    // LIFE-CYCLE CALLBACKS:

    onLoad() {
        Game.NotificationController.On(Game.Define.EVENT_KEY.CONNECT_TO_GATESERVER, this, this.onLoginComplete);
    },

    start() {

    },

    update(dt) {

    },

    onDestroy() {
        Game.NotificationController.Off(Game.Define.EVENT_KEY.CONNECT_TO_GATESERVER, this, this.onLoginComplete);
    },

    onStartGame() {
        if (this.accEditBox.string == '') {
            return;
        }
        this.onLoginPlatfrom();
    },

    onLoginPlatfrom() {
        let loginInfo = {
            account: this.accEditBox.string,
            nickname: this.accEditBox.string,
            face: '',
            token: ''
        }
        Game.UserModel.loginInfo = loginInfo;
        Game.LoginController.ConnectToLoginServer(function () {
            Game.NetWorkController.Send('msg.C2L_ReqLogin', loginInfo);
        }.bind(this));
    },

    onLoginComplete(msgid, data) {
        cc.director.loadScene("StartScene");
    },
});
