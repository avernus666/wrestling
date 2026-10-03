package models

import "time"

type Requirement struct {
	Name   string `json:"name"`
	Muscle string `json:"muscle"`
}

type Element struct {
	ID              int           `json:"id"`
	Name            string        `json:"name"`
	Icon            string        `json:"icon"`
	Category        string        `json:"category"`
	Subcategory     string        `json:"subcategory,omitempty"`
	Difficulty      string        `json:"difficulty"`
	DifficultyLabel string        `json:"difficultyLabel"`
	Reps            string        `json:"reps"`
	Muscles         []string      `json:"muscles"`
	Image           string        `json:"image"`
	Video           string        `json:"video,omitempty"`
	VideoSearch     string        `json:"videoSearch,omitempty"`
	Description     string        `json:"description"`
	Requirements    []Requirement `json:"requirements"`
	Steps           []string      `json:"steps"`
	Tips            string        `json:"tips"`
	DemoAvailable   bool          `json:"demoAvailable,omitempty"`
}

type Bar struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Icon        string   `json:"icon"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Features    []string `json:"features"`
	Tips        string   `json:"tips"`
}

type BaseBlock struct {
	Title   string   `json:"title"`
	Content string   `json:"content,omitempty"`
	Items   []string `json:"items,omitempty"`
}

type SafetyItem struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Icon        string   `json:"icon"`
	Category    string   `json:"category"`
	Image       string   `json:"image"`
	Description string   `json:"description"`
	Benefits    []string `json:"benefits"`
	Tips        string   `json:"tips"`
}

type Stats struct {
	Total        int            `json:"total"`
	ByCategory   map[string]int `json:"byCategory"`
	ByDifficulty map[string]int `json:"byDifficulty"`
	Categories   []string       `json:"categories"`
	Difficulties []string       `json:"difficulties"`
}

type User struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	CreatedAt   time.Time `json:"createdAt"`
}
type ProgressItem struct {
	ElementID int  `json:"elementId"`
	Completed bool `json:"completed"`
	Favorite  bool `json:"favorite"`
}
type Progress struct {
	Items          []ProgressItem `json:"items"`
	CompletedCount int            `json:"completedCount"`
	FavoriteCount  int            `json:"favoriteCount"`
}
type WorkoutSession struct {
	ID                 string     `json:"id"`
	Goal               string     `json:"goal"`
	DurationMinutes    int        `json:"durationMinutes"`
	Equipment          string     `json:"equipment"`
	StartedAt          time.Time  `json:"startedAt"`
	CompletedAt        *time.Time `json:"completedAt,omitempty"`
	TotalExercises     int        `json:"totalExercises"`
	CompletedExercises int        `json:"completedExercises"`
}
type SkillProgress struct {
	Key      string `json:"key"`
	Level    int    `json:"level"`
	MaxLevel int    `json:"maxLevel"`
}
type Analytics struct {
	Sessions           int `json:"sessions"`
	CompletedExercises int `json:"completedExercises"`
	Minutes            int `json:"minutes"`
	CurrentStreak      int `json:"currentStreak"`
	BestStreak         int `json:"bestStreak"`
}

type CommunityElement struct {
	ID             string    `json:"id"`
	AuthorID       string    `json:"authorId"`
	AuthorName     string    `json:"authorName"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Category       string    `json:"category"`
	Direction      string    `json:"direction"`
	YouTubeURL     string    `json:"youtubeUrl"`
	YouTubeVideoID string    `json:"youtubeVideoId"`
	CreatedAt      time.Time `json:"createdAt"`
	Comments       int       `json:"comments"`
}

type CommunityComment struct {
	ID         string    `json:"id"`
	ElementID  string    `json:"elementId"`
	AuthorID   string    `json:"authorId"`
	AuthorName string    `json:"authorName"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"createdAt"`
}

type CommunityReport struct {
	ID        string `json:"id"`
	ElementID string `json:"elementId"`
	Reason    string `json:"reason"`
}
