package apihttp

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/tuvisolutions/tuvi-website-backend/internal/consultations"
)

type Handlers struct {
	svc       *consultations.Service
	apiToken  string
}

func NewHandlers(svc *consultations.Service, apiToken string) *Handlers {
	return &Handlers{svc: svc, apiToken: apiToken}
}

func (h *Handlers) Register(app *fiber.App) {
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	v1 := app.Group("/api/v1", h.authMiddleware)
	v1.Get("/consultations/availability", h.getAvailability)
	v1.Get("/consultations/availability/check", h.checkAvailability)
	v1.Post("/consultations", h.bookConsultation)
}

func (h *Handlers) authMiddleware(c *fiber.Ctx) error {
	auth := c.Get("Authorization")
	expected := "Bearer " + h.apiToken
	if auth != expected {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "unauthorized",
		})
	}
	return c.Next()
}

func (h *Handlers) getAvailability(c *fiber.Ctx) error {
	date := strings.TrimSpace(c.Query("date"))
	days := parseIntDefault(c.Query("days"), 0)

	result, err := h.svc.GetAvailability(c.Context(), date, days)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}
	return c.JSON(result)
}

func (h *Handlers) checkAvailability(c *fiber.Ctx) error {
	date := strings.TrimSpace(c.Query("date"))
	timeStr := strings.TrimSpace(c.Query("time"))
	if date == "" || timeStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "date and time query params are required",
		})
	}

	result, err := h.svc.CheckSlot(c.Context(), date, timeStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}
	return c.JSON(result)
}

func (h *Handlers) bookConsultation(c *fiber.Ctx) error {
	var req consultations.BookRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "invalid request body",
		})
	}

	success, conflict, err := h.svc.Book(c.Context(), req)
	if conflict != nil {
		return c.Status(fiber.StatusConflict).JSON(conflict)
	}
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(success)
}

func parseIntDefault(raw string, fallback int) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return n
}
