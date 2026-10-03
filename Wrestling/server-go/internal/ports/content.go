package ports

import (
	"context"
	"time"
	"wrestling/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, email, passwordHash, displayName string) (models.User, error)
	FindByEmail(ctx context.Context, email string) (models.User, string, error)
	CreateSession(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error
	UserByToken(ctx context.Context, tokenHash string) (models.User, error)
	DeleteSession(ctx context.Context, tokenHash string) error
	GetProgress(ctx context.Context, userID string) (models.Progress, error)
	SetProgress(ctx context.Context, userID string, elementID int, completed, favorite bool) error
	CreateWorkout(ctx context.Context, userID, goal string, duration int, equipment string, elementIDs []int) (models.WorkoutSession, error)
	CompleteWorkout(ctx context.Context, userID, sessionID string, completedIDs []int) (models.WorkoutSession, error)
	History(ctx context.Context, userID string, limit int) ([]models.WorkoutSession, error)
	Analytics(ctx context.Context, userID string) (models.Analytics, error)
	Skills(ctx context.Context, userID string) ([]models.SkillProgress, error)
	SetSkill(ctx context.Context, userID, key string, level int) error
}

type ElementRepository interface {
	All(ctx context.Context, filter ElementQuery) ([]models.Element, error)
	ByID(ctx context.Context, id int) (models.Element, bool, error)
	ByCategory(ctx context.Context, category string) ([]models.Element, error)
	Stats(ctx context.Context) (models.Stats, error)
}
type ElementQuery struct{ Category, Subcategory, Difficulty, Search string }
type ContentRepository interface {
	ListCommunity(ctx context.Context, category, search string, limit, offset int) ([]models.CommunityElement, error)
	CreateCommunity(ctx context.Context, item models.CommunityElement) error
	ListComments(ctx context.Context, elementID string, limit int) ([]models.CommunityComment, error)
	CreateComment(ctx context.Context, comment models.CommunityComment) error
	CreateReport(ctx context.Context, report models.CommunityReport, reporterID string) error
	Elements() ElementRepository
	Bars(ctx context.Context) ([]models.Bar, error)
	Base(ctx context.Context) ([]models.BaseBlock, error)
	Safety(ctx context.Context) ([]models.SafetyItem, error)
}

type CommunityRepository interface {
	ListCommunity(ctx context.Context, category, search string, limit, offset int) ([]models.CommunityElement, error)
	CreateCommunity(ctx context.Context, item models.CommunityElement) error
	ListComments(ctx context.Context, elementID string, limit int) ([]models.CommunityComment, error)
	CreateComment(ctx context.Context, comment models.CommunityComment) error
	CreateReport(ctx context.Context, report models.CommunityReport, reporterID string) error
}
