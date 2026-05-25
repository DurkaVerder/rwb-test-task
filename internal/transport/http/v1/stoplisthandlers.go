package v1

import (
	"context"
	"errors"
	"net/http"

	customErr "github.com/DurkaVerder/rwb-test-task/internal/errors"
	"github.com/gin-gonic/gin"
)

type StopListService interface {
	GetAllStopWords(ctx context.Context) ([]string, error)
	AddStopWord(ctx context.Context, word string) error
	RemoveStopWord(ctx context.Context, word string) error
}

type StopListHandlers struct {
	stopListService StopListService
}

func NewStopListHandlers(s StopListService) *StopListHandlers {
	return &StopListHandlers{
		stopListService: s,
	}
}

type StopListResponse struct {
	StopWords []string `json:"stop_words"`
}

type StopWordRequest struct {
	Word string `json:"word" binding:"required"`
}

func (h *StopListHandlers) GetStopList(ctx *gin.Context) {
	stopWords, err := h.stopListService.GetAllStopWords(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, StopListResponse{StopWords: stopWords})
}

func (h *StopListHandlers) AddStopWord(ctx *gin.Context) {
	var req StopWordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.stopListService.AddStopWord(ctx.Request.Context(), req.Word); err != nil {
		if errors.Is(err, customErr.InvalidWordError) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusOK)
}

func (h *StopListHandlers) RemoveStopWord(ctx *gin.Context) {
	var req StopWordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.stopListService.RemoveStopWord(ctx.Request.Context(), req.Word); err != nil {
		if errors.Is(err, customErr.InvalidWordError) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, customErr.WordNotFoundError) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusOK)
}
