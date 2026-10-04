package book

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// handler 只做"翻译"：把 HTTP 输入交给 service，把 service 的错误翻译成状态码。
// 它本身不埋点——server span 由 otelgin 中间件统一负责。
type handler struct {
	svc *Service
}

// RegisterRoutes 注册路由。
func RegisterRoutes(r gin.IRouter, svc *Service) {
	h := &handler{svc: svc}

	r.GET("/healthz", h.health)

	api := r.Group("/api/v1")
	api.GET("/books", h.list)
	api.GET("/books/:id", h.get)
	api.POST("/books", h.create)
}

func (h *handler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *handler) list(c *gin.Context) {
	books, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": books})
}

func (h *handler) get(c *gin.Context) {
	// 关键：把 gin.Context 里的 request context 往下传，
	// 它携带着当前 span —— 断在这里，后面所有子 span 都会变成孤儿。
	book, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	switch {
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	default:
		c.JSON(http.StatusOK, book)
	}
}

type createBookRequest struct {
	Title  string `json:"title" binding:"required"`
	Author string `json:"author" binding:"required"`
}

func (h *handler) create(c *gin.Context) {
	var req createBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	book, err := h.svc.Create(c.Request.Context(), req.Title, req.Author)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, book)
}
