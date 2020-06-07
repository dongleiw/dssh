package dssh

import (
	"fmt"
	"path"
	"time"
	//"regexp"
	//"strconv"
)

// 切换到account账号, 获取filepath文件的md5. 返回map: addr->md5
// 如果addrs中有重复, map会去重. 此时len(map)<len(addrs)
func Md5OfFile(addrs []string, account string, filepath string, opt *Option) (map[string]string, error) {
	var rets = SudoRuns(addrs, account, fmt.Sprintf("md5sum %v | awk '{print $1}'", filepath), opt)
	var map_host_md5 = map[string]string{}
	for _, ret := range rets {
		if ret.IsSuccess() {
			map_host_md5[ret.Addr] = ret.GetStripStdout()
		} else {
			return nil, ret.GetError()
		}
	}
	return map_host_md5, nil
}

// 切换到account账号, 获取filepath文件的md5. 返回map: md5 -> 出现次数
// 如果addrs中有重复, map会去重. 此时len(map)<len(addrs)
func Md5CountOfFile(addrs []string, account string, filepath string, opt *Option) (map[string]int, error) {
	var rets = SudoRuns(addrs, account, fmt.Sprintf("md5sum %v | awk '{print $1}'", filepath), opt)
	var map_md5_count = map[string]int{}
	for _, ret := range rets {
		if ret.IsSuccess() {
			map_md5_count[ret.GetStripStdout()]++
		} else {
			return nil, ret.GetError()
		}
	}
	return map_md5_count, nil
}

// 切换到account账号, 检查filepath文件是否存在.
// 如果无异常且所有机器上都存在返回true,nil
func FileExist(addrs []string, account string, filepath string, opt *Option) (bool, error) {
	var rets = SudoRuns(addrs, account, fmt.Sprintf("[ -f '%v' ]", filepath), opt)
	for _, ret := range rets {
		if ret.IsSuccess() {
		} else {
			return false, ret.GetError()
		}
	}
	return true, nil
}

// 功能类似FileExist. 如果有机器上不存在返回也error
func FileMustExist(addrs []string, account string, filepath string, opt *Option) error {
	var exist, err = FileExist(addrs, account, filepath, opt)
	if err != nil {
		return err
	}
	if !exist {
		return fmt.Errorf("file[%v] not exist", filepath)
	}
	return nil
}

// 功能类似FileExist. 如果有机器上不存在返回也error
// 将文件filepath备份到所在路径下. 假设原名为A, 新文件名为A.{unixtime}.XXXXXX
// @must_exist 备份目标是否必须存在. 如果不符合预期, 则报错
func BackupPath(addrs []string, account string, filepath string, must_exist bool, opt *Option) error {
	var name = path.Base(filepath)
	var dir = path.Dir(filepath)
	var unixtime = time.Now().Unix()
	var cmd = ""
	var not_exist_code = 0
	if must_exist {
		not_exist_code = 1
	} else {
		not_exist_code = 0
	}
	cmd = fmt.Sprintf(` 
			if [ -d %[1]v -o -f %[1]v ]; then
				 cd %[2]v && mv %[3]v $(mktemp -d %[3]v.%[4]v.XXXXXX) 
			else
				exit %[5]v
			fi
			`, filepath, dir, name, unixtime, not_exist_code)
	var rets = SudoRuns(addrs, account, cmd, opt)
	for _, ret := range rets {
		if ret.IsSuccess() {
		} else {
			return ret.GetError()
		}
	}
	return nil
}
