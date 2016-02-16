package route

import (
	"centerserver/tools"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"text/template"

	"github.com/tealeg/xlsx"
)

var tempPath string = "template"

type responeconfig struct {
	TempPath   string
	PathConfig map[string][]string
}

func config(w http.ResponseWriter, r *http.Request) {
	gameid := r.FormValue("gameid")
	path := tempPath + "/" + gameid + "/"
	tempName, err := getServerTempale(path)
	if err != nil {
		http.Error(w, err.Error(), 558)
		return
	}
	G, err := Exist(gameid)
	if err != nil {
		http.Error(w, err.Error(), 556)
		return
	}
	configInfo := responeconfig{path + tempName, G.BaseInfo.Configpath}
	buf, err := json.Marshal(configInfo)
	if err != nil {
		http.Error(w, "Marshal config faild", 556)
		return
	}
	w.Write(buf)
}

func getServerTempale(path string) (string, error) {
	Files, err := ioutil.ReadDir(path)
	if err != nil {
		return "", err
	}
	var list info
	for _, v := range Files {
		if v.IsDir() {
			continue
		}
		if !tools.CheckValidZip(path+v.Name()) || !strings.Contains(v.Name(), ".zip") {
			continue
		}
		list = append(list, v)
	}
	if len(list) == 0 {
		return "", errors.New("can't find valid file.")
	}
	list.Sort()
	return list[0].Name(), nil
}

func install(w http.ResponseWriter, r *http.Request) {
	k := r.FormValue("key")
	gameID := r.FormValue("gameid")
	if k == "" || gameID == "" {
		http.Error(w, "key or gameid can't null", 555)
		return
	}
	G, err := Exist(gameID)
	if err != nil {
		http.Error(w, err.Error(), 556)
		return
	}
	serverID := r.FormValue("serverid")
	if serverID == "" {
		serverID = strings.Split(r.RemoteAddr, ":")[0]
	}
	M := G.BaseInfo.Get(serverID)
	if len(M) == 0 {
		http.Error(w, "can't find variables.", 557)
		return
	}
	G.Templates.ExecuteTemplate(w, k, M)
}

var lock *sync.RWMutex = new(sync.RWMutex)

type installConfig struct {
	Configpath       map[string][]string
	RelationVariable map[string][]string
}

type GameInstallConfig struct {
	BaseInfo  *installConfig
	Templates *template.Template
}

var GameConfigRelationMap map[string]*GameInstallConfig = make(map[string]*GameInstallConfig)

func Exist(id string) (*GameInstallConfig, error) {
	value, ok := GameConfigRelationMap[id]
	if ok {
		return value, nil
	}
	install, err := initConfig(tempPath, id)
	if err != nil {
		return nil, err
	}
	T, err := install.GetTemplates(tempPath + "/" + id + "/config/")
	if err != nil {
		return nil, err
	}
	lock.Lock()
	defer lock.Unlock()
	GameConfigRelationMap[id] = &GameInstallConfig{install, T}
	return GameConfigRelationMap[id], nil
}

func initConfig(Tempconfigpath, id string) (*installConfig, error) {
	path := Tempconfigpath + "/" + id + "/" + "server.xlsx"
	return relationConfig(path)
}

func (self *installConfig) GetTemplates(path string) (*template.Template, error) {
	var list []string
	for k, _ := range self.Configpath {
		k = path + k
		list = append(list, k)
	}
	return template.ParseFiles(list...)
}

func (self *installConfig) Get(match string) map[string]string {
	lock.RLock()
	defer lock.RUnlock()
	value, ok := self.RelationVariable[match]
	m := make(map[string]string)
	if ok {
		for k, v := range self.RelationVariable["relationVariable"] {
			m[v] = value[k]
		}
		return m
	}
	for _, variable := range self.RelationVariable {
		for _, i := range variable {
			if i == match {
				for k, v := range self.RelationVariable["relationVariable"] {
					m[v] = variable[k]
				}
				return m
			}
		}
	}
	return map[string]string{}
}

func relationConfig(path string) (*installConfig, error) {
	//文件类型必须是xlsx格式的.
	file, err := xlsx.OpenFile(path)
	if err != nil {
		return nil, err
	}
	if len(file.Sheets) < 2 {
		return nil, errors.New("config format error.")
	}
	configpath, ok := file.Sheet["configpath"]
	if !ok {
		return nil, errors.New("can't find configpath sheet.")
	}
	configPath := make(map[string][]string)
	for _, row := range configpath.Rows {
		if len(row.Cells) < 2 {
			continue
		}
		var list []string
		for _, v := range row.Cells[1:] {
			if strings.TrimSpace(v.Value) == "" {
				continue
			}
			list = append(list, v.Value)
		}
		if len(list) == 0 {
			continue
		}
		configPath[row.Cells[0].Value] = list
	}
	variable, ok := file.Sheet["variable"]
	if !ok {
		return nil, errors.New("can't find variable sheet.")
	}
	if len(variable.Rows) < 1 {
		return nil, errors.New("config format error.")
	}
	relationVariable := make(map[string][]string)
	var list []string
	for _, v := range variable.Rows[0].Cells {
		list = append(list, v.Value)
	}
	//变量关系表中不能出现relationVariable,不然会替换掉key的值.
	relationVariable["relationVariable"] = list
	for _, row := range variable.Rows[1:] {
		if len(row.Cells) != len(variable.Rows[0].Cells) {
			continue
		}
		var list []string
		for _, cell := range row.Cells {
			list = append(list, cell.Value)
		}
		relationVariable[row.Cells[0].Value] = list
	}
	lock.Lock()
	defer lock.Unlock()
	var gameConfig *installConfig = new(installConfig)
	gameConfig.Configpath = configPath
	gameConfig.RelationVariable = relationVariable
	return gameConfig, nil
}
