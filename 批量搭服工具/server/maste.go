package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"tool"
)

type Config struct {
	Listen            string `json:listen`
	Port              string `json:port`
	TemplateName      string `json:templatename`
	CfgTempRelational string `json:cfgtemprelational`
	RelationalTable   string `json:relationaltable`
}

var config Config

func init() {
	fmt.Printf("%s Start init config.\n", tool.GetNow())
	c := Flag()
	ParseConfig(c)
	fmt.Printf("%s Parse config ok.\n%s\n", tool.GetNow(), config)
	str := tool.Md5(fmt.Sprintf("template/%s", config.TemplateName))
	if len(str) <= 0 {
		os.Exit(1)
	}
	tool.TemplateInfo.Md5 = str
	tool.TemplateInfo.Name = config.TemplateName
	fmt.Printf("%s start parse template relational \n",tool.GetNow())
	tool.GetPathConfig(config.CfgTempRelational)
	fmt.Printf("%s parse template relational ok\n",tool.GetNow())
	tool.ParseServerConfig(config.RelationalTable)
	tool.TemplateInfo.ConfigList = tool.ConfigTemplate
}

func main() {
	tool.Server(fmt.Sprintf("%s:%s", config.Listen, config.Port))
}

func ParseConfig(path string) {
	buf, err := tool.ReadFile(path)
	if err != nil {
		fmt.Printf("%s Open %s error.", tool.GetNow(), path)
		os.Exit(1)
	}
	err = json.Unmarshal(buf, &config)
	if err != nil {
		fmt.Printf("%s Parse %s faild.\n%s\n", tool.GetNow(), path, err)
		os.Exit(1)
	}
}

func Flag() string {
	configPath := flag.String("c", "cfg.json", "Specify config path.")
	flag.Parse()
	return *configPath
}
