package dssh

import (
	"fmt"
	//"regexp"
	//"strconv"
	"strings"
)

// 多个机器上的单个账号存在返回true, 有错误返回error. 否则返回false
func AccountExistOnAllHosts(addrs []string, account string, opt *Option) (bool, error) {
	var rets = Runs(addrs, fmt.Sprintf("id %v", account), opt)
	for _, ret := range rets {
		if ret.IsSuccess() {
		} else if ret.IsExecFail() {
			return false, nil
		} else {
			return false, ret.GetError()
		}
	}
	return true, nil
}

// 功能类似AccountExistOnAllHosts. 仅仅是为了方便调用
// 多个机器上的单个账号存在返回nil, 不存在或者有其他错误返回error
func AccountMustExistOnAllHosts(addrs []string, account string, opt *Option) error {
	var yes, err = AccountExistOnAllHosts(addrs, account, opt)
	if err != nil {
		return err
	}
	if !yes {
		return fmt.Errorf("account[%v] not exist", account)
	}
	return nil
}

// 在所有机器上创建指定账号(创建home). 已经存在则跳过. 返回error.
func CreateAccount(addrs []string, account string, groups []string, opt *Option) error {
	var cmd = fmt.Sprintf(`
		if id %[1]v >/dev/null 2>&1; then
			:
		else
			sudo /usr/sbin/useradd -m '%[1]v'
		fi
	`, account)
	if len(groups) > 0 {
		cmd += fmt.Sprintf("\nsudo usermod -G '%v' '%v'", strings.Join(groups, ","), account)
	}
	return RunsErrOnFail(addrs, cmd, opt, fmt.Sprintf("failed to create account[%v]", account))
}

// 设置账号和密码永不过期
func SetAccountPassNeverExpire(addrs []string, account string, opt *Option) error {
	var cmd = fmt.Sprintf(` sudo chage -E -1 -M -1 %v `, account)
	return RunsErrOnFail(addrs, cmd, opt, fmt.Sprintf("failed to set expire of account[%v]", account))
}

// 在所有机器上的指定账号下添加pubkey. 账号不存在视作异常. 返回error.
func AddSSHAuthKey(addrs []string, account string, pubkey string, opt *Option) error {
	var cmd = fmt.Sprintf(`
		if id %[1]v >/dev/null 2>&1; then
			mkdir -p -m 700 /home/%[1]v/.ssh/ 
			echo '%[2]v' >> /home/%[1]v/.ssh/authorized_keys
			chmod 600 /home/%[1]v/.ssh/authorized_keys
			chown -R %[1]v:%[1]v /home/%[1]v/.ssh
		else
			echo "account not exists"
			exit 1
		fi
	`, account, pubkey)
	return SudoRunsErrOnFail(addrs, account, cmd, opt, "")
}
