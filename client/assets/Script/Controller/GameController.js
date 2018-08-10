let Define = require('../Util/Define');
let ChickenDefine = require('../Util/ChickenDefine');
let Tools = require('../Util/Tools');

let NotificationController = require('../Controller/NotificationController');

var GameController = function () {
    this.state = 0;
    this.winWidth = 0;
    this.winHeight = 0;
    this.bets = 0;
}

GameController.prototype.Init = function (cb) {
    this.winWidth = Math.min(cc.winSize.width, 860);
    this.winHeight = Math.min(cc.winSize.height, 1280);
    Tools.InvokeCallback(cb, null);
}

GameController.prototype.RestartGame = function () {
    this.state = ChickenDefine.GAME_STATE.STATE_PENDING;
    this.bets = 0;
}

GameController.prototype.ChangeState = function (state) {
    if (this.state == state) {
        return;
    }
    this.state = state;
    NotificationController.Emit(Define.EVENT_KEY.CHANGE_GAMESTATE, state);
}

module.exports = new GameController();