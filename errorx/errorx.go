package errorx

import "errors"

var (
	ErrorAcceptSocket   = errors.New("accept socket failed")
	ErrorEngineShutdown = errors.New("engine shutdown")
)
