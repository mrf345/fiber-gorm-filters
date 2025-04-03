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

func TestDefaultSortScope(t *testing.T) {
	assert := assert.New(t)
	req := httptest.NewRequest(http.MethodGet, "/test-sort", nil)
	rows := [][]driver.Value{
		{1, "Testing name 1", 22},
		{2, "Testing name 2", 42},
	}

	Mock.ExpectQuery("SELECT .* FROM `test_models` ORDER BY `name`").
		WillReturnRows(sqlmock.
			NewRows([]string{"id", "name", "age"}).
			AddRow(rows[0]...).
			AddRow(rows[1]...),
		)

	resp, err := App.Test(req, TestTimeoutMS)

	assert.Nil(err)
	assert.Equal(fiber.StatusOK, resp.StatusCode)
}

func TestRequestSortScope(t *testing.T) {
	assert := assert.New(t)
	req := httptest.NewRequest(http.MethodGet, "/test-sort?sort=-name,age", nil)
	rows := [][]driver.Value{
		{1, "Testing name 1", 22},
		{2, "Testing name 2", 42},
	}

	Mock.ExpectQuery("SELECT .* FROM `test_models` ORDER BY `name` DESC,`age`").
		WillReturnRows(sqlmock.
			NewRows([]string{"id", "name", "age"}).
			AddRow(rows[0]...).
			AddRow(rows[1]...),
		)

	resp, err := App.Test(req, TestTimeoutMS)

	assert.Nil(err)
	assert.Equal(fiber.StatusOK, resp.StatusCode)
}

func TestRequestSortScopeWrongField(t *testing.T) {
	assert := assert.New(t)
	req := httptest.NewRequest(http.MethodGet, "/test-sort?sort=occupation,wrong", nil)
	rows := [][]driver.Value{
		{1, "Testing name 1", 22},
		{2, "Testing name 2", 42},
	}

	Mock.ExpectQuery("SELECT .* FROM `test_models`").
		WillReturnRows(sqlmock.
			NewRows([]string{"id", "name", "age"}).
			AddRow(rows[0]...).
			AddRow(rows[1]...),
		)

	resp, err := App.Test(req, TestTimeoutMS)

	assert.Nil(err)
	assert.Equal(fiber.StatusOK, resp.StatusCode)
}
