package dssh

import (
	"fmt"
	"os"
	"testing"
)

var g_hosts []string

func init() {
	if err := Initialize(true); err != nil {
		fmt.Printf("failed to init: %v", err)
	}

	var local, err = os.Hostname()
	if err != nil {
		fmt.Printf("failed to get hostname: %v", err)
	}
	var another = os.Getenv("test_another_host")
	g_hosts = []string{local, another}
}
func Test_AccountExistOnAllHosts(t *testing.T) {
	var account = "dsshtestAccountExistOnAllHosts"

	// 测试不存在
	var exist, err = AccountExistOnAllHosts(g_hosts, account, nil)
	tassert_err(t, err)
	tassert_bool(t, exist == false)

	// 创建
	tassert_err(t, RunsErrOnFail(g_hosts, fmt.Sprintf("sudo useradd %v", account), nil, "failed to create tmpaccount"))

	// 测试存在
	exist, err = AccountExistOnAllHosts(g_hosts, account, nil)
	tassert_err(t, err)
	tassert_bool(t, exist)

	// 测试存在
	tassert_err(t, AccountMustExistOnAllHosts(g_hosts, account, nil))

	// 删除
	tassert_err(t, RunsErrOnFail(g_hosts, fmt.Sprintf("sudo userdel %v", account), nil, "failed to rm tmpaccount"))
}
func Test_CreateAccount(t *testing.T) {
	var account = "dsshtestAccountExistOnAllHosts"

	// 测试不存在
	var exist, err = AccountExistOnAllHosts(g_hosts, account, nil)
	tassert_err(t, err)
	tassert_bool(t, exist == false)

	// 创建
	tassert_err(t, CreateAccount(g_hosts, account, nil, nil))

	// 删除
	tassert_err(t, RunsErrOnFail(g_hosts, fmt.Sprintf("sudo userdel %v", account), nil, "failed to rm tmpaccount"))
}
