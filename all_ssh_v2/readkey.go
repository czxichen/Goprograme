package all_ssh

import (
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/ssh"
)

func ReadKey(keypath []string) {
	for _, v := range keypath {
		buf, err := ioutil.ReadFile(v)
		if err != nil {
			fmt.Printf("读取key文件%s失败:\n%s\n", v, err)
			os.Exit(1)
		}
		signer, err := ssh.ParsePrivateKey(buf)
		if err != nil {
			fmt.Printf("解析key文件%s失败:\n%s\n", v, err)
			os.Exit(1)
		}
		privateKey = append(privateKey, ssh.PublicKeys(signer))
	}
}
