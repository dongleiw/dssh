package dssh

import (
	"fmt"
)

type ErrTimeout struct {
	str string
}

func newErrTimeout(user string, addr string, cmd string) *ErrTimeout {
	return &ErrTimeout{
		str: fmt.Sprintf("execute timeout"),
	}
}

func (self ErrTimeout) Error() string {
	return self.str
}
