package interceptor

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	"github.com/brunoluiz/x/apperr"
	"github.com/samber/oops"
)

func ErrorLogger(logger *slog.Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			l := logger
			level := slog.LevelInfo
			startedAt := time.Now()
			resp, err := next(ctx, req)
			finishedAt := time.Now()

			// Handle oops errors
			if oopsErr, ok := oops.AsOops(err); ok {
				var connectErr *connect.Error
				publicErr := errors.New(oopsErr.Public())
				attrs := []slog.Attr{
					slog.String("message", oopsErr.Error()),
					slog.String("summary", oopsErr.Public()),
				}

				switch oopsErr.Code() {
				case apperr.CodeNotFound.String():
					level = slog.LevelWarn
					connectErr = connect.NewError(connect.CodeNotFound, publicErr)
				case apperr.CodeValidation.String():
					level = slog.LevelWarn
					connectErr = connect.NewError(connect.CodeInvalidArgument, publicErr)
				case apperr.CodeConflict.String():
					level = slog.LevelWarn
					connectErr = connect.NewError(connect.CodeAlreadyExists, publicErr)
				default:
					level = slog.LevelError
					attrs = append(attrs, slog.String("stack", oopsErr.Stacktrace()))
					connectErr = connect.NewError(connect.CodeInternal, publicErr)
				}

				l = l.With(slog.GroupAttrs("error", attrs...))
				err = connectErr
			} else if err != nil {
				level = slog.LevelError
				l = l.With(slog.GroupAttrs("error", slog.String("message", err.Error())))
			}

			l.Log(ctx, level, "connectrpc request",
				slog.String("method", req.Spec().Procedure),
				slog.Duration("duration", finishedAt.Sub(startedAt)),
			)

			return resp, err
		}
	}
}
