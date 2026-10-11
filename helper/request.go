package helper

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// CursorQuery menyimpan parameter pagination umum dari query string
type CursorQuery struct {
	Limit  int
	Cursor string
	Search string
}

// ParamID membaca parameter :id dari path URL, validasi angka positif
func ParamID(c *fiber.Ctx) (int, error) {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		return 0, BadRequest("id harus berupa angka positif")
	}
	return id, nil
}

// ParseCursorQuery membaca limit, cursor, dan search dari query string
func ParseCursorQuery(c *fiber.Ctx) CursorQuery {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return CursorQuery{
		Limit:  limit,
		Cursor: c.Query("cursor"),
		Search: c.Query("search"),
	}
}

// RequestContext membuat context dengan timeout 5 detik untuk query database
func RequestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}