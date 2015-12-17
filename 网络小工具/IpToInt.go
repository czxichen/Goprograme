package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type IntIP struct {
	IP    string
	Intip int
}

func main() {
	var x *IntIP = &IntIP{IP: "192.168.1.1"}
	fmt.Println(x)
	x.ToIntIp()
	fmt.Println(*x)
}

func (self *IntIP) String() string {
	return self.IP
}

func (self *IntIP) ToIntIp() (int, error) {
	Intip, err := ConvertToIntIP(self.IP)
	if err != nil {
		return 0, err
	}
	self.Intip = Intip
	return Intip, nil
}

func (self *IntIP) ToString() (string, error) {
	i4 := self.Intip & 255
	i3 := self.Intip >> 8 & 255
	i2 := self.Intip >> 16 & 255
	i1 := self.Intip >> 24 & 255
	if i1 > 255 || i2 > 255 || i3 > 255 || i4 > 255 {
		return "", errors.New("Isn't a IntIP Type.")
	}
	ipstring := fmt.Sprintf("%d.%d.%d.%d", i4, i3, i2, i1)
	self.IP = ipstring
	return ipstring, nil
}
func ConvertToIntIP(ip string) (int, error) {
	ips := strings.Split(ip, ".")
	E := errors.New("Not A IP.")
	if len(ips) != 4 {
		return 0, E
	}
	var intIP int
	for k, v := range ips {
		i, err := strconv.Atoi(v)
		if err != nil || i > 255 {
			return 0, E
		}
		intIP = intIP | i<<uint(8*(3-k))
	}
	return intIP, nil
}
