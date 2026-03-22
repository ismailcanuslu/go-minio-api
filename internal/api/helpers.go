package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

func respondJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("json encode failed: %v", err)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func objectKey(path, routePrefix string) string {
	return strings.TrimPrefix(path, routePrefix)
}

func extFromKey(key string) string {
	idx := strings.LastIndex(key, ".")
	if idx == -1 {
		return ""
	}
	return key[idx:]
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(started).String())
	})
}
