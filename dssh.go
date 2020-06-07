/*
	对"golang.org/x/crypto/ssh"做了简单封装. 提供批量并行执行的接口. 提供一些常用的功能(比如检测端口进程文件存在性,获取md5等等)

	dssh库使用之前, 必须进行初始化:
		dssh.Initialize(true); // 初始化, 启用连接缓存
		dssh.Initialize(false); // 初始化, 不启用连接缓存
	在一批机器执行某个指令
		var addrs = []string{A,B,C,D}
		dssh.Runs(addrs, "df", nil); // 执行df
		dssh.SudoRuns(addrs, "root", "netstat -ntpl", nil); // 以root名义执行netstat -ntpl. 等价于 `sudo netstat -ntpl`
	指定超时时间
		dssh.Runs(addrs, "df", NewOption().SetConnTimeout(3*time.Secon)); // 3s连接超时
		dssh.SudoRuns(addrs, "root", "netstat -ntpl", NewOption().SetExecTimeout(10*time.Second)); // 10秒执行超时
	默认情况下, 指令开头会自动添加'errexit,pipefail,nounset', 可以关闭
		dssh.Runs(addrs, "grep -i error ~/result | awk '{print $3}' ", NewOption().SetBashOpt(false)); // grep不到error也不认为是错误
	一些常用功能
		PortExist(addrs, []int{8080}, dssh.PortExist_AllOnAll, nil); // 检测是否所有机器都有8080端口存在
		ProcExist(addrs, []string{"myservice", "myservice2"}, dssh.ProcExist_0On0, nil); // 检测是否所有机器都没有任何服务存在


*/
package dssh

import (
	"fmt"
	"os/user"
	"time"
)

/*
	一些ssh相关详细参数
*/
type Option struct {
	Conn_timeout time.Duration // 连接超时时间. 默认3s
	Exec_timeout time.Duration // 执行超时时间. 默认不超时
	Bash_opt     bool          // 是否添加bash选项: (errexit,nounset,pipefail). 默认是
}

func (self *Option) SetConnTimeout(timeout time.Duration) *Option {
	self.Conn_timeout = timeout
	return self
}

func (self *Option) SetExecTimeout(timeout time.Duration) *Option {
	self.Exec_timeout = timeout
	return self
}
func (self *Option) SetBashOpt(enable bool) *Option {
	self.Bash_opt = enable
	return self
}

func NewOption() *Option {
	return &Option{
		Conn_timeout: 3 * time.Second,
		Exec_timeout: 0,
		Bash_opt:     true,
	}
}
func get_opt(opt *Option) *Option {
	if opt != nil {
		return opt
	} else {
		return NewOption()
	}
}

var g_sshclient_pool *SSHClientPool
var g_enable_conn_cache = true

var g_cur_username string

/*
	lib库初始化函数. 成功调用后, 其他函数才能使用

	enable_conn_cache: 是否开启连接缓存
*/
func Initialize(enable_conn_cache bool) error {
	g_enable_conn_cache = enable_conn_cache

	var u, err = user.Current()
	if err != nil {
		return err
	}
	g_cur_username = u.Username

	g_sshclient_pool = new_ssh_client_pool()
	return nil
}

// 在单个机器执行指令
func Run(addr string, cmd string, opt *Option) *CmdResult {
	return SudoRun(addr, "", cmd, opt)
}

// 以sudo用户执行
func SudoRun(addr string, sudo string, cmd string, opt *Option) *CmdResult {
	var finish_chan = make(chan *CmdResult, 1)
	sudo_run_nb(addr, sudo, cmd, opt, finish_chan)
	return <-finish_chan
}

// 并行批量执行
//
// 不保证返回results顺序和addrs顺序一致
func Runs(addrs []string, cmd string, opt *Option) []*CmdResult {
	return SudoRuns(addrs, "", cmd, opt)
}

