let Game = require('../../Game');
cc.Class({
    extends: cc.Component,

    properties: {
        playerNode: { default: null, type: cc.Node },
        answer: { default: 0, type: cc.Integer },
        playerSkeleton: { default: null, type: sp.Skeleton },
        headSkeleton: { default: null, type: sp.Skeleton },
    },

    onLoad() {
    },

    start() {
        this._playRunAction();
        let head = this.playerSkeleton.findSlot('head');
        console.log(head);
        let headHead = this.headSkeleton.findSlot('head');
        console.log(headHead);
        cc.loader.load('http://jump.cdn.giantfun.cn/cdn/jumphead/tx%20(18).jpg', function (err, res) {
            console.log(res);
            headHead.attachment.region.texture._texture = res;
            head.setAttachment(headHead.attachment);
        });
    },

    update(dt) {
    },

    onDestroy() {
    },
    UpdateAnswer(answer) {
        this.answer = answer;
    },
    _playRunAction() {
        this.playerSkeleton.setAnimation(0, 'run1', true);
    }
});
