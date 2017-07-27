package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	sendurl   = `https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=`
	get_token = `https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=`
)

var requestError = errors.New("request error,check url or network")

type access_token struct {
	Access_token string `json:"access_token"`
	Expires_in   int    `json:"expires_in"`
}

type send_msg struct {
	Touser  string            `json:"touser"`
	Toparty string            `json:"toparty"`
	Totag   string            `json:"totag"`
	Msgtype string            `json:"msgtype"`
	Agentid int               `json:"agentid"`
	Text    map[string]string `json:"text"`
	Safe    int               `json:"safe"`
}

type send_msg_error struct {
	Errcode int    `json:"errcode`
	Errmsg  string `json:"errmsg"`
}

func main() {
	mfile := flag.String("m", "", "-m msg.txt �������ļ���ȡ���÷�����Ϣ")
	touser := flag.String("t", "@all", "-t user ֱ�ӽ�����Ϣ���û��ǳ�")
	agentid := flag.Int("i", 0, "-i 0 ָ��agentid")
	content := flag.String("c", "Hello world", "-c 'Hello world' ָ��Ҫ���͵�����")
	corpid := flag.String("p", "", "-p corpid ����ָ��")
	corpsecret := flag.String("s", "", "-s corpsecret ����ָ��")
	flag.Parse()

	if *corpid == "" || *corpsecret == "" {
		flag.Usage()
		return
	}

	var m send_msg = send_msg{Touser: *touser, Msgtype: "text", Agentid: *agentid, Text: map[string]string{"content": *content}}

	if *mfile != "" {
		buf, err := Parse(*mfile)
		if err != nil {
			println(err.Error())
			return
		}
		err = json.Unmarshal(buf, &m)
		if err != nil {
			println(err)
			return
		}
	}
	///-p "wx2468f5838693e374" -s "JbjkM1jYq8g3GaHjOTgj27y4n4_7Dsv4FV94I5BMRSrBsm_aTsMUVJMhGu_s9jkH"
	token, err := Get_token(*corpid, *corpsecret)
	if err != nil {
		println(err.Error())
		return
	}
	buf, err := json.Marshal(m)
	if err != nil {
		return
	}
	err = Send_msg(token.Access_token, buf)
	if err != nil {
		println(err.Error())
	}
}

func Send_msg(Access_token string, msgbody []byte) error {
	body := bytes.NewBuffer(msgbody)
	resp, err := http.Post(sendurl+Access_token, "application/json", body)
	if resp.StatusCode != 200 {
		return requestError
	}
	buf, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	var e send_msg_error
	err = json.Unmarshal(buf, &e)
	if err != nil {
		return err
	}
	if e.Errcode != 0 && e.Errmsg != "ok" {
		return errors.New(string(buf))
	}
	return nil
}

func Get_token(corpid, corpsecret string) (at access_token, err error) {
	resp, err := http.Get(get_token + corpid + "&corpsecret=" + corpsecret)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = requestError
		return
	}
	buf, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(buf, &at)
	if at.Access_token == "" {
		err = errors.New("corpid or corpsecret error.")
	}
	return
}

func Parse(jsonpath string) ([]byte, error) {
	var zs = []byte("//")
	File, err := os.Open(jsonpath)
	if err != nil {
		return nil, err
	}
	defer File.Close()
	var buf []byte
	b := bufio.NewReader(File)
	for {
		line, _, err := b.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}
		line = bytes.TrimSpace(line)
		if len(line) <= 0 {
			continue
		}
		index := bytes.Index(line, zs)
		if index == 0 {
			continue
		}
		if index > 0 {
			line = line[:index]
		}
		buf = append(buf, line...)
	}
	return buf, nil
}
