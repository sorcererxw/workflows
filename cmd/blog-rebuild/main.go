package main

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	urlpkg "net/url"
	"sync"
	"time"
)

func get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	fmt.Printf("request %s\n", req.URL.String())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get: %s", resp.Status)
	}
	buf := make([]byte, 1024*1024)
	n, err := resp.Body.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	return buf[:n], nil
}

func run(ctx context.Context) error {
	var sitemap struct {
		URL []struct {
			Loc string `xml:"loc"`
		} `xml:"url"`
	}
	{
		body, err := get(ctx, "https://sorcererxw.com/sitemap.xml")
		if err != nil {
			return fmt.Errorf("failed to get sitemap: %w", err)
		}
		if err := xml.NewDecoder(bytes.NewReader(body)).Decode(&sitemap); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	var g sync.WaitGroup

	for _, u := range sitemap.URL {
		url := u.Loc
		g.Add(1)
		go func() {
			defer g.Done()
			uri, err := urlpkg.Parse(url)
			if err != nil {
				fmt.Printf("failed to parse url: %s, %v\n", url, err)
				return
			}

			if _, err := get(ctx, "https://sorcererxw.com/api/revalidate?"+urlpkg.Values{
				"path": []string{uri.Path},
			}.Encode()); err != nil {
				fmt.Printf("failed to revalidate: %s, %v\n", url, err)
				return
			}

			time.Sleep(time.Second)

			if _, err := get(ctx, url); err != nil {
				fmt.Printf("failed to get: %s, %v\n", url, err)
				return
			}
		}()
	}

	g.Wait()

	return nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := run(ctx); err != nil {
		panic(err)
	}
}
