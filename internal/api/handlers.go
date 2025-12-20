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

const (
	// 30 days
	defaultFromDateDifference = -time.Hour * 24 * 30
)

type RegisterRequest struct {
	Name     string `json:"name" example:"arch_linux_user"`
	Password string `json:"password" example:"secret_password"`
}

type LoginRequest struct {
	Name     string `json:"name" example:"arch_linux_user"`
	Password string `json:"password" example:"secret_password"`
}

type CreateHabitRequest struct {
	Title       string `json:"title" example:"LEG DAY"`
	Description string `json:"desc" example:"hit my legs very hard"`
}

type GetHabitsResponse struct {
	UserID string          `json:"uid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Page   int             `json:"page" example:"1"`
	Limit  int             `json:"limit" example:"10"`
	Habits []*entity.Habit `json:"habits"`
}

type UIDResponse struct {
	UserID string `json:"uid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Token  string `json:"token,omitempty" example:"xxxx.yyyy.zzzz"`
}

type GetChecksResponse struct {
	Values []entity.HabitCheck `json:"values"`
}

// Register godoc
// @Summary Register a new user
// @Description Recieves username and password, registers new user
// @Description and saves in DB.
// @Tags Users
// @Accept json
// @Produce json
// @Param credentials body RegisterRequest true "User's credentials"
// @Success 201 {object} UIDResponse "Response with user ID"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 409 {object} map[string]string "Registering already existed user"
// @Failure 500 {object} map[string]string "Something went wrong internally (in services, repos etc.)"
// @Router /auth/register [post]
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
	httputil.WriteJSONResponse(w, http.StatusCreated, UIDResponse{
		UserID: user.ID.String(),
	})
	logger.Info("successful registration")
}

// Login godoc
// @Summary Authentication with providing token
// @Description Recieves user's credentials and on success returns user ID and auth token.
// @Description Gives back error if user doesn't exist or password is wrong, etc.
// @Tags Users
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "User's credentials"
// @Success 200 {object} UIDResponse "Response with user ID and auth token"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 404 {object} map[string]string "User doesn't exist"
// @Failure 403 {object} map[string]string "Wrong credentials"
// @Failure 500 {object} map[string]string "Something went wrong internally (in services, repos etc.)"
// @Router /auth/login [post]
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
	httputil.WriteJSONResponse(w, http.StatusOK, UIDResponse{
		UserID: user.ID.String(),
		Token:  token,
	})
	logger.Info("successful login")
}

// CreateHabit godoc
// @Summary Creates new user's habit
// @Description Recieves habits' title and description, create new one
// @Description and returns its ID.
// @Tags Habits
// @Accept json
// @Produce json
// @Param Authorization header string true "Access token"
// @Param Habit body CreateHabitRequest true "Habit title and description"
// @Success 201 {object} map[string]string "Response with habit_id"
// @Failure 401 {object} map[string]string "Authorization failed"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 409 {object} map[string]string "Habit with such title already exists"
// @Failure 404 {object} map[string]string "Owner (user) doesn't exist"
// @Failure 500 {object} map[string]string "Something went wrong internally (in services, repos etc.)"
// @Router /habits [post]
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

// GetHabits godoc
// @Summary Provides list of habits
// @Description Provides list of user's habits with pagination in query params (page, limit).
// @Tags Habits
// @Produce json
// @Param Authorization header string true "Access token"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Limit of habits by page" default(10)
// @Success 200 {object} GetHabitsResponse "Response with md (uid, page, limit) and habits list"
// @Failure 401 {object} map[string]string "Authorization failed"
// @Failure 500 {object} map[string]string "Something went wrong internally (in services, repos etc.)"
// @Router /habits [get]
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

