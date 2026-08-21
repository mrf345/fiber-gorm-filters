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
	// scope specific number of items to return per page (overrides [PageSize])
	PageSize int
	// scope specific maximum number of items that can be returned per page (overrides [MaxPageSize])
	MaxPageSize int
	// optimization flag if enabled will set page numbers post scope execution
	SingleQuery bool

	current, previous, next, pageSize int
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

// sets the current, previous and next page numbers
func (p *PageScope) SetPages() {
	max := p.maxPage()

	if max == 0 || p.current <= 0 {
		p.current = 1
	} else if p.current > max {
		p.current = max
	}

	if p.current > 1 {
		p.previous = p.current - 1
	}

	if max > p.current {
		p.next = p.current + 1
	}
}

func (p *PageScope) maxPage() int {
	return int(math.Ceil(float64(p.Total) / float64(p.pageSize)))
}

// generates the GORM scope for pagination
func (p *PageScope) Scope() GScope {
	if p.Ctx == nil {
		panic("PageScope.Ctx is not set")
	}

	p.current = p.Ctx.QueryInt(PageParam, 0)
	p.pageSize = p.Ctx.QueryInt(PageSizeParam, p.DefaultPageSize())

	if !p.SingleQuery {
		p.SetPages()
	}

	if p.pageSize > p.DefaultMaxPageSize() {
		p.pageSize = p.DefaultMaxPageSize()
	} else if p.pageSize <= 0 {
		p.pageSize = PageSize
	}

	return func(db *gorm.DB) *gorm.DB {
		return db.Offset((int(p.current) - 1) * p.pageSize).Limit(p.pageSize)
	}
}

// returns default page size to fallback to
func (p *PageScope) DefaultPageSize() int {
	if p.PageSize != 0 {
		return p.PageSize
	}

	return PageSize
}

// returns default maximum page size to fallback to
func (p *PageScope) DefaultMaxPageSize() int {
	if p.MaxPageSize != 0 {
		return p.MaxPageSize
	}

	return MaxPageSize
}

// returns populated response body, pulled into a separate method for ease of overriding
func (p *PageScope) RespBody(results any) any {
	if p.SingleQuery {
		p.SetPages()
	}

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
