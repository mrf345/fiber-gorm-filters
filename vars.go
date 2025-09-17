package fgf

import "gorm.io/gorm"

type GScope func(db *gorm.DB) *gorm.DB

var (
	// maximum number of items that can be returned per page
	MaxPageSize = 200
	// default number of items to return per page
	PageSize = 20
	// query param for the current page
	PageParam = "page"
	// query param for the number of items per page
	PageSizeParam = "page_size"
	// query param for the sort order (comma separated list of fields, with optional - prefix to reverse the sort order)
	SortParam = "sort"
)
