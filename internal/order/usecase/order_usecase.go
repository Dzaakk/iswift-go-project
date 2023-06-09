package order

import (
	"errors"
	cartUseCase "iswift-go-project/internal/cart/usecase"
	discountEntity "iswift-go-project/internal/discount/entity"
	discountUseCase "iswift-go-project/internal/discount/usecase"
	dto "iswift-go-project/internal/order/dto"
	entity "iswift-go-project/internal/order/entity"
	repository "iswift-go-project/internal/order/repository"
	orderDetailEntity "iswift-go-project/internal/order_detail/entity"
	orderDetailUseCase "iswift-go-project/internal/order_detail/usecase"
	paymentDto "iswift-go-project/internal/payment/dto"
	paymentUseCase "iswift-go-project/internal/payment/usecase"
	productEntity "iswift-go-project/internal/product/entity"
	productUseCase "iswift-go-project/internal/product/usecase"
	"strconv"

	"github.com/google/uuid"
)

type OrderUseCase interface {
	FindAll(offset, limit int) []entity.Order
	FindAllByUserId(offset, limit, userId int) []entity.Order
	FindById(id int) (*entity.Order, error)
	FindByExternalId(externalId string) (*entity.Order, error)
	Count() int
	Create(dto dto.OrderRequestBody) (*entity.Order, error)
	Update(id int, dto dto.OrderRequestBody) (*entity.Order, error)
}

type OrderUseCaseImpl struct {
	repository         repository.OrderRepository
	cartUseCase        cartUseCase.CartUseCase
	discountUseCase    discountUseCase.DiscountUseCase
	productUseCase     productUseCase.ProductUseCase
	orderDetailUseCase orderDetailUseCase.OrderDetailUsecase
	paymentUseCase     paymentUseCase.PaymentUseCase
}

// Count implements OrderUseCase.
func (usecase *OrderUseCaseImpl) Count() int {
	return usecase.repository.Count()
}

// FindAllByUserId implements OrderUseCase.
func (usecase *OrderUseCaseImpl) FindAllByUserId(offset int, limit int, userId int) []entity.Order {
	return usecase.repository.FindAllByUserId(offset, limit, userId)
}

// Update implements OrderUseCase.
func (usecase *OrderUseCaseImpl) Update(id int, dto dto.OrderRequestBody) (*entity.Order, error) {
	order, err := usecase.repository.FindById(id)

	if err != nil {
		return nil, err
	}

	order.Status = dto.Status

	updateOrder, err := usecase.repository.Update(*order)

	if err != nil {
		return nil, err
	}

	return updateOrder, nil
}

// FindByExternalId implements OrderUseCase.
func (usecase *OrderUseCaseImpl) FindByExternalId(externalId string) (*entity.Order, error) {
	return usecase.repository.FindOneByExternalId(externalId)
}

// Create implements OrderUseCase.
func (usecase *OrderUseCaseImpl) Create(dto dto.OrderRequestBody) (*entity.Order, error) {
	price := 0
	totalPrice := 0
	description := ""
	var products []productEntity.Product

	order := entity.Order{
		UserID: dto.UserID,
		Status: "pending",
	}

	var dataDiscount *discountEntity.Discount

	// Cari data carts berdasarkan user id
	carts := usecase.cartUseCase.FindByUserId(int(dto.UserID), 0, 999)

	// Check apakah keranjang kosong atau tidak
	if len(carts) == 0 {
		//Jika kosong kita akan melakukan pemeriksaan product idnya apakah akan dikirimkan atau tidak
		if dto.ProductID == nil {
			return nil, errors.New("cart anda kosong atau anda belum memasukan product id")
		}
	}

	// Check data discount
	if dto.DiscountCode != nil {
		dataDiscount, err := usecase.discountUseCase.FindByCode(*dto.DiscountCode)

		if err != nil {
			return nil, errors.New("code discount sudah tidak berlaku")
		}

		if dataDiscount.RemainingQuantity == 0 {
			return nil, errors.New("kode discount sudah tidak berlaku")
		}
		// Validasi lainnya (ex: check start date dan end date)
	}

	if len(carts) > 0 {
		// Jika menggunakan keranjang
		for _, cart := range carts {
			product, err := usecase.productUseCase.FindById(int(cart.ProductID))

			if err != nil {
				return nil, err
			}

			products = append(products, *product)

		}
	} else if dto.ProductID != nil {
		// Jika user langsung melakukan checkout
		product, err := usecase.productUseCase.FindById(int(*dto.ProductID))

		if err != nil {
			return nil, err
		}

		products = append(products, *product)
	}

	// Kalkulasi price serta membuat description ke xendit
	for index, product := range products {
		price += int(product.Price)

		i := strconv.Itoa(index + 1)

		description = i + ". Product : " + product.Title + "<br/>"
	}

	totalPrice = price

	if dataDiscount != nil {
		// Hitung Logic Discount
		if dataDiscount.Type == "rebate" {
			totalPrice = price - int(dataDiscount.Value)
		} else if dataDiscount.Type == "percent" {
			totalPrice = price - (price / 100 * int(dataDiscount.Value))
		}

		order.DiscountID = &dataDiscount.ID
	}

	order.Price = int64(price)           // Harga asli
	order.TotalPrice = int64(totalPrice) //Harga sudah dikurangi discount
	order.CreatedByID = dto.UserID

	externalId := uuid.New().String()

	order.ExternalID = externalId

	// Insert to table order
	data, err := usecase.repository.Create(order)

	if err != nil {
		return nil, err
	}

	//Insert to table order details
	for _, product := range products {
		orderDetail := orderDetailEntity.OrderDetail{
			OrderID:     data.ID,
			ProductID:   product.ID,
			CreatedByID: order.UserID,
			Price:       product.Price,
		}

		usecase.orderDetailUseCase.Create(orderDetail)
	}

	// Hit payment xendit
	dataPayment := paymentDto.PaymentRequestBody{
		ExternalID:  externalId,
		Amount:      int64(totalPrice),
		PayerEmail:  dto.Email,
		Description: description,
	}

	payment, err := usecase.paymentUseCase.Create(dataPayment)
	if err != nil {
		return nil, err
	}

	data.CheckoutLink = payment.InvoiceURL

	usecase.repository.Update(*data)

	// Update remaining quantity discount
	if dto.DiscountCode != nil {
		_, err := usecase.discountUseCase.UpdateRemainingQuantity(int(dataDiscount.ID), 1, "-")
		if err != nil {
			return nil, err
		}
	}

	// Delete carts
	err = usecase.cartUseCase.DeleteByUserId(int(dto.UserID))
	if err != nil {
		return nil, err
	}

	return data, nil
}

// FindAll implements OrderUseCase.
func (usecase *OrderUseCaseImpl) FindAll(offset int, limit int) []entity.Order {
	return usecase.repository.FindAll(offset, limit)
}

// FindById implements OrderUseCase.
func (usecase *OrderUseCaseImpl) FindById(id int) (*entity.Order, error) {
	return usecase.repository.FindById(id)
}

func NewOrderUseCase(
	repository repository.OrderRepository,
	cartUseCase cartUseCase.CartUseCase,
	discountUseCase discountUseCase.DiscountUseCase,
	productUseCase productUseCase.ProductUseCase,
	orderDetailUseCase orderDetailUseCase.OrderDetailUsecase,
	paymentUseCase paymentUseCase.PaymentUseCase,
) OrderUseCase {

	return &OrderUseCaseImpl{
		repository,
		cartUseCase,
		discountUseCase,
		productUseCase,
		orderDetailUseCase,
		paymentUseCase,
	}
}
