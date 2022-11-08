package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

var tasks = map[string]string{
	"solidot":         "https://jike-rss.vercel.app/job/solidot",
	"meme":            "https://jike-rss.vercel.app/job/meme",
	"programmerhumor": "https://jike-rss.vercel.app/job/programmer_humor",
	"yearprogress":    "https://jike-rss.vercel.app/job/year_progress",
}

func main() {
	var g errgroup.Group
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	for name, url := range tasks {
		name, url := name, url
		g.Go(func() error {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				return errors.Wrapf(err, "failed to create request for %s", name)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return errors.Wrapf(err, "failed to do request for %s", name)
			}
			defer resp.Body.Close()
			if resp.StatusCode >= 400 {
				return errors.Errorf("failed to get %s: %s", name, resp.Status)
			}
			log.Printf("success for %s", name)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		panic(err)
	}
}
