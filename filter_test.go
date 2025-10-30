package fgf_test

import (
	"database/sql/driver"
	"log"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

type AnyArgIn struct {
	V []driver.Value
}

func (a AnyArgIn) Match(v driver.Value) bool {
	log.Println("match:", v)
	return slices.Contains(a.V, v)
}

func TestRequestFilterScope(t *testing.T) {
	assert := assert.New(t)
	rows := [][]driver.Value{
		{1, "Testing name 1", 22},
		{2, "Testing name 2", 42},
	}
	req := httptest.NewRequest(
		http.MethodGet,
		"/test-filter?name__contains=2&age=22",
		nil,
	)

	m := AnyArgIn{V: []driver.Value{"%2%", int64(22)}}

	Mock.ExpectQuery("SELECT (.+) FROM `test_models` WHERE").
		WithArgs(m, m).
		WillReturnRows(sqlmock.
			NewRows([]string{"id", "name", "age"}).
			AddRow(rows[0]...).
			AddRow(rows[1]...),
		)

	resp, err := App.Test(req, TestTimeoutMS)

	assert.Nil(err)
	assert.Equal(fiber.StatusOK, resp.StatusCode)
}

func TestRequestFilterScopeWrongFilter(t *testing.T) {
	assert := assert.New(t)
	rows := [][]driver.Value{
		{1, "Testing name 1", 22},
		{2, "Testing name 2", 42},
	}
	req := httptest.NewRequest(
		http.MethodGet,
		"/test-filter?age__gt=1&name__has=2",
		nil,
	)

	Mock.ExpectQuery("SELECT .* FROM `test_models` WHERE `age` < (.+)").
		WillReturnRows(sqlmock.
			NewRows([]string{"id", "name", "age"}).
			AddRow(rows[0]...).
			AddRow(rows[1]...),
		)

	resp, err := App.Test(req, TestTimeoutMS)

	assert.Nil(err)
	assert.Equal(fiber.StatusOK, resp.StatusCode)
}

func TestRequestFilterScopeBoolFilter(t *testing.T) {
	assert := assert.New(t)
	rows := [][]driver.Value{{1, 22, true}}
	req := httptest.NewRequest(
		http.MethodGet,
		"/test-filter?active=true",
		nil,
	)

	Mock.ExpectQuery("SELECT .* FROM `test_models` WHERE `active` = (.+)").
		WithArgs(true).
		WillReturnRows(sqlmock.
			NewRows([]string{"id", "age", "active"}).
			AddRow(rows[0]...),
		)

	resp, err := App.Test(req, TestTimeoutMS)

	assert.Nil(err)
	assert.Equal(fiber.StatusOK, resp.StatusCode)
}

func TestRequestFilterScopeSpecialFilter(t *testing.T) {
	assert := assert.New(t)
	rows := [][]driver.Value{
		{1, "Testing name 1", 22},
		{2, "Testing name 2", 42},
	}
	req := httptest.NewRequest(
		http.MethodGet,
		"/test-special-filter?age__neq=1",
		nil,
	)

	Mock.ExpectQuery("SELECT .* FROM `test_models` WHERE `age` != (.+)").
		WithArgs(2).
		WillReturnRows(sqlmock.
			NewRows([]string{"id", "name", "age"}).
			AddRow(rows[0]...).
			AddRow(rows[1]...),
		)

	resp, err := App.Test(req, TestTimeoutMS)

	assert.Nil(err)
	assert.Equal(fiber.StatusOK, resp.StatusCode)
}

func TestRequestFilterScopeIsNullFilter(t *testing.T) {
	assert := assert.New(t)
	rows := [][]driver.Value{{1, 22, true}}
	req := httptest.NewRequest(
		http.MethodGet,
		"/test-filter?active__isnull=true",
		nil,
	)

	Mock.ExpectQuery("SELECT .* FROM `test_models` WHERE `active` IS NULL").
		WillReturnRows(sqlmock.
			NewRows([]string{"id", "age", "active"}).
			AddRow(rows[0]...),
		)

	resp, err := App.Test(req, TestTimeoutMS)

	assert.Nil(err)
	assert.Equal(fiber.StatusOK, resp.StatusCode)
}

func TestRequestFilterScopeNotNullFilter(t *testing.T) {
	assert := assert.New(t)
	rows := [][]driver.Value{{1, 22, true}}
	req := httptest.NewRequest(
		http.MethodGet,
		"/test-filter?active__isnull=false",
		nil,
	)

	Mock.ExpectQuery("SELECT .* FROM `test_models` WHERE `active` IS NOT NULL").
		WillReturnRows(sqlmock.
			NewRows([]string{"id", "age", "active"}).
			AddRow(rows[0]...),
		)

	resp, err := App.Test(req, TestTimeoutMS)

	assert.Nil(err)
	assert.Equal(fiber.StatusOK, resp.StatusCode)
}
