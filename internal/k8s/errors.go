package k8s

import "errors"

var (
	ErrBadRequest        = errors.New("bad request")
	ErrBadResolveRequest = errors.Join(ErrBadRequest, errors.New("invalid resolve request, missing required input"))
)
