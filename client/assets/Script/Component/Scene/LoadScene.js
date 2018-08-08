let Game = require('../../Game');

cc.Class({
    extends: cc.Component,

    properties: {
        percentLabel: { default: null, type: cc.Label },
        loaded: { default: [], type: cc.Boolean },
        targetCanvas: { default: null, type: cc.Canvas },
        loadProgressBar: { default: null, type: cc.ProgressBar },
    },

    onLoad() {
        let canvas = this.targetCanvas;
        let designResolution = canvas.designResolution
        var viewSize = cc.view.getFrameSize()
        if (viewSize.width / viewSize.height > designResolution.width / designResolution.height) {
            canvas.fitHeight = true;
            canvas.fitWidth = false;
        }
        else {
            canvas.fitHeight = false;
            canvas.fitWidth = true
        }

        this.loaded = false;

        Game.NotificationController.On(Game.Define.EVENT_KEY.CONNECT_TO_GATESERVER, this, this.onLoginComplete);
    },

    onDestroy() {
        Game.NotificationController.Off(Game.Define.EVENT_KEY.CONNECT_TO_GATESERVER, this, this.onLoginComplete);
    },

    start() {
    },

    update(dt) {
        if (Game.GameInstance == null || Game.GameInstance.totalCount == 0 || this.loaded) {
            return;
        }
        let percent = (Game.GameInstance.loadingCount / Game.GameInstance.totalCount).toFixed(2);
        this.percentLabel.string = Math.ceil(percent * 100) + '%';
        this.loadProgressBar.progress = percent;
        if (Game.GameInstance.loadingCount == Game.GameInstance.totalCount) {
            //加载完了
            this.loaded = true;
            if (Game.Platform.PLATFORM == 'Normal') {
                cc.director.loadScene("LoginScene");
            } else {
                var tvmTimeout = setTimeout(function () {
                    Game.LoginController._showNetFailed();
                }.bind(this), 2000);
                Game.UserModel.GetUser(function (usr) {
                    clearTimeout(tvmTimeout);
                    if (usr == null) {
                        cc.director.loadScene("LoginScene");
                        return;
                    }
                    let loginInfo = { token: usr.token, account: usr.tvmid, face: usr.avatar, nickname: usr.nickname }
                    Game.UserModel.loginInfo = loginInfo;
                    Game.LoginController.ConnectToLoginServer(function () {
                        Game.NetWorkController.Send('msg.C2L_ReqLogin', loginInfo);
                    }.bind(this));
                }.bind(this));
            }

        }
    },

    onLoginComplete() {
        cc.director.loadScene("GameScene");
    },
});