// DeleteHabit godoc
// @Summary Deletes habit
// @Description Recieves habit ID in path, deletes it if user is owner.
// @Tags Habits
// @Produce json
// @Param Authorization header string true "Access token"
// @Param id path string true "Habit ID"
// @Success 200
// @Failure 401 {object} map[string]string "Authorization failed"
// @Failure 400 {object} map[string]string "Invalid id param in path"
// @Failure 404 {object} map[string]string "Habit doesn't exist or authorizated user is not its owner"
// @Failure 500 {object} map[string]string "Something went wrong internally (in services, repos etc.)"
// @Router /habits/{id} [delete]
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
			logger.Error("habit deletion error: service error", slog.String("error", err.Error()))
			httputil.WriteErrorResponse(w, http.StatusInternalServerError, "internal error while deleting habit", nil)
		}
		return
	}
	httputil.WriteJSONResponse(w, http.StatusNoContent, map[string]any{
		"message": "habit deleted",
	})
	logger.Info("habit deleted")
}

// CheckHabit godoc
// @Summary Checking habit
// @Description Recieves habit ID in path, checks it if user is owner.
// @Tags Checks
// @Produce json
// @Param Authorization header string true "Access token"
// @Param id path string true "Habit ID"
// @Param date query string false "Check date (yyyy-mm-dd)"
// @Success 200
// @Failure 401 {object} map[string]string "Authorization failed"
// @Failure 400 {object} map[string]string "Invalid id param in path or query date param invalid"
// @Failure 404 {object} map[string]string "Habit doesn't exist or authorizated user is not its owner"
// @Failure 409 {object} map[string]string "Check already exists or date is in future"
// @Failure 500 {object} map[string]string "Something went wrong internally (in services, repos etc.)"
// @Router /habits/check/{id} [post]
func (s *Server) CheckHabit(w http.ResponseWriter, r *http.Request) {
	logger := GetLoggerFromCtx(r.Context())
	uid, err := GetUIDFromContext(r)
	if err != nil {
		logger.Error("check habit error: unauthorized")
		httputil.WriteErrorResponse(w, http.StatusUnauthorized, "no authorization", nil)
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		logger.Error("checking habit error: invalid id in path value")
		httputil.WriteErrorResponse(w, http.StatusBadRequest, "invalid habit id in path value", nil)
		return
	}
	var date time.Time
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		date = time.Now()
	} else {
		date, err = time.Parse(time.DateOnly, dateStr)
		if err != nil {
			logger.Error("checking habit error: invalid date in query params")
			httputil.WriteErrorResponse(w, http.StatusBadRequest, "invalid checking date in query params", nil)
			return
		}
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
	defer cancel()
	err = s.habitCheckService.CheckHabit(ctx, id, uid, date)
	if err != nil {
		switch {
		case errors.Is(err, errorvalues.ErrHabitNotFound):
			logger.Error("checking habit error: habit not found")
			httputil.WriteErrorResponse(w, http.StatusNotFound, "habit not found", nil)
		case errors.Is(err, errorvalues.ErrWrongOwner):
			logger.Error("checking habit error: habits' owner doesn't match")
			httputil.WriteErrorResponse(w, http.StatusNotFound, "habit not found", nil)
		case errors.Is(err, errorvalues.ErrCheckExist):
			logger.Error("checking habit error: habit already checked on that date")
			httputil.WriteErrorResponse(w, http.StatusConflict, "habit already checked", nil)
		case errors.Is(err, errorvalues.ErrCheckDateNotAllowed):
			logger.Error("checking habit error: checking habit on unallowed date")
			httputil.WriteErrorResponse(w, http.StatusConflict, "habit check date not allowed", nil)
		default:
			logger.Error("checking habit error: service error", slog.String("error", err.Error()))
			httputil.WriteErrorResponse(w, http.StatusInternalServerError, "internal error", err)
		}
		return
	}
	httputil.WriteJSONResponse(w, http.StatusCreated, map[string]any{
		"message": "habit checked",
	})
	logger.Info("habbit checked")
}

