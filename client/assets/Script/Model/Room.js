let _ = require('lodash');
let Define = require('../Util/Define');
let ChickenDefine = require('../Util/ChickenDefine');
let Tools = require('../Util/Tools');
let NetWorkController = require('../Controller/NetWorkController');
let NotificationController = require('../Controller/NotificationController');
let GameController = require('../Controller/GameController');

var RoomModel = function () {
    this.sumreward = 0;
    this.members = [];
    this.round = 0;
    this.coldDownTime = 0;
    this.question = '';
    this.answer = 0;
    this.startTime = 0;
    this.left = 0;
}

RoomModel.prototype.Init = function (cb) {
    NetWorkController.AddListener('msg.GW2C_UpdateRoomInfo', this, this.onGW2C_UpdateRoomInfo);
    NetWorkController.AddListener('msg.GW2C_StartGame', this, this.onGW2C_StartGame);
    NetWorkController.AddListener('msg.GW2C_QuestionInfo', this, this.onGW2C_QuestionInfo);
    NetWorkController.AddListener('msg.GW2C_AnswerInfo', this, this.onGW2C_AnswerInfo);
    NetWorkController.AddListener('msg.GW2C_GameOver', this, this.onGW2C_GameOver);
    NetWorkController.AddListener('msg.GW2C_JoinOk', this, this.onGW2C_JoinOk);
    Tools.InvokeCallback(cb);
}

/**
 * 对外接口
 */
RoomModel.prototype.RestartGame = function () {
    this.sumreward = 0;
    this.members = [];
    this.round = 0;
    this.coldDownTime = 0;
    this.question = '';
    this.answer = 0;
    this.startTime = 0;
    this.left = 0;
}
/**
 * 消息处理接口
 */
RoomModel.prototype.onGW2C_UpdateRoomInfo = function (msgid, data) {
    console.log(data);
    this.sumreward = data.sumreward;
    let newList = [];
    let updateList = [];
    let removeList = [];
    for (let i = 0; i < data.members.length; i++) {
        let member = _.cloneDeep(data.members[i]);
        let index = _.findIndex(this.members, { uid: member.uid });
        if (index == -1) {
            this.members.push(member);
            newList.push(member);
        } else {
            let oldInfo = this.members[index];
            if (oldInfo.answer != member.answer) {
                this.members[index] = member;
                updateList.push(member);
            }
        }
    }

    removeList = _.remove(this.members, function (o) {
        return _.findIndex(data.members, { uid: o.uid }) == -1;
    });

    NotificationController.Emit(Define.EVENT_KEY.ROOMINFO_UPDATEINFO, newList, updateList, removeList);
}

RoomModel.prototype.onGW2C_StartGame = function (msgid, data) {
    NotificationController.Emit(Define.EVENT_KEY.ROOMINFO_STARTGAME);
}

RoomModel.prototype.onGW2C_QuestionInfo = function (msgid, data) {
    this.question = data.txt;
    this.round = data.round;
    this.coldDownTime = data.time;
    this.left = data.left;
    GameController.ChangeState(ChickenDefine.GAME_STATE.STATE_ASKING);
    NotificationController.Emit(Define.EVENT_KEY.ROOMINFO_UPDATEQUESTION);

}
RoomModel.prototype.onGW2C_AnswerInfo = function (msgid, data) {
    this.answer = data.answer;
    GameController.ChangeState(ChickenDefine.GAME_STATE.STATE_FILTER);
    NotificationController.Emit(Define.EVENT_KEY.ROOMINFO_UPDATEANSWER, data.delids || []);
}
RoomModel.prototype.onGW2C_GameOver = function (msgid, data) {
    NotificationController.Emit(Define.EVENT_KEY.ROOMINFO_GAMEOVER, data.reward);
}

RoomModel.prototype.onGW2C_JoinOk = function (msgid, data) {
    this.startTime = data.starttime;
}

module.exports = new RoomModel();