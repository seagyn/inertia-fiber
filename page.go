package inertia

import "github.com/gofiber/fiber/v2"

// Page type.
type Page struct {
	Component string                 `json:"component"`
	Routes    []*fiber.Route         `json:"routes"`
	Props     map[string]interface{} `json:"props"`
	URL       string                 `json:"url"`
	Version   string                 `json:"version"`
}
