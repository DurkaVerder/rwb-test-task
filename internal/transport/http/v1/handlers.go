package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Service interface {
	TopNRequests(n int) ([]string, error)
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

func (h *Handlers) GetTopNRequests(ctx *gin.Context) {
	nRaw := ctx.Query("n")
	if nRaw == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "missing n query parameter"})
		return
	}

	n, err := strconv.Atoi(nRaw)
	if err != nil || n <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "n must be a positive integer"})
		return
	}

	requests, err := h.Service.TopNRequests(n)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, TopNResponse{Requests: requests})
}
