package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	urlpkg "net/url"
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
			uri, err := urlpkg.Parse(url)
			if err != nil {
				fmt.Printf("failed to parse url: %s, %v\n", url, err)
				return
			}

			req, err := http.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"https://sorcererxw.com/api/revalidate",
				nil,
			)
			if err != nil {
				fmt.Printf("failed to create request: %s\n", err)
				return
			}
			{
				q := uri.Query()
				q.Add("path", uri.Path)
				req.URL.RawQuery = q.Encode()
			}
			fmt.Println(req.URL.String())
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Printf("failed to do request: %s\n", err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("failed to revalidate: %s\n", resp.Status)
				return
			}
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
