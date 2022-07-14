package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"sync"
	"time"
)

func run(ctx context.Context) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://sorcererxw.com/sitemap.xml",
		nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()
	var sitemap struct {
		URL []struct {
			Loc string `xml:"loc"`
		} `xml:"url"`
	}
	if err := xml.NewDecoder(resp.Body).Decode(&sitemap); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	var g sync.WaitGroup

	for _, u := range sitemap.URL {
		url := u.Loc
		g.Add(1)
		go func() {
			defer g.Done()
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				fmt.Printf("failed to create request: %s\n", err)
				return
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Printf("failed to do request: %s\n", err)
				return
			}
			defer resp.Body.Close()
		}()
	}

	g.Wait()

	return nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := run(ctx); err != nil {
		panic(err)
	}
}
