package meter

import (
	"errors"
	"log"
	"net/http"
	"sync"
	"time"
)

// Ограничение на потенциально возможное число параллельных запросов к Url
const maxConcurrency = 512

type Meter struct {
	requestTimeout        time.Duration
	throttleMultiplier    float64
	throttleMinimalMargin time.Duration

	getter func(url string, num int) bool
}

type Result struct {
	Url         string
	Concurrency int
	Duration    time.Duration
}

// New создает новую структуру для поиска числа максимальных запросов
// 	requestTimeout - предельное время на запрос к Url
//	throttleMultiplier - коэффициент для определения, начался ли троттлинг (но при условии, что задержка выросла болле чем на throttleMinimalMargin)
func New(requestTimeout time.Duration, throttleMultiplier float64, throttleMinimalMargin time.Duration) *Meter {
	m := &Meter{
		requestTimeout:        requestTimeout,
		throttleMultiplier:    throttleMultiplier,
		throttleMinimalMargin: throttleMinimalMargin,
	}
	m.getter = m.get
	return m
}

func (m *Meter) Check(url string, results chan Result) {
	lastSuccessConcurrency := 1 // сколько можно делать безопасных запросов параллельно

	// Засекаем длительность при запросе в один поток, относительно нее считаем не стали ли нас троттлить
	started := time.Now()
	if !m.getter(url, 1) {
		results <- Result{
			Url:         url,
			Concurrency: 0,
			Duration:    0,
		}
		return
	}
	referenceDuration := time.Now().Sub(started)
	throttledDuration1 := time.Duration(float64(referenceDuration) * m.throttleMultiplier)
	throttledDuration2 := referenceDuration + m.throttleMinimalMargin

	// Выберем максимальый порог из throttledDuration1 и throttledDuration2
	throttledDuration := throttledDuration1
	if throttledDuration2 > throttledDuration1 {
		throttledDuration = throttledDuration2
	}

	results <- Result{
		Url:         url,
		Concurrency: lastSuccessConcurrency,
		Duration:    referenceDuration,
	}

	var concurrency int
	var overflowConcurrency int
	for concurrency = 2; concurrency <= maxConcurrency; concurrency *= 2 {
		duration, err := m.getInParallel(url, concurrency)

		if err != nil || duration > throttledDuration {
			// Сайт не выдержал Concurrency запросов
			overflowConcurrency = concurrency
			break
		}

		results <- Result{
			Url:         url,
			Concurrency: concurrency,
			Duration:    duration,
		}
		lastSuccessConcurrency = concurrency
	}

	// Не было троттлинга или ошибок?
	if overflowConcurrency == 0 {
		return
	}

	// Допустимое число потоков где-то в интервале между lastSuccessConcurrency и overflowConcurrency
	checkFrom := lastSuccessConcurrency + 1
	checkUntil := overflowConcurrency - 1

	// Не осталось вариантов
	if checkFrom > checkUntil {
		return
	}

	// Пройдемся в 10 шагов в поисках более точного предела числа параллельных запросов
	step := (checkUntil - checkFrom) / 10
	if step < 1 {
		step = 1
	}

	for concurrency = checkFrom; concurrency <= checkUntil; concurrency += step {
		duration, err := m.getInParallel(url, concurrency)

		if err != nil || duration > throttledDuration {
			// Сайт не выдержал Concurrency запросов
			break
		}

		results <- Result{
			Url:         url,
			Concurrency: concurrency,
			Duration:    duration,
		}
	}
}

// Допустимое число потоков где-то между lastSuccessConcurrency и Concurrency,
// для ускорения поиска будем искать бисекцией
//if lastSuccessConcurrency < Concurrency {
//	bSearch := binarysearch.New(lastSuccessConcurrency, Concurrency)
//
//	for {
//		Concurrency := bSearch.GetMidPoint()
//		Duration, err := m.getInParallel(Url, Concurrency)
//
//		if err != nil || Duration > throttledDuration {
//			if !bSearch.Down() {
//				break
//			}
//			continue
//		}
//
//		lastSuccessConcurrency = Concurrency
//
//		if !bSearch.Up() {
//			break
//		}
//	}
//}

// getInParallel возвращает наибольшую длительность из num HTTP-запросов к Url,
// либо ошибку, если хотя бы один запрос был не успешным (не 200 ОК)
func (m *Meter) getInParallel(url string, num int) (time.Duration, error) {
	failed := false

	var wg sync.WaitGroup

	started := time.Now()

	for i := 0; i < num; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if !m.getter(url, num) {
				failed = true
			}
		}()
	}

	wg.Wait()

	if failed {
		return 0, errors.New("bad response")
	}

	duration := time.Now().Sub(started)

	return duration, nil
}

func (m *Meter) get(url string, _ int) bool {
	client := http.Client{
		Timeout: m.requestTimeout,
		Transport: &http.Transport{
			DisableKeepAlives: true, // чтобы не оставались открытые коннекшены от прошлых запросов
		},
	}

	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			log.Println(url, err, resp.Status)
		} else {
			log.Println(url, err)
		}
		return false
	}

	return true
}
