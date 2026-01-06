package check

import (
	"github.com/mkch/gg/errorcheck"
	"github.com/mkch/gg/errortrace"
)

func Must[T any](val T, err error) T {
	return errorcheck.Must(errortrace.Panic, val, err)
}

func MustOK(err error) {
	errorcheck.MustOK(errortrace.Panic, err)
}
