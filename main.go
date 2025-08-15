package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

const metricTemplate = `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`

type apiConfig struct {
	fileserverHits atomic.Int32
}

type ReturnMsg struct {
	Valid bool `json:"valid"`
}

type BodyMsg struct {
	Body string `json:"body"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	body := BodyMsg{}
	err := decoder.Decode(&body)
	if err != nil {
		respondWithError(w, 500, "Something went wrong", err)
		return
	}

	if len(body.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long", nil)
		return
	}

	respondWithJSON(w, 200, ReturnMsg{Valid: true})
}

func (cfg *apiConfig) getHits(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Header().Add("Content-Type", "text/html")
	w.Write([]byte(fmt.Sprintf(metricTemplate, cfg.fileserverHits.Load())))
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
	mux.Handle("GET /api/healthz", conf.middlewareMetricsInc(http.HandlerFunc(myHandler)))
	mux.Handle("POST /api/validate_chirp", http.HandlerFunc(validateChirp))
	mux.Handle("GET /admin/metrics", http.HandlerFunc(conf.getHits))
	mux.Handle("POST /admin/reset", http.HandlerFunc(conf.resetHits))

	server := http.Server{Addr: ":" + port, Handler: mux}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
