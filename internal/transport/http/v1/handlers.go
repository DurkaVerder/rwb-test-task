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
	n := ctx.Query("n")
	if n == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'n' is required"})
		return
	}

	nInt, err := strconv.Atoi(n)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'n' must be a valid integer"})
		return
	}

	requests, err := h.Service.TopNRequests(nInt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, TopNResponse{Requests: requests})
}
