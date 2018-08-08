package main
import (
	"fmt"
	"os"
	"time"
	"math/rand"
	"syscall"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/util"
)


var g_KeyBordInput util.KeyBordInput
var g_Signal *util.Signal = nil
var g_CmdArgs CmdlineArgument

type CmdlineArgument struct {
	Conf        string
	Logname     string
	Logpath     string
	Loglvl		string
	Daemon      bool
	EventStat	bool
	PProf		int64
	Ver			string
}

// 命令行参数解析
func init() {
	fmt.Println("main.init()")

	util.DoCmdLineParse(&g_CmdArgs)
	if g_CmdArgs.Conf == ""         { panic("-----input '-conf' cmdline argu-----")     }
	if g_CmdArgs.Logname == ""      { panic("-----input '-logname' cmdline argu-----")  }
	if g_CmdArgs.Logpath == ""      { panic("-----input '-logpath' cmdline argu-----")  }
	if g_CmdArgs.Loglvl == ""       { panic("-----input '-loglvl' cmdline argu-----")   }
	if g_CmdArgs.PProf == 0 		{ panic("-----input '-pprof' cmdline argu-----")   }
}

//TODO: 多Gate/Room会监听同一个端口
func main() {
	//util.OpenPProf(main_pprof)
	util.OpenNetPProf("localhost", int32(g_CmdArgs.PProf) ,main_pprof)
}

func main_pprof() 	{
//func main() 		{

	// 宕机日志
	defer util.RecoverPanic(nil, nil)

	// 初始随机数
	rand.Seed(time.Now().Unix())

	// 初始化日志
	log.Init(g_CmdArgs.Logpath, g_CmdArgs.Logname, g_CmdArgs.Loglvl)
	defer log.Flush()
	defer log.Info("main quit")
	log.Info("初始化日志ok")
	log.Info("启动参数: %v", os.Args)


	// 信号注册
	g_Signal = util.CreateNewSignal()
	g_Signal.Init()
	g_Signal.Regist(syscall.SIGHUP,  SignalHup)
	g_Signal.Regist(syscall.SIGINT,  SignalInt)
	g_Signal.Regist(syscall.SIGTERM, SignalTerm)
	g_Signal.Regist(syscall.SIGQUIT, SignalCoreDump)
	g_Signal.Regist(syscall.SIGSEGV, SignalCoreDump)
	g_Signal.Regist(syscall.SIGILL,  SignalCoreDump)
	g_Signal.Regist(syscall.SIGBUS,  SignalCoreDump)
	g_Signal.Start()


	// 初始化键盘输入
	g_KeyBordInput.Init()
	if g_CmdArgs.Daemon == false {
		g_KeyBordInput.Start()
	}


	// 初始化配置 
	server := NewGateServer()
	if server.Init(g_CmdArgs.Conf) == false {
		log.Info("初始化配置失败")
		return
	}

	// 初始化Redis
	if server.StartRedis() == false {
		log.Info("初始化Redis失败")
		return
	}

	// 初始化网络
	if server.StartNetWork() == false {
		log.Info("启动网络失败")
		return
	}

	// 服务器启动完成回调
	server.OnStart()

	//主线程
	runMainLoop(server)

	//
	server.OnStop()

	time.Sleep(time.Millisecond * 100)
}

func runMainLoop(server *GateServer) {
	for {
		time.Sleep(time.Microsecond * 1000)
		select {
		case cmd, open := <-g_KeyBordInput.C:
			if cmd == "exit" || cmd == "quit" || open == false {
				log.Info("keybord recv cmd:'%v' open:'%v'", cmd, open)
				server.Quit()
				return
			}
			server.DoInputCmd(cmd)
		default:
			server.Run()
		}
	}
}

var gTicker *time.Ticker = nil
func doEventStatistics(server *GateServer) {
	if g_CmdArgs.EventStat == true {
		if gTicker == nil { gTicker = time.NewTicker(time.Second *1) }
		select {
		case <-gTicker.C:
			if size := server.net.EventQueueSize(); size != 0 {
				log.Info("netevent chan current size = %d", size)
			}
		default:
		}
	}
}
