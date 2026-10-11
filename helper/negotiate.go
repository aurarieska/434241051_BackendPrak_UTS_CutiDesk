package helper

import (
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Negotiate memeriksa header Accept dan menentukan format response
// Kembalikan "json" atau "csv", atau error 406 jika format tidak tersedia
func Negotiate(c *fiber.Ctx) (string, error) {
	accept := c.Get("Accept", "application/json")

	if accept == "" || accept == "*/*" || strings.Contains(accept, "application/json") {
		return "json", nil
	}
	if strings.Contains(accept, "text/csv") {
		return "csv", nil
	}

	return "", NotAcceptable(
		"format yang diminta tidak tersedia, gunakan application/json atau text/csv",
	)
}

// TulisCSV menulis data sebagai file CSV ke response
// header: nama kolom, baris: isi data per baris
func TulisCSV(c *fiber.Ctx, namaFile string, header []string, baris [][]string) error {
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, namaFile))

	var sb strings.Builder
	writer := csv.NewWriter(&sb)

	// Tulis baris header
	if err := writer.Write(header); err != nil {
		return Internal("gagal menulis header CSV", err)
	}

	// Tulis baris data
	for _, b := range baris {
		if err := writer.Write(b); err != nil {
			return Internal("gagal menulis baris CSV", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return Internal("gagal flush CSV", err)
	}

	return c.SendString(sb.String())
}