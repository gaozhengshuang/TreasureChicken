let Game = require('../../Game');
let PlayerView = require('../View/PlayerView');
let PlayerGroupView = require('../View/PlayerGroupView');
const bgHeight = 1100;
const moveSpeed = 300;
cc.Class({
    extends: cc.Component,

    properties: {
        bg1: { default: null, type: cc.Node },
        bg2: { default: null, type: cc.Node },
        resultButton: { default: [], type: [cc.Button] },
        result: { default: 1, type: cc.Integer },
        playerViewPrefab: { default: null, type: cc.Prefab },
        playerViews: { default: [], type: [PlayerView] },
        selfPlayerView: { default: null, type: PlayerView },
        playerViewParent: { default: null, type: cc.Node },
        leftPlayerGroupView: { default: null, type: PlayerGroupView },
        rightPlayerGroupView: { default: null, type: PlayerGroupView },
        countDownLabel: { default: null, type: cc.Label },
        questionLabel: { default: null, type: cc.Label },
        lastLabel: { default: null, type: cc.Label },
        coldDownNum: { default: 0, type: cc.Integer },
        matchingNode: { default: null, type: cc.Node },
        sumRewardLabel: { default: null, type: cc.Label }
    },

    onLoad() {
    },

    start() {
        //轮播背景
        this.bg1.y = 0;
        this.bg2.y = -bgHeight;
        this.bg1.runAction(cc.repeatForever(cc.sequence([
            cc.moveTo((bgHeight - this.bg1.y) / moveSpeed, 0, bgHeight),
            cc.callFunc(function () {
                this.bg1.y = -bgHeight;
            }, this),
            cc.moveTo((this.bg1.y + bgHeight) / moveSpeed, 0, 0),
        ])));
        this.bg2.runAction(cc.repeatForever(cc.sequence([
            cc.moveTo((bgHeight - this.bg2.y) / moveSpeed, 0, bgHeight),
            cc.callFunc(function () {
                this.bg2.y = -bgHeight;
            }, this),
        ])));
        //初始化玩家咯
        for (let i = 0; i < Game.RoomModel.members.length; i++) {
            let member = Game.RoomModel.members[i];
            let playerView = this._createPlayerView(member);
            if (member.uid == Game.UserModel.GetUserId()) {
                this.selfPlayerView = playerView
                this.result = member.answer;
            } else {
                this.playerViews.push(playerView);
            }
        }
        this.lastLabel.string = '剩余' + (this.playerViews.length + 1) + '人';
        this.sumRewardLabel.string = '总奖池\n' + Game.RoomModel.sumreward + '金币';
        this._updateResultButton();
        this.onGameStateChange(Game.GameController.state);
        Game.NotificationController.On(Game.Define.EVENT_KEY.ROOMINFO_UPDATEINFO, this, this.onUpdateRoomInfo);
        Game.NotificationController.On(Game.Define.EVENT_KEY.ROOMINFO_UPDATEQUESTION, this, this.onUpdateQuestion);
        Game.NotificationController.On(Game.Define.EVENT_KEY.ROOMINFO_GAMEOVER, this, this.onGameOver);
        Game.NotificationController.On(Game.Define.EVENT_KEY.ROOMINFO_UPDATEANSWER, this, this.onUpdateAnswer);
        Game.NotificationController.On(Game.Define.EVENT_KEY.CHANGE_GAMESTATE, this, this.onGameStateChange);
        Game.NotificationController.On(Game.Define.EVENT_KEY.CONNECT_TO_GATESERVER, this, this.onLoginComplete);
    },

    update(dt) {
        if (Game.GameController.state == Game.ChickenDefine.GAME_STATE.STATE_ANSWERING || Game.GameController.state == Game.ChickenDefine.GAME_STATE.STATE_ASKING) {
            let lastTime = this._getLastTime();
            if (lastTime > 0) {
                this.countDownLabel.string = lastTime
            } else {
                this.countDownLabel.string = '';
            }
        } else if (Game.GameController.state == Game.ChickenDefine.GAME_STATE.STATE_PENDING) {
            let lastToStartTime = this._getLastToStartTime();
            if (lastToStartTime >= 0) {
                this.countDownLabel.string = lastToStartTime
            } else {
                this.matchingNode.active = false;
                this.countDownLabel.string = '';
            }
        } else {
            this.countDownLabel.string = '';
        }
    },

    onDestroy() {
        Game.NotificationController.Off(Game.Define.EVENT_KEY.ROOMINFO_UPDATEINFO, this, this.onUpdateRoomInfo);
        Game.NotificationController.Off(Game.Define.EVENT_KEY.ROOMINFO_UPDATEQUESTION, this, this.onUpdateQuestion);
        Game.NotificationController.Off(Game.Define.EVENT_KEY.ROOMINFO_GAMEOVER, this, this.onGameOver);
        Game.NotificationController.Off(Game.Define.EVENT_KEY.ROOMINFO_UPDATEANSWER, this, this.onUpdateAnswer);
        Game.NotificationController.Off(Game.Define.EVENT_KEY.CHANGE_GAMESTATE, this, this.onGameStateChange);
        Game.NotificationController.Off(Game.Define.EVENT_KEY.CONNECT_TO_GATESERVER, this, this.onLoginComplete);
    },

    onResultClick(event, customData) {
        customData = parseInt(customData);
        this.result = customData;
        this._updateResultButton();
        this._updatePlayerView(this.selfPlayerView, customData, true);
        Game.NetWorkController.Send('msg.C2GW_Answer', { answer: customData });
    },
    onUpdateRoomInfo(newList, updateList) {
        for (let i = 0; i < newList.length; i++) {
            let info = newList[i];
            let playerView = this._createPlayerView(info);
            if (info.uid == Game.UserModel.GetUserId()) {
                this.selfPlayerView = playerView
            } else {
                this.playerViews.push(playerView);
            }
        }
        for (let i = 0; i < updateList.length; i++) {
            let info = updateList[i];
            if (info.uid != Game.UserModel.GetUserId()) {
                //更新咯
                let playerView = Game._.find(this.playerViews, function (o) {
                    return o.playerInfo.uid == info.uid;
                });
                if (playerView) {
                    playerView.UpdateInfo(info, info.uid == Game.UserModel.GetUserId() ? 'http://jump.cdn.giantfun.cn/cdn/jumphead/tx%20(36).jpg' : null);
                    this._updatePlayerView(playerView, info.answer);
                }
            }
        }
        this.lastLabel.string = '剩余' + (this.playerViews.length + 1) + '人';
        this.sumRewardLabel.string = '总奖池\n' + Game.RoomModel.sumreward + '金币';
    },
    onUpdateQuestion() {
        this.questionLabel.string = Game.RoomModel.question;
        this.coldDownNum = Game.RoomModel.coldDownTime;
        this.questionLabel.node.opacity = 0;
        this.questionLabel.node.runAction(cc.sequence([
            cc.fadeTo(0.5, 255),
            cc.callFunc(function () {
                Game.GameController.ChangeState(Game.ChickenDefine.GAME_STATE.STATE_ANSWERING);
            }, this)
        ]));
    },
    onGameOver() {
        //获奖咯，现在直接跳回开始界面吧
        cc.director.loadScene("StartScene");
    },
    onGameFail() {
        //失败咯
        cc.director.loadScene("StartScene");
    },
    onUpdateAnswer(delids) {
        //看看自己有没有被压死
        for (let i = 0; i < delids.length; i++) {
            let id = delids[i];
            if (id == Game.UserModel.GetUserId()) {
                this.selfPlayerView._playDieAction(this.onGameFail.bind(this));
            } else {
                let playerView = Game._.find(this.playerViews, function (o) {
                    return o.playerInfo.uid == id;
                });
                if (playerView) {
                    playerView._playDieAction();
                }
                Game._.remove(this.playerViews, function (o) {
                    return o.playerInfo.uid == id;
                });
            }
        }
        if (Game.RoomModel.answer == Game.ChickenDefine.GAME_RESULT.RESULT_RIGHT) {
            this.rightPlayerGroupView.ClearGroup();
        } else {
            this.leftPlayerGroupView.ClearGroup();
        }
        //TODO 从this.playerViews中移除
        this.lastLabel.string = '剩余' + (this.playerViews.length + 1) + '人';
    },
    onGameStateChange(state) {
        if (Game.ChickenDefine.GAME_STATE.STATE_ANSWERING == state) {
            //显示选择按钮
            this._activeResultButton(true);
        } else {
            this._activeResultButton(false);
        }

        if (Game.ChickenDefine.GAME_STATE.STATE_PENDING != state) {
            this.matchingNode.active = false;
        }
    },
    onLoginComplete() {
        cc.director.loadScene("StartScene");
    },
    _updateResultButton() {
        for (let i = 0; i < this.resultButton.length; i++) {
            let btn = this.resultButton[i];
            btn.interactable = !((i + 1) == this.result);
        }
    },
    _activeResultButton(active) {
        for (let i = 0; i < this.resultButton.length; i++) {
            let btn = this.resultButton[i];
            btn.node.active = (active && Game.GameController.state == Game.ChickenDefine.GAME_STATE.STATE_ANSWERING);
        }
    },
    _createPlayerView(info) {
        let node = cc.instantiate(this.playerViewPrefab);
        let playerView = node.getComponent(PlayerView);
        let index = 0;
        let targetPos = cc.Vec2.ZERO;
        playerView.UpdateInfo(info, info.uid == Game.UserModel.GetUserId() ? 'http://jump.cdn.giantfun.cn/cdn/jumphead/tx%20(36).jpg' : null);
        if (info.answer == Game.ChickenDefine.GAME_RESULT.RESULT_RIGHT) {
            index = this.leftPlayerGroupView.EnterGroup(playerView);
            targetPos = this.leftPlayerGroupView.GetPositionByIndex(index);
            console.log('左边第' + index + '个');
        } else {
            index = this.rightPlayerGroupView.EnterGroup(playerView);
            targetPos = this.rightPlayerGroupView.GetPositionByIndex(index);
            console.log('右边第' + index + '个');
        }
        node.setLocalZOrder(-index);
        node.position = this._randomOutScreenPos();
        node.runAction(cc.moveTo(1, targetPos));
        this.playerViewParent.addChild(node);

        return playerView;
    },
    _updatePlayerView(playerView, answer, hideButtons = false) {
        let index = 0;
        let targetPos = cc.Vec2.ZERO;
        if (answer == Game.ChickenDefine.GAME_RESULT.RESULT_RIGHT) {
            //现在选的是对 那从错跑过来
            this.rightPlayerGroupView.LeaveGroup(playerView);
            index = this.leftPlayerGroupView.EnterGroup(playerView);
            targetPos = this.leftPlayerGroupView.GetPositionByIndex(index);
        } else {
            this.leftPlayerGroupView.LeaveGroup(playerView);
            index = this.rightPlayerGroupView.EnterGroup(playerView);
            targetPos = this.rightPlayerGroupView.GetPositionByIndex(index);
        }
        playerView.node.setLocalZOrder(-index);
        playerView.node.stopAllActions();
        playerView.node.runAction(
            cc.sequence([
                cc.callFunc(function () {
                    // if (hideButtons) {
                    //     this._activeResultButton(false);
                    // }
                }, this),
                cc.moveTo(0.5, targetPos),
                cc.callFunc(function () {
                    // if (hideButtons) {
                    //     this._activeResultButton(true);
                    // }
                }, this),
            ])
        );
    },
    _randomOutScreenPos() {
        let x = 0;
        let y = 0;
        if (Game.Tools.GetRandomResult()) {
            //左右
            if (Game.Tools.GetRandomResult()) {
                //左
                x = -Game.GameController.winWidth / 2 - 100;
            } else {
                //右
                x = Game.GameController.winWidth / 2 + 100;
            }
            y = Game.Tools.GetRandomInt(-Game.GameController.winHeight / 2, Game.GameController.winHeight / 2);
        } else {
            //下
            y = -Game.GameController.winHeight / 2 - 100;
            x = Game.Tools.GetRandomInt(-Game.GameController.winWidth / 2, Game.GameController.winWidth / 2);
        }
        return cc.v2(x, y);
    },
    _getLastTime() {
        let lastTime = Math.max(0, this.coldDownNum - Game.TimeController.GetCurTime());
        return lastTime;
    },
    _getLastToStartTime() {
        let lastTime = Math.max(-1, Game.RoomModel.startTime - Game.TimeController.GetCurTime());
        return lastTime;
    }
});
