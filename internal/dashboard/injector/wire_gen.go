// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package dashboard

import (
	"gorm.io/gorm"
	"iswift-go-project/internal/admin/repository"
	admin2 "iswift-go-project/internal/admin/usecase"
	"iswift-go-project/internal/cart/repository"
	cart2 "iswift-go-project/internal/cart/usecase"
	"iswift-go-project/internal/dashboard/delivery/http"
	dashboard2 "iswift-go-project/internal/dashboard/usecase"
	"iswift-go-project/internal/discount/repository"
	discount2 "iswift-go-project/internal/discount/usecase"
	"iswift-go-project/internal/order/repository"
	order2 "iswift-go-project/internal/order/usecase"
	"iswift-go-project/internal/order_detail/repository"
	order_detail2 "iswift-go-project/internal/order_detail/usecase"
	"iswift-go-project/internal/payment/usecase"
	"iswift-go-project/internal/product/repository"
	"iswift-go-project/internal/product/usecase"
	"iswift-go-project/internal/user/repository"
	user2 "iswift-go-project/internal/user/usecase"
	"iswift-go-project/pkg/fileupload/cloudinary"
)

// Injectors from wire.go:

func InitializedService(db *gorm.DB) *dashboard.DashboardHandler {
	userRepository := user.NewUserRepository(db)
	userUseCase := user2.NewUserUseCase(userRepository)
	adminRepository := admin.NewAdminRepository(db)
	adminUseCase := admin2.NewAdminUseCase(adminRepository)
	orderRepository := order.NewOrderRepository(db)
	cartRepository := cart.NewCartRepository(db)
	cartUseCase := cart2.NewCartUseCase(cartRepository)
	discountRepository := discount.NewDiscountRepository(db)
	discountUseCase := discount2.NewDiscountUseCase(discountRepository)
	productRepository := product.NewProductRepository(db)
	fileUpload := fileupload.NewFileUpload()
	productUseCase := products.NewProductUsecase(productRepository, fileUpload)
	orderDetailRepository := order_detail.NewOrderDetailRepository(db)
	orderDetailUsecase := order_detail2.NewOrderDetailUseCase(orderDetailRepository)
	paymentUseCase := payment.NewPaymentUseCase()
	orderUseCase := order2.NewOrderUseCase(orderRepository, cartUseCase, discountUseCase, productUseCase, orderDetailUsecase, paymentUseCase)
	dashboardUseCase := dashboard2.NewDashboardUseCase(userUseCase, adminUseCase, orderUseCase, productUseCase)
	dashboardHandler := dashboard.NewDashboardHandler(dashboardUseCase)
	return dashboardHandler
}
