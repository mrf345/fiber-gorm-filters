package fgf_test

import (
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

type AnyReqArg struct{}

func (a AnyReqArg) Match(v driver.Value) bool {
	if v == "%2%" || v == "22" {
		return true
	}

	return false
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

	Mock.ExpectQuery("SELECT (.+) FROM `test_models` WHERE").
		WithArgs(AnyReqArg{}, AnyReqArg{}).
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
	Mock.ExpectQuery("SELECT .* FROM `test_models` WHERE age < (.+)").
		WillReturnRows(sqlmock.
			NewRows([]string{"id", "name", "age"}).
			AddRow(rows[0]...).
			AddRow(rows[1]...),
		)

	resp, err := App.Test(req, TestTimeoutMS)

	assert.Nil(err)
	assert.Equal(fiber.StatusOK, resp.StatusCode)
}
