import Define from '../Util/Define';
import Tools from '../Util/Tools';

import NotificationController from '../Controller/NotificationController';

var GameController = function () {
    this.state = 0;
}

GameController.prototype.Init = function (cb) {
    Tools.InvokeCallback(cb, null);
}

GameController.prototype.RestartRound = function () {
}

GameController.prototype.ChangeState = function (state) {
    if (this.state == state) {
        return;
    }
    this.state = state;
    NotificationController.Emit(Define.EVENT_KEY.CHANGE_GAMESTATE, state);
}

module.exports = new GameController();