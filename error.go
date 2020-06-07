package dssh

import (
	"fmt"
)

type ErrTimeout struct {
	str string
}

func newErrTimeout(user string, addr string, cmd string) *ErrTimeout {
	return &ErrTimeout{
		str: fmt.Sprintf("%v@%v [%v] execute timeout", user, addr, cmd),
	}
}

func (self ErrTimeout) Error() string {
	return self.str
}
