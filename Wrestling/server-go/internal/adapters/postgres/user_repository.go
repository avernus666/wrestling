package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"wrestling/internal/models"
	"wrestling/internal/ports"
)

type UserRepository struct{ db *pgxpool.Pool }

func NewUserRepository(db *pgxpool.Pool) ports.UserRepository { return &UserRepository{db: db} }

func (r *UserRepository) Create(ctx context.Context, email, passwordHash, displayName string) (models.User, error) {
	id := newID()
	var u models.User
	err := r.db.QueryRow(ctx, `INSERT INTO users(id,email,password_hash,display_name) VALUES($1,$2,$3,$4) RETURNING id,email,display_name,created_at`, id, email, passwordHash, displayName).Scan(&u.ID, &u.Email, &u.DisplayName, &u.CreatedAt)
	return u, err
}
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (models.User, string, error) {
	var u models.User
	var hash string
	err := r.db.QueryRow(ctx, `SELECT id,email,password_hash,display_name,created_at FROM users WHERE email=$1`, strings.ToLower(strings.TrimSpace(email))).Scan(&u.ID, &u.Email, &hash, &u.DisplayName, &u.CreatedAt)
	return u, hash, err
}
func (r *UserRepository) CreateSession(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, _ = r.db.Exec(ctx, `SELECT cleanup_expired_sessions()`)
	_, err := r.db.Exec(ctx, `INSERT INTO sessions(id,user_id,token_hash,expires_at) VALUES($1,$2,$3,$4)`, newID(), userID, tokenHash, expiresAt)
	return err
}
func (r *UserRepository) UserByToken(ctx context.Context, tokenHash string) (models.User, error) {
	var u models.User
	err := r.db.QueryRow(ctx, `SELECT u.id,u.email,u.display_name,u.created_at FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.token_hash=$1 AND s.expires_at>now() AND s.revoked_at IS NULL`, tokenHash).Scan(&u.ID, &u.Email, &u.DisplayName, &u.CreatedAt)
	if err == nil {
		_, _ = r.db.Exec(ctx, `UPDATE sessions SET last_seen_at=now() WHERE token_hash=$1`, tokenHash)
	}
	return u, err
}
func (r *UserRepository) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := r.db.Exec(ctx, `UPDATE sessions SET revoked_at=now() WHERE token_hash=$1`, tokenHash)
	return err
}
func (r *UserRepository) GetProgress(ctx context.Context, userID string) (models.Progress, error) {
	rows, err := r.db.Query(ctx, `SELECT element_id,completed,favorite FROM user_progress WHERE user_id=$1 ORDER BY element_id`, userID)
	if err != nil {
		return models.Progress{}, err
	}
	defer rows.Close()
	p := models.Progress{Items: []models.ProgressItem{}}
	for rows.Next() {
		var x models.ProgressItem
		if err := rows.Scan(&x.ElementID, &x.Completed, &x.Favorite); err != nil {
			return p, err
		}
		p.Items = append(p.Items, x)
		if x.Completed {
			p.CompletedCount++
		}
		if x.Favorite {
			p.FavoriteCount++
		}
	}
	return p, rows.Err()
}
func (r *UserRepository) SetProgress(ctx context.Context, userID string, elementID int, completed, favorite bool) error {
	_, err := r.db.Exec(ctx, `INSERT INTO user_progress(user_id,element_id,completed,favorite) VALUES($1,$2,$3,$4) ON CONFLICT(user_id,element_id) DO UPDATE SET completed=excluded.completed,favorite=excluded.favorite,updated_at=now()`, userID, elementID, completed, favorite)
	return err
}
func (r *UserRepository) CreateWorkout(ctx context.Context, userID, goal string, duration int, equipment string, ids []int) (models.WorkoutSession, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return models.WorkoutSession{}, err
	}
	defer tx.Rollback()
	id := newID()
	now := time.Now().UTC()
	var w models.WorkoutSession
	err = tx.QueryRow(ctx, `INSERT INTO workout_sessions(id,user_id,goal,duration_minutes,equipment,started_at,total_exercises) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id,goal,duration_minutes,equipment,started_at,total_exercises,completed_exercises`, id, userID, goal, duration, equipment, now, len(ids)).Scan(&w.ID, &w.Goal, &w.DurationMinutes, &w.Equipment, &w.StartedAt, &w.TotalExercises, &w.CompletedExercises)
	if err != nil {
		return w, err
	}
	for i, eid := range ids {
		if _, err = tx.Exec(ctx, `INSERT INTO workout_session_exercises(session_id,element_id,position) VALUES($1,$2,$3)`, id, eid, i+1); err != nil {
			return w, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return w, err
	}
	return w, nil
}
func (r *UserRepository) CompleteWorkout(ctx context.Context, userID, sessionID string, completedIDs []int) (models.WorkoutSession, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return models.WorkoutSession{}, err
	}
	defer tx.Rollback()
	var w models.WorkoutSession
	err = tx.QueryRow(ctx, `SELECT id,goal,duration_minutes,equipment,started_at,total_exercises,completed_exercises FROM workout_sessions WHERE id=$1 AND user_id=$2 FOR UPDATE`, sessionID, userID).Scan(&w.ID, &w.Goal, &w.DurationMinutes, &w.Equipment, &w.StartedAt, &w.TotalExercises, &w.CompletedExercises)
	if err != nil {
		return w, err
	}
	for _, eid := range completedIDs {
		if _, err = tx.Exec(ctx, `UPDATE workout_session_exercises SET completed=true WHERE session_id=$1 AND element_id=$2`, sessionID, eid); err != nil {
			return w, err
		}
	}
	err = tx.QueryRow(ctx, `UPDATE workout_sessions SET completed_exercises=(SELECT count(*) FROM workout_session_exercises WHERE session_id=$1 AND completed=true),completed_at=now() WHERE id=$1 RETURNING completed_exercises,completed_at`, sessionID).Scan(&w.CompletedExercises, &w.CompletedAt)
	if err != nil {
		return w, err
	}
	if err = tx.Commit(ctx); err != nil {
		return w, err
	}
	return w, nil
}
func (r *UserRepository) History(ctx context.Context, userID string, limit int) ([]models.WorkoutSession, error) {
	if limit < 1 || limit > 100 {
		limit = 30
	}
	rows, err := r.db.Query(ctx, `SELECT id,goal,duration_minutes,equipment,started_at,completed_at,total_exercises,completed_exercises FROM workout_sessions WHERE user_id=$1 ORDER BY started_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.WorkoutSession{}
	for rows.Next() {
		var w models.WorkoutSession
		if err := rows.Scan(&w.ID, &w.Goal, &w.DurationMinutes, &w.Equipment, &w.StartedAt, &w.CompletedAt, &w.TotalExercises, &w.CompletedExercises); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}
func (r *UserRepository) Analytics(ctx context.Context, userID string) (models.Analytics, error) {
	var a models.Analytics
	err := r.db.QueryRow(ctx, `SELECT count(*),COALESCE(sum(completed_exercises),0),COALESCE(sum(duration_minutes),0) FROM workout_sessions WHERE user_id=$1`, userID).Scan(&a.Sessions, &a.CompletedExercises, &a.Minutes)
	if err != nil {
		return a, err
	}
	rows, err := r.db.Query(ctx, `SELECT started_at::date FROM workout_sessions WHERE user_id=$1 AND completed_at IS NOT NULL GROUP BY started_at::date ORDER BY started_at::date DESC`, userID)
	if err != nil {
		return a, err
	}
	defer rows.Close()
	var dates []time.Time
	for rows.Next() {
		var d time.Time
		if err := rows.Scan(&d); err != nil {
			return a, err
		}
		dates = append(dates, d)
	}
	a.CurrentStreak = streak(dates)
	a.BestStreak = bestStreak(dates)
	return a, rows.Err()
}
func streak(dates []time.Time) int {
	if len(dates) == 0 {
		return 0
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	if dates[0].After(today) {
		return 0
	}
	if today.Sub(dates[0]) > 24*time.Hour {
		return 0
	}
	n := 1
	for i := 1; i < len(dates); i++ {
		if dates[i-1].Sub(dates[i]) == 24*time.Hour {
			n++
		} else {
			break
		}
	}
	return n
}
func bestStreak(dates []time.Time) int {
	if len(dates) == 0 {
		return 0
	}
	best, cur := 1, 1
	for i := 1; i < len(dates); i++ {
		if dates[i-1].Sub(dates[i]) == 24*time.Hour {
			cur++
		} else {
			if cur > best {
				best = cur
			}
			cur = 1
		}
	}
	if cur > best {
		best = cur
	}
	return best
}
func (r *UserRepository) Skills(ctx context.Context, userID string) ([]models.SkillProgress, error) {
	rows, err := r.db.Query(ctx, `SELECT skill_key,level FROM user_skills WHERE user_id=$1 ORDER BY skill_key`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.SkillProgress{}
	for rows.Next() {
		var x models.SkillProgress
		if err := rows.Scan(&x.Key, &x.Level); err != nil {
			return nil, err
		}
		x.MaxLevel = 5
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *UserRepository) SetSkill(ctx context.Context, userID, key string, level int) error {
	if level < 0 {
		level = 0
	}
	if level > 5 {
		level = 5
	}
	_, err := r.db.Exec(ctx, `INSERT INTO user_skills(user_id,skill_key,level) VALUES($1,$2,$3) ON CONFLICT(user_id,skill_key) DO UPDATE SET level=excluded.level,updated_at=now()`, userID, key, level)
	return err
}
func newID() string {
	b := make([]byte, 16)
	_, _ = cryptoRand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%s-%s-%s-%s-%s", hex.EncodeToString(b[:4]), hex.EncodeToString(b[4:6]), hex.EncodeToString(b[6:8]), hex.EncodeToString(b[8:10]), hex.EncodeToString(b[10:]))
}

var cryptoRand = rand.Reader
