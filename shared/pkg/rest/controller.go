package rest

import "github.com/go-chi/chi/v5"

// Controller defines the interface for REST controllers
type Controller interface {
	RegisterRoutes(r chi.Router)
}
