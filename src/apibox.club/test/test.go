package main

import (
	"apibox.club/utils"
	"bytes"
	"fmt"
	"io/ioutil"
	// "os"
	// "strings"
	"syscall"
)

func GetPID() (int, error) {
	b, err := ioutil.ReadFile(apibox.PidPath)
	if nil != err {
		return -1, err
	}
	b = bytes.TrimSpace(b)
	pid, err := apibox.StringUtils(string(b)).Int()
	if nil != err {
		return -1, err
	}
	return pid, nil
}

func main() {
	pid, err := GetPID()
	if nil != err {
		panic(err)
	}

	if err := syscall.Kill(pid, 0); nil != err {
		fmt.Println(err.Error())
	} else {
		fmt.Println("sssss")
	}
}
