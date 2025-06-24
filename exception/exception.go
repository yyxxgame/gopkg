package exception

import (
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
)

type Exception interface {
}

type TryStruct struct {
	catches map[string]HandlerEx
	try     func()
}

func Try(tryHandler func()) *TryStruct {
	tryStruct := TryStruct{
		catches: make(map[string]HandlerEx),
		try:     tryHandler,
	}
	return &tryStruct
}

type HandlerEx func(Exception)

func (sel *TryStruct) Catches(exceptionId string, catch func(Exception)) *TryStruct {
	sel.catches[exceptionId] = catch
	return sel
}

func (sel *TryStruct) Catch(catch func(Exception)) *TryStruct {
	sel.catches["default"] = catch
	return sel
}

func (sel *TryStruct) Finally(finally func()) {
	defer func() {
		if e := recover(); nil != e {
			exception := ""
			switch e.(type) {
			case string:
				exception = e.(string)
			case error:
				err := e.(error)
				exception = err.Error()
			}

			if catch, ok := sel.catches[exception]; ok {
				catch(e)
			} else if catch, ok = sel.catches["default"]; ok {
				catch(e)
			}

			logx.ErrorStack(e)
		}
		finally()
	}()

	sel.try()
}

func Throw(err interface{}) Exception {
	switch err.(type) {
	case string:
		err = errors.New(err.(string))
	case int32:
		errStr := fmt.Sprintf("%d", err)
		err = errors.New(errStr)
	}
	panic(err)
}

func MakeError(errCode int32, errStr string) error {
	s := fmt.Sprintf("%d@%s", errCode, errStr)
	return errors.New(s)
}
