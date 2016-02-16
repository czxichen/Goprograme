package client

import (
	"centerserver/log"
	"flag"
)

var Log *log.Log
var Config clientConfig

var (
	configPath string
	Action     string
	updateDir  string
)

func init() {
	flag.StringVar(&configPath, "f", "cfg.json", "-f cfg.json")
	flag.StringVar(&Action, "a", "", "指定要做的操作 -a update")
	flag.StringVar(&updateDir, "u", "", "指定更新目录 -u 20160101")
	flag.StringVar(&Config.MasteUrl, "m", "127.0.0.1:1789", "指定服务端IP端口 -m 127.0.0.1:1789")
	flag.StringVar(&Config.RequestMode, "r", "https", "指定请求的模式 -r https")
	flag.StringVar(&Config.GameId, "g", "40", "指定游戏的gameid -g 40")
	flag.StringVar(&Config.TmpDir, "t", "tmp", "下载文件的临时目录 -t tmp")
	flag.StringVar(&Config.ProgramHome, "p", "/test", "指定解压的根目录 -p /test")
	flag.BoolVar(&Config.CheckMd5, "c", true, "是否检查文件md5 -c true")
	flag.StringVar(&Config.Primary, "P", "", "指定请求的关键字 -P 7400001")
	flag.StringVar(&Config.LogPath, "l", "run.log", "指定log的文件名.")
	flag.IntVar(&Config.LogLevel, "L", 1, "指定日志的级别 -L 1")
	flag.Parse()
	Log = log.NewLog(Config.LogPath, Config.LogLevel)
	Log.PrintfI("config file parse end:%v\n", Config)
}
