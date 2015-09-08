package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	_ "github.com/lunny/godbc"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type NxServerState struct {
	GameID         int       `xorm:"not null 'GameID'"`
	IssuerId       int       `xorm:"not null IssuerId"`
	ServerID       int       `xorm:"not null ServerID"`
	ServerName     string    `xorm:"nvarchar 'ServerName'"`
	OnlineNum      int       `xorm:"not null OnlineNum"`
	MaxOnlineNum   int       `xorm:"not null MaxOnlineNum"`
	ServerIP       string    `xorm:"not null ServerIP"`
	Port           int       `xorm:"not null Port"`
	IsRuning       int       `xorm:"not null IsRuning"`
	ServerStyle    int       `xorm:"ServerStyle"`
	IsStartIPWhile int       `xorm:"not null IsStartIPWhile"`
	LogTime        time.Time `xorm:"LogTime"`
	UpdateTime     time.Time `xorm:"UpdateTime"`
	OrderBy        int       `xorm:"not null OrderBy"`
}

var debug bool

func main() {
	flag.BoolVar(&debug, "d", false, "-d=true打印debug信息")
	flag.Parse()

	var readline int
	fmt.Printf("1:备份线上内容到本地.\n2:更新server.xlsx中update的内容\n3:插入server.xlsx中insert的内容\n4:存在更新不存在插入server.xlsx中insertAndupdate的内容\n")
	fmt.Print("输入操作选项：")
	fmt.Scan(&readline)
	File, e := xlsx.OpenFile("server.xlsx")
	if e != nil {
		fmt.Println(e)
		return
	}
	engines := Engines()
	switch readline {
	case 1:
		fmt.Println("备份线上内容到本地")
		getcsv()
	case 2:
		fmt.Println("更新server.xlsx中update的内容")
		update(engines, File.Sheet["update"])
	case 3:
		fmt.Println("插入server.xlsx中insert的内容")
		insert(engines, File.Sheet["insert"])
	case 4:
		fmt.Println("存在更新不存在插入server.xlsx中insertAndupdate的内容")
		insertAndupdate(engines, File.Sheet["insertAndupdate"])
	}
}

func Engines() *xorm.Engine {
	buf, err := ioutil.ReadFile("cfg.conf")
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	str := strings.Trim(string(buf), "")
	Engine, err := xorm.NewEngine("odbc", str)
	if err != nil {
		fmt.Println("新建引擎错误：", err)
		return nil
	}
	if err := Engine.Ping(); err != nil {
		fmt.Println(err)
		return nil
	}
	Engine.SetTableMapper(core.SameMapper{})
	if debug {
		Engine.ShowSQL = true
	}
	return Engine
}

func insertAndupdate(engines *xorm.Engine, File *xlsx.Sheet) {
	for _, row := range File.Rows {
		var list []*xlsx.Cell
		for _, cell := range row.Cells {
			list = append(list, cell)
		}
		date := parse(list)
		if date != nil {
			n, err := engines.Where("IssuerId = ? and ServerID = ?", date.IssuerId, date.ServerID).Update(date)
			if n == 0 || err != nil {
				fmt.Printf("更新:%s出错.\n", fmt.Sprint(*date))
				fmt.Printf("尝试直接插入：%s", fmt.Sprint(*date))
				n, err := engines.Insert(date)
				if n == 0 || err != nil {
					fmt.Printf("插入:%s出错.\n", fmt.Sprint(*date))
					continue
				}
				engines.Query(fmt.Sprintf("UPDATE NxServerState SET ServerName=N'%s' where IssuerId = %d and ServerID = %d", date.ServerName, date.IssuerId, date.ServerID))
				continue
			}
			engines.Query(fmt.Sprintf("UPDATE NxServerState SET ServerName=N'%s' where IssuerId = %d and ServerID = %d", date.ServerName, date.IssuerId, date.ServerID))
		}
	}
}

