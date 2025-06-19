package app

import (
	"github.com/Ian-zy0329/go-mall/config"
	"github.com/gin-gonic/gin"
	"strconv"
)

type Pagination struct {
	Page      int `json:"page"`
	PageSize  int `json:"page_Size"`
	TotalRows int `json:"total_rows"`
}

func NewPagination(c *gin.Context) *Pagination {
	page, _ := strconv.Atoi(c.Query("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if pageSize <= 0 {
		pageSize = config.App.Pagination.DefaultSize
	}
	if pageSize > config.App.Pagination.MaxSize {
		pageSize = config.App.Pagination.MaxSize
	}
	return &Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}

func (p *Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}

func (p *Pagination) GetPage() int {
	return p.Page
}

func (p *Pagination) GetPageSize() int {
	return p.PageSize
}

func (p *Pagination) SetTotalRows(total int) {
	p.TotalRows = total
}
