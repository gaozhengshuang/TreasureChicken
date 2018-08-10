let Game = require('../../Game');
let PlayerView = require('../View/PlayerView');
let PlayerGroupView = require('../View/PlayerGroupView');
let ResultView = require('../View/ResultView');
const bgHeight = 1560;
const moveSpeed = 300;
const stoneSpeed = 600;
const playerSpeed = 300;
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
        sumRewardLabel: { default: null, type: cc.Label },
        winResultView: { default: null, type: ResultView },
        loseResultView: { default: null, type: ResultView },
        winNode: { default: null, type: cc.Node },
        loseNode: { default: null, type: cc.Node },
        matchSprite: { default: null, type: cc.Sprite },
        countDownSpriteFrame: { default: [], type: [cc.SpriteFrame] },
        leftQuestionSpriteFrame: { default: [], type: [cc.SpriteFrame] },
        stoneNode: { default: null, type: cc.Node },
        leftLightNode: { default: null, type: cc.Node },
        rightLightNode: { default: null, type: cc.Node },

        startingCountDown: { default: false }
    },

    onLoad() {
        this.stoneStartY = -Game.GameController.winHeight / 2 - 200;
        this.stoneEndY = Game.GameController.winHeight / 2 + 200;
        this.stoneTime = (this.stoneEndY - this.stoneStartY) / stoneSpeed;
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
        this.sumRewardLabel.string = Game.RoomModel.sumreward + '金币';
        this._updateResultButton();
        this.onGameStateChange(Game.GameController.state);
        Game.AudioController.StopAllEffect();
        Game.AudioController.SetMusicVolume(1);
        Game.AudioController.PlayMusic('Audio/game');
        Game.NotificationController.On(Game.Define.EVENT_KEY.ROOMINFO_UPDATEINFO, this, this.onUpdateRoomInfo);
        Game.NotificationController.On(Game.Define.EVENT_KEY.ROOMINFO_UPDATEQUESTION, this, this.onUpdateQuestion);
        Game.NotificationController.On(Game.Define.EVENT_KEY.ROOMINFO_GAMEOVER, this, this.onGameOver);
        Game.NotificationController.On(Game.Define.EVENT_KEY.ROOMINFO_UPDATEANSWER, this, this.onUpdateAnswer);
        Game.NotificationController.On(Game.Define.EVENT_KEY.CHANGE_GAMESTATE, this, this.onGameStateChange);
        Game.NotificationController.On(Game.Define.EVENT_KEY.CONNECT_TO_GATESERVER, this, this.onLoginComplete);
        Game.NetWorkController.AddListener('msg.GW2C_AnswerOk', this, this.onGW2C_AnswerOk);
    },

    update(dt) {
        if (Game.GameController.state == Game.ChickenDefine.GAME_STATE.STATE_ANSWERING || Game.GameController.state == Game.ChickenDefine.GAME_STATE.STATE_ASKING) {
            let lastTime = this._getLastTime();
            if (lastTime >= 0) {
                this.countDownLabel.string = lastTime
            } else {
                this.countDownLabel.string = 0;
            }
            if (lastTime <= 1) {
                this._activeResultButton(false);
            }
        } else if (Game.GameController.state == Game.ChickenDefine.GAME_STATE.STATE_PENDING) {
            let lastToStartTime = this._getLastToStartTime();
            if (lastToStartTime > 0) {
                this.countDownLabel.string = lastToStartTime
            } else {
                if (!this.startingCountDown) {
                    this._runCountDown();
                    this.startingCountDown = true;
                }
                this.matchingNode.active = false;
                this.countDownLabel.string = 0;
            }
        } else {
            this.countDownLabel.string = 0;
        }
    },

    onDestroy() {
        Game.NotificationController.Off(Game.Define.EVENT_KEY.ROOMINFO_UPDATEINFO, this, this.onUpdateRoomInfo);
        Game.NotificationController.Off(Game.Define.EVENT_KEY.ROOMINFO_UPDATEQUESTION, this, this.onUpdateQuestion);
        Game.NotificationController.Off(Game.Define.EVENT_KEY.ROOMINFO_GAMEOVER, this, this.onGameOver);
        Game.NotificationController.Off(Game.Define.EVENT_KEY.ROOMINFO_UPDATEANSWER, this, this.onUpdateAnswer);
        Game.NotificationController.Off(Game.Define.EVENT_KEY.CHANGE_GAMESTATE, this, this.onGameStateChange);
        Game.NotificationController.Off(Game.Define.EVENT_KEY.CONNECT_TO_GATESERVER, this, this.onLoginComplete);
        Game.NetWorkController.RemoveListener('msg.GW2C_AnswerOk', this, this.onGW2C_AnswerOk);
    },

    onResultClick(event, customData) {
        customData = parseInt(customData);
        Game.NetWorkController.Send('msg.C2GW_Answer', { answer: customData });
    },
    onUpdateRoomInfo(newList, updateList, removeList) {
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

        for (let i = 0; i < removeList.length; i++) {
            let info = removeList[i];
            if (info.uid != Game.UserModel.GetUserId()) {
                //掉队咯
                let playerViews = Game._.remove(this.playerViews, function (o) {
                    return o.playerInfo.uid == info.uid;
                });
                for (let j = 0; j < playerViews.length; j++) {
                    let view = playerViews[j];
                    let pos = this._randomOutScreenPos();
                    //强行从两边移除
                    this.leftPlayerGroupView.LeaveGroup(view);
                    this.rightPlayerGroupView.LeaveGroup(view);
                    //跑出界外
                    view.node.runAction(cc.sequence([
                        cc.moveTo(0.6, pos),
                        cc.removeSelf(true)
                    ]))
                }
            }
        }
        this.lastLabel.string = '剩余' + (this.playerViews.length + 1) + '人';
        this.sumRewardLabel.string = Game.RoomModel.sumreward + '金币';
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
        //播放动画
        let spriteFrame = this.leftQuestionSpriteFrame[Game.RoomModel.left];
        if (spriteFrame != null) {
            this.matchSprite.spriteFrame = spriteFrame;
            this.matchSprite.node.runAction(
                cc.sequence([
                    cc.callFunc(function () {
                        this.matchSprite.node.opacity = 0;
                        this.matchSprite.node.scaleX = 3;
                        this.matchSprite.node.scaleY = 3;
                    }, this),
                    cc.spawn([
                        cc.scaleTo(0.3, 1, 1).easing(new cc.easeIn(3)),
                        cc.fadeTo(0.3, 255).easing(new cc.easeIn(3)),
                    ]),
                    cc.spawn([
                        cc.scaleTo(0.4, 0.8, 0.8).easing(new cc.easeIn(2)),
                        cc.fadeTo(0.4, 0).easing(new cc.easeIn(2)),
                    ]),
                ])
            );
        }

    },
    onGameOver(reward) {
        //获奖咯，现在直接跳回开始界面吧
        if (reward == 0) {
            this.node.runAction(cc.sequence([
                cc.delayTime(this.stoneTime * 1.5),
                cc.callFunc(function () {
                    this.onGameFail();
                }, this)
            ]));
            return;
        }
        Game.AudioController.SetMusicVolume(0);
        Game.AudioController.PlayEffect('Audio/win', function () {
            Game.AudioController.SetMusicVolume(1);
        });
        this.winResultView.node.active = true;
        this.winResultView.Init(reward);
        this.winNode.scaleX = 0;
        this.winNode.scaleY = 0;
        this.winNode.runAction(cc.sequence([
            cc.scaleTo(0.3, 1.3, 1.2),
            cc.scaleTo(0.05, 0.8, 0.8),
            cc.scaleTo(0.05, 1.1, 1.1),
            cc.scaleTo(0.05, 1, 1),
            cc.callFunc(function () {
                this.winResultView.Update(true)
            }, this)
        ]));
    },
    onGameFail() {
        //失败咯
        Game.AudioController.SetMusicVolume(0);
        Game.AudioController.PlayEffect('Audio/lose', function () {
            Game.AudioController.SetMusicVolume(1);
        });
        this.loseResultView.node.active = true;
        this.loseResultView.Init(Game.GameController.bets);
        this.loseNode.scaleX = 0;
        this.loseNode.scaleY = 0;
        this.loseNode.runAction(cc.sequence([
            cc.scaleTo(1.5, 1, 1),
            cc.callFunc(function () {
                this.loseResultView.Update(true)
            }, this)
        ]));
    },
    onUpdateAnswer(delids) {
        //播放石头碾压动画
        let x = 0;
        if (Game.RoomModel.answer == Game.ChickenDefine.GAME_RESULT.RESULT_RIGHT) {
            x = 150;
        } else {
            x = -150;
        }
        this.stoneNode.position = cc.v2(x, this.stoneStartY);
        this.stoneNode.runAction(cc.moveTo(this.stoneTime, x, this.stoneEndY));
        //看看自己有没有被压死
        for (let i = 0; i < delids.length; i++) {
            let id = delids[i];
            if (id == Game.UserModel.GetUserId()) {
                this.selfPlayerView.node.stopAllActions();
                this.selfPlayerView.node.runAction(cc.sequence([
                    cc.delayTime(Game.Tools.GetPercent(this.stoneStartY, this.stoneEndY, this.selfPlayerView.node.y) * this.stoneTime),
                    cc.callFunc(function () {
                        this.selfPlayerView._playDieAction();
                    }, this),
                    cc.moveTo((this.stoneEndY - this.selfPlayerView.node.y) / playerSpeed, this.selfPlayerView.node.x, this.stoneEndY),
                ]))
            } else {
                let playerViews = Game._.remove(this.playerViews, function (o) {
                    return o.playerInfo.uid == id;
                });
                for (let j = 0; j < playerViews.length; j++) {
                    let playerView = playerViews[j];
                    if (playerView) {
                        playerView.node.stopAllActions();
                        playerView.node.runAction(cc.sequence([
                            cc.delayTime(Game.Tools.GetPercent(this.stoneStartY, this.stoneEndY, playerView.node.y) * this.stoneTime),
                            cc.callFunc(function () {
                                playerView._playDieAction();
                            }, this),
                            cc.moveTo((this.stoneEndY - playerView.node.y) / playerSpeed, playerView.node.x, this.stoneEndY),
                            cc.removeSelf()
                        ]))
                    }
                }
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
    onGW2C_AnswerOk(msgid, data) {
        console.log(this);
        this.result = data.answer;
        this._updateResultButton();
        this._updatePlayerView(this.selfPlayerView, this.result, true);
        // 播动画
        if (this.result == Game.ChickenDefine.GAME_RESULT.RESULT_RIGHT) {
            this.leftLightNode.runAction(cc.sequence([
                cc.fadeTo(0.2, 255),
                cc.fadeTo(0.3, 0),
            ]));
        } else {
            this.rightLightNode.runAction(cc.sequence([
                cc.fadeTo(0.2, 255),
                cc.fadeTo(0.3, 0),
            ]));
        }
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
    },
    _runCountDown() {
        this.matchSprite.node.active = true;
        this.Indexs = 3;
        this.matchSprite.node.runAction(
            cc.repeat(
                cc.sequence([
                    cc.callFunc(function () {
                        this.matchSprite.node.opacity = 0;
                        this.matchSprite.node.scaleX = 5;
                        this.matchSprite.node.scaleY = 5;
                        this.matchSprite.spriteFrame = this.countDownSpriteFrame[this.Indexs];
                    }, this),
                    cc.spawn([
                        cc.scaleTo(0.3, 1, 1).easing(new cc.easeIn(3)),
                        cc.fadeTo(0.3, 255).easing(new cc.easeIn(3)),
                    ]),
                    cc.spawn([
                        cc.scaleTo(0.3, 0.3, 0.3).easing(new cc.easeIn(2)),
                        cc.fadeTo(0.3, 0).easing(new cc.easeIn(2)),
                    ]),
                    cc.callFunc(function () {
                        this.Indexs--;
                    }, this),
                    cc.delayTime(0.4)
                ]), 4)
        );
    }
});
