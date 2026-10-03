package services

import (
	"context"
	"sort"
	"strings"
	"testing"
	"wrestling/internal/models"
	"wrestling/internal/ports"
)

type fakeElementRepository struct{ items []models.Element }

func (f fakeElementRepository) All(ctx context.Context, filter ports.ElementQuery) ([]models.Element, error) {
	result := make([]models.Element, 0, len(f.items))
	for _, item := range f.items {
		if filter.Category != "" && filter.Category != "все" && item.Category != filter.Category {
			continue
		}
		if filter.Subcategory != "" && item.Subcategory != filter.Subcategory {
			continue
		}
		if filter.Difficulty != "" && item.Difficulty != filter.Difficulty {
			continue
		}
		if filter.Search != "" {
			q := strings.ToLower(filter.Search)
			matched := strings.Contains(strings.ToLower(item.Name), q) || strings.Contains(strings.ToLower(item.Description), q)
			for _, m := range item.Muscles {
				matched = matched || strings.Contains(strings.ToLower(m), q)
			}
			if !matched {
				continue
			}
		}
		result = append(result, item)
	}
	return result, nil
}
func (f fakeElementRepository) ByID(ctx context.Context, id int) (models.Element, bool, error) {
	for _, item := range f.items {
		if item.ID == id {
			return item, true, nil
		}
	}
	return models.Element{}, false, nil
}
func (f fakeElementRepository) ByCategory(ctx context.Context, category string) ([]models.Element, error) {
	result := []models.Element{}
	for _, item := range f.items {
		if item.Category == category {
			result = append(result, item)
		}
	}
	return result, nil
}

func (f fakeElementRepository) Stats(ctx context.Context) (models.Stats, error) {
	stats := models.Stats{ByCategory: map[string]int{}, ByDifficulty: map[string]int{}}
	for _, item := range f.items {
		stats.Total++
		stats.ByCategory[item.Category]++
		stats.ByDifficulty[item.Difficulty]++
	}
	for k := range stats.ByCategory {
		stats.Categories = append(stats.Categories, k)
	}
	for k := range stats.ByDifficulty {
		stats.Difficulties = append(stats.Difficulties, k)
	}
	sort.Strings(stats.Categories)
	sort.Strings(stats.Difficulties)
	return stats, nil
}

func testElements() []models.Element {
	return []models.Element{
		{ID: 1, Name: "Pull Up", Category: "стойка", Subcategory: "подтягивания", Difficulty: "medium", Description: "Силовое упражнение", Muscles: []string{"Спина"}},
		{ID: 2, Name: "Handstand", Category: "стойка", Subcategory: "стойки", Difficulty: "hard", Description: "Стойка", Muscles: []string{"Плечи"}},
		{ID: 3, Name: "360", Category: "партер", Subcategory: "градусы", Difficulty: "easy", Description: "Поворот", Muscles: []string{"Ноги"}},
	}
}

func TestElementServiceList(t *testing.T) {
	service := NewElementService(fakeElementRepository{items: testElements()})
	cases := []struct {
		name   string
		filter ElementFilter
		want   int
	}{
		{"all", ElementFilter{}, 3},
		{"category", ElementFilter{Category: "стойка"}, 2},
		{"all category", ElementFilter{Category: "все"}, 3},
		{"subcategory", ElementFilter{Subcategory: "стойки"}, 1},
		{"difficulty", ElementFilter{Difficulty: "hard"}, 1},
		{"search name", ElementFilter{Search: "pull"}, 1},
		{"search description", ElementFilter{Search: "стойка"}, 1},
		{"search muscle", ElementFilter{Search: "плечи"}, 1},
		{"combined", ElementFilter{Category: "стойка", Difficulty: "hard", Search: "плечи"}, 1},
		{"no match", ElementFilter{Search: "unknown"}, 0},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			items, err := service.List(context.Background(), tt.filter)
			if err != nil {
				t.Fatal(err)
			}
			if got := len(items); got != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestElementServiceGetAndCategory(t *testing.T) {
	service := NewElementService(fakeElementRepository{items: testElements()})
	if item, ok, err := service.Get(context.Background(), 2); err != nil || !ok || item.Name != "Handstand" {
		t.Fatalf("unexpected item: %+v %v", item, ok)
	}
	if _, ok, err := service.Get(context.Background(), 999); err != nil || ok {
		t.Fatal("expected missing item")
	}
	if got, err := service.ByCategory(context.Background(), "партер"); err != nil || len(got) != 1 || got[0].ID != 3 {
		t.Fatalf("unexpected category result: %+v", got)
	}
}

func TestElementServiceStatsAreDeterministic(t *testing.T) {
	service := NewElementService(fakeElementRepository{items: testElements()})
	stats, err := service.Stats(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if stats.Total != 3 {
		t.Fatalf("expected total 3, got %d", stats.Total)
	}
	if stats.ByCategory["стойка"] != 2 || stats.ByDifficulty["hard"] != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	if len(stats.Categories) != 2 || stats.Categories[0] != "партер" || stats.Categories[1] != "стойка" {
		t.Fatalf("categories not sorted: %v", stats.Categories)
	}
	if len(stats.Difficulties) != 3 || stats.Difficulties[0] != "easy" {
		t.Fatalf("difficulties not sorted: %v", stats.Difficulties)
	}
}

func BenchmarkElementServiceList(b *testing.B) {
	service := NewElementService(fakeElementRepository{items: testElements()})
	filter := ElementFilter{Category: "стойка", Search: "плечи"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.List(context.Background(), filter)
	}
}
