package api

import (
	_ "embed"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"kraken-trader/internal/decision"
	"kraken-trader/internal/market"
)

//go:embed openapi.json
var openAPISpec []byte

// Server implements HTTP handlers for the trader API (market + decisions).
type Server struct {
	collector *market.Collector
	engine    *decision.Engine
}

// NewServer builds an API server backed by market data and the decision engine.
func NewServer(collector *market.Collector, engine *decision.Engine) *Server {
	return &Server{collector: collector, engine: engine}
}

// GinServerOptions mirrors oapi-codegen options; reserved for future use (e.g. base path).
type GinServerOptions struct{}

// RegisterHandlersWithOptions registers all routes on the Gin engine.
func RegisterHandlersWithOptions(router *gin.Engine, srv *Server, _ GinServerOptions) {
	market.NewServer(srv.collector).RegisterGin(router)
	router.POST("/v1/decide", srv.handleDecide)
}

// GetSwagger returns the embedded OpenAPI document as generic JSON for gin.JSON.
func GetSwagger() (map[string]interface{}, error) {
	var doc map[string]interface{}
	if err := json.Unmarshal(openAPISpec, &doc); err != nil {
		return nil, err
	}
	return doc, nil
}

type decideRequest struct {
	Pairs []string `json:"pairs"`
}

func (s *Server) handleDecide(c *gin.Context) {
	var body decideRequest
	if err := c.ShouldBindJSON(&body); err != nil || len(body.Pairs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pairs is required and must be non-empty"})
		return
	}

	decisions, err := s.engine.Decide(c.Request.Context(), body.Pairs)
	if err != nil {
		log.Error().Err(err).Msg("API /v1/decide failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"decisions": decisions})
}
