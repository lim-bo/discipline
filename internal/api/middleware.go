package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	errorvalues "github.com/limbo/discipline/internal/error_values"
	"github.com/limbo/discipline/pkg/httputil"
)

var (
	requestIDKContextKey = "Request-ID"
	loggerContextKey     = "Logger"
	uidContextKey        = "User-ID"
)

func (s *Server) RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := uuid.New()
		ctx := context.WithValue(r.Context(), requestIDKContextKey, reqID.String())
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) SettingUpLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := slog.Default()
		reqID, ok := r.Context().Value(requestIDKContextKey).(string)
		if ok && reqID != "" {
			logger = logger.With(slog.String("request_id", reqID))
		}
		logger = logger.With(slog.String("from", r.RemoteAddr))
		ctx := context.WithValue(r.Context(), loggerContextKey, logger)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) LoggerExtensionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := GetLoggerFromCtx(r.Context())
		userID, ok := r.Context().Value(uidContextKey).(string)
		if ok && userID != "" {
			logger = logger.With(slog.String("uid", userID))
		}
		ctx := context.WithValue(r.Context(), loggerContextKey, logger)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := GetLoggerFromCtx(r.Context())
		// Getting token from header
		tokenString, err := GetTokenFromHeader(r)
		if err != nil {
			logger.Error("auth failed: invalid token")
			httputil.WriteErrorResponse(w, http.StatusUnauthorized, "authorization failed: invalid token", nil)
			return
		}
		// Getting claims from token string
		tokenClaims, err := s.jwtService.ParseToken(tokenString)
		if err != nil {
			switch {
			case errors.Is(err, errorvalues.ErrInvalidToken):
				logger.Error("auth failed: error parsing token")
				httputil.WriteErrorResponse(w, http.StatusUnauthorized, "authorization failed: invalid token", nil)
				return
			default:
				logger.Error("auth failed: internal error while parsing token", slog.String("error", err.Error()))
				httputil.WriteErrorResponse(w, http.StatusInternalServerError, "error parsing token", nil)
				return
			}
		}
		// Assuring if token is alive
		now := time.Now()
		if tokenClaims.ExpiresAt.Time.Before(now) || tokenClaims.NotBefore.Time.After(now) {
			logger.Error("tried to auth with expired or not ready token")
			httputil.WriteErrorResponse(w, http.StatusUnauthorized, "token expired or not ready", nil)
			return
		}
		uid, err := uuid.Parse(tokenClaims.UserID)
		if err != nil {
			logger.Error("invalid uid in token claims")
			httputil.WriteErrorResponse(w, http.StatusUnauthorized, "invalid token payload", nil)
			return
		}
		// Assuring if user still exists
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		_, err = s.userService.GetByID(ctx, uid)
		if err != nil {
			if errors.Is(err, errorvalues.ErrUserNotFound) {
				logger.Error("user doesn't exist")
				httputil.WriteErrorResponse(w, http.StatusNotFound, "auth failed: user not found", nil)
				return
			}
			logger.Error("error while searching for user", slog.String("error", err.Error()))
			httputil.WriteErrorResponse(w, http.StatusInternalServerError, "internal error while searching for user", nil)
			return
		}
		ctx = context.WithValue(r.Context(), uidContextKey, uid)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func GetLoggerFromCtx(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(loggerContextKey).(*slog.Logger)
	if ok {
		return logger
	}
	return slog.Default()
}

func GetTokenFromHeader(r *http.Request) (string, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		return "", errorvalues.ErrInvalidToken
	}
	parts := strings.Split(token, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errorvalues.ErrInvalidToken
	}
	return parts[1], nil
}

func GetUIDFromContext(r *http.Request) (uuid.UUID, error) {
	uid, ok := r.Context().Value(uidContextKey).(uuid.UUID)
	if !ok {
		return uuid.UUID{}, errors.New("uid invalid or doesn't exists")
	}
	return uid, nil
}
