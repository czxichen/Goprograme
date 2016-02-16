package server

import (
	"centerserver/log"
	"centerserver/route"
	"net/http"
)

var Log *log.Log

func init() {
	if cfg == nil {
		ParseConfig("cfg.json")
		Log = log.NewLog(GetConfig().LogPath, 0)
		Log.PrintfI("config file parse end:%v\n", GetConfig())
	}
}

func Listen() {
	http.HandleFunc("/", route.Router)
	switch GetConfig().Mode {
	case "http":
		err := http.ListenAndServe(GetConfig().IP, nil)
		if err != nil {
			Log.PrintfE("%s\n", err)
		}
	case "https":
		err := http.ListenAndServeTLS(GetConfig().IP, GetConfig().CertFile,
			GetConfig().KeyFile, nil)
		if err != nil {
			Log.PrintfE("%s\n", err)
		}
	}
	return
}
