package client

type updateResult struct {
	Path  string   `json:path`
	Files []string `json:files`
}

type clientConfig struct {
	MasteUrl    string `json:masteurl`
	RequestMode string `json:requestmode`
	GameId      string `json:gameid`
	TmpDir      string `json:tmpdir`
	ProgramHome string `json:programhome`
	CheckMd5    bool   `json:checkmd5`
	Primary     string `json:primary`
	LogPath     string `json:logpath`
	LogLevel    int    `json:loglevel`
}

//服务端响应配置文件的格式
type responeconfig struct {
	TempPath   string
	PathConfig map[string][]string
}
