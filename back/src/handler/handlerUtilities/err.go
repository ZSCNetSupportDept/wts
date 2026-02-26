package hutil

import "errors"

var ErrUnknown = errors.New("unknown error:")
var ErrUnknownInner = errors.New("sorry, but there is No further information")

type WtsErr struct {
	Current error
	inner   error //如果不是未捕获错误，那么这个字段没有太大用处，因为设计上logic会封装已知的错误写进Current
}

func (e *WtsErr) Error() string {
	if e.Current != ErrUnknown {
		return e.Current.Error()
	}
	return ErrUnknown.Error() + e.inner.Error()
}

// implements Unwrap interface
func (e *WtsErr) Unwrap() error {
	if e.inner == nil {
		return ErrUnknownInner
	}
	return e.inner
}

func NewWtsErr(current error, inner error) *WtsErr {
	return &WtsErr{
		Current: current,
		inner:   inner,
	}
}
func NewUnknownErr(inner error) *WtsErr {
	return &WtsErr{
		Current: ErrUnknown,
		inner:   inner,
	}
}

func IsKnownErr(err error) bool {
	if err == nil {
		return true
	}
	wtsErr, ok := err.(*WtsErr)
	if !ok {
		return false
	}
	return wtsErr.Current != ErrUnknown
}
