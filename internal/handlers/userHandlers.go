package handlers

import (
	"fmt"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"github.com/kurt4ins/taskmanager/internal/repo"
	"github.com/kurt4ins/taskmanager/internal/utils"
)

func validateUser(user repo.RequestUser) ([]string, bool) {
	var msg []string

	if user.Name == "" {
		msg = append(msg, "name wasn't provided")
	}

	if user.Email == "" {
		msg = append(msg, "email wasn't provided")
	} else if _, err := mail.ParseAddress(user.Email); err != nil {
		msg = append(msg, "invalid email")
	}

	if user.Password == "" {
		msg = append(msg, "password wasn't provided")
	}

	return msg, len(msg) == 0
}

type UserHandler struct {
	repo repo.UserRepository
}

func NewUserHandler(repo repo.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var data repo.RequestUser
	if err := utils.ReadJSON(w, r, &data); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if msg, ok := validateUser(data); !ok {
		utils.WriteError(w, http.StatusBadRequest, strings.Join(msg, "; "))
		return
	}

	user, err := h.repo.Create(r.Context(), data)
	if err != nil {
		fmt.Println(err.Error())
		utils.WriteError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, user)
}

func (h *UserHandler) GetById(w http.ResponseWriter, r *http.Request) {
	strId := r.PathValue("id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid id provided")
		return
	}

	user, err := h.repo.GetById(r.Context(), id)
	if err != nil {
		fmt.Println(err.Error())
		utils.WriteError(w, http.StatusInternalServerError, "failed to fetch user")
		return
	}
	if user == nil {
		utils.WriteError(w, http.StatusNotFound, fmt.Sprintf("user with id %d doesn't exist", id))
		return
	}

	utils.WriteJSON(w, http.StatusOK, user)
}

func (h *UserHandler) GetByEmail(w http.ResponseWriter, r *http.Request) {
	email := r.PathValue("email")

	if _, err := mail.ParseAddress(email); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid email")
		return
	}

	user, err := h.repo.GetByEmail(r.Context(), email)
	if err != nil {
		fmt.Println(err.Error())
		utils.WriteError(w, http.StatusInternalServerError, "failed to fetch user")
		return
	}
	if user == nil {
		utils.WriteError(w, http.StatusNotFound, fmt.Sprintf("user with email %s doesn't exist", email))
		return
	}

	utils.WriteJSON(w, http.StatusOK, user)
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := 20, 0

	if strLimit := r.URL.Query().Get("limit"); strLimit != "" {
		l, err := strconv.Atoi(strLimit)
		if err != nil || l <= 0 {
			utils.WriteError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = l
	}

	if strOffset := r.URL.Query().Get("offset"); strOffset != "" {
		o, err := strconv.Atoi(strOffset)
		if err != nil || o < 0 {
			utils.WriteError(w, http.StatusBadRequest, "invalid offset")
			return
		}
		offset = o
	}

	users, err := h.repo.List(r.Context(), limit, offset)
	if err != nil {
		fmt.Println(err.Error())
		utils.WriteError(w, http.StatusInternalServerError, "failed to fetch users")
		return
	}

	if users == nil {
		users = []repo.User{}
	}

	utils.WriteJSON(w, http.StatusOK, users)
}
