package multimeter

import (
	"sync"
	"time"
	"ttt/internal/pkg/meter"
	"ttt/internal/pkg/parser"
)

type Multimeter struct {
	requestTimeout        time.Duration
	throttleMultiplier    float64
	throttleMinimalMargin time.Duration
	cacheTtl              time.Duration

	results chan meter.Result
	cache   map[string]meter.Result
	parser  *parser.Parser

	mu *sync.RWMutex
}

// New создает новую структуру для поиска числа максимальных запросов по поисковой строке
// 	requestTimeout - предельное время на запрос к url
//	throttleMultiplier - коэффициент для определения, начался ли троттлинг (но при условии, что задержка выросла болле чем на throttleMinimalMargin)
func New(requestTimeout time.Duration, throttleMultiplier float64, throttleMinimalMargin time.Duration, cacheTtl time.Duration) *Multimeter {
	m := &Multimeter{
		requestTimeout:        requestTimeout,
		throttleMultiplier:    throttleMultiplier,
		throttleMinimalMargin: throttleMinimalMargin,
		cacheTtl:              cacheTtl,

		results: make(chan meter.Result, 1000),
		cache:   make(map[string]meter.Result),
		parser:  parser.New(cacheTtl),

		mu: &sync.RWMutex{},
	}

	go m.run()

	return m
}

func (m *Multimeter) Query(search string) error {
	urls, err := m.parser.Search(search)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(len(urls))
	for _, url := range urls {
		go func(url string) {
			defer wg.Done()
			checker := meter.New(m.requestTimeout, m.throttleMultiplier, m.throttleMinimalMargin)
			checker.Check(url, m.results)
		}(url)
	}

	wg.Wait()

	return nil
}

func (m *Multimeter) GetConcurrency(search string) (map[string]int, bool, error) {
	urls, err := m.parser.Search(search)
	if err != nil {
		return nil, false, err
	}

	concurrences := make(map[string]int, len(urls))
	allInitialized := true // по всем ли урлам просчитана конкурентность запросов

	m.mu.RLock()
	for _, url := range urls {
		if item, ok := m.cache[url]; ok {
			concurrences[url] = item.Concurrency
		} else {
			allInitialized = false
			concurrences[url] = 0
		}
	}
	m.mu.RUnlock()

	return concurrences, allInitialized, nil
}

func (m *Multimeter) run() {
	for result := range m.results {
		m.mu.Lock()
		m.cache[result.Url] = result
		m.mu.Unlock()
	}
}
