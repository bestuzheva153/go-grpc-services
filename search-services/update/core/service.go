package core

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

type Service struct {
	log         *slog.Logger
	db          DB
	xkcd        XKCD
	words       Words
	concurrency int
	mu          sync.Mutex
	status      ServiceStatus
}

func NewService(
	log *slog.Logger, db DB, xkcd XKCD, words Words, concurrency int,
) (*Service, error) {
	if concurrency < 1 {
		return nil, fmt.Errorf("wrong concurrency specified: %d", concurrency)
	}
	return &Service{
		log:         log,
		db:          db,
		xkcd:        xkcd,
		words:       words,
		concurrency: concurrency,
		status:      StatusIdle,
	}, nil
}

func (s *Service) Update(ctx context.Context) error {
	// статус
	s.mu.Lock()
	if s.status == StatusRunning {
		s.mu.Unlock()
		return ErrAlreadyExists
	}
	s.status = StatusRunning
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.status = StatusIdle
		s.mu.Unlock()
	}()
	lastID, err := s.xkcd.LastID(ctx)
	if err != nil {
		return err
	}
	existing, err := s.db.IDs(ctx)
	if err != nil {
		return err
	}
	exist := make(map[int]struct{}, len(existing))
	for _, id := range existing {
		exist[id] = struct{}{}
	}
	jobs := make(chan int, s.concurrency)
	var wg sync.WaitGroup
	for i := 0; i < s.concurrency; i++ {
		wg.Add(1)
		go func(ctx context.Context) {
			defer wg.Done()
			for id := range jobs {
				comic, err := s.xkcd.Get(ctx, id)
				if err != nil {
					continue
				}
				text := comic.Title + " " + comic.Description
				words, err := s.words.Norm(ctx, text)
				if err != nil {
					continue
				}
				_ = s.db.Add(ctx, Comics{
					ID:    comic.ID,
					URL:   comic.URL,
					Words: words,
				})
			}
		}(ctx)
	}
	for id := 1; id <= lastID; id++ {
		if _, ok := exist[id]; ok {
			continue
		}
		jobs <- id
	}
	close(jobs)
	wg.Wait()
	return nil
}

func (s *Service) Stats(ctx context.Context) (ServiceStats, error) {
	dbStats, err := s.db.Stats(ctx)
	if err != nil {
		return ServiceStats{}, err
	}

	total, err := s.xkcd.LastID(ctx)
	if err != nil {
		return ServiceStats{}, err
	}

	return ServiceStats{
		DBStats:     dbStats,
		ComicsTotal: total,
	}, nil
}

func (s *Service) Status(ctx context.Context) ServiceStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status
}

func (s *Service) Drop(ctx context.Context) error {
	return s.db.Drop(ctx)
}