// UncheckHabit godoc
// @Summary Unchecking habit
// @Description Recieves habit ID in path, deletes check on it if user is owner.
// @Tags Checks
// @Produce json
// @Param Authorization header string true "Access token"
// @Param id path string true "Habit ID"
// @Param date query string false "Check date (yyyy-mm-dd)"
// @Success 200
// @Failure 401 {object} map[string]string "Authorization failed"
// @Failure 400 {object} map[string]string "Invalid id param in path"
// @Failure 404 {object} map[string]string "Habit doesn't exist or authorizated user is not its owner"
// @Failure 409 {object} map[string]string "Check doesn't exist"
// @Failure 500 {object} map[string]string "Something went wrong internally (in services, repos etc.)"
// @Router /habits/check/{id} [delete]
func (s *Server) UncheckHabit(w http.ResponseWriter, r *http.Request) {
	logger := GetLoggerFromCtx(r.Context())
	uid, err := GetUIDFromContext(r)
	if err != nil {
		logger.Error("deleting check on habit error: unauthorized")
		httputil.WriteErrorResponse(w, http.StatusUnauthorized, "no authorization", nil)
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		logger.Error("deleting check on habit error: invalid id in path value")
		httputil.WriteErrorResponse(w, http.StatusBadRequest, "invalid habit id in path value", nil)
		return
	}
	var date time.Time
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		date = time.Now()
	} else {
		date, err = time.Parse(time.DateOnly, dateStr)
		if err != nil {
			logger.Error("deleting check on habit error: invalid date in query params")
			httputil.WriteErrorResponse(w, http.StatusBadRequest, "invalid checking date in query params", nil)
			return
		}
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
	defer cancel()
	err = s.habitCheckService.UncheckHabit(ctx, id, uid, date)
	if err != nil {
		switch {
		case errors.Is(err, errorvalues.ErrHabitNotFound):
			logger.Error("deleting check on habit error: habit not found")
			httputil.WriteErrorResponse(w, http.StatusNotFound, "habit not found", nil)
		case errors.Is(err, errorvalues.ErrWrongOwner):
			logger.Error("deleting check on habit error: habits' owner doesn't match")
			httputil.WriteErrorResponse(w, http.StatusNotFound, "habit not found", nil)
		case errors.Is(err, errorvalues.ErrCheckNotFound):
			logger.Error("deleting check on habit error: habit haven't checked on that date")
			httputil.WriteErrorResponse(w, http.StatusConflict, "habit already checked", nil)
		default:
			logger.Error("deleting check on habit error: service error", slog.String("error", err.Error()))
			httputil.WriteErrorResponse(w, http.StatusInternalServerError, "internal error", err)
		}
		return
	}
	httputil.WriteJSONResponse(w, http.StatusNoContent, map[string]any{
		"message": "habit unchecked",
	})
	logger.Info("habbit unchecked")
}

