package feedfetch

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		DialContext:         (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
		IdleConnTimeout:     90 * time.Second,
	},
}

// The caller controls the request lifetime via ctx
// (used for per-fetch timeouts by the scheduler).
func Fetch(ctx context.Context, feedURL string) (*gofeed.Feed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "murmur/0.1 (+https://github.com/nielwyn/murmur)")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d fetching %s", resp.StatusCode, feedURL)
	}

	feed, err := gofeed.NewParser().Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing feed %s: %w", feedURL, err)
	}
	return feed, nil
}
