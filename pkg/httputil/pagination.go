package httputil

import (
	"net/http"
	"strconv"

	"gorm.io/gorm"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

func ParsePagination(r *http.Request) (page, pageSize, offset int) {
	page = DefaultPage
	pageSize = DefaultPageSize

	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			pageSize = v
			if pageSize > MaxPageSize {
				pageSize = MaxPageSize
			}
		}
	}

	offset = (page - 1) * pageSize
	return page, pageSize, offset
}

// PaginateQuery applies offset/limit and returns total count metadata.
func PaginateQuery(query *gorm.DB, r *http.Request) (*gorm.DB, PaginatedMeta, error) {
	page, pageSize, offset := ParsePagination(r)

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, PaginatedMeta{}, err
	}

	paginated := query.Offset(offset).Limit(pageSize)
	return paginated, BuildMeta(page, pageSize, total), nil
}

func BuildMeta(page, pageSize int, total int64) PaginatedMeta {
	totalPages := 0
	if pageSize > 0 {
		totalPages = int(total) / pageSize
		if int(total)%pageSize > 0 {
			totalPages++
		}
	}
	return PaginatedMeta{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}