// GetHabitStats godoc
// @Summary Habit statistics
// @Description Recieves habit ID in path and provides habit stats (habitID, total checks,
// @Description current streak, max streak)
// @Tags Checks
// @Produce json
// @Param Authorization header string true "Access token"
// @Param id path string true "Habit ID"
// @Success 200 {object} entity.HabitStats "Habit statistics"
// @Failure 401 {object} map[string]string "Authorization failed"
// @Failure 400 {object} map[string]string "Invalid id param in path"
// @Failure 404 {object} map[string]string "Habit doesn't exist or authorizated user is not its owner"
// @Failure 500 {object} map[string]string "Something went wrong internally (in services, repos etc.)"
// @Router /habits/stat/{id} [get]
func (s *Server) GetHabitStats(w http.ResponseWriter, r *http.Request) {
	logger := GetLoggerFromCtx(r.Context())
	uid, err := GetUIDFromContext(r)
	if err != nil {
		logger.Error("getting habit stats error: unauthorized")
		httputil.WriteErrorResponse(w, http.StatusUnauthorized, "no authorization", nil)
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		logger.Error("getting habit stats error: invalid id in path value")
		httputil.WriteErrorResponse(w, http.StatusBadRequest, "invalid habit id in path value", nil)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
	defer cancel()
	stats, err := s.habitCheckService.GetHabitStats(ctx, id, uid)
	if err != nil {
		switch {
		case errors.Is(err, errorvalues.ErrHabitNotFound):
			logger.Error("getting habit stats error: habit not found")
			httputil.WriteErrorResponse(w, http.StatusNotFound, "habit not found", nil)
		case errors.Is(err, errorvalues.ErrWrongOwner):
			logger.Error("getting habit stats error: habit not found")
			httputil.WriteErrorResponse(w, http.StatusNotFound, "habit not found", nil)
		default:
			logger.Error("getting habit stats error: service error", slog.String("error", err.Error()))
			httputil.WriteErrorResponse(w, http.StatusInternalServerError, "internal error", err)
		}
		return
	}
	httputil.WriteJSONResponse(w, http.StatusOK, stats)
	logger.Info("stats provided")
}

// GetChecks godoc
// @Summary Habit checks list
// @Description Recieves habit ID in path and provides habit checks list
// @Tags Checks
// @Produce json
// @Param Authorization header string true "Access token"
// @Param id path string true "Habit ID"
// @Param from query string false "Checks time bound in the past, default - 30-days ago (yyyy-mm-dd)"
// @Param to query string false "Checks higher time bound, default - today (yyyy-mm-dd)"
// @Success 200 {object} GetChecksResponse "Checks list"
// @Failure 401 {object} map[string]string "Authorization failed"
// @Failure 400 {object} map[string]string "Invalid id param in path"
// @Failure 404 {object} map[string]string "Habit doesn't exist or authorizated user is not its owner"
// @Failure 500 {object} map[string]string "Something went wrong internally (in services, repos etc.)"
// @Router /habits/check/{id} [get]
func (s *Server) GetChecks(w http.ResponseWriter, r *http.Request) {
	logger := GetLoggerFromCtx(r.Context())
	uid, err := GetUIDFromContext(r)
	if err != nil {
		logger.Error("getting habit stats error: unauthorized")
		httputil.WriteErrorResponse(w, http.StatusUnauthorized, "no authorization", nil)
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		logger.Error("getting habit stats error: invalid id in path value")
		httputil.WriteErrorResponse(w, http.StatusBadRequest, "invalid habit id in path value", nil)
		return
	}
	var fromDate, toDate time.Time
	now := time.Now()
	fromDate, err = time.Parse(time.DateOnly, r.URL.Query().Get("from"))
	// If from query param is invalid or not provided, setting to 30 days-ago date
	if err != nil {
		fromDate = now.Add(defaultFromDateDifference)
	}
	toDate, err = time.Parse(time.DateOnly, r.URL.Query().Get("to"))
	// If to query param is invalid or not provided, setting to now
	if err != nil {
		toDate = now
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
	defer cancel()
	checks, err := s.habitCheckService.GetHabitChecks(ctx, id, uid, fromDate, toDate)
	if err != nil {
		switch {
		case errors.Is(err, errorvalues.ErrHabitNotFound):
			logger.Error("getting habit checks error: habit not found")
			httputil.WriteErrorResponse(w, http.StatusNotFound, "habit not found", nil)
		case errors.Is(err, errorvalues.ErrWrongOwner):
			logger.Error("getting habit checks error: wrong habit owner")
			httputil.WriteErrorResponse(w, http.StatusNotFound, "habit not found", nil)
		default:
			logger.Error("getting habit checks error: service error", slog.String("error", err.Error()))
			httputil.WriteErrorResponse(w, http.StatusInternalServerError, "internal error", err)
		}
		return
	}
	httputil.WriteJSONResponse(w, http.StatusOK, GetChecksResponse{
		Values: checks,
	})
	logger.Info("checks provided")
}
