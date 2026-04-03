package market

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Server struct {
	collector *Collector
	mux       *http.ServeMux
}

func NewServer(col *Collector) *Server {
	s := &Server{
		collector: col,
		mux:       http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /subscriptions", s.handleListSubscriptions)
	s.mux.HandleFunc("POST /subscriptions", s.handleAddSubscription)
	s.mux.HandleFunc("DELETE /subscriptions/{symbol}", s.handleRemoveSubscription)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// RegisterGin mounts the same REST routes as ServeHTTP on a Gin engine.
// Gin does not populate http.Request.PathValue; DELETE uses a cloned request with SetPathValue.
func (s *Server) RegisterGin(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		s.handleHealth(c.Writer, c.Request)
	})
	r.GET("/subscriptions", func(c *gin.Context) {
		s.handleListSubscriptions(c.Writer, c.Request)
	})
	r.POST("/subscriptions", func(c *gin.Context) {
		s.handleAddSubscription(c.Writer, c.Request)
	})
	r.DELETE("/subscriptions/:symbol", func(c *gin.Context) {
		req := c.Request.Clone(c.Request.Context())
		req.SetPathValue("symbol", c.Param("symbol"))
		s.handleRemoveSubscription(c.Writer, req)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleListSubscriptions(w http.ResponseWriter, r *http.Request) {
	subs := s.collector.ListSubscriptions()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"subscriptions": subs,
		"count":         len(subs),
	})
}

func (s *Server) handleAddSubscription(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Symbol string `json:"symbol"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Symbol == "" {
		http.Error(w, "symbol is required", http.StatusBadRequest)
		return
	}

	wsSymbol := s.collector.formatSymbol(req.Symbol)
	log.Info().Str("symbol", wsSymbol).Msg("Adding subscription via API")

	ok := s.collector.AddSubscription(wsSymbol)
	if !ok {
		http.Error(w, "symbol already subscribed", http.StatusConflict)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status": "subscribed",
		"symbol": wsSymbol,
	})
}

func (s *Server) handleRemoveSubscription(w http.ResponseWriter, r *http.Request) {
	symbol := r.PathValue("symbol")
	if symbol == "" {
		http.Error(w, "symbol is required", http.StatusBadRequest)
		return
	}

	wsSymbol := s.collector.formatSymbol(symbol)
	log.Info().Str("symbol", wsSymbol).Msg("Removing subscription via API")

	ok := s.collector.RemoveSubscription(wsSymbol)
	if !ok {
		http.Error(w, "symbol not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status": "unsubscribed",
		"symbol": wsSymbol,
	})
}