func insert(engines *xorm.Engine, File *xlsx.Sheet) {
	for _, row := range File.Rows {
		var list []*xlsx.Cell
		for _, cell := range row.Cells {
			list = append(list, cell)
		}
		date := parse(list)
		if date != nil {
			b, err := engines.Where("IssuerId = ? and ServerID = ?", date.IssuerId, date.ServerID).Get(&NxServerState{})
			if b {
				fmt.Println(*date, "  已存在")
				continue
			}
			n, err := engines.Insert(date)
			if n == 0 || err != nil {
				fmt.Printf("插入:%s出错.\n", fmt.Sprint(*date))
				continue
			}
			engines.Query(fmt.Sprintf("UPDATE NxServerState SET ServerName=N'%s' where IssuerId = %d and ServerID = %d", date.ServerName, date.IssuerId, date.ServerID))
		}
	}
}

func update(engines *xorm.Engine, File *xlsx.Sheet) {
	for _, row := range File.Rows {
		var list []*xlsx.Cell
		for _, cell := range row.Cells {
			list = append(list, cell)
		}
		date := parse(list)
		if date != nil {
			n, err := engines.Where("IssuerId = ? and ServerID = ?", date.IssuerId, date.ServerID).Update(date)
			if n == 0 || err != nil {
				fmt.Printf("更新:%s出错.\n", fmt.Sprint(*date))
				continue
			}
			engines.Query(fmt.Sprintf("UPDATE NxServerState SET ServerName=N'%s' where IssuerId = %d and ServerID = %d", date.ServerName, date.IssuerId, date.ServerID))
		}
	}
}

func parse(list []*xlsx.Cell) *NxServerState {
	if !check(list) {
		return nil
	}
	s_GameID, _ := list[1].Int()
	s_IssuerId, _ := list[2].Int()
	s_ServerID, _ := list[3].Int()
	s_ServerName := list[4].Value
	s_OnlineNum, _ := list[5].Int()
	s_MaxOnlineNum, _ := list[6].Int()
	s_ServerIP := list[7].Value
	s_Port, _ := list[8].Int()
	s_IsRuning, _ := list[9].Int()
	s_ServerStyle, _ := list[10].Int()
	s_IsStartIPWhile, _ := list[11].Int()
	s_OrderBy, _ := list[14].Int()
	return &NxServerState{s_GameID,
		s_IssuerId,
		s_ServerID,
		s_ServerName,
		s_OnlineNum,
		s_MaxOnlineNum,
		s_ServerIP,
		s_Port,
		s_IsRuning,
		s_ServerStyle,
		s_IsStartIPWhile,
		time.Now(),
		time.Now(),
		s_OrderBy,
	}
}

func check(cells []*xlsx.Cell) bool {
	for _, v := range cells {
		if len(v.Value) == 0 {
			return false
		}
	}
	return true
}

func getcsv() {
	File, _ := os.Create("statuslist.csv")
	defer File.Close()
	Csv := csv.NewWriter(File)
	Engine := Engines()
	result := new(NxServerState)
	lines, _ := Engine.Rows(result)
	defer lines.Close()
	lines.Next()
	r := new(NxServerState)
	var i int = 1
	for {
		err := lines.Scan(r)
		if err != nil {
			return
		}
		Csv.Write([]string{fmt.Sprint(i),
			fmt.Sprint(r.GameID),
			fmt.Sprint(r.IssuerId),
			fmt.Sprint(r.ServerID),
			fmt.Sprint(r.ServerName),
			fmt.Sprint(r.OnlineNum),
			fmt.Sprint(r.MaxOnlineNum),
			fmt.Sprint(r.ServerIP),
			fmt.Sprint(r.Port),
			fmt.Sprint(r.IsRuning),
			fmt.Sprint(r.ServerStyle),
			fmt.Sprint(r.IsStartIPWhile),
			fmt.Sprint(r.LogTime),
			fmt.Sprint(r.UpdateTime),
			fmt.Sprint(r.OrderBy)})
		Csv.Flush()
		if !lines.Next() {
			break
		}
		i++
	}
}
