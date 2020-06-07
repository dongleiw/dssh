package dssh

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"strings"
	"time"
)

type ResultStatus int

var (
	RS_SUCCESS ResultStatus = 0 // 执行成功. exitcode=0
	RS_TIMEOUT ResultStatus = 1 // 等待超时
	RS_RET     ResultStatus = 2 // exitcode!=0
	RS_CONN    ResultStatus = 3 // 连接错误
	RS_OTHERS  ResultStatus = 4 // 其他异常

	result_status_desc_map = map[ResultStatus]string{
		RS_SUCCESS: "success",
		RS_TIMEOUT: "timeout",
		RS_RET:     "errexit",
		RS_CONN:    "connerr",
		RS_OTHERS:  "othererr",
	}
)

// 指令执行结果
type CmdResult struct {
	User string
	Addr string
	Sudo string
	Cmd  string

	RS       ResultStatus
	Beg_time time.Time
	End_time time.Time
	ExitCode int
	Stdout   string
	Stderr   string
	Err      error
}

func newCmdResult(user string, addr string, sudo string, cmd string) *CmdResult {
	var cr = &CmdResult{
		User:     user,
		Addr:     addr,
		Sudo:     sudo,
		Cmd:      cmd,
		Beg_time: time.Now(),
		ExitCode: 1,
		RS:       RS_OTHERS,
	}
	return cr
}

func (self *CmdResult) end_connerr(err error) {
	self.End_time = time.Now()

	if err == nil {
		panic("err is nil")
	}
	self.Err = err
	self.RS = RS_CONN
}
func (self *CmdResult) End(err error) {
	self.End_time = time.Now()
	if err != nil {
		self.Err = err
		if exiterr, ok := err.(*ssh.ExitError); ok {
			self.RS = RS_RET
			self.ExitCode = exiterr.Waitmsg.ExitStatus()
		} else if _, ok := err.(*ErrTimeout); ok {
			self.RS = RS_TIMEOUT
		} else {
			self.RS = RS_OTHERS
		}
	} else {
		self.RS = RS_SUCCESS
		self.ExitCode = 0
	}
}

func (self *CmdResult) IsSuccess() bool   { return self.RS == RS_SUCCESS }
func (self *CmdResult) IsExecFail() bool  { return self.RS == RS_RET }
func (self *CmdResult) GetStatus() string { return result_status_desc_map[self.RS] }

func (self *CmdResult) GetStdout() string      { return self.Stdout }
func (self *CmdResult) GetStripStdout() string { return strings.TrimSpace(self.Stdout) }
func (self *CmdResult) GetNonEmptyLinesStdout() []string {
	var lines []string
	for _, line := range strings.Split(strings.Replace(self.Stdout, "\r\n", "\n", -1), "\n") {
		var sline = strings.TrimSpace(line)
		if len(sline) > 0 {
			lines = append(lines, sline)
		}
	}
	return lines
}
func (self *CmdResult) GetError() error {
	if self.IsSuccess() {
		return nil
	}
	var status_desc = self.GetStatus()
	return fmt.Errorf("cmd[%v] on [%v@%v] sudo[%v] failed[%v]: stdout[%v] stderr[%v] err[%v] exitcode[%v]", self.Cmd, self.User, self.Addr, self.Sudo, status_desc, self.Stdout, self.Stderr, self.Err, self.ExitCode)
}
func (self *CmdResult) GetErrorMsg() string {
	return self.GetError().Error()
}
func (self *CmdResult) ToString() string {
	return fmt.Sprintf("%+v", self)
}
