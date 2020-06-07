package dssh

import (
	"fmt"
	"path"
	"runtime"
	"strings"
	"testing"
)

func intarray_contains(a []int, k int) bool {
	for _, v := range a {
		if v == k {
			return true
		}
	}
	return false
}

func tassert_bool(t *testing.T, b bool) {
	if !b {
		var filename = ""
		var linenum = 0
		var caller_name = ""

		var pc, _, _, ok = runtime.Caller(1)
		if ok {
			fn := runtime.FuncForPC(pc)
			var sps = strings.Split(fn.Name(), ".")
			caller_name = sps[len(sps)-1]
			filename, linenum = fn.FileLine(pc)
			filename = path.Base(filename)
		}

		t.Fatalf("[%v:%v %v]", filename, linenum, caller_name)
	}
}
func tassert_err(t *testing.T, err error) {
	if err != nil {
		var filename = ""
		var linenum = 0
		var caller_name = ""

		var pc, _, _, ok = runtime.Caller(1)
		if ok {
			fn := runtime.FuncForPC(pc)
			var sps = strings.Split(fn.Name(), ".")
			caller_name = sps[len(sps)-1]
			filename, linenum = fn.FileLine(pc)
			filename = path.Base(filename)
		}

		t.Fatalf("[%v:%v %v] %v", filename, linenum, caller_name, err)
	}
}

func error_on_false(b bool, msg string) error {
	if b {
		return nil
	} else {
		return fmt.Errorf(msg)
	}
}
