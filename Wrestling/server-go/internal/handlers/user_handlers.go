package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"wrestling/internal/models"
	"wrestling/internal/services"
)

type AuthService interface {
	Register(context.Context, string, string, string) (models.User, string, error)
	Login(context.Context, string, string) (models.User, string, error)
	Authenticate(context.Context, string) (models.User, error)
	Logout(context.Context, string) error
}
type UserService interface {
	Progress(context.Context, string) (models.Progress, error)
	SetProgress(context.Context, string, int, bool, bool) error
	CreateWorkout(context.Context, string, string, int, string, []int) (models.WorkoutSession, error)
	CompleteWorkout(context.Context, string, string, []int) (models.WorkoutSession, error)
	History(context.Context, string, int) ([]models.WorkoutSession, error)
	Analytics(context.Context, string) (models.Analytics, error)
	Skills(context.Context, string) ([]models.SkillProgress, error)
	SetSkill(context.Context, string, string, int) error
}

func (h *Handler) authUser(r *http.Request) (models.User, bool) {
	if h.auth == nil {
		return models.User{}, false
	}
	v, err := h.auth.Authenticate(r.Context(), strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	return v, err == nil
}
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var in struct{ Email, Password, Name string }
	if !decodeJSON(w, r, &in) {
		return
	}
	u, t, err := h.auth.Register(r.Context(), in.Email, in.Password, in.Name)
	if errors.Is(err, services.ErrEmailExists) {
		respondError(w, 409, "Email уже зарегистрирован")
		return
	}
	if err != nil {
		respondError(w, 400, "Некорректные данные регистрации")
		return
	}
	respondJSON(w, map[string]any{"user": u, "token": t})
}
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var in struct{ Email, Password string }
	if !decodeJSON(w, r, &in) {
		return
	}
	u, t, err := h.auth.Login(r.Context(), in.Email, in.Password)
	if err != nil {
		respondError(w, 401, "Неверный email или пароль")
		return
	}
	respondJSON(w, map[string]any{"user": u, "token": t})
}
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	u, ok := h.authUser(r)
	if !ok {
		respondError(w, 401, "Требуется авторизация")
		return
	}
	respondJSON(w, u)
}
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.authUser(r); !ok {
		respondError(w, 401, "Требуется авторизация")
		return
	}
	_ = h.auth.Logout(r.Context(), strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	w.WriteHeader(http.StatusNoContent)
}
func (h *Handler) GetProgress(w http.ResponseWriter, r *http.Request) {
	u, ok := h.authUser(r)
	if !ok {
		respondError(w, 401, "Требуется авторизация")
		return
	}
	v, err := h.users.Progress(r.Context(), u.ID)
	if err != nil {
		respondError(w, 500, "Не удалось загрузить прогресс")
		return
	}
	respondJSON(w, v)
}
func (h *Handler) SetProgress(w http.ResponseWriter, r *http.Request) {
	u, ok := h.authUser(r)
	if !ok {
		respondError(w, 401, "Требуется авторизация")
		return
	}
	var in struct {
		ElementID int  `json:"elementId"`
		Completed bool `json:"completed"`
		Favorite  bool `json:"favorite"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.ElementID < 1 {
		respondError(w, 400, "Некорректный элемент")
		return
	}
	if err := h.users.SetProgress(r.Context(), u.ID, in.ElementID, in.Completed, in.Favorite); err != nil {
		respondError(w, 500, "Не удалось сохранить прогресс")
		return
	}
	respondJSON(w, map[string]bool{"ok": true})
}
func (h *Handler) CreateWorkout(w http.ResponseWriter, r *http.Request) {
	u, ok := h.authUser(r)
	if !ok {
		respondError(w, 401, "Требуется авторизация")
		return
	}
	var in struct {
		Goal       string `json:"goal"`
		Duration   int    `json:"duration"`
		Equipment  string `json:"equipment"`
		ElementIDs []int  `json:"elementIds"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	v, err := h.users.CreateWorkout(r.Context(), u.ID, in.Goal, in.Duration, in.Equipment, in.ElementIDs)
	if err != nil {
		respondError(w, 400, "Не удалось создать тренировку")
		return
	}
	respondJSON(w, v)
}
func (h *Handler) CompleteWorkout(w http.ResponseWriter, r *http.Request) {
	u, ok := h.authUser(r)
	if !ok {
		respondError(w, 401, "Требуется авторизация")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/workouts/")
	var in struct {
		CompletedIDs []int `json:"completedIds"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	v, err := h.users.CompleteWorkout(r.Context(), u.ID, id, in.CompletedIDs)
	if err != nil {
		respondError(w, 400, "Не удалось завершить тренировку")
		return
	}
	respondJSON(w, v)
}
func (h *Handler) History(w http.ResponseWriter, r *http.Request) {
	u, ok := h.authUser(r)
	if !ok {
		respondError(w, 401, "Требуется авторизация")
		return
	}
	v, err := h.users.History(r.Context(), u.ID, 30)
	if err != nil {
		respondError(w, 500, "Не удалось загрузить историю")
		return
	}
	respondJSON(w, v)
}
func (h *Handler) Analytics(w http.ResponseWriter, r *http.Request) {
	u, ok := h.authUser(r)
	if !ok {
		respondError(w, 401, "Требуется авторизация")
		return
	}
	v, err := h.users.Analytics(r.Context(), u.ID)
	if err != nil {
		respondError(w, 500, "Не удалось загрузить аналитику")
		return
	}
	respondJSON(w, v)
}
func (h *Handler) Skills(w http.ResponseWriter, r *http.Request) {
	u, ok := h.authUser(r)
	if !ok {
		respondError(w, 401, "Требуется авторизация")
		return
	}
	v, err := h.users.Skills(r.Context(), u.ID)
	if err != nil {
		respondError(w, 500, "Не удалось загрузить навыки")
		return
	}
	respondJSON(w, v)
}
func (h *Handler) SetSkill(w http.ResponseWriter, r *http.Request) {
	u, ok := h.authUser(r)
	if !ok {
		respondError(w, 401, "Требуется авторизация")
		return
	}
	var in struct {
		Key   string `json:"key"`
		Level int    `json:"level"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := h.users.SetSkill(r.Context(), u.ID, in.Key, in.Level); err != nil {
		respondError(w, 400, "Не удалось сохранить навык")
		return
	}
	respondJSON(w, map[string]bool{"ok": true})
}
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
		respondError(w, 405, "Метод не поддерживается")
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		respondError(w, 400, "Некорректный JSON")
		return false
	}
	return true
}

func (h *Handler) CreateCommunity(w http.ResponseWriter, r *http.Request) {
	u, ok := h.authUser(r)
	if !ok {
		respondError(w, 401, "Требуется авторизация")
		return
	}
	var in struct{ Name, Description, Category, Direction, YouTubeURL string }
	if !decodeJSON(w, r, &in) {
		return
	}
	name := strings.TrimSpace(in.Name)
	desc := strings.TrimSpace(in.Description)
	cat := strings.TrimSpace(in.Category)
	dir := strings.TrimSpace(in.Direction)
	url := strings.TrimSpace(in.YouTubeURL)
	if len([]rune(name)) < 2 || len([]rune(name)) > 120 || len([]rune(desc)) < 10 || len([]rune(desc)) > 3000 || (cat != "стойка" && cat != "партер" && cat != "офп") || len([]rune(dir)) < 2 || len([]rune(dir)) > 80 {
		respondError(w, 400, "Некорректные данные элемента")
		return
	}
	videoID, ok := youtubeID(url)
	if !ok {
		respondError(w, 400, "Разрешена только корректная ссылка на YouTube")
		return
	}
	if containsAdultTerms(name + " " + desc) {
		respondError(w, 400, "Материал не прошел автоматическую проверку безопасности")
		return
	}
	verifiedTitle := ""
	var verifyErr error
	if h.youtube != nil {
		verifiedTitle, verifyErr = h.youtube.Verify(r.Context(), videoID)
	}
	if verifyErr != nil || containsAdultTerms(verifiedTitle) {
		respondError(w, 400, "Видео YouTube не прошло проверку безопасности или недоступно")
		return
	}
	item := models.CommunityElement{ID: newUUID(), AuthorID: u.ID, AuthorName: u.DisplayName, Name: name, Description: desc, Category: cat, Direction: dir, YouTubeURL: canonicalYouTubeURL(videoID), YouTubeVideoID: videoID}
	if err := h.content.CreateCommunity(r.Context(), item); err != nil {
		respondError(w, 500, "Не удалось сохранить элемент")
		return
	}
	respondJSON(w, item)
}
func (h *Handler) CommunityComments(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		id := strings.TrimPrefix(r.URL.Path, "/api/community/")
		id = strings.TrimSuffix(id, "/comments")
		items, err := h.content.Comments(r.Context(), id, 100)
		if err != nil {
			respondError(w, 500, "Не удалось загрузить комментарии")
			return
		}
		respondJSON(w, items)
		return
	}
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}
	u, ok := h.authUser(r)
	if !ok {
		respondError(w, 401, "Требуется авторизация")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/community/")
	id = strings.TrimSuffix(id, "/comments")
	var in struct {
		Body string `json:"body"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	body := strings.TrimSpace(in.Body)
	if len([]rune(body)) < 1 || len([]rune(body)) > 1200 {
		respondError(w, 400, "Некорректный комментарий")
		return
	}
	if containsAdultTerms(body) {
		respondError(w, 400, "Комментарий не прошел автоматическую проверку безопасности")
		return
	}
	c := models.CommunityComment{ID: newUUID(), ElementID: id, AuthorID: u.ID, AuthorName: u.DisplayName, Body: body}
	if err := h.content.CreateComment(r.Context(), c); err != nil {
		respondError(w, 500, "Не удалось сохранить комментарий")
		return
	}
	respondJSON(w, c)
}
func newUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return hex.EncodeToString(b)
	}
	return hex.EncodeToString(b[:4]) + "-" + hex.EncodeToString(b[4:6]) + "-" + hex.EncodeToString(b[6:8]) + "-" + hex.EncodeToString(b[8:10]) + "-" + hex.EncodeToString(b[10:])
}
func youtubeID(raw string) (string, bool) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" {
		return "", false
	}
	host := strings.ToLower(u.Hostname())
	if host == "youtu.be" {
		id := strings.Trim(u.Path, "/")
		return validVideoID(id)
	}
	if host == "www.youtube.com" || host == "youtube.com" || host == "m.youtube.com" {
		if id := u.Query().Get("v"); id != "" {
			return validVideoID(id)
		}
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) == 2 && (parts[0] == "shorts" || parts[0] == "embed") {
			return validVideoID(parts[1])
		}
	}
	return "", false
}
func validVideoID(id string) (string, bool) {
	if len(id) != 11 {
		return "", false
	}
	for _, r := range id {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_') {
			return "", false
		}
	}
	return id, true
}
func canonicalYouTubeURL(id string) string { return "https://www.youtube.com/watch?v=" + id }
func containsAdultTerms(s string) bool {
	t := strings.ToLower(s)
	banned := []string{"porn", "xxx", "pornography", "эрот", "порн", "секс", "sex", "nude", "голая", "голые", "18+"}
	for _, x := range banned {
		if strings.Contains(t, x) {
			return true
		}
	}
	return false
}

func (h *Handler) ReportCommunity(w http.ResponseWriter, r *http.Request) {
	u, ok := h.authUser(r)
	if !ok {
		respondError(w, 401, "Требуется авторизация")
		return
	}
	id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/community/"), "/report")
	var in struct {
		Reason string `json:"reason"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	reason := strings.TrimSpace(in.Reason)
	if len(reason) < 3 || len(reason) > 80 {
		respondError(w, 400, "Некорректная причина жалобы")
		return
	}
	if err := h.content.Report(r.Context(), models.CommunityReport{ID: newUUID(), ElementID: id, Reason: reason}, u.ID); err != nil {
		respondError(w, 500, "Не удалось сохранить жалобу")
		return
	}
	respondJSON(w, map[string]bool{"ok": true})
}
