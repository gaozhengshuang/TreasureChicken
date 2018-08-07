let Game = require('../../Game');

cc.Class({
    extends: cc.Component,

    properties: {
        loadingCount: { default: 0, type: cc.Integer },
        totalCount: { default: 0, type: cc.Integer },
        loaded: { default: [], type: cc.Boolean },
    },

    onLoad() {
        // cc.director.setDisplayStats(false);
        cc.game.addPersistRootNode(this.node);
        Game.GameInstance = this;
        //初始化游戏
        this.loaded = false;
        this.ctls = [
            Game.ConfigController,
            Game.NotificationController,
            Game.NetWorkController,
            Game.GameController,
            Game.AudioController,
            Game.LoginController,
            Game.ResController,
            Game.TimeController,
        ];

        this.models = [
            Game.UserModel,
            Game.RoomModel
        ];
        this.totalCount = this.ctls.length + this.models.length;
        Game.async.waterfall([
            function (anext) {
                //初始化controller
                Game.async.timesSeries(this.ctls.length, function (n, tnext) {
                    this.loadingCount++;
                    let ctl = this.ctls[n];
                    if (Game._.isFunction(ctl.Init)) {
                        ctl.Init(function (err) {
                            tnext(err);
                        });
                    } else {
                        console.log(ctl);
                        console.log('[警告] 该controller没有init方法');
                        tnext(null);
                    }
                }.bind(this), function (err) {
                    anext(err);
                }.bind(this));
            }.bind(this),
            function (anext) {
                //初始化model
                Game.async.timesSeries(this.models.length, function (n, tnext) {
                    this.loadingCount++;
                    let model = this.models[n];
                    if (Game._.isFunction(model.Init)) {
                        model.Init(function (err) {
                            tnext(err);
                        });
                    } else {
                        console.log(model);
                        console.log('[警告] 该model没有init方法');
                        tnext(null);
                    }
                }.bind(this), function (err) {
                    anext(err);
                }.bind(this));
            }.bind(this),
        ], function (err) {
            if (err) {
                console.log(err);
            } else {
                this.loaded = true;
            }
        }.bind(this));
    },

    start() {
    },

    update(dt) {
        if (this.loaded) {
            for (let i = 0; i < this.ctls.length; i++) {
                let ctl = this.ctls[i];
                if (Game._.isFunction(ctl.Update)) {
                    ctl.Update(dt);
                }
            }
        }
    },
});
