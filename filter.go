package fgf

import (
	"fmt"
	"log"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/stoewer/go-strcase"
	"gorm.io/gorm"
)

// fixed set of supported filters
type Filter string

// map of special filter handlers, keyed by filter name (i.e. age__neq)
type SFilters map[string]func(value any, db *gorm.DB) *gorm.DB

type filterQueryMap map[Filter]func(string, any) (string, any)

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
	NotIn  Filter = "not_in"
	IsNull Filter = "isnull"
)

// converts filter to string type
func (f Filter) Str() string {
	return string(f)
}

// maps filter request key and value to ready to use query string and value
func (f Filter) Map(field string, value any) (q string, v any, ok bool) {
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
	Contains: func(field string, value any) (string, any) {
		return fmt.Sprintf("`%s` LIKE ?", field), fmt.Sprintf("%%%v%%", value)
	},
	Equals: func(field string, value any) (string, any) {
		return fmt.Sprintf("`%s` = ?", field), value
	},
	NotEquals: func(field string, value any) (string, any) {
		return fmt.Sprintf("`%s` != ?", field), value
	},
	Greater: func(field string, value any) (string, any) {
		return fmt.Sprintf("`%s` < ?", field), value
	},
	GreaterEquals: func(field string, value any) (string, any) {
		return fmt.Sprintf("`%s` <= ?", field), value
	},
	Lesser: func(field string, value any) (string, any) {
		return fmt.Sprintf("`%s` > ?", field), value
	},
	LesserEquals: func(field string, value any) (string, any) {
		return fmt.Sprintf("`%s` >= ?", field), value
	},
	StartsWith: func(field string, value any) (string, any) {
		return fmt.Sprintf("`%s` LIKE ?", field), fmt.Sprintf("%v%%", value)
	},
	EndsWith: func(field string, value any) (string, any) {
		return fmt.Sprintf("`%s` LIKE ?", field), fmt.Sprintf("%%%v", value)
	},
	In: func(field string, value any) (string, any) {
		if value, ok := value.(string); ok {
			return fmt.Sprintf("`%s` IN (?)", field), strings.Split(value, ",")
		}

		return fmt.Sprintf("`%s` IN (?)", field), value
	},
	NotIn: func(field string, value any) (string, any) {
		if value, ok := value.(string); ok {
			return fmt.Sprintf("`%s` NOT IN (?)", field), strings.Split(value, ",")
		}

		return fmt.Sprintf("`%s` NOT IN (?)", field), value
	},
	IsNull: func(field string, value any) (string, any) {
		if value == true || value == "true" || value == "1" {
			return fmt.Sprintf("`%s` IS NULL", field), nil
		}

		return fmt.Sprintf("`%s` IS NOT NULL", field), nil
	},
}

// scope that enables filtering the results by [FilterScope.Fields] if a [Filter] is present in the request.
// (i.e. ?name__contains=John&age__gt=18)
type FilterScope struct {
	// fiber's request context
	Ctx *fiber.Ctx
	// fields to allow filtering by (i.e. name, age)
	Fields []string
	// map of filter special handlers keyed with <field>__<filter> (i.e. name__contains, age__gt)
	Special SFilters
	// convert all datetime fields with date function
	ForceDate bool
	// optional uri to parse the query string from instead of [FilterScope.Ctx]
	FromUri string
	// optional table alias to use in the query (i.e. users)
	Alias string

	db            *gorm.DB
	specialValues map[string]any
}

// generates the GORM scope for filtering
func (f *FilterScope) Scope() GScope {
	if len(f.Special) > 0 && f.specialValues == nil {
		f.specialValues = make(map[string]any)
	}

	return func(db *gorm.DB) *gorm.DB {
		f.db = db
		queries, values := f.getQueriesAndValues()

		if len(queries) > 0 {
			db = db.Where(
				strings.Join(queries, " AND "),
				values...,
			)
		}

		if len(f.Special) > 0 {
			for k, scope := range f.Special {
				if v, ok := f.specialValues[k]; ok {
					db = scope(v, db)
				}
			}
		}

		return db
	}
}

func (f *FilterScope) getQueriesAndValues() (queries []string, values []any) {
	var model reflect.Value
	var params map[string]string
	var err error

	if f.db.Statement.Model != nil {
		model = reflect.Indirect(reflect.ValueOf(f.db.Statement.Model))
	}

	if params, err = f.getQueryParams(); err != nil {
		log.Println("FilterScope: failed to parse query params:", err.Error())
	}

	for q, v := range params {
		var (
			query string
			value any
			ok    bool
			err   error
		)

		if _, ok = f.Special[q]; ok {
			f.specialValues[q] = v
			continue
		}

		if slices.Contains(f.Fields, q) {
			if value, err = f.convertValue(model, q, v); err != nil {
				continue
			}

			query, value, _ = Equals.Map(q, value)
			query = f.mapQuery(query)
			queries = append(queries, query)
			values = append(values, value)
			continue
		}

		chunks := strings.Split(q, "__")

		if len(chunks) != 2 {
			continue
		}

		if !slices.Contains(f.Fields, chunks[0]) {
			continue
		}

		if value, err = f.convertValue(model, chunks[0], v); err != nil {
			continue
		}

		if query, value, ok = Filter(chunks[1]).Map(chunks[0], value); !ok {
			continue
		}

		query = f.mapQuery(query)

		if f.ForceDate {
			query = f.convertField(model, chunks[0], query)
		}

		queries = append(queries, query)

		if value != nil {
			values = append(values, value)
		}
	}

	return
}

func (f *FilterScope) mapQuery(query string) string {
	if len(f.Alias) > 0 {
		query = fmt.Sprintf("`%s`.%s", f.Alias, query)
	}
	return query
}

func (f *FilterScope) convertValue(model reflect.Value, field, value string) (o any, err error) {
	if !model.IsValid() {
		return value, nil
	}

	modelField := strcase.UpperCamelCase(field)
	kind := model.FieldByName(modelField).Kind()

	switch kind {
	case reflect.Bool:
		o, err = strconv.ParseBool(value)
	case reflect.Invalid:
		log.Println("FilterScope: field not found:", modelField)
		o = value
	default:
		o = value
	}

	if err != nil {
		log.Println("FilterScope: failed to convert value:", err.Error())
	}

	return
}

func (f *FilterScope) convertField(model reflect.Value, field, query string) string {
	if !model.IsValid() {
		return query
	}

	modelField := strcase.UpperCamelCase(field)
	value := model.FieldByName(modelField)
	kind := value.Kind()

	if kind == reflect.Struct &&
		value.Type().String() == "time.Time" {
		return strings.ReplaceAll(
			query,
			fmt.Sprintf("`%s`", field),
			fmt.Sprintf("DATE(`%s`)", field),
		)
	}

	return query
}

func (f *FilterScope) getQueryParams() (map[string]string, error) {
	if len(f.FromUri) > 0 {
		var uri *url.URL
		var err error
		var vs url.Values

		if uri, err = url.ParseRequestURI(f.FromUri); err != nil {
			return nil, err
		}

		if vs, err = url.ParseQuery(uri.RawQuery); err != nil {
			return nil, err
		}

		params := make(map[string]string, len(vs))

		for k, v := range vs {
			if len(v) > 0 {
				params[k] = v[len(v)-1]
			}
		}

		return params, nil
	}

	return f.Ctx.Queries(), nil
}
