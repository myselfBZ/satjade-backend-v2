package main

func (a *api) notFoundLog(method string, path string, err error) {
	a.logger.Warnw("not found error", "method", method, "path", path, "error", err.Error(), "type", "notfound")
}

func (a *api) badRequestLog(method string, path string, err error) {
	a.logger.Errorw("bad request error", "method", method, "path", path, "error", err.Error(), "type", "badRequest")
}

func (a *api) internalErrLog(method string, path string, err error) {
	a.logger.Errorw("internal error", "method", method, "path", path, "error", err.Error(), "type","internalError")
}

func (a *api) unauthorizedLog(method string, path string, err error) {
	a.logger.Errorw("unauthorized request", "method", method, "path", path, "error", err.Error(), "type", "unauthorized")
}

func (a *api) conflictLog(method string, path string, err error) {
	a.logger.Errorw("conflicted request", "method", method, "path", path, "error", err.Error(), "type", "conflict")
}

