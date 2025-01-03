package httputil

import (
	"context"
)

type contextKey string

const (
	PathParamsKey contextKey = "path_params"
)

func GetPathParam(ctx context.Context, param string) string {
	params, ok := ctx.Value(PathParamsKey).(map[string]string)
	if !ok {
		return ""
	}
	return params[param]
}
