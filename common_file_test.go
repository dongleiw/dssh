package dssh

import (
	"crypto/md5"
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
func Test_BackupPath(t *testing.T) {
	// 创建临时文件
	var file = "/tmp/dssh.test.backupfile"
	tassert_err(t, RunsErrOnFail(g_hosts, fmt.Sprintf("[ ! -f %[1]v ] && touch %[1]v", file), nil, "failed to create backupfile"))

	// 测试
	tassert_err(t, BackupPath(g_hosts, "", file, true, nil))
}
func Test_FileExist(t *testing.T) {
	// 创建临时文件
	var file = "/tmp/dssh.test.fileexist"
	tassert_err(t, RunsErrOnFail(g_hosts, fmt.Sprintf("[ ! -f %[1]v ] && touch %[1]v", file), nil, "failed to create tmpfiel"))

	// 测试
	var exist, err = FileExist(g_hosts, "", file, nil)
	tassert_err(t, err)
	tassert_bool(t, exist)

	// 清理
	tassert_err(t, RunsErrOnFail(g_hosts, fmt.Sprintf("rm %v", file), nil, "failed to rm tmpfile"))
}
func Test_FileMustExist(t *testing.T) {
	// 创建临时文件
	var file = "/tmp/dssh.test.filemustexist"
	tassert_err(t, RunsErrOnFail(g_hosts, fmt.Sprintf("[ ! -f %[1]v ] && touch %[1]v", file), nil, "failed to create tmpfile"))

	// 测试
	tassert_err(t, FileMustExist(g_hosts, "", file, nil))

	// 清理
	tassert_err(t, RunsErrOnFail(g_hosts, fmt.Sprintf("rm %v", file), nil, "failed to rm tmpfile"))
}
func Test_Md5OfFile(t *testing.T) {
	// 创建临时文件
	var file = "/tmp/dssh.test.md5offile"
	var data = "abcd"
	var data_md5 = fmt.Sprintf("%x", md5.Sum([]byte(data)))
	tassert_err(t, RunsErrOnFail(g_hosts, fmt.Sprintf("[ ! -f %[1]v ] && echo -n '%[2]v' > %[1]v", file, data), nil, "failed to create tmpfile"))

	// 测试
	var map_host_md5, err = Md5OfFile(g_hosts, "", file, nil)
	tassert_err(t, err)
	for _, md5 := range map_host_md5 {
		tassert_bool(t, md5 == data_md5)
	}

	// 清理
	tassert_err(t, RunsErrOnFail(g_hosts, fmt.Sprintf("rm %v", file), nil, "failed to rm tmpfile"))
}
