package dssh

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net"
	"time"
)

func sudo_run_nb(addr string, sudo string, cmd string, opt *Option, finish_chan chan<- *CmdResult) {
	opt = get_opt(opt)

	var client *ssh.Client
	var session *ssh.Session
	var err error
	var result = newCmdResult(g_cur_username, addr, sudo, cmd)

	if g_enable_conn_cache {
		session, err = g_sshclient_pool.AllocSession(addr, opt.Conn_timeout)
		if err != nil {
			result.end_connerr(err)
			finish_chan <- result
			return
		}
	} else {
		client, err = conn(addr, opt.Conn_timeout)
		if err != nil {
			result.end_connerr(err)
			finish_chan <- result
			return
		}
		session, err = client.NewSession()
		if err != nil {
			client.Conn.Close()
			result.end_connerr(err)
			finish_chan <- result
			return
		}
	}

	session_run(session, result, addr, sudo, cmd, opt.Exec_timeout, opt.Bash_opt)

	if client != nil {
		client.Conn.Close()
	}
	finish_chan <- result
}
func ssh_client_callback(hostname string, remote net.Addr, key ssh.PublicKey) error {
	return nil
}

func conn(addr string, conn_timeout time.Duration) (*ssh.Client, error) {
	var pubkeyfile = fmt.Sprintf("/home/%s/.ssh/id_rsa", g_cur_username)
	var content, err = ioutil.ReadFile(pubkeyfile)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(content)
	if err != nil {
		return nil, err
	}
	var ssh_config = ssh.ClientConfig{
		User:            g_cur_username,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh_client_callback,
		Timeout:         conn_timeout,
	}
	var address_port = fmt.Sprintf("%s:22", addr)
	client, err := ssh.Dial("tcp", address_port, &ssh_config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// 执行指令
// cmd 需要执行的指令
// exec_timeout 等待指令完成超时时间. <=0表示永不超时
func session_run(session *ssh.Session, result *CmdResult, addr string, sudo string, cmd string, exec_timeout time.Duration, bash_opt bool) {
	defer session.Close()

	var realcmd = "export PATH=$PATH:/usr/sbin; " + cmd
	if bash_opt {
		realcmd = "set -o errexit && set -o pipefail && set -o nounset;" + realcmd
	}
	if len(sudo) > 0 {
		realcmd = fmt.Sprintf("sudo -iu %s <<\"EOF\"\n %s \nEOF", sudo, cmd)
	}

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	var execute_err error
	if exec_timeout > 0 {
		var finish = make(chan error)
		go func() {
			var err = session.Run(realcmd)
			finish <- err
		}()
		select {
		case execute_err = <-finish:
		case <-time.After(exec_timeout):
			execute_err = newErrTimeout(g_cur_username, addr, realcmd)
		}
	} else {
		execute_err = session.Run(realcmd)
	}

	result.End(execute_err)

	result.Stdout = stdout.String()
	result.Stderr = stderr.String()
}
