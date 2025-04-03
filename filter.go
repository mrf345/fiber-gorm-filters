package fgf

import (
	"fmt"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// fixed set of supported filters
type Filter string
type filterQueryMap map[Filter]func(string, string) (string, any)

const (
	// if field value contains (i.e. ?name__contains=John)
	Contains Filter = "contains"
	// if field value equals (i.e. ?name=John or ?name__eq=John)
	Equals Filter = "eq"
	// if field value equals (i.e. ?name__neq=John)
	NotEquals Filter = "neq"
	// if field value greater than (i.e. ?age__gt=18)
	Greater Filter = "gt"
	// if field value greater than or equal (i.e. ?age__gte=18)
	GreaterEquals Filter = "gte"
	// if field value less than (i.e. ?age__lt=18)
	Lesser Filter = "lt"
	// if field value is less than or equal to (i.e. ?age__lte=18)
	LesserEquals Filter = "lte"
	// if field value starts with (i.e. ?name__startswith=John)
	StartsWith Filter = "startswith"
	// if field value ends with (i.e. ?name__endswith=John)
	EndsWith Filter = "endswith"
	// if field value in comma separated list of values  (i.e. ?name__in=John,Oliver)
	In Filter = "in"
	// if field value not in comma separated list of values  (i.e. ?name__not_in=John,Oliver)
	NotIn Filter = "not_in"
)

// converts filter to string type
func (f Filter) Str() string {
	return string(f)
}

// maps filter request key and value to ready to use query string and value
func (f Filter) Map(field, value string) (q string, v any, ok bool) {
	for filter, resolve := range filterQueryMapper {
		if filter.Str() == f.Str() {
			q, v = resolve(field, value)
			ok = true
			return
		}
	}

	return
}

var filterQueryMapper = filterQueryMap{
	Contains: func(field, value string) (string, any) {
		return fmt.Sprintf("%s LIKE ?", field), "%" + value + "%"
	},
	Equals: func(field, value string) (string, any) {
		return fmt.Sprintf("%s = ?", field), value
	},
	NotEquals: func(field, value string) (string, any) {
		return fmt.Sprintf("%s != ?", field), value
	},
	Greater: func(field, value string) (string, any) {
		return fmt.Sprintf("%s < ?", field), value
	},
	GreaterEquals: func(field, value string) (string, any) {
		return fmt.Sprintf("%s <= ?", field), value
	},
	Lesser: func(field, value string) (string, any) {
		return fmt.Sprintf("%s > ?", field), value
	},
	LesserEquals: func(field, value string) (string, any) {
		return fmt.Sprintf("%s >= ?", field), value
	},
	StartsWith: func(field, value string) (string, any) {
		return fmt.Sprintf("%s LIKE ?", field), value + "%"
	},
	EndsWith: func(field, value string) (string, any) {
		return fmt.Sprintf("%s LIKE ?", field), "%" + value
	},
	In: func(field, value string) (string, any) {
		return fmt.Sprintf("%s IN (?)", field), strings.Split(value, ",")
	},
	NotIn: func(field, value string) (string, any) {
		return fmt.Sprintf("%s NOT IN (?)", field), strings.Split(value, ",")
	},
}

// scope that enables filtering the results by [FilterScope.Fields] if a [Filter] is present in the request.
// (i.e. ?name__contains=John&age__gt=18)
type FilterScope struct {
	// fiber's request context
	Ctx *fiber.Ctx
	// fields to allow filtering by (i.e. name, age)
	Fields []string
}

// generates the GORM scope for filtering
func (f FilterScope) Scope() GScope {
	var queries []string
	var values []any

	for q, v := range f.Ctx.Queries() {
		var (
			query string
			value any
			ok    bool
		)

		if slices.Contains(f.Fields, q) {
			query, value, _ = Equals.Map(q, v)
			queries = append(queries, query)
			values = append(values, value)
			continue
		} else if !strings.Contains(q, "__") {
			continue
		}

		chunks := strings.Split(q, "__")

		if len(chunks) != 2 {
			continue
		}

		if !slices.Contains(f.Fields, chunks[0]) {
			continue
		}

		if query, value, ok = Filter(chunks[1]).Map(chunks[0], v); !ok {
			continue
		}

		queries = append(queries, query)
		values = append(values, value)
	}

	return func(db *gorm.DB) *gorm.DB {
		if len(queries) == 0 {
			return db
		}

		return db.Where(
			strings.Join(queries, " AND "),
			values...,
		)
	}
}
