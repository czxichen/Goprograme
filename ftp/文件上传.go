package main

import (
	"fmt"
	iconv "github.com/djimenez/iconv-go"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	"github.com/jlaffaye/ftp"
	_ "github.com/lunny/godbc"
	"os"
	"time"
)

type info struct {
	path string
	id   int
}

func main() {
	//配置ＦＴＰ文件路径,为方便本地保持一致.
	pathlist := []info{{"appstore/server_list.ini", 10000},
		{"HUN/common/server_list.ini", 10001},
		{"YHLM/common/server_list.ini", 10002},
		{"YYBS/common/server_list.ini", 10003}}
	Sql(pathlist)
	Ftp(pathlist)
	fmt.Printf("20秒后自动退出,也可直接关闭此窗口\n")
	for i := 20; i > 0; i-- {
		fmt.Printf("\r剩余%d秒...", i)
		time.Sleep(1e9)
	}
}

func Ftp(pathlist []info) {
	//配置FTP的地址.
	con, err := ftp.Dial("127.0.0.1:21")
	if err != nil {
		fmt.Println("连接FTP错误:", err)
		return
	}
	//下面是配置FTP的账户.
	if err := con.Login("root", "123456"); err != nil {
		fmt.Println("登录错误:", err)
	}
	defer con.Logout()

	for _, v := range pathlist {
		File, err := os.Open(v.path)
		if err != nil {
			fmt.Printf("打开文件:%s错误 %s\n", v.path, err)
			continue
		}
		defer File.Close()
		err = con.Stor(v.path, File)
		if err != nil {
			fmt.Println("上传文件出错：", err)
			continue
		}
		fmt.Println(v.path, "上传成功")
	}
}

type NxServerState struct {
	ServerID    int    `xorm:"not null ServerID"`
	ServerName  string `xorm:"nvarchar 'ServerName'"`
	ServerIP    string `xorm:"not null ServerIP"`
	Port        int    `xorm:"not null Port"`
	IsRuning    int    `xorm:"not null IsRuning"`
	ServerStyle int    `xorm:"ServerStyle"`
}

func Sql(list []info) {
	engine := Engines()
	if engine == nil {
		return
	}
	for _, v := range list {
		rows, err := engine.Where("IssuerId=? and IsRuning=1", v.id).Rows(&NxServerState{})
		if err != nil {
			fmt.Printf("查询%s %s\n", v.id, err)
			continue
		}
		defer rows.Close()
		File, err := os.Create(v.path)
		if err != nil {
			fmt.Printf("新建文件:%s失败 %s\n", v.path, err)
			continue
		}
		result := new(NxServerState)
		defer File.Close()
		for rows.Next() {
			rows.Scan(result)
			s := fmt.Sprintf("%d/%s/%s/%d/%d/%d/0\r\r\n", result.ServerID,
				result.ServerName,
				result.ServerIP,
				result.Port,
				result.ServerStyle,
				result.IsRuning)
			File.Write(convert([]byte(s)))
		}
	}
}
func Engines() *xorm.Engine {
	//下面是配置数据库的连接串.
	Engine, err := xorm.NewEngine("odbc", "driver={SQL Server};Server=127.0.0.1;Database=gyc_status_s;uid=sa;pwd=123456;")
	if err != nil {
		fmt.Println("新建引擎错误：", err)
		return nil
	}
	if err := Engine.Ping(); err != nil {
		fmt.Println(err)
		return nil
	}
	Engine.SetTableMapper(core.SameMapper{})
	return Engine
}
func convert(src []byte) []byte {
	x := make([]byte, 100)
	_, n, _ := iconv.Convert(src, x, "GB18030", "utf-8")
	return x[:n]
}
