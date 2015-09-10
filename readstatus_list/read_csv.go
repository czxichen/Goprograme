package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/djimenez/iconv-go"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	_ "github.com/lunny/godbc"
	"github.com/tealeg/xlsx"
)

type NxServerState struct {
	GameID         int    `xorm:"not null 'GameID'"`
	IssuerId       int    `xorm:"not null IssuerId"`
	ServerID       int    `xorm:"not null ServerID"`
	ServerName     string `xorm:"nvarchar 'ServerName'"`
	OnlineNum      int    `xorm:"not null OnlineNum"`
	MaxOnlineNum   int    `xorm:"not null MaxOnlineNum"`
	ServerIP       string `xorm:"not null ServerIP"`
	Port           int    `xorm:"not null Port"`
	IsRuning       int    `xorm:"not null IsRuning"`
	ServerStyle    int    `xorm:"ServerStyle"`
	IsStartIPWhile int    `xorm:"not null IsStartIPWhile"`
	OrderBy        int    `xorm:"not null OrderBy"`
}

var debug bool
var head = []string{"GameID", "IssuerId",
	"ServerID", "ServerName", "OnlineNum",
	"MaxOnlineNum", "ServerIP", "Port",
	"IsRuning", "ServerStyle", "IsStartIPWhile", "OrderBy"}

func main() {
	flag.BoolVar(&debug, "d", false, "-d=true打印debug信息")
	flag.Parse()
	var readline int
	fmt.Printf("1:同步线上状态服务器列表\n2:更新server.xlsx中update的内容\n3:插入中server.xlsx中的insert的内容\n")
	fmt.Print("输入操作选项：")
	fmt.Scan(&readline)
	engines := Engines()
	switch readline {
	case 1:
		getOnLineList()
		writehead()
		fmt.Println("已同步线上最新列表,保存在server.xlsx")
	case 2:
		File, e := xlsx.OpenFile("server.xlsx")
		if e != nil {
			fmt.Println(e)
			return
		}
		fmt.Println("更新server.xlsx中update的内容")
		update(engines, File.Sheet["update"])
	case 3:
		File, e := xlsx.OpenFile("server.xlsx")
		if e != nil {
			fmt.Println(e)
			return
		}
		fmt.Println("插入server.xlsx中insert的内容")
		insert(engines, File.Sheet["insert"])
	default:
		fmt.Println("使用方法：输入想要操作的选项（例如：1）")
	}
	fmt.Printf("\n\n20秒后自动退出,也可直接关闭\n")
	for i := 20; i > 0; i-- {
		fmt.Printf("\r剩余%d秒...", i)
		time.Sleep(1e9)
	}
}

func update(engines *xorm.Engine, File *xlsx.Sheet) {
	var num int = 0
	if len(File.Rows) < 1 {
		fmt.Println("检查更新列表.")
		return
	}
	for _, row := range File.Rows[1:] {
		var list []*xlsx.Cell
		for _, cell := range row.Cells {
			list = append(list, cell)
		}
		date := parse(list)
		fmt.Println(date)
		if date != nil {
			n, err := engines.Where("IssuerId = ? and ServerID = ?", date.IssuerId,
				date.ServerID).Cols("OnlineNum", "IsRuning", "ServerStyle",
				"IsStartIPWhile", "OrderBy").Update(date)
			if n == 0 || err != nil {
				fmt.Printf("更新:%s出错.\n", fmt.Sprint(*date))
				continue
			}
			engines.Query(fmt.Sprintf("UPDATE NxServerState SET ServerName=N'%s' where IssuerId = %d and ServerID = %d", date.ServerName, date.IssuerId, date.ServerID))
		}
		num++
	}
	fmt.Printf("一共更新%d条数据\n", num)
}

func insert(engines *xorm.Engine, File *xlsx.Sheet) {
	var num int = 0
	if len(File.Rows) < 1 {
		fmt.Println("检查更新列表.")
		return
	}
	for _, row := range File.Rows[1:] {
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
		num++
	}
	fmt.Printf("一共插入%d条数据\n", num)
}

func parse(list []*xlsx.Cell) *NxServerState {
	if !check(list) {
		return nil
	}
	s_GameID, _ := list[0].Int()
	s_IssuerId, _ := list[1].Int()
	s_ServerID, _ := list[2].Int()
	s_ServerName := list[3].Value
	s_OnlineNum, _ := list[4].Int()
	s_MaxOnlineNum, _ := list[5].Int()
	s_ServerIP := list[6].Value
	s_Port, _ := list[7].Int()
	s_IsRuning, _ := list[8].Int()
	s_ServerStyle, _ := list[9].Int()
	s_IsStartIPWhile, _ := list[10].Int()
	s_OrderBy, _ := list[11].Int()
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
		s_OrderBy,
	}
}

func getOnLineList() {
	File := xlsx.NewFile()
	sheet := File.AddSheet("online")
	File.AddSheet("update")
	File.AddSheet("insert")
	row := sheet.AddRow()
	row.WriteSlice(&head, len(head))
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
		r.ServerName = convert([]byte(r.ServerName))
		row := sheet.AddRow()
		row.WriteStruct(r, len(head))
		if !lines.Next() {
			break
		}
		i++
	}
	File.Save("server.xlsx")
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

func convert(src []byte) string {
	x := make([]byte, 40)
	_, n, _ := iconv.Convert(src, x, "GB18030", "utf-8")
	return string(x[:n])
}

func check(cells []*xlsx.Cell) bool {
	for _, v := range cells {
		if len(v.Value) == 0 {
			return false
		}
	}
	return true
}

func writehead() {
	File, err := xlsx.OpenFile("server.xlsx")
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	Sheet := File.Sheet["update"]
	row := Sheet.AddRow()
	for _, v := range head {
		row.AddCell().SetString(v)
	}
	Sheet = File.Sheet["insert"]
	row = Sheet.AddRow()
	for _, v := range head {
		row.AddCell().SetString(v)
	}
	File.Save("server.xlsx")
}