// 并行批量执行
//
// 不保证返回results顺序和addrs顺序一致
func SudoRuns(addrs []string, sudo string, cmd string, opt *Option) []*CmdResult {
	var finish_chan = make(chan *CmdResult, 10)
	for _, addr := range addrs {
		go sudo_run_nb(addr, sudo, cmd, opt, finish_chan)
	}

	var results = make([]*CmdResult, 0, len(addrs))
	for i := 0; i < len(addrs); i++ {
		results = append(results, <-finish_chan)
	}
	return results
}

// 并行批量执行不同命令.
//
// 不保证返回results顺序和addrs顺序一致
func RunsDiff(addrs []string, cmds []string, opt *Option) []*CmdResult {
	return SudoRunsDiff(addrs, "", cmds, opt)
}

/*
	并行批量执行不同命令

	不保证返回results顺序和addrs顺序一致

	len(addrs) 必须等于 len(cmds), 否则直接panic
*/
func SudoRunsDiff(addrs []string, sudo string, cmds []string, opt *Option) []*CmdResult {
	if len(addrs) != len(cmds) {
		panic("bug: len(addrs)!=len(cmds)")
	}
	var finish_chan = make(chan *CmdResult, len(addrs))
	for idx, addr := range addrs {
		go sudo_run_nb(addr, sudo, cmds[idx], opt, finish_chan)
	}

	var results = make([]*CmdResult, 0, len(addrs))
	for i := 0; i < len(addrs); i++ {
		results = append(results, <-finish_chan)
	}
	return results
}

// 执行是否返回错误. 成功返回nil.
func RunsErrOnFail(addrs []string, cmd string, opt *Option, failmsg string) error {
	return SudoRunsErrOnFail(addrs, "", cmd, opt, failmsg)
}
func SudoRunsErrOnFail(addrs []string, sudo string, cmd string, opt *Option, failmsg string) error {
	var rets = SudoRuns(addrs, sudo, cmd, opt)
	var combine_err = ""
	for _, ret := range rets {
		if ret.IsSuccess() {
		} else if ret.IsExecFail() {
			combine_err += fmt.Sprintf("\n%v %v %v", ret.Addr, failmsg, ret.GetError())
		} else {
			combine_err += fmt.Sprintf("\n%v", ret.GetErrorMsg())
		}
	}
	if len(combine_err) > 0 {
		return fmt.Errorf("%v", combine_err[1:])
	} else {
		return nil
	}
}

/*
	分arch执行不同cmd. 相同arch并行批量执行

	map_arch_cmd: arch -> cmd. 如果有机器的arch在该map中找不到对应cmd, 则失败. 这种情况下所有cmd都不会执行

	不保证返回results顺序和addrs顺序一致
*/
func RunsByArch(addrs []string, map_arch_cmd map[string]string, opt *Option) ([]*CmdResult, error) {
	return SudoRunsByArch(addrs, "", map_arch_cmd, opt)
}

// 类似RunsByArch
func SudoRunsByArch(addrs []string, sudo string, map_arch_cmd map[string]string, opt *Option) ([]*CmdResult, error) {
	// 获取arch
	var arch_rets = Runs(addrs, "arch", nil)
	var map_arch_hosts = map[string][]string{}
	for _, ret := range arch_rets {
		if ret.IsSuccess() {
			var arch = ret.GetStripStdout()
			var hosts = map_arch_hosts[arch]
			map_arch_hosts[arch] = append(hosts, ret.Addr)
		} else {
			return nil, ret.GetError()
		}
	}
	// 检查arch存在性
	for arch, _ := range map_arch_hosts {
		if _, e := map_arch_cmd[arch]; !e {
			return nil, fmt.Errorf("unknown arch[%v]", arch)
		}
	}
	// 执行
	var rets []*CmdResult
	for arch, hosts := range map_arch_hosts {
		var cmd = map_arch_cmd[arch]
		rets = append(rets, SudoRuns(hosts, sudo, cmd, opt)...)
	}
	if len(rets) != len(addrs) {
		panic("bug")
	}
	return rets, nil
}
