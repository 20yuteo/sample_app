package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"commerce-lab/backend/internal/db"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Server struct {
	store *db.Store
}

func New(store *db.Store, allowedOrigins []string) http.Handler {
	server := &Server{store: store}
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Timeout(10 * time.Second))
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
		MaxAge:         300,
	}))

	router.Get("/healthz", server.healthz)
	router.Get("/api/categories", server.listCategories)
	router.Get("/api/products", server.listProducts)

	return router
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) listCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := s.store.ListCategories(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, categories)
}

func (s *Server) listProducts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	products, err := s.store.ListProducts(r.Context(), db.ProductFilters{
		Query:      query.Get("q"),
		CategoryID: int64Param(query.Get("categoryId")),
		MinPrice:   int32Param(query.Get("minPrice")),
		MaxPrice:   int32Param(query.Get("maxPrice")),
		Limit:      int32Param(query.Get("limit")),
		Offset:     int32Param(query.Get("offset")),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, products)
}

func int64Param(value string) int64 {
	parsed, _ := strconv.ParseInt(value, 10, 64)
	return parsed
}

func int32Param(value string) int32 {
	parsed, _ := strconv.ParseInt(value, 10, 32)
	return int32(parsed)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
