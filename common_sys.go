package dssh

import (
	"fmt"
	//"regexp"
	//"strconv"
)

func SetPerformance(hosts []string, opt *Option) error {
	const change_perf_keyword = "change_perf"
	var cmd = fmt.Sprintf(`
        if cpupower  frequency-info | grep 'unknown cpufreq driver'; then
            exit 0
        fi
        sed --follow-symlinks -i '/%[1]v/d' /etc/rc.local
        chmod a+x /etc/rc.local
        echo 'sleep 100 && echo performance | tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor #%[1]v' >> /etc/rc.local
        echo performance | tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor
    `, change_perf_keyword)
	return SudoRunsErrOnFail(hosts, "root", cmd, nil, "failed to set performance")
}
func CheckOSVersion(addrs []string, v string, opt *Option) (bool, error) {
	var cmd = "cat /etc/redhat-release"
	var rets = Runs(addrs, cmd, opt)
	for _, ret := range rets {
		if ret.IsSuccess() {
			if ret.GetStripStdout() != v {
				return false, nil
			}
		} else {
			return false, ret.GetError()
		}
	}
	return true, nil
}
func AssertOSVersion(addrs []string, v string, opt *Option) error {
	var cmd = "cat /etc/redhat-release"
	var rets = Runs(addrs, cmd, opt)
	for _, ret := range rets {
		if ret.IsSuccess() {
			if ret.GetStripStdout() != v {
				return fmt.Errorf("[%v] wrong osversion. expect[%v]", ret.Addr, v)
			}
		} else {
			return ret.GetError()
		}
	}
	return nil
}
