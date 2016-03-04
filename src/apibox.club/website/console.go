package website

import (
	"apibox.club/utils"
	"bytes"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/net/websocket"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

var (
	aesKey string = "$hejGRT^$*#@#12o"
)

type ssh struct {
	user    string
	pwd     string
	addr    string
	client  *gossh.Client
	session *gossh.Session
}

func (s *ssh) Connect() (*ssh, error) {
	config := &gossh.ClientConfig{}
	config.SetDefaults()
	config.User = s.user
	config.Auth = []gossh.AuthMethod{gossh.Password(s.pwd)}
	client, err := gossh.Dial("tcp", s.addr, config)
	if nil != err {
		return nil, err
	}
	s.client = client
	session, err := client.NewSession()
	if nil != err {
		return nil, err
	}
	s.session = session
	return s, nil
}

func (s *ssh) Exec(cmd string) (string, error) {
	var buf bytes.Buffer
	s.session.Stdout = &buf
	s.session.Stderr = &buf
	err = s.session.Run(cmd)
	if err != nil {
		return "", err
	}
	defer s.session.Close()
	stdout := buf.String()
	apibox.Log_Debug("Stdout:", stdout)
	return stdout, nil
}

func chkSSHSrvAddr(ssh_addr, key string) (string, string, error) {
	u, err := url.Parse(ssh_addr)
	if nil != err {
		return "", "", err
	}
	var new_url, new_host string
	if "" == u.Host {
		new_host = u.String()
	} else {
		new_host = u.Host
	}
	urls := strings.Split(new_host, ":")
	if len(urls) != 2 {
		new_url = new_host + ":22"
	} else {
		new_url = new_host
	}
	addr, err := net.ResolveTCPAddr("tcp4", new_url)
	if nil != err {
		return "", "", err
	}
	en_addr, err := apibox.AESEncode(addr.String(), key)
	if nil != err {
		return "", "", err
	}
	return addr.String(), en_addr, nil
}

func SSHWebSocketHandler(ws *websocket.Conn) {
	ctx := NewContext(nil, ws.Request())

	vm_info := ctx.GetFormValue("vm_info")
	cols := ctx.GetFormValue("cols")
	rows := ctx.GetFormValue("rows")

	apibox.Log_Debug("VM Info:", vm_info, "Cols:", cols, "Rows:", rows)

	de_vm_info, err := apibox.AESDecode(vm_info, aesKey)
	if nil != err {
		apibox.Log_Err("AESDecode:", err)
		return
	} else {
		de_vm_info_arr := strings.Split(de_vm_info, "\n")
		if len(de_vm_info_arr) == 3 {
			user_name := strings.TrimSpace(de_vm_info_arr[0])
			user_pwd := strings.TrimSpace(de_vm_info_arr[1])
			vm_addr := strings.TrimSpace(de_vm_info_arr[2])

			apibox.Log_Debug("VM Addr:", vm_addr)

			sh := &ssh{
				user: user_name,
				pwd:  user_pwd,
				addr: vm_addr,
			}
			sh, err = sh.Connect()
			if nil != err {
				apibox.Log_Err(err)
				return
			}

			ptyCols, err := apibox.StringUtils(cols).Int()
			if nil != err {
				apibox.Log_Err(err)
				return
			}
			ptyRows, err := apibox.StringUtils(rows).Int()
			if nil != err {
				apibox.Log_Err(err)
				return
			}

			session := sh.session
			defer session.Close()
			modes := gossh.TerminalModes{
				gossh.ECHO:          1,
				gossh.TTY_OP_ISPEED: 14400,
				gossh.TTY_OP_OSPEED: 14400,
			}

			if err = session.RequestPty("xterm-256color", ptyRows, ptyCols, modes); err != nil {
				apibox.Log_Err(err)
				return
			}

			w, err := session.StdinPipe()
			if nil != err {
				apibox.Log_Err(err)
				return
			}
			go func() {
				io.Copy(w, ws)
			}()

			r, err := session.StdoutPipe()
			if nil != err {
				apibox.Log_Err(err)
				return
			}
			go func() {
				io.Copy(ws, r)
			}()

			er, err := session.StderrPipe()
			if nil != err {
				apibox.Log_Err(err)
				return
			}
			go func() {
				io.Copy(ws, er)
			}()

			if err := session.Shell(); nil != err {
				apibox.Log_Err(err)
				return
			}

			if err := session.Wait(); nil != err {
				apibox.Log_Err(err)
				return
			}
		} else {
			apibox.Log_Err("Unable to parse the data.")
			return
		}
	}
}

type Console struct {
}

type LoginPageData struct {
	VM_Name    string `json:"vm_name" xml:"vm_name"`
	VM_Addr    string `json:"vm_addr" xml:"vm_addr"`
	EN_VM_Name string `json:"en_vm_name" xml:"en_vm_name"`
	EN_VM_Addr string `json:"en_vm_addr" xml:"en_vm_addr"`
	Token      string `json:"token" xml:"token"`
}

func (c *Console) ConsoleLoginPage(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(w, r)
	vm_addr := ctx.GetFormValue("vm_addr")

	de_vm_addr, vm_addr_err := apibox.AESDecode(vm_addr, aesKey)
	if vm_addr == "" || nil != vm_addr_err {
		ctx.OutHtml("login", nil)
	} else {
		lpd := LoginPageData{
			VM_Addr:    de_vm_addr,
			EN_VM_Addr: vm_addr,
			Token:      apibox.StringUtils("sss").Base64Encode(),
		}
		ctx.OutHtml("console/console_login", lpd)
	}
}

type ConsoleMainPageData struct {
	Token    string `json:"token" xml:"token"`
	UserName string `json:"user_name" xml:"user_name"`
	UserPwd  string `json:"user_pwd" xml:"user_pwd"`
	VM_Name  string `json:"vm_name" xml:"vm_name"`
	VM_Addr  string `json:"vm_addr" xml:"vm_addr"`
	WS_Addr  string `json:"ws_addr" xml:"ws_addr"`
}

func (c *Console) ConsoleMainPage(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(w, r)

	vm_info := ctx.GetFormValue("vm_info")

	apibox.Log_Debug("VM Info:", vm_info)

	de_vm_info, err := apibox.AESDecode(vm_info, aesKey)
	if nil != err {
		apibox.Log_Err("AESDecode:", err)
		ctx.OutHtml("login", nil)
	} else {
		de_vm_info_arr := strings.Split(de_vm_info, "\n")
		if len(de_vm_info_arr) == 3 {
			user_name := strings.TrimSpace(de_vm_info_arr[0])
			user_pwd := strings.TrimSpace(de_vm_info_arr[1])
			vm_addr := strings.TrimSpace(de_vm_info_arr[2])

			cmpd := ConsoleMainPageData{
				UserName: user_name,
				UserPwd:  user_pwd,
				VM_Addr:  vm_addr,
			}

			wsScheme := "ws://"
			if Conf.Web.EnableTLS {
				wsScheme = "wss://"
			}
			wsAddr := wsScheme + r.Host + "/console/sshws/" + vm_info
			apibox.Log_Debug("WS Addr:", wsAddr)
			cmpd.WS_Addr = wsAddr
			ctx.OutHtml("console/console_main", cmpd)
		} else {
			ctx.OutHtml("login", nil)
		}
	}
}

func (c *Console) ConsoleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(w, r)

	user_name := ctx.GetFormValue("user_name")
	user_pwd := ctx.GetFormValue("user_pwd")
	vm_addr := ctx.GetFormValue("vm_addr")

	var err error
	boo := true

	de_vm_addr, err := apibox.AESDecode(vm_addr, aesKey)
	if nil != err {
		boo = false
	}

	vm_addr_arr := strings.Split(de_vm_addr, ":")

	if len(vm_addr_arr) != 2 {
		boo = false
	}

	result := &Result{}
	if boo {
		sh := &ssh{
			user: user_name,
			pwd:  user_pwd,
			addr: de_vm_addr,
		}
		sh, err = sh.Connect()
		if nil != err {
			result.Ok = false
			result.Msg = "无法连接到远端主机，请确认远端主机已开机且保证口令的正确性。"
		} else {
			_, err := sh.Exec("true")
			if nil != err {
				result.Ok = false
				result.Msg = "用户无权限访问到远端主机，请联系系统管理员。"
			} else {
				ssh_info := make([]string, 0, 0)
				ssh_info = append(ssh_info, user_name)
				ssh_info = append(ssh_info, user_pwd)
				ssh_info = append(ssh_info, de_vm_addr)
				b64_ssh_info, err := apibox.AESEncode(strings.Join(ssh_info, "\n"), aesKey)
				if nil != err {
					apibox.Log_Err("AESEncode:", err)
					result.Ok = false
					result.Msg = "内部错误，请联系管理员（postmaster@apibox.club）。"
				} else {
					result.Ok = true
					result.Data = "/console/main/" + b64_ssh_info
				}
			}
		}
	} else {
		result.Ok = false
		result.Msg = "内部错误，请联系管理员（postmaster@apibox.club）。"
	}
	ctx.OutJson(result)
}

