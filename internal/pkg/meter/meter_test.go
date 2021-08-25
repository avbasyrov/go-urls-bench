package meter

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMeter_CheckThrottleWithFailedFirstRequest(t *testing.T) {
	results := make(chan Result, 10)

	m := New(5*time.Second, 1.3, 200*time.Millisecond)
	m.getter = func(url string, num int) bool {
		return false
	}

	var result Result
	go func() {
		for result = range results {
		}
	}()

	m.Check("fake_url", results)

	time.Sleep(10 * time.Millisecond) // чтобы успели дочитать с канала
	assert.Equal(t, "fake_url", result.Url)
	assert.Equal(t, 0, result.Concurrency)
	assert.Equal(t, time.Duration(0), result.Duration)
}

func TestMeter_CheckThrottleWithFailedConcurrentRequests(t *testing.T) {
	results := make(chan Result, 10)

	m := New(5*time.Second, 1.3, 200*time.Millisecond)
	m.getter = func(url string, num int) bool {
		if num > 1 {
			return false
		}

		return true
	}

	var result Result
	go func() {
		for result = range results {
		}
	}()

	m.Check("fake_url", results)

	time.Sleep(10 * time.Millisecond) // чтобы успели дочитать с канала
	assert.Equal(t, "fake_url", result.Url)
	assert.Equal(t, 1, result.Concurrency)
}

func TestMeter_CheckThrottleWithOverflow(t *testing.T) {
	results := make(chan Result, 10)

	m := New(5*time.Second, 1.3, 200*time.Millisecond)
	m.getter = func(url string, num int) bool {
		// имитируем длительность запроса в 100мс
		time.Sleep(100 * time.Millisecond)

		// а когда параллельных запросов более 9 - троттлим на дополнительную время
		if num > 9 {
			time.Sleep(300 * time.Millisecond)
		}

		return true
	}

	var result Result
	go func() {
		for result = range results {
		}
	}()

	m.Check("fake_url", results)

	time.Sleep(10 * time.Millisecond) // чтобы успели дочитать с канала
	assert.Equal(t, "fake_url", result.Url)
	assert.Equal(t, 9, result.Concurrency)
}

func TestMeter_CheckThrottleWithoutOverflow(t *testing.T) {
	results := make(chan Result, 10)

	m := New(5*time.Second, 1.3, 500*time.Millisecond)
	m.getter = func(url string, num int) bool {
		// имитируем длительность запроса в 50мс
		time.Sleep(50 * time.Millisecond)

		return true
	}

	var result Result
	go func() {
		for result = range results {
		}
	}()

	m.Check("fake_url", results)

	time.Sleep(10 * time.Millisecond) // чтобы успели дочитать с канала
	assert.Equal(t, "fake_url", result.Url)
	assert.Equal(t, maxConcurrency, result.Concurrency)
}
