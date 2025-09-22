package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	errorvalues "github.com/limbo/discipline/internal/error_values"
	"github.com/limbo/discipline/internal/service"
	"github.com/limbo/discipline/pkg/httputil"
)

type RegisterRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (s *Server) Register(w http.ResponseWriter, r *http.Request) {
	logger := GetLoggerFromCtx(r.Context())
	var req RegisterRequest
	err := sonic.ConfigDefault.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Error("register request with invalid body")
		httputil.WriteErrorResponse(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	user, err := s.userService.Register(ctx, &service.RegisterRequest{
		Name:     req.Name,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, errorvalues.ErrUserExists) {
			logger.Error("registration request for existed user")
			httputil.WriteErrorResponse(w, http.StatusConflict, "user with such name already exists", nil)
			return
		}
		logger.Error("service error on registration", slog.String("error", err.Error()))
		httputil.WriteErrorResponse(w, http.StatusInternalServerError, "internal error during registration", nil)
		return
	}
	httputil.WriteJSONResponse(w, http.StatusCreated, map[string]any{
		"uid": user.ID.String(),
	})
	logger.Info("successful registration")
}

func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	logger := GetLoggerFromCtx(r.Context())
	var req LoginRequest
	err := sonic.ConfigDefault.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Error("login request with invalid body")
		httputil.WriteErrorResponse(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	user, err := s.userService.Login(ctx, req.Name, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, errorvalues.ErrUserNotFound):
			logger.Error("login to unexist user")
			httputil.WriteErrorResponse(w, http.StatusNotFound, "user with such name doesn't exist", nil)
			return
		case errors.Is(err, errorvalues.ErrWrongCredentials):
			logger.Error("login with wrong password")
			httputil.WriteErrorResponse(w, http.StatusForbidden, "invalid username or password", nil)
			return
		default:
			logger.Error("service error on login", slog.String("error", err.Error()))
			httputil.WriteErrorResponse(w, http.StatusInternalServerError, "internal error during login", nil)
			return
		}
	}
	token, err := s.jwtService.GenerateToken(user)
	if err != nil {
		logger.Error("error while generating token", slog.String("error", err.Error()))
		httputil.WriteErrorResponse(w, http.StatusInternalServerError, "error creating token", nil)
		return
	}
	httputil.WriteJSONResponse(w, http.StatusOK, map[string]any{
		"uid":   user.ID.String(),
		"token": token,
	})
	logger.Info("successful login")
}
