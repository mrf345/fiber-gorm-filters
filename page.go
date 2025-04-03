package fgf

import (
	"math"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// scope that paginates the results by [PageSize] based on the [PageParam] passed in the request.
// also we can override [PageSize] with [PageSizeParam] if passed in the request,
// and it limits the maximum [PageSize] with the [MaxPageSize] value for sanity sake.
type PageScope struct {
	// fiber's request context
	Ctx *fiber.Ctx
	// the expected total number of results
	Total int64

	current  int
	previous int
	next     int
}

// default paginated response format
type PaginatedResponse[T any] struct {
	Total   int `json:"total"`
	Results T   `json:"results"`
	Page    int `json:"page"`
	Next    int `json:"next,omitempty"`
	Prev    int `json:"prev,omitempty"`
}

// returns the current page number
func (p *PageScope) Current() int {
	return p.current
}

// returns the previous page number
func (p *PageScope) Previous() int {
	return p.previous
}

// returns the next page number
func (p *PageScope) Next() int {
	return p.next
}

// generates the GORM scope for pagination
func (p *PageScope) Scope() GScope {
	p.current = p.Ctx.QueryInt(PageParam, 0)
	pageSize := p.Ctx.QueryInt(PageSizeParam, PageSize)
	maxPage := int(math.Ceil(float64(p.Total) / float64(pageSize)))

	if p.current <= 0 {
		p.current = 1
	} else if p.current > maxPage {
		p.current = maxPage
	}

	if maxPage > p.current {
		p.next = p.current + 1
	}

	if p.current > 1 {
		p.previous = p.current - 1
	}

	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	} else if pageSize <= 0 {
		pageSize = PageSize
	}

	return func(db *gorm.DB) *gorm.DB {
		return db.Offset((int(p.current) - 1) * pageSize).Limit(pageSize)
	}
}

// return populated response body, pulled into a separate method for ease of overriding
func (p *PageScope) RespBody(results any) any {
	return PaginatedResponse[any]{
		Results: results,
		Page:    p.Current(),
		Prev:    p.Previous(),
		Next:    p.Next(),
		Total:   int(p.Total),
	}
}

// sends a JSON paginated response (default format: [PaginatedResponse])
func (p *PageScope) Resp(results any) error {
	return p.Ctx.JSON(p.RespBody(results))
}
