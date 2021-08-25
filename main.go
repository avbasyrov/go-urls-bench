package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"ttt/internal/pkg/config"
	"ttt/internal/pkg/multimeter"
)

var m *multimeter.Multimeter
var cfg *config.Params

func main() {
	var err error

	cfg, err = config.Parse()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(cfg)
	log.Println(cfg.MaxResponseTime - time.Second)

	m = multimeter.New(cfg.RequestTimeout, cfg.ThrottleMultiplier, cfg.ThrottleMinimalMargin, cfg.QueryCacheTtl)

	http.HandleFunc("/sites", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["search"]
	if !ok || len(keys[0]) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Url Param 'search' is missing"))
		return
	}
	search := keys[0]

	timeout := time.After(cfg.MaxResponseTime - 500*time.Millisecond)

	concurrences, initialized, err := m.GetConcurrency(search)
	if err != nil {
		log.Println("Cache hit")
		showError(w, err)
		return
	}

	log.Println("Cache miss")

	if initialized {
		showResult(w, concurrences)
		return
	}

	errChannel := make(chan error, 1)
	go func() {
		errChannel <- m.Query(search)
	}()

	select {
	case err := <-errChannel:
		log.Println("<-errChannel")
		if err != nil {
			showError(w, err)
			return
		}
	case <-timeout:
		log.Println("timeout")
	}

	concurrences, _, err = m.GetConcurrency(search)
	if err != nil {
		showError(w, err)
		return
	}

	showResult(w, concurrences)
}

func showError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(err.Error()))
}

func showResult(w http.ResponseWriter, concurrences map[string]int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	asJson, err := json.Marshal(concurrences)
	if err != nil {
		log.Fatal(err)
	}
	_, _ = w.Write(asJson)
}
