package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) getHits(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(fmt.Sprintf("Hits: %v", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) resetHits(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(200)
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("OK"))
}

func myHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("OK"))
}

func main() {
	const port = "8080"

	conf := apiConfig{}

	mux := http.NewServeMux()

	mux.Handle("/app/", conf.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("./www/")))))
	mux.Handle("GET /healthz", conf.middlewareMetricsInc(http.HandlerFunc(myHandler)))
	mux.Handle("GET /metrics", http.HandlerFunc(conf.getHits))
	mux.Handle("POST /reset", http.HandlerFunc(conf.resetHits))

	server := http.Server{Addr: ":" + port, Handler: mux}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
