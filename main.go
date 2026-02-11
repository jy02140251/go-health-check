package main

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"
)

type Config struct {
    Timeout     time.Duration
    MaxRetries  int
    Concurrency int
}

type Service struct {
    config Config
    mu     sync.RWMutex
    cache  map[string]interface{}
}

func NewService(cfg Config) *Service {
    return &Service{
        config: cfg,
        cache:  make(map[string]interface{}),
    }
}

func (s *Service) Get(key string) (interface{}, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    val, ok := s.cache[key]
    return val, ok
}

func (s *Service) Set(key string, value interface{}) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.cache[key] = value
}

func (s *Service) Process(ctx context.Context, data []string) ([]string, error) {
    results := make([]string, 0, len(data))
    sem := make(chan struct{}, s.config.Concurrency)
    
    var wg sync.WaitGroup
    var mu sync.Mutex
    
    for _, item := range data {
        wg.Add(1)
        go func(d string) {
            defer wg.Done()
            sem <- struct{}{}
            defer func() { <-sem }()
            
            result := fmt.Sprintf("processed: %s", d)
            mu.Lock()
            results = append(results, result)
            mu.Unlock()
        }(item)
    }
    
    wg.Wait()
    return results, nil
}

func main() {
    cfg := Config{
        Timeout:     30 * time.Second,
        MaxRetries:  3,
        Concurrency: 10,
    }
    
    svc := NewService(cfg)
    log.Printf("Service started with config: %+v", cfg)
    
    ctx := context.Background()
    data := []string{"a", "b", "c"}
    results, _ := svc.Process(ctx, data)
    log.Printf("Results: %v", results)
}