package route

import (
	"centerserver/log"
	"net/http"
)

var (
	LogPath      string   = "http_access.log"
	Log          *log.Log = log.NewLog(LogPath, 0)
	DownLoadPath string   = "download/"
)

func Router(w http.ResponseWriter, r *http.Request) {
	Log.PrintfI("%s%s\n", r.RemoteAddr, r.RequestURI)
	switch r.URL.Path {
	case "/":
		index(w)
	case "/update":
		update(w, r)
	case "/install":
		install(w, r)
	case "/download":
		download(w, r)
	case "/config":
		config(w, r)
	default:
		http.Error(w, "NotFound", 404)
	}
	return
}
