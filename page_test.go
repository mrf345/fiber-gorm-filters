package fgf_test

import (
	"database/sql/driver"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	fgf "github.com/mrf345/fiber-gorm-filters"
	"github.com/stretchr/testify/assert"
)

func TestPageScope(t *testing.T) {
	assert := assert.New(t)
	total := 50
	rows := [][]driver.Value{}
	req := httptest.NewRequest(
		http.MethodGet,
		"/test-page",
		nil,
	)

	for i := range total {
		rows = append(rows, []driver.Value{
			i + 1,
			"Testing name 1",
			i * 10,
		})
	}

	Mock.ExpectQuery("SELECT .* FROM `test_models` LIMIT (.+)").
		WillReturnRows(sqlmock.
			NewRows([]string{"id", "name", "age"}).
			AddRows(rows...),
		)

	resp, err := App.Test(req, TestTimeoutMS)
	data := GetRespParsedBody[fgf.PaginatedResponse[[]TestModel]](resp)

	assert.Nil(err)
	assert.Equal(fiber.StatusOK, resp.StatusCode)
	assert.Equal(data.Total, total)
	assert.Equal(data.Next, 2)
	assert.Equal(data.Prev, 0)
}

func TestRequestPageScope(t *testing.T) {
	assert := assert.New(t)
	total := 50
	rows := [][]driver.Value{}
	page := 3
	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/test-page?page=%d&page_size=10", page),
		nil,
	)

	for i := range total {
		rows = append(rows, []driver.Value{
			i + 1,
			"Testing name 1",
			i * 10,
		})
	}

	Mock.ExpectQuery("SELECT .* FROM `test_models` LIMIT (.+)").
		WillReturnRows(sqlmock.
			NewRows([]string{"id", "name", "age"}).
			AddRows(rows...),
		)

	resp, err := App.Test(req, TestTimeoutMS)
	data := GetRespParsedBody[fgf.PaginatedResponse[[]TestModel]](resp)

	assert.Nil(err)
	assert.Equal(fiber.StatusOK, resp.StatusCode)
	assert.Equal(data.Total, total)
	assert.Equal(data.Next, page+1)
	assert.Equal(data.Prev, page-1)
}
