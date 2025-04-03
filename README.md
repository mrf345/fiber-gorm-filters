<h2></h2>
<h1>
fiber-gorm-filters (fgf)
<a href='https://github.com/mrf345/fiber-gorm-filters/actions/workflows/ci.yml'>
  <img src='https://github.com/mrf345/fiber-gorm-filters/actions/workflows/ci.yml/badge.svg' alt='build status'>
</a>
<a href="https://pkg.go.dev/github.com/mrf345/fiber-gorm-filters">
  <img src='https://img.shields.io/badge/reference-blue.svg?logo=go&logoColor=white' alt='Go Reference'>
</a>
</h1>

Provides a simple way to add filtering, sorting and pagination to  fiber/gorm applications. Inspired by [django-filter](https://github.com/carltongibson/django-filter)


### Install

```shell
go get https://github.com/mrf345/fiber-gorm-filters@latest
```

### Examples

Assuming you have the following GORM model defined:

```go
type User struct {
    ID         uint
    Name       string
    Age        uint
    Occupation string
}
```

#### Sorting

```go
import (
    "github.com/gofiber/fiber/v2"
    fgf "github.com/mrf345/fiber-gorm-filters"
)

// Within your fiber endpoint handler:
func ListUsers(c *fiber.Ctx) error {
    var users []User

    // Default sort by name descending, and allow request sorting by name and age
    var sort = fgf.SortScope{Ctx: c, Default: []string{"-name"}, Fields: []string{"age"}}

    // Apply sorting scope
    if err := DB.Scopes(sort.Scope()).Find(&users).Error; err != nil {
        log.Println(err)
        _ = c.SendStatus(fiber.StatusInternalServerError)
        return err
    }

    return c.JSON(users)
}
```

#### Filtering

```go
import (
    "github.com/gofiber/fiber/v2"
    fgf "github.com/mrf345/fiber-gorm-filters"
)

// Within your fiber endpoint handler:
func ListUsers(c *fiber.Ctx) error {
    var users []User

    // Enable filtering by age and name
    var filter = fgf.FilterScope{Ctx: c, Fields: []string{"age", "name"}}

    // Apply filtering scope
    if err := DB.Scopes(filter.Scope()).Find(&users).Error; err != nil {
        log.Println(err)
        _ = c.SendStatus(fiber.StatusInternalServerError)
        return err
    }

    return c.JSON(users)
}
```

#### Pagination

```go
import (
    "github.com/gofiber/fiber/v2"
    fgf "github.com/mrf345/fiber-gorm-filters"
)

// Within your fiber endpoint handler:
func ListUsers(c *fiber.Ctx) error {
    var users []User
    var page = fgf.PageScope{Ctx: c}

    // Get and set total count of users
    if err = database.I.DB.
        Model(&User{}).
        Count(&page.Total).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(
            fiber.Map{
                "message": err.Error(),
            },
        )
    }

    // Apply pagination scope
    if err := DB.Scopes(page.Scope()).Find(&users).Error; err != nil {
        log.Println(err)
        _ = c.SendStatus(fiber.StatusInternalServerError)
        return err
    }

    // Return paginated response
    return page.Resp(users)
}
```

### Usage

#### Settings

By default the package has the following settings:

```go

var (
    // maximum number of results that can be returned per page
    MaxPageSize = 200
    // default number of results to return per page
    PageSize = 20
    // query param for the current page
    PageParam = "page"
    // query param for the number of results per page
    PageSizeParam = "page_size"
    // query param for the sort order (comma separated list of fields, with optional - prefix to reverse the sort order)
    SortParam = "sort"
)
```

You can update them before running your app, example:

```go
import (
    fgf "github.com/mrf345/fiber-gorm-filters"
)

func main() {
    // increase maximum page size allowed
    fgf.MaxPageSize = 500
    // change the default page size GET parameter
    fgf.PageSizeParam = "size"

    app := fiber.New()
    log.Fatal(app.Listen(":8080"))
}
```

#### Filters

List of filters that can be used, and their corresponding query parameters:

| Filter | Query Param | Description |
| --- | --- | --- |
| `Equals` | `<field>` | Equals to the value provided. Example: `/users?name=John` or `name__eq` will return all users with name equal to John. |
| `NotEquals` | `<field>__neq` | Not equals to the value provided. Example: `/users?age__neq=20` will return all users that are not 20 years old. |
| `Greater` | `<field>__gt` | Greater than the value provided. Example: `/users?age__gt=18` will return all users older than 18. |
| `Lesser` | `<field>__lt` | Less than the value provided. Example: `/users?age__lt=20` will return all users younger than 20. |
| `GreaterEquals` | `<field>__gte` | Greater or equal to the value provided. Example: `/users?age__gte=18` will return all users older or equal to 18. |
| `LesserEquals` | `<field>__lte` | Less or equal to the value provided. Example: `/users?age__lte=20` will return all users younger or equal to 20. |
| `In` | `<field>__in` | In a list of values provided. Example: `/users?name__in=John,Jack,Jane` will return all users with name John, Jack and Jane. |
| `NotIn` | `<field>__not_in` | Not in a list of values provided. Example: `/users?name__not_in=John,Jack,Jane` will return all users that are not named John, Jack or Jane. |
| `Contains` | `<field>__contains` | Contains the value provided. Example: `/users?occupation__contains=developer` will return all users with occupation containing "developer". |
| `StartsWith` | `<field>__startswith` | Starts with the value provided. Example: `/users?name__startswith=john` will return all users that start their name with John. |
| `EndsWith` | `<field>__endswith` | Ends with the value provided. Example: `/users?name__endswith=john` will return all users that end their name with John. |
