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
func Test_Runs(t *testing.T) {
	// 测试
	var rets = Runs(g_hosts, "true", nil)
	for _, ret := range rets {
		fmt.Printf("%+v\n", ret)
	}
}
