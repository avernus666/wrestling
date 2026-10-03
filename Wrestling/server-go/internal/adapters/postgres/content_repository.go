package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"wrestling/internal/models"
	"wrestling/internal/ports"
)

type postgresContentRepository struct {
	db       *pgxpool.Pool
	elements *postgresElementRepository
}

type postgresElementRepository struct {
	db *pgxpool.Pool
}

func NewContentRepository(db *pgxpool.Pool) ports.ContentRepository {
	return &postgresContentRepository{
		db:       db,
		elements: &postgresElementRepository{db: db},
	}
}

func (r *postgresContentRepository) Elements() ports.ElementRepository { return r.elements }

func (r *postgresContentRepository) Bars(ctx context.Context) ([]models.Bar, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, icon, type, description, features, tips
		FROM bars ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("query bars: %w", err)
	}
	defer rows.Close()

	result := make([]models.Bar, 0)
	for rows.Next() {
		var item models.Bar
		if err := rows.Scan(&item.ID, &item.Name, &item.Icon, &item.Type, &item.Description, &item.Features, &item.Tips); err != nil {
			return nil, fmt.Errorf("scan bar: %w", err)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bars: %w", err)
	}
	return result, nil
}

func (r *postgresContentRepository) Base(ctx context.Context) ([]models.BaseBlock, error) {
	rows, err := r.db.Query(ctx, `
		SELECT title, content, items
		FROM base_blocks ORDER BY sort_order, id`)
	if err != nil {
		return nil, fmt.Errorf("query base blocks: %w", err)
	}
	defer rows.Close()

	result := make([]models.BaseBlock, 0)
	for rows.Next() {
		var item models.BaseBlock
		if err := rows.Scan(&item.Title, &item.Content, &item.Items); err != nil {
			return nil, fmt.Errorf("scan base block: %w", err)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate base blocks: %w", err)
	}
	return result, nil
}

func (r *postgresContentRepository) Safety(ctx context.Context) ([]models.SafetyItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, icon, category, image, description, benefits, tips
		FROM safety_items ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("query safety items: %w", err)
	}
	defer rows.Close()

	result := make([]models.SafetyItem, 0)
	for rows.Next() {
		var item models.SafetyItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Icon, &item.Category, &item.Image, &item.Description, &item.Benefits, &item.Tips); err != nil {
			return nil, fmt.Errorf("scan safety item: %w", err)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate safety items: %w", err)
	}
	return result, nil
}

func (r *postgresElementRepository) All(ctx context.Context, filter ports.ElementQuery) ([]models.Element, error) {
	query, args := buildElementQuery(`
		SELECT id, name, icon, category, subcategory, difficulty, difficulty_label,
		       reps, muscles, image, video, video_search, description, requirements, steps, tips, demo_available
		FROM elements`, filter)
	query += " ORDER BY id"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query elements: %w", err)
	}
	defer rows.Close()
	return scanElements(rows)
}

func (r *postgresElementRepository) ByID(ctx context.Context, id int) (models.Element, bool, error) {
	var item models.Element
	var requirements []byte
	err := r.db.QueryRow(ctx, `
		SELECT id, name, icon, category, subcategory, difficulty, difficulty_label,
		       reps, muscles, image, video, video_search, description, requirements, steps, tips, demo_available
		FROM elements WHERE id = $1`, id).
		Scan(&item.ID, &item.Name, &item.Icon, &item.Category, &item.Subcategory,
			&item.Difficulty, &item.DifficultyLabel, &item.Reps, &item.Muscles,
			&item.Image, &item.Video, &item.VideoSearch, &item.Description,
			&requirements, &item.Steps, &item.Tips, &item.DemoAvailable)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.Element{}, false, nil
		}
		return models.Element{}, false, fmt.Errorf("query element %d: %w", id, err)
	}
	if err := json.Unmarshal(requirements, &item.Requirements); err != nil {
		return models.Element{}, false, fmt.Errorf("decode requirements for element %d: %w", id, err)
	}
	return item, true, nil
}

func (r *postgresElementRepository) ByCategory(ctx context.Context, category string) ([]models.Element, error) {
	return r.All(ctx, ports.ElementQuery{Category: category})
}

