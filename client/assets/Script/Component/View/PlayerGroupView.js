let Game = require('../../Game');
const width = 90;
const height = 60;
const gridpreline = 3;
const randomoffset = 7;
cc.Class({
    extends: cc.Component,

    properties: {
    },

    onLoad() {
        this.playerViews = new Array(27);
    },

    start() {
    },

    update(dt) {
    },

    onDestroy() {
    },

    GetPositionByIndex(i) {
        return cc.v2((i % gridpreline) * width + Game.Tools.GetRandomInt(-randomoffset, randomoffset), Math.floor(i / gridpreline) * height).add(this.node.position);
    },

    EnterGroup(playerView) {
        //先看看是不是已经在这里了
        let oldIndex = Game._.findIndex(this.playerViews, function (o) {
            return Game._.get(o, 'playerInfo.uid', 0) == Game._.get(playerView, 'playerInfo.uid', 1);
        });
        if (oldIndex != -1) {
            //已经有这个id了
            console.log('已经有咯 ' + Game._.get(playerView, 'playerInfo.uid', 1) + oldIndex);
            return oldIndex;
        }

        for (let i = 0; i < this.playerViews.length; i++) {
            let info = this.playerViews[i];
            if (info == null) {
                this.playerViews[i] = playerView;
                return i;
            }
        }
    },
    LeaveGroup(playerView) {
        for (let i = 0; i < this.playerViews.length; i++) {
            Game._.get(this.playerViews[i], 'playerInfo.uid')
            if (Game._.get(this.playerViews[i], 'playerInfo.uid', 0) == Game._.get(playerView, 'playerInfo.uid', 1)) {
                this.playerViews[i] = null;
                return i;
            }
        }
    },
    ClearGroup() {
        Game._.fill(this.playerViews, null);
    }
});
