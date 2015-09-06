package main

import (
	"crypto/hmac"
	"crypto/sha512"
	"fmt"
	"strconv"
	"time"
)

type Key struct {
	gkey string
	skey string
	date func() int64
}

const (
	Gkey = "What"
)

func main() {
	K := &Key{gkey: Gkey, date: getdate}
	b := hmac.New(sha512.New, []byte(K.Hmac("Hello World")))
	B := b.Sum(nil)
	offset := B[len(B)-1] & 0xf
	x := ((int(B[offset+1]) & 0xff) << 16) | ((int(B[offset+2]) & 0xff) << 8) | (int(B[offset+3]) & 0xff)
	z := fmt.Sprint(x % 1000000)
	for len(z) < 6 {
		z = fmt.Sprintf("0%s", z)
	}
	fmt.Println(z)
}

func (self *Key) Hmac(Skey string) string {
	return fmt.Sprintf("%s%s%d", self.gkey, Skey, self.date())
}

func getdate() int64 {
	T := time.Now()
	format_T := T.Format("0601021504")
	unix_T := T.Unix() - T.Unix()%30
	num, _ := strconv.Atoi(format_T)
	x := num % 10
	return unix_T<<uint(x) ^ int64(num)>>uint(x)
}
