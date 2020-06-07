package dssh

import (
	"fmt"
	"os"
	"testing"
)

type PortInfo struct {
	common_exist_ports []int // 两个机器都有的端口
	uniq_exist_ports   []int // 只有本机有的端口
	exist_ports        []int // 存在的端口
	not_exist_ports    []int // 不存在的端口
}

var g_map_host_portinfo = map[string]*PortInfo{}
var g_hosts []string

func init() {
	if err := Initialize(true); err != nil {
		fmt.Printf("failed to init: %v", err)
	}

	// 解析参数
	var another = os.Getenv("test_another_host")

	// 本机
	var local, err = os.Hostname()
	if err != nil {
		fmt.Printf("failed to get hostname: %v", err)
	}
	var port_info = &PortInfo{
		common_exist_ports: []int{65530},
		uniq_exist_ports:   []int{65531},
		exist_ports:        []int{65530, 65531},
		not_exist_ports:    []int{65532},
	}
	g_map_host_portinfo[local] = port_info
	g_hosts = append(g_hosts, local)
	for _, port := range port_info.exist_ports {
		var ret = Run(local, fmt.Sprintf("ncat -l localhost %v </dev/null >/dev/null &", port), nil)
		if !ret.IsSuccess() {
			panic(ret.GetError())
		}
	}

	// 另一台
	port_info = &PortInfo{
		common_exist_ports: []int{65530},
		uniq_exist_ports:   []int{65532},
		exist_ports:        []int{65530, 65532},
		not_exist_ports:    []int{65531},
	}
	g_map_host_portinfo[another] = port_info
	g_hosts = append(g_hosts, another)
	for _, port := range port_info.exist_ports {
		var ret = Run(another, fmt.Sprintf("ncat -l localhost %v </dev/null >/dev/null &", port), nil)
		if !ret.IsSuccess() {
			panic(ret.GetError())
		}
	}

}

func Test_PortsOnHost(t *testing.T) {
	eports, err := PortsOnHost(g_hosts[0], nil)
	tassert_err(t, err)
	for _, port := range g_map_host_portinfo[g_hosts[0]].exist_ports {
		tassert_bool(t, intarray_contains(eports, port))
	}
	for _, port := range g_map_host_portinfo[g_hosts[0]].not_exist_ports {
		tassert_bool(t, !intarray_contains(eports, port))
	}
}
func Test_PortsOnHosts(t *testing.T) {
	map_host_eports, err := PortsOnHosts(g_hosts, nil)
	tassert_err(t, err)
	tassert_bool(t, len(g_hosts) == len(map_host_eports))
	for _, host := range g_hosts {
		eports, e := map_host_eports[host]
		tassert_bool(t, e)
		for _, port := range g_map_host_portinfo[host].exist_ports {
			tassert_bool(t, intarray_contains(eports, port))
		}
	}
}
func Test_FilterPortsOnHosts(t *testing.T) {
	var ports []int
	ports = append(ports, g_map_host_portinfo[g_hosts[0]].exist_ports...)
	ports = append(ports, g_map_host_portinfo[g_hosts[0]].not_exist_ports...)
	ports = append(ports, g_map_host_portinfo[g_hosts[1]].exist_ports...)
	ports = append(ports, g_map_host_portinfo[g_hosts[1]].not_exist_ports...)

	map_host_eports, err := FilterPortsOnHosts(g_hosts, ports, nil)
	tassert_err(t, err)
	tassert_bool(t, len(g_hosts) == len(map_host_eports))
	for _, host := range g_hosts {
		eports, e := map_host_eports[host]
		tassert_bool(t, e)
		for _, port := range g_map_host_portinfo[host].exist_ports {
			tassert_bool(t, intarray_contains(eports, port))
		}
		for _, port := range g_map_host_portinfo[host].not_exist_ports {
			tassert_bool(t, !intarray_contains(eports, port))
		}
	}
}
func Test_PortExist(t *testing.T) {
	//
}
