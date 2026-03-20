package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kurt4ins/taskmanager/internal/repo"
	"github.com/kurt4ins/taskmanager/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	repo   repo.UserRepository
	secret []byte
}

func NewAuthHandler(repo repo.UserRepository, secret []byte) *AuthHandler {
	return &AuthHandler{repo: repo, secret: secret}
}

func (h *AuthHandler) getTokenPair(userId int) (string, string, error) {
	access, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  userId,
		"type": "access",
		"exp":  time.Now().Add(2 * time.Hour).Unix(),
	}).SignedString(h.secret)
	if err != nil {
		return "", "", err
	}

	refresh, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  userId,
		"type": "refresh",
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
	}).SignedString(h.secret)
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var data map[string]string
	var msg []string

	if err := utils.ReadJSON(w, r, &data); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if _, ok := data["email"]; !ok {
		msg = append(msg, "email wasn't provided")
	} else if _, err := mail.ParseAddress(data["email"]); err != nil {
		msg = append(msg, "invalid email")
	}

	if _, ok := data["password"]; !ok {
		msg = append(msg, "password wasn't provided")
	}

	if len(msg) != 0 {
		utils.WriteError(w, http.StatusBadRequest, strings.Join(msg, "; "))
		return
	}

	user, err := h.repo.GetByEmail(r.Context(), data["email"])
	if err != nil {
		fmt.Println(err.Error())
		utils.WriteError(w, http.StatusInternalServerError, "failed to fetch user")
		return
	}
	if user == nil {
		utils.WriteError(w, http.StatusNotFound, fmt.Sprintf("user with email %s doesn't exist", data["email"]))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(data["password"])); err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	access, refresh, err := h.getTokenPair(user.Id)
	if err != nil {
		fmt.Println(err.Error())
		utils.WriteError(w, http.StatusInternalServerError, "failed to sign token")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"access_token":  access,
		"refresh_token": refresh,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var data map[string]string

	if err := utils.ReadJSON(w, r, &data); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	tokenStr, ok := data["refresh_token"]
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, "refresh token missing")
		return
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return h.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			utils.WriteError(w, http.StatusUnauthorized, "token expired")
		} else {
			utils.WriteError(w, http.StatusUnauthorized, "invalid token")
		}
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "invalid token claims")
		return
	}

	userId, ok := claims["sub"].(float64)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "invalid token claims")
		return
	}

	if claims["type"] != "refresh" {
		utils.WriteError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	access, _, err := h.getTokenPair(int(userId))
	if err != nil {
		fmt.Println(err.Error())
		utils.WriteError(w, http.StatusInternalServerError, "failed to sign token")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"access_token": access})
}
