package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"net/mail"
	"strings"
	"time"
	"wrestling/internal/models"
	"wrestling/internal/ports"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrEmailExists = errors.New("email already exists")

const bcryptCost = 12

type AuthService struct {
	repo       ports.UserRepository
	sessionTTL time.Duration
}

func NewAuthService(repo ports.UserRepository, ttl time.Duration) *AuthService {
	return &AuthService{repo: repo, sessionTTL: ttl}
}
func (s *AuthService) Register(ctx context.Context, email, password, name string) (models.User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)
	if _, err := mail.ParseAddress(email); err != nil || len(email) > 254 || len(password) < 10 || len(password) > 128 || name == "" || len([]rune(name)) > 80 {
		return models.User{}, "", ErrInvalidCredentials
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return models.User{}, "", err
	}
	u, err := s.repo.Create(ctx, email, string(hash), name)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return models.User{}, "", ErrEmailExists
		}
		return models.User{}, "", err
	}
	token, err := newToken()
	if err != nil {
		return models.User{}, "", err
	}
	if err = s.repo.CreateSession(ctx, u.ID, hashToken(token), time.Now().Add(s.sessionTTL)); err != nil {
		return models.User{}, "", err
	}
	return u, token, nil
}
func (s *AuthService) Login(ctx context.Context, email, password string) (models.User, string, error) {
	u, hash, err := s.repo.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return models.User{}, "", ErrInvalidCredentials
	}
	token, err := newToken()
	if err != nil {
		return models.User{}, "", err
	}
	if err = s.repo.CreateSession(ctx, u.ID, hashToken(token), time.Now().Add(s.sessionTTL)); err != nil {
		return models.User{}, "", err
	}
	return u, token, nil
}
func (s *AuthService) Authenticate(ctx context.Context, token string) (models.User, error) {
	if token == "" {
		return models.User{}, ErrInvalidCredentials
	}
	return s.repo.UserByToken(ctx, hashToken(token))
}
func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.repo.DeleteSession(ctx, hashToken(token))
}
func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
func hashToken(token string) string {
	b := sha256.Sum256([]byte(token))
	return hex.EncodeToString(b[:])
}

type UserService struct{ repo ports.UserRepository }

func NewUserService(repo ports.UserRepository) *UserService { return &UserService{repo: repo} }
func (s *UserService) Progress(ctx context.Context, userID string) (models.Progress, error) {
	return s.repo.GetProgress(ctx, userID)
}
func (s *UserService) SetProgress(ctx context.Context, userID string, elementID int, completed, favorite bool) error {
	return s.repo.SetProgress(ctx, userID, elementID, completed, favorite)
}
func (s *UserService) CreateWorkout(ctx context.Context, userID, goal string, duration int, equipment string, ids []int) (models.WorkoutSession, error) {
	return s.repo.CreateWorkout(ctx, userID, goal, duration, equipment, ids)
}
func (s *UserService) CompleteWorkout(ctx context.Context, userID, sessionID string, ids []int) (models.WorkoutSession, error) {
	return s.repo.CompleteWorkout(ctx, userID, sessionID, ids)
}
func (s *UserService) History(ctx context.Context, userID string, limit int) ([]models.WorkoutSession, error) {
	return s.repo.History(ctx, userID, limit)
}
func (s *UserService) Analytics(ctx context.Context, userID string) (models.Analytics, error) {
	return s.repo.Analytics(ctx, userID)
}
func (s *UserService) Skills(ctx context.Context, userID string) ([]models.SkillProgress, error) {
	return s.repo.Skills(ctx, userID)
}
func (s *UserService) SetSkill(ctx context.Context, userID, key string, level int) error {
	return s.repo.SetSkill(ctx, userID, key, level)
}

