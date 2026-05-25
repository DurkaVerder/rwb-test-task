package v1

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/DurkaVerder/rwb-test-task/internal/metrics"
	"github.com/gin-gonic/gin"
)

type Service interface {
	TopNQueries(ctx context.Context, n int) ([]string, error)
}

type Handlers struct {
	Service Service
}

type TopNResponse struct {
	Requests []string `json:"requests"`
}

func NewHandlers(s Service) *Handlers {
	return &Handlers{
		Service: s,
	}
}

func (h *Handlers) GetTopNQueries(ctx *gin.Context) {
	start := time.Now()
	status := http.StatusOK
	defer func() {
		metrics.ObserveTopRequests(time.Since(start), status)
	}()

	nRaw := ctx.Query("n")
	if nRaw == "" {
		status = http.StatusBadRequest
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "missing n query parameter"})
		return
	}

	n, err := strconv.Atoi(nRaw)
	if err != nil || n <= 0 {
		status = http.StatusBadRequest
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "n must be a positive integer"})
		return
	}

	queries, err := h.Service.TopNQueries(ctx.Request.Context(), n)
	if err != nil {
		status = http.StatusInternalServerError
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, TopNResponse{Requests: queries})
}
