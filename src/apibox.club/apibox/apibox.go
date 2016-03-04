package main

import (
	"apibox.club/utils"
	"apibox.club/website"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"syscall"
	"time"
)

const (
	START  string = "start"
	STOP   string = "stop"
	STATUS string = "status"
)

type Apibox struct {
	PID int
}

func (a *Apibox) GetPID() (*Apibox, error) {
	b, err := ioutil.ReadFile(apibox.PidPath)
	if nil != err {
		return nil, err
	}
	b = bytes.TrimSpace(b)
	pid, err := apibox.StringUtils(string(b)).Int()
	if nil != err {
		return nil, err
	}
	a.PID = pid
	return a, nil
}

func (a *Apibox) Start() error {
	time.Sleep(time.Duration(1 * time.Second))
	_, err := a.GetPID()
	if nil != err {
		return err
	}
	if err := syscall.Kill(a.PID, 0); nil != err {
		apibox.Set_log_level(apibox.LevelDebug)
		website.Run()
	} else {
		fmt.Fprintf(os.Stderr, "The program is running, turn off the start again.\n")
	}
	return nil
}

func (a *Apibox) Stop() error {
	time.Sleep(time.Duration(1 * time.Second))
	_, err := a.GetPID()
	if nil != err {
		return err
	}
	p, err := os.FindProcess(a.PID)
	if nil != err {
		return err
	}
	err = p.Kill()
	if nil != err {
		return err
	}
	return nil
}

func (a *Apibox) Status() (bool, error) {
	_, err := a.GetPID()
	if nil != err {
		return false, err
	}
	if err := syscall.Kill(a.PID, 0); nil != err {
		return false, nil
	} else {
		return true, nil
	}
}

func main() {
	flag.Parse()
	var cmd string = flag.Arg(0)
	cmd = strings.ToLower(cmd)
	switch strings.TrimSpace(cmd) {
	case START:
		a := &Apibox{}
		err := a.Start()
		if nil != err {
			fmt.Fprintf(os.Stderr, err.Error()+"\n")
		}
	case STOP:
		a := &Apibox{}
		err := a.Stop()
		if nil != err {
			fmt.Fprintf(os.Stderr, err.Error()+"\n")
		}
	case STATUS:
		a := &Apibox{}
		t, err := a.Status()
		if nil != err {
			fmt.Fprintf(os.Stderr, err.Error()+"\n")
		}
		if !t {
			fmt.Fprintf(os.Stdout, "Stop.\n")
		} else {
			fmt.Fprintf(os.Stdout, "Running...\n")
		}
	default:
		fmt.Fprintf(os.Stderr, "Usage: %s command <start|stop|status>\n", os.Args[0])
	}
}
