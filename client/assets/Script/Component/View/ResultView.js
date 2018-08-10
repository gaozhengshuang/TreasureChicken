cc.Class({
    extends: cc.Component,

    properties: {
        canTrig: { default: false },
        rewardLabel: { default: null, type: cc.Label }
    },

    onLoad() {
        this.node.on(cc.Node.EventType.MOUSE_UP, this.onMouseUp, this);
        this.node.on(cc.Node.EventType.TOUCH_END, this.onMouseUp, this);
    },

    onDestroy() {
        this.node.off(cc.Node.EventType.MOUSE_UP, this.onMouseUp, this);
        this.node.off(cc.Node.EventType.TOUCH_END, this.onMouseUp, this);
    },

    onMouseUp() {
        //要等动画播放完
        //返回到大厅吧 
        if (this.canTrig) {
            cc.director.loadScene("StartScene");
            this.canTrig = false;
        }
    },
    Init(reward) {
        if (this.rewardLabel != null) {
            this.rewardLabel.string = reward;
        }
    },
    Update(canTrig) {
        this.canTrig = canTrig;
    }
});
