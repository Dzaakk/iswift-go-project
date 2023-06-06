// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package cart

import (
	"gorm.io/gorm"
	"iswift-go-project/internal/cart/delivery/http"
	cart2 "iswift-go-project/internal/cart/repository"
	cart3 "iswift-go-project/internal/cart/usecase"
)

// Injectors from wire.go:

func InitializedService(db *gorm.DB) *cart.CartHandler {
	cartRepository := cart2.NewCartRepository(db)
	cartUseCase := cart3.NewCartUseCase(cartRepository)
	cartHandler := cart.NewCartHandler(cartUseCase)
	return cartHandler
}