func (r *postgresElementRepository) Stats(ctx context.Context) (models.Stats, error) {
	var stats models.Stats
	stats.ByCategory = map[string]int{}
	stats.ByDifficulty = map[string]int{}

	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM elements`).Scan(&total); err != nil {
		return models.Stats{}, fmt.Errorf("count elements: %w", err)
	}
	stats.Total = total

	rows, err := r.db.Query(ctx, `
		SELECT category, count(*) FROM elements GROUP BY category ORDER BY category`)
	if err != nil {
		return models.Stats{}, fmt.Errorf("stats categories: %w", err)
	}
	for rows.Next() {
		var key string
		var count int
		if err := rows.Scan(&key, &count); err != nil {
			rows.Close()
			return models.Stats{}, fmt.Errorf("scan category stats: %w", err)
		}
		stats.ByCategory[key] = count
		stats.Categories = append(stats.Categories, key)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return models.Stats{}, fmt.Errorf("iterate category stats: %w", err)
	}
	rows.Close()

	rows, err = r.db.Query(ctx, `
		SELECT difficulty, count(*) FROM elements GROUP BY difficulty ORDER BY difficulty`)
	if err != nil {
		return models.Stats{}, fmt.Errorf("stats difficulties: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var key string
		var count int
		if err := rows.Scan(&key, &count); err != nil {
			return models.Stats{}, fmt.Errorf("scan difficulty stats: %w", err)
		}
		stats.ByDifficulty[key] = count
		stats.Difficulties = append(stats.Difficulties, key)
	}
	if err := rows.Err(); err != nil {
		return models.Stats{}, fmt.Errorf("iterate difficulty stats: %w", err)
	}
	return stats, nil
}

func buildElementQuery(base string, filter ports.ElementQuery) (string, []any) {
	conditions := make([]string, 0, 4)
	args := make([]any, 0, 6)

	add := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}

	if category := strings.TrimSpace(filter.Category); category != "" && category != "все" {
		add("category = $%d", category)
	}
	if subcategory := strings.TrimSpace(filter.Subcategory); subcategory != "" {
		add("subcategory = $%d", subcategory)
	}
	if difficulty := strings.TrimSpace(filter.Difficulty); difficulty != "" {
		add("difficulty = $%d", difficulty)
	}
	if search := strings.TrimSpace(filter.Search); search != "" {
		pattern := "%" + search + "%"
		n := len(args) + 1
		conditions = append(conditions, fmt.Sprintf(`(name ILIKE $%d OR description ILIKE $%d OR EXISTS (
			SELECT 1 FROM unnest(muscles) AS muscle WHERE muscle ILIKE $%d
		))`, n, n+1, n+2))
		args = append(args, pattern, pattern, pattern)
	}
	if len(conditions) == 0 {
		return base, args
	}
	return base + " WHERE " + strings.Join(conditions, " AND "), args
}

func scanElements(rows pgx.Rows) ([]models.Element, error) {
	result := make([]models.Element, 0)
	for rows.Next() {
		var item models.Element
		var requirements []byte
		if err := rows.Scan(&item.ID, &item.Name, &item.Icon, &item.Category, &item.Subcategory,
			&item.Difficulty, &item.DifficultyLabel, &item.Reps, &item.Muscles, &item.Image,
			&item.Video, &item.VideoSearch, &item.Description, &requirements, &item.Steps,
			&item.Tips, &item.DemoAvailable); err != nil {
			return nil, fmt.Errorf("scan element: %w", err)
		}
		if err := json.Unmarshal(requirements, &item.Requirements); err != nil {
			return nil, fmt.Errorf("decode requirements for element %d: %w", item.ID, err)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate elements: %w", err)
	}
	return result, nil
}

func (r *postgresContentRepository) ListCommunity(ctx context.Context, category, search string, limit, offset int) ([]models.CommunityElement, error) {
	if limit < 1 || limit > 50 {
		limit = 24
	}
	if offset < 0 {
		offset = 0
	}
	query := `SELECT ce.id, ce.author_id, u.display_name, ce.name, ce.description, ce.category, ce.direction, ce.youtube_url, ce.youtube_video_id, ce.created_at,
		(SELECT count(*) FROM community_comments cc WHERE cc.element_id=ce.id)
		FROM community_elements ce JOIN users u ON u.id=ce.author_id WHERE ce.moderation_status='approved'`
	args := []any{}
	if category == "стойка" || category == "партер" || category == "офп" {
		args = append(args, category)
		query += fmt.Sprintf(" AND ce.category=$%d", len(args))
	}
	if search = strings.TrimSpace(search); search != "" {
		args = append(args, "%"+search+"%")
		n := len(args)
		query += fmt.Sprintf(" AND (ce.name ILIKE $%d OR ce.description ILIKE $%d OR ce.direction ILIKE $%d)", n, n, n)
	}
	args = append(args, limit, offset)
	query += fmt.Sprintf(" ORDER BY ce.created_at DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list community: %w", err)
	}
	defer rows.Close()
	out := make([]models.CommunityElement, 0)
	for rows.Next() {
		var x models.CommunityElement
		if err := rows.Scan(&x.ID, &x.AuthorID, &x.AuthorName, &x.Name, &x.Description, &x.Category, &x.Direction, &x.YouTubeURL, &x.YouTubeVideoID, &x.CreatedAt, &x.Comments); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *postgresContentRepository) CreateCommunity(ctx context.Context, item models.CommunityElement) error {
	_, err := r.db.Exec(ctx, `INSERT INTO community_elements(id,author_id,name,description,category,direction,youtube_url,youtube_video_id,moderation_status,moderation_reason) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, item.ID, item.AuthorID, item.Name, item.Description, item.Category, item.Direction, item.YouTubeURL, item.YouTubeVideoID, "approved", "")
	return err
}
func (r *postgresContentRepository) ListComments(ctx context.Context, elementID string, limit int) ([]models.CommunityComment, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}
	rows, err := r.db.Query(ctx, `SELECT cc.id,cc.element_id,cc.author_id,u.display_name,cc.body,cc.created_at FROM community_comments cc JOIN users u ON u.id=cc.author_id WHERE cc.element_id=$1 ORDER BY cc.created_at ASC LIMIT $2`, elementID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.CommunityComment, 0)
	for rows.Next() {
		var c models.CommunityComment
		if err := rows.Scan(&c.ID, &c.ElementID, &c.AuthorID, &c.AuthorName, &c.Body, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
func (r *postgresContentRepository) CreateComment(ctx context.Context, comment models.CommunityComment) error {
	_, err := r.db.Exec(ctx, `INSERT INTO community_comments(id,element_id,author_id,body) VALUES($1,$2,$3,$4)`, comment.ID, comment.ElementID, comment.AuthorID, comment.Body)
	return err
}

func (r *postgresContentRepository) CreateReport(ctx context.Context, report models.CommunityReport, reporterID string) error {
	_, err := r.db.Exec(ctx, `INSERT INTO community_reports(id,element_id,reporter_id,reason) VALUES($1,$2,$3,$4) ON CONFLICT(element_id,reporter_id) DO NOTHING`, report.ID, report.ElementID, reporterID, report.Reason)
	return err
}
