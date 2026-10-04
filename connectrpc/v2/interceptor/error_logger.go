package interceptor

import (
	"context"
	stderrors "errors"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	"github.com/brunoluiz/x/core/v2/errcode"
	"github.com/samber/oops"
)

type Logger interface {
	Info(string, ...any)
	Warn(string, ...any)
	Error(string, ...any)
}

func ErrorLogger(logger Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			level := slog.LevelInfo
			startedAt := time.Now()
			resp, err := next(ctx, req)
			finishedAt := time.Now()

			// Handle oops errors
			if oopsErr, ok := oops.AsOops(err); ok {
				var connectErr *connect.Error
				publicErr := stderrors.New(oopsErr.Public())
				attrs := []slog.Attr{
					slog.String("message", oopsErr.Error()),
					slog.String("summary", oopsErr.Public()),
				}

				switch oopsErr.Code() {
				case errcode.CodeNotFound.String():
					level = slog.LevelWarn
					connectErr = connect.NewError(connect.CodeNotFound, publicErr)
				case errcode.CodeValidation.String():
					level = slog.LevelWarn
					connectErr = connect.NewError(connect.CodeInvalidArgument, publicErr)
				case errcode.CodeConflict.String():
					level = slog.LevelWarn
					connectErr = connect.NewError(connect.CodeAlreadyExists, publicErr)
				default:
					level = slog.LevelError
					attrs = append(attrs, slog.String("stack", oopsErr.Stacktrace()))
					connectErr = connect.NewError(connect.CodeInternal, publicErr)
				}

				err = connectErr
				log(logger, level, "connectrpc request", "error", slog.GroupAttrs("error", attrs...), "method", req.Spec().Procedure, "duration", finishedAt.Sub(startedAt))
			} else if err != nil {
				level = slog.LevelError
				log(logger, level, "connectrpc request", "error", slog.Group("error", "message", err.Error()), "method", req.Spec().Procedure, "duration", finishedAt.Sub(startedAt))
			} else {
				log(logger, level, "connectrpc request", "method", req.Spec().Procedure, "duration", finishedAt.Sub(startedAt))
			}

			return resp, err
		}
	}
}

func log(logger Logger, level slog.Level, message string, args ...any) {
	switch {
	case level >= slog.LevelError:
		logger.Error(message, args...)
	case level >= slog.LevelWarn:
		logger.Warn(message, args...)
	default:
		logger.Info(message, args...)
	}
}
