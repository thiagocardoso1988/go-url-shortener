package routes

import (
	"github.com/thiagocardoso1988/go-url-shortener/api/helpers"

	"time"

	"github.com/gofiber/fiber/v2"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short`
	Expire      time.Duration `json:"expire"`
}

type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expire          time.Duration `json:"expire"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) error {
	body := new(request)

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	// implement rate limit

	// check if the input is an actual url
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid URL"})
	}

	// check for domain error
	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "You can't access this"})
	}

	// enforce https, SSL
	body.URL = helpers.EnforceHTTP(body.URL)
}
