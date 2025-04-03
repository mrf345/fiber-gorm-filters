package fgf

import (
	"slices"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// scope that sorts the results by [SortScope.Default] and overrides it with the [SortParam] if it is present in the request
type SortScope struct {
	// fiber's request context
	Ctx *fiber.Ctx
	// fields to allow sorting by (i.e. id, updated_at)
	Fields []string
	// default fields to sort by if [SortParam] is not present in the request (i.e. id, -updated_at)
	Default []string
}

// generates the GORM scope for sorting
func (s SortScope) Scope() GScope {
	var fields = s.Default
	var params = s.Ctx.Query(SortParam, "")

	if params != "" {
		fields = strings.Split(params, ",")
	}

	return func(db *gorm.DB) *gorm.DB {
		query := db

		if len(fields) == 0 {
			return query
		}

		for _, field := range fields {
			var desc bool

			if strings.HasPrefix(field, "-") {
				desc = true
				field = field[1:]
			}

			if !slices.Contains(s.Fields, field) && !slices.Contains(s.Default, field) {
				continue
			}

			// log.Fatalln(fields)

			query = query.Order(clause.OrderByColumn{
				Desc:   desc,
				Column: clause.Column{Name: field},
			})
		}

		return query
	}
}
