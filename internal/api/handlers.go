package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	errorvalues "github.com/limbo/discipline/internal/error_values"
	"github.com/limbo/discipline/internal/service"
	"github.com/limbo/discipline/pkg/entity"
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

type CreateHabitRequest struct {
	Title       string `json:"title"`
	Description string `json:"desc"`
}

type GetHabitsResponse struct {
	UserID string          `json:"uid"`
	Page   int             `json:"page"`
	Limit  int             `json:"limit"`
	Habits []*entity.Habit `json:"habits"`
}

func (s *Server) Register(w http.ResponseWriter, r *http.Request) {
	logger := GetLoggerFromCtx(r.Context())
	var req RegisterRequest
	defer r.Body.Close()
	err := sonic.ConfigDefault.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Error("registering error: invalid body")
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
			logger.Error("registering error: existed user")
			httputil.WriteErrorResponse(w, http.StatusConflict, "user with such name already exists", nil)
			return
		}
		logger.Error("registering error: service error", slog.String("error", err.Error()))
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
	defer r.Body.Close()
	err := sonic.ConfigDefault.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Error("login error: invalid body")
		httputil.WriteErrorResponse(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	user, err := s.userService.Login(ctx, req.Name, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, errorvalues.ErrUserNotFound):
			logger.Error("login error: unexist user")
			httputil.WriteErrorResponse(w, http.StatusNotFound, "user with such name doesn't exist", nil)
			return
		case errors.Is(err, errorvalues.ErrWrongCredentials):
			logger.Error("login error: wrong password")
			httputil.WriteErrorResponse(w, http.StatusForbidden, "invalid username or password", nil)
			return
		default:
			logger.Error("login error: service error", slog.String("error", err.Error()))
			httputil.WriteErrorResponse(w, http.StatusInternalServerError, "internal error during login", nil)
			return
		}
	}
	token, err := s.jwtService.GenerateToken(user)
	if err != nil {
		logger.Error("login error: generating token error", slog.String("error", err.Error()))
		httputil.WriteErrorResponse(w, http.StatusInternalServerError, "error creating token", nil)
		return
	}
	httputil.WriteJSONResponse(w, http.StatusOK, map[string]any{
		"uid":   user.ID.String(),
		"token": token,
	})
	logger.Info("successful login")
}

func (s *Server) CreateHabit(w http.ResponseWriter, r *http.Request) {
	logger := GetLoggerFromCtx(r.Context())
	uid, err := GetUIDFromContext(r)
	if err != nil {
		logger.Error("create habit error: unauthorized")
		httputil.WriteErrorResponse(w, http.StatusUnauthorized, "no authorization", nil)
		return
	}
	var req CreateHabitRequest
	defer r.Body.Close()
	err = sonic.ConfigDefault.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Error("create habit error: invalid request body")
		httputil.WriteErrorResponse(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	habit, err := s.habitService.CreateHabit(ctx, uid, service.CreateHabitRequest{
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		switch {
		case errors.Is(err, errorvalues.ErrUserHasHabit):
			logger.Error("create habit error: attempt to create existed habit")
			httputil.WriteErrorResponse(w, http.StatusConflict, "habit already exists", nil)
		case errors.Is(err, errorvalues.ErrUserNotFound):
			logger.Error("create habit error: unexist user")
			httputil.WriteErrorResponse(w, http.StatusNotFound, "couldn't create habit: user doesn't exists", nil)
		default:
			logger.Error("create habit error: service error", slog.String("error", err.Error()))
			httputil.WriteErrorResponse(w, http.StatusInternalServerError, "internal error while creating habit", nil)
		}
		return
	}
	httputil.WriteJSONResponse(w, http.StatusCreated, map[string]any{"habit_id": habit.ID.String()})
	logger.Info("habit created")
}

func (s *Server) GetHabits(w http.ResponseWriter, r *http.Request) {
	logger := GetLoggerFromCtx(r.Context())
	uid, err := GetUIDFromContext(r)
	if err != nil {
		logger.Error("get habits error: unauthorized")
		httputil.WriteErrorResponse(w, http.StatusUnauthorized, "no authorization", nil)
		return
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	offset := (page - 1) * limit
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	habits, err := s.habitService.GetUserHabits(ctx, uid, service.PaginationOpts{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		logger.Error("getting habits list error", slog.String("error", err.Error()))
		httputil.WriteErrorResponse(w, http.StatusInternalServerError, "error while getting habits list", nil)
		return
	}
	httputil.WriteJSONResponse(w, http.StatusOK, GetHabitsResponse{
		UserID: uid.String(),
		Page:   page,
		Limit:  limit,
		Habits: habits,
	})
	logger.Info("habits provided")
}

func (s *Server) DeleteHabit(w http.ResponseWriter, r *http.Request) {
	logger := GetLoggerFromCtx(r.Context())
	uid, err := GetUIDFromContext(r)
	if err != nil {
		logger.Error("habit deletion error: unauthorized")
		httputil.WriteErrorResponse(w, http.StatusUnauthorized, "no authorization", nil)
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		logger.Error("habit deletion error: invalid id in path value")
		httputil.WriteErrorResponse(w, http.StatusBadRequest, "invalid habit id in path value", nil)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err = s.habitService.DeleteHabit(ctx, id, uid)
	if err != nil {
		switch {
		case errors.Is(err, errorvalues.ErrHabitNotFound):
			logger.Error("habit deletion error: unexist habit")
			httputil.WriteErrorResponse(w, http.StatusNotFound, "habit doesn't exist", nil)
		case errors.Is(err, errorvalues.ErrWrongOwner):
			logger.Error("habit deletion error: habit has different owner")
			httputil.WriteErrorResponse(w, http.StatusNotFound, "habit doesn't exist", nil)
		default:
			logger.Error("habit deletion error: service error")
			httputil.WriteErrorResponse(w, http.StatusInternalServerError, "internal error while deleting habit", nil)
		}
		return
	}
}
