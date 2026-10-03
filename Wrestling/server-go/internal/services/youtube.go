package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type YouTubeVerifier struct {
	APIKey string
	Client *http.Client
}

func NewYouTubeVerifier(apiKey string) *YouTubeVerifier {
	return &YouTubeVerifier{APIKey: strings.TrimSpace(apiKey), Client: &http.Client{Timeout: 4 * time.Second}}
}
func (v *YouTubeVerifier) Verify(ctx context.Context, id string) (string, error) {
	if v.APIKey != "" {
		endpoint := "https://www.googleapis.com/youtube/v3/videos?part=snippet,contentDetails&id=" + url.QueryEscape(id) + "&key=" + url.QueryEscape(v.APIKey)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return "", err
		}
		resp, err := v.Client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(io.LimitReader(resp.Body, 256<<10))
		if err != nil {
			return "", err
		}
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("youtube api status %d", resp.StatusCode)
		}
		var data struct {
			Items []struct {
				Snippet struct {
					Title string `json:"title"`
				} `json:"snippet"`
				Content struct {
					Rating struct {
						YTRating string `json:"ytRating"`
					} `json:"contentRating"`
				} `json:"contentDetails"`
			} `json:"items"`
		}
		if err := json.Unmarshal(body, &data); err != nil {
			return "", err
		}
		if len(data.Items) == 0 {
			return "", errors.New("youtube video not found")
		}
		if data.Items[0].Content.Rating.YTRating == "ytAgeRestricted" {
			return "", errors.New("age restricted youtube video")
		}
		return data.Items[0].Snippet.Title, nil
	}
	raw := "https://www.youtube.com/watch?v=" + url.QueryEscape(id)
	endpoint := "https://www.youtube.com/oembed?url=" + url.QueryEscape(raw) + "&format=json"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Wrestling")
	resp, err := v.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("youtube oembed status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 128<<10))
	if err != nil {
		return "", err
	}
	var data struct {
		Title string `json:"title"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}
	if strings.TrimSpace(data.Title) == "" {
		return "", errors.New("missing youtube title")
	}
	return data.Title, nil
}
