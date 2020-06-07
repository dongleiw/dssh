package dssh

import (
	//doputils "dop/utils"
	"fmt"
	"regexp"
	"strconv"
)

var (
	PortExist_AllOnAll = 1 // 所有端口在所有机器都存在
	PortExist_0On0     = 2 // 所有端口在所有机器都不存在
	PortExist_AllOn1   = 3 // 所有端口在其中一个机器都存在, 在其他所有机器都不存在
)

// 获取一个机器监听的端口信息
func PortsOnHost(host string, opt *Option) ([]int, error) {
	var cmd = " netstat -ntpl | awk '{print $4}' "
	/*
		cmd 输出格式如下. 之所以不在remote直接处理好, 是因为grep匹配不到会返回失败, 不太好处理
			(only
			Local
			127.0.0.1:9000
			0.0.0.0:22
	*/
	var ret = Run(host, cmd, opt)
	if ret.IsSuccess() {
		var ports []int
		var re, err = regexp.Compile(":([0-9]+)$")
		if err != nil {
			return nil, err
		}
		for _, line := range ret.GetNonEmptyLinesStdout() {
			var reret = re.FindStringSubmatch(line)
			if len(reret) <= 1 {
			} else {
				if port, err := strconv.ParseInt(reret[1], 0, 0); err != nil {
					return nil, err
				} else {
					ports = append(ports, int(port))
				}
			}
		}
		return ports, nil
	} else if ret.IsExecFail() {
		// 正常不应该失败的. 如果执行失败, 那应该是遇到未知情况了
		return nil, ret.GetError()
	} else {
		return nil, ret.GetError()
	}

}

// 获取一批机器监听端口信息
func PortsOnHosts(hosts []string, opt *Option) (map[string][]int, error) {
	var re, err = regexp.Compile(":([0-9]+)$")
	if err != nil {
		return nil, err
	}
	var map_host_ports = map[string][]int{}

	var cmd = " netstat -ntpl | awk '{print $4}' "
	/*
		cmd 输出格式如下. 之所以不在remote直接处理好, 是因为grep匹配不到会返回失败, 不太好处理
			(only
			Local
			127.0.0.1:9000
			0.0.0.0:22
	*/
	var rets = Runs(hosts, cmd, opt)

	for _, ret := range rets {
		if ret.IsSuccess() {
			var ports []int
			for _, line := range ret.GetNonEmptyLinesStdout() {
				var reret = re.FindStringSubmatch(line)
				if len(reret) <= 1 {
				} else {
					if port, err := strconv.ParseInt(reret[1], 0, 0); err != nil {
						return nil, err
					} else {
						ports = append(ports, int(port))
					}
				}
			}
			map_host_ports[ret.Addr] = ports
		} else if ret.IsExecFail() {
			// 正常不应该失败的. 如果执行失败, 那应该是遇到未知情况了
			return nil, ret.GetError()
		} else {
			return nil, ret.GetError()
		}
	}

	return map_host_ports, nil
}

// 获取一批机器上指定端口集合中存在的. 有错误返回error. 没有错误返回监听中的端口
func FilterPortsOnHosts(hosts []string, ports []int, opt *Option) (map[string][]int, error) {
	var map_host_ports, err = PortsOnHosts(hosts, opt)
	if err != nil {
		return nil, err
	}
	var map_host_filter = map[string][]int{}

	for _, host := range hosts {
		var exists_ports = map_host_ports[host]
		var filter []int
		for _, port := range ports {
			if intarray_contains(exists_ports, port) {
				filter = append(filter, port)
			}
		}
		map_host_filter[host] = filter
	}

	return map_host_filter, nil
}

func all_ports_exit_on_all_hosts(hosts []string, ports []int, opt *Option) (bool, error) {
	var map_host_filterport, err = FilterPortsOnHosts(hosts, ports, opt)
	if err != nil {
		return false, err
	}
	for _, eports := range map_host_filterport {
		if len(eports) != len(ports) {
			return false, nil
		}
	}
	return true, nil
}
func none_of_ports_exit_on_any_hosts(hosts []string, ports []int, opt *Option) (bool, error) {
	var map_host_filterport, err = FilterPortsOnHosts(hosts, ports, opt)
	if err != nil {
		return false, err
	}
	for _, eports := range map_host_filterport {
		if len(eports) > 0 {
			return false, nil
		}
	}
	return true, nil
}
func all_ports_exit_on_one_of_hosts(hosts []string, ports []int, opt *Option) (bool, error) {
	var map_host_filterport, err = FilterPortsOnHosts(hosts, ports, opt)
	if err != nil {
		return false, err
	}
	var exist_count = 0
	for _, eports := range map_host_filterport {
		if len(eports) == len(ports) {
			exist_count++
		} else if len(eports) == 0 {
			// slave
		} else {
			// 只有部分端口存在
			return false, nil
		}
	}
	return exist_count == 1, nil
}

func PortExist(hosts []string, ports []int, checktype int, opt *Option) (bool, error) {
	switch checktype {
	case PortExist_AllOnAll:
		return all_ports_exit_on_all_hosts(hosts, ports, opt)
	case PortExist_0On0:
		return none_of_ports_exit_on_any_hosts(hosts, ports, opt)
	case PortExist_AllOn1:
		return all_ports_exit_on_one_of_hosts(hosts, ports, opt)
	default:
		return false, fmt.Errorf("unknown type[%v]", checktype)
	}
}
func PortExistErrOnFalse(hosts []string, ports []int, checktype int, opt *Option) error {
	var yes, err = PortExist(hosts, ports, checktype, opt)
	if err != nil {
		return err
	}
	// 到这里checktype必然是合法的
	switch checktype {
	case PortExist_AllOnAll:
		return error_on_false(yes, fmt.Sprintf("all ports[%v] should exist on all hosts", ports))
	case PortExist_0On0:
		return error_on_false(yes, fmt.Sprintf("none of ports[%v] should exist on any hosts", ports))
	case PortExist_AllOn1:
		return error_on_false(yes, fmt.Sprintf("there should be only one host that has ports[%v]", ports))
	}
	return nil
}
