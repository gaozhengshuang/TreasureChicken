let Game = require('../../Game');
cc.Class({
    extends: cc.Component,

    properties: {
        goldLabel: { default: null, type: cc.Label },
        playButtonSprite: { default: null, type: cc.Sprite },
        buttonSpriteFrames: { default: [], type: [cc.SpriteFrame] },
        bets: { default: 1 }
    },

    // LIFE-CYCLE CALLBACKS:

    onLoad() {
        Game.NetWorkController.AddListener('msg.GW2C_JoinOk', this, this.onGW2C_JoinOk);
        Game.NotificationController.On(Game.Define.EVENT_KEY.USERINFO_UPDATEYUANBAO, this, this.onUpdateCostCurrency);
        Game.NotificationController.On(Game.Define.EVENT_KEY.USERINFO_UPDATECOINS, this, this.onUpdateCostCurrency);
    },

    start() {
        Game.AudioController.StopAllEffect();
        Game.AudioController.SetMusicVolume(1);
        Game.AudioController.PlayMusic('Audio/hall');
        this.onUpdateCostCurrency();
    },

    update(dt) {

    },

    onDestroy() {
        Game.NetWorkController.RemoveListener('msg.GW2C_JoinOk', this, this.onGW2C_JoinOk);
        Game.NotificationController.Off(Game.Define.EVENT_KEY.USERINFO_UPDATEYUANBAO, this, this.onUpdateCostCurrency);
        Game.NotificationController.Off(Game.Define.EVENT_KEY.USERINFO_UPDATECOINS, this, this.onUpdateCostCurrency);
    },
    onUpdateCostCurrency() {
        this.goldLabel.string = Game.UserModel.GetCostCurrency();
        if (Game.UserModel.GetCostCurrency() >= 1000) {
            this.bets = 1;
        } else {
            this.bets = 0;
        }
        this._updatePlayButton();
    },
    onStartGame(event) {
        Game.RoomModel.RestartGame();
        Game.GameController.RestartGame();
        switch (this.bets) {
            case 0: {
                Game.GameController.bets = 100;
                break;
            }
            case 1: {
                Game.GameController.bets = 1000;
                break;
            }
            case 2: {
                Game.GameController.bets = 3000;
                break;
            }
            default:
                break;
        }
        Game.NetWorkController.Send('msg.C2GW_JoinGame', { type: this.bets });
    },
    onGW2C_JoinOk() {
        //收到自己的消息了 进入游戏吧
        cc.director.loadScene("GameScene");
    },
    _updatePlayButton() {
        this.playButtonSprite.spriteFrame = this.buttonSpriteFrames[this.bets];
    }
});