func (c *Console) ConsoleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(w, r)
	ctx.OutHtml("login", nil)
}

func (c *Console) ChkSSHSrvAddr(w http.ResponseWriter, r *http.Request) {
	result := &Result{}
	ctx := NewContext(w, r)
	vm_addr := ctx.GetFormValue("vm_addr")
	if vm_addr == "" {
		result.Ok = false
		result.Msg = "Invalid host address."
	} else {
		sshd_addr, en_addr, err := chkSSHSrvAddr(vm_addr, aesKey)
		if nil != err {
			result.Ok = false
			result.Msg = "Unable to resolve host address."
		} else {
			chkMap := make(map[string]string)
			chkMap["sshd_addr"] = sshd_addr
			chkMap["en_addr"] = en_addr

			result.Ok = true
			result.Data = chkMap
		}
	}
	ctx.OutJson(result)
}

func init() {
	aesKey, _ = apibox.StringUtils("").UUID16()
	console := &Console{}
	Add_HandleFunc("get,post", "/", console.ConsoleLoginPage)
	Add_HandleFunc("get,post", "/console/chksshdaddr", console.ChkSSHSrvAddr)
	Add_HandleFunc("get,post", "/console/login/:vm_addr", console.ConsoleLoginPage)
	Add_HandleFunc("post", "/console/login", console.ConsoleLogin)
	Add_HandleFunc("get,post", "/console/logout", console.ConsoleLogout)
	Add_HandleFunc("get,post", "/console/main/:vm_info", console.ConsoleMainPage)
	Add_Handle("get", "/console/sshws/:vm_info", websocket.Handler(SSHWebSocketHandler))
}
