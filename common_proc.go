package dssh

import (
	"fmt"
)

var (
	ProcExist_AllOnAll = 1 // 所有进程在所有机器都存在
	ProcExist_0On0     = 2 // 所有进程在所有机器都不存在
)

// 从给定集合中, 筛选出存在的进程
func FilterProcsOnHosts(addrs []string, procs []string, opt *Option) (map[string][]string, error) {
	var cmd = ""
	for _, proc := range procs {
		cmd += fmt.Sprintf("\nif pgrep -x '%[1]v'>/dev/null; then echo '%[1]v'; fi", proc)
	}
	var rets = Runs(addrs, cmd, opt)
	//fmt.Println(cmd)
	var map_host_procs = map[string][]string{}
	for _, ret := range rets {
		if ret.IsSuccess() {
			map_host_procs[ret.Addr] = ret.GetNonEmptyLinesStdout()
		} else if ret.IsExecFail() {
			return nil, ret.GetError()
		} else {
			return nil, ret.GetError()
		}
	}

	return map_host_procs, nil
}

func all_procs_exist_on_all_hosts(hosts []string, procs []string, opt *Option) (bool, error) {
	var map_host_procs, err = FilterProcsOnHosts(hosts, procs, opt)
	if err != nil {
		return false, err
	}
	for _, eprocs := range map_host_procs {
		if len(eprocs) != len(procs) {
			return false, nil
		}
	}
	return true, nil
}

func none_of_procs_exist_on_any_hosts(hosts []string, procs []string, opt *Option) (bool, error) {
	var map_host_procs, err = FilterProcsOnHosts(hosts, procs, opt)
	if err != nil {
		return false, err
	}
	for _, eprocs := range map_host_procs {
		if len(eprocs) > 0 {
			return false, nil
		}
	}
	return true, nil
}

func ProcExist(hosts []string, procs []string, checktype int, opt *Option) (bool, error) {
	switch checktype {
	case ProcExist_AllOnAll:
		return all_procs_exist_on_all_hosts(hosts, procs, opt)
	case ProcExist_0On0:
		return none_of_procs_exist_on_any_hosts(hosts, procs, opt)
	default:
		return false, fmt.Errorf("unknown type[%v]", checktype)
	}
}
func ProcExistErrOnFalse(hosts []string, procs []string, checktype int, opt *Option) error {
	var yes, err = ProcExist(hosts, procs, checktype, opt)
	if err != nil {
		return err
	}
	// 到这里checktype必然是合法的
	switch checktype {
	case ProcExist_AllOnAll:
		return error_on_false(yes, fmt.Sprintf("all procs[%v] should exist on all hosts", procs))
	case ProcExist_0On0:
		return error_on_false(yes, fmt.Sprintf("none of procs[%v] should exist on any hosts", procs))
	}
	return nil
}
