package helper

import (
	"cutidesk/app/model"

	"github.com/gofiber/fiber/v2"
)

// Success mengirim response sukses dengan data (200 OK)
func Success(c *fiber.Ctx, pesan string, data interface{}) error {
	return c.JSON(model.WebResponse{
		Success: true,
		Message: pesan,
		Data:    data,
	})
}

// SuccessCursor mengirim response sukses dengan data dan cursor pagination
func SuccessCursor(c *fiber.Ctx, pesan string, data interface{}, meta model.CursorMeta) error {
	return c.JSON(model.WebResponse{
		Success: true,
		Message: pesan,
		Data:    data,
		Meta:    meta,
	})
}

// Created mengirim response 201 Created dengan header Location
func Created(c *fiber.Ctx, pesan string, data interface{}, lokasi string) error {
	c.Set("Location", lokasi)
	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{
		Success: true,
		Message: pesan,
		Data:    data,
	})
}

// NoContent mengirim response 204 No Content tanpa body
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}