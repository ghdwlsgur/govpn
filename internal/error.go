package internal

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"govpn/wraperror"
)

var (
	ErrInvalidParams = errors.New("[err] invalid params")
	ErrUnknown       = errors.New("[err] unknown")
)

func WrapError(err error) error {
	if err != nil {
		pc, _, line, _ := runtime.Caller(1)
		fn := runtime.FuncForPC(pc).Name()
		details := strings.Split(fn, "/")
		fn = details[len(details)-1]
		chainErr := wraperror.Error(err)
		return chainErr.Wrap(fmt.Errorf("[err][%s:%d]", fn, line))
	}
	return nil
}