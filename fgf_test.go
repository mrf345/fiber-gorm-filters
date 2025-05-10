package fgf_test

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	fgf "github.com/mrf345/fiber-gorm-filters"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	TestTimeoutMS = 1000
)

var (
	Mock sqlmock.Sqlmock
	App  *fiber.App
	DB   *gorm.DB
)

type TestModel struct {
	ID         uint
	Name       string
	Age        uint
	Occupation string
	Active     bool
	Created    time.Time
}

func GetRespParsedBody[T any](resp *http.Response) (respData T) {
	data := make([]byte, resp.ContentLength)
	_, _ = resp.Body.Read(data)
	_ = json.Unmarshal(data, &respData)

	if reflect.ValueOf(respData).IsZero() {
		log.Println("GetRespParsedBody failed:", string(data))
	}

	_ = resp.Body.Close()
	return
}

func TestMain(m *testing.M) {
	var db *sql.DB
	DB, db, Mock = setupTestDB()
	App = fiber.New()
	setupRoutes(App)
	m.Run()
	_ = db.Close()
}

func setupTestDB() (gdb *gorm.DB, db *sql.DB, mock sqlmock.Sqlmock) {
	var err error
	db, mock, err = sqlmock.New()

	if err != nil {
		log.Fatalln("failed to setup test database")
	}

	con := mysql.New(mysql.Config{
		DSN:                       "sqlmock_db_0",
		DriverName:                "mysql",
		Conn:                      db,
		SkipInitializeWithVersion: true,
	})

	gdb, err = gorm.Open(con, &gorm.Config{})

	if err != nil {
		log.Fatalln("failed to setup test database")
	}

	return
}

func setupRoutes(app *fiber.App) {
	app.Get("/test-sort", func(c *fiber.Ctx) error {
		var items []TestModel
		var sort = fgf.SortScope{Ctx: c, Default: []string{"name"}, Fields: []string{"age"}}

		if err := DB.Scopes(sort.Scope()).Find(&items).Error; err != nil {
			log.Println(err)
			_ = c.SendStatus(fiber.StatusInternalServerError)
			return err
		}

		return c.JSON(items)
	})

	app.Get("/test-filter", func(c *fiber.Ctx) error {
		var items []TestModel
		var filter = fgf.FilterScope{
			Ctx:       c,
			Fields:    []string{"age", "name", "active", "created"},
			ForceDate: true,
		}

		if err := DB.
			Model(&TestModel{}).
			Scopes(filter.Scope()).
			Find(&items).Error; err != nil {
			log.Println(err)
			_ = c.SendStatus(fiber.StatusInternalServerError)
			return err
		}

		return c.JSON(items)
	})

	app.Get("/test-special-filter", func(c *fiber.Ctx) error {
		var items []TestModel
		var filter = fgf.FilterScope{Ctx: c, Special: fgf.SFilters{
			"age__neq": func(value any, db *gorm.DB) *gorm.DB {
				v, _ := strconv.ParseInt(value.(string), 10, 64)
				return db.Where("`age` != ?", v+1)
			},
		}}

		if err := DB.
			Model(&TestModel{}).
			Scopes(filter.Scope()).
			Find(&items).Error; err != nil {
			log.Println(err)
			_ = c.SendStatus(fiber.StatusInternalServerError)
			return err
		}

		return c.JSON(items)
	})

	app.Get("/test-page", func(c *fiber.Ctx) error {
		var items []TestModel
		var page = fgf.PageScope{Ctx: c, Total: 35}

		if err := DB.Scopes(page.Scope()).Find(&items).Error; err != nil {
			log.Println(err)
			_ = c.SendStatus(fiber.StatusInternalServerError)
			return err
		}

		return page.Resp(items)
	})
}
