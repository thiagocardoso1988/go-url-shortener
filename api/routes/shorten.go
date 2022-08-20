package routes

import (
	"strconv"

	"github.com/thiagocardoso1988/go-url-shortener/api/database"
	"github.com/thiagocardoso1988/go-url-shortener/api/helpers"

	"os"
	"time"

	"github.com/go-redis/redis/v8"
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
	r2 := database.CreateClient((1))
	defer r2.Close()

	value, err := r2.Get(database.Ctx, c.IP()).Result()
	if err == redis.Nil {
		_ = r2.Set(database.Ctx, c.IP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
	}

	value, _ = r2.Get(database.Ctx, c.IP()).Result()
	valueInt, _ := strconv.Atoi(value)

	if valueInt <= 0 {
		limit, _ := r2.TTL(database.Ctx, c.IP()).Result()
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error":            "Rate limit exceeded",
			"rate_limit_reset": limit / time.Nanosecond / time.Minute,
		})
	}

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
	r2.Decr(database.Ctx, c.IP())

	return nil
}
