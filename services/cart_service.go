package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator"
	"github.com/sirupsen/logrus"
	"github.com/tensuqiuwulu/be-service-teman-bunda/config"
	"github.com/tensuqiuwulu/be-service-teman-bunda/exceptions"
	"github.com/tensuqiuwulu/be-service-teman-bunda/models/entity"
	"github.com/tensuqiuwulu/be-service-teman-bunda/models/http/request"
	"github.com/tensuqiuwulu/be-service-teman-bunda/models/http/response"
	"github.com/tensuqiuwulu/be-service-teman-bunda/repository/mysql"
	"github.com/tensuqiuwulu/be-service-teman-bunda/utilities"
	"gorm.io/gorm"
)

type CartServiceInterface interface {
	AddProductToCart(requestId string, idUser string, addProductToCartRequest *request.AddProductToCartRequest) (addProductToCartResponse response.AddProductToCartResponse)
	FindCartByIdUser(requestId string, idUser string, IdKelurahan int) (cartResponses response.FindCartByIdUserResponse)
	CartPlusQtyProduct(requestId string, updateQtyProductInCartRequest *request.UpdateQtyProductInCartRequest) (updateProductQtyInCartResponse response.UpdateProductQtyInCartResponse)
	CartMinQtyProduct(requestId string, updateQtyProductInCartRequest *request.UpdateQtyProductInCartRequest) (updateProductQtyInCartResponse response.UpdateProductQtyInCartResponse)
	UpdateQtyProductInCart(requestId string, updateQtyProductInCartRequest *request.UpdateQtyProductInCartRequest) (updateProductQtyInCartResponse response.UpdateProductQtyInCartResponse)
	// DeleteProductInCart(requestId string, deleteProductInCartRequest *request.DeleteProductInCartRequest)
}

type CartServiceImplementation struct {
	ConfigWebserver             config.Webserver
	DB                          *gorm.DB
	Validate                    *validator.Validate
	Logger                      *logrus.Logger
	CartRepositoryInterface     mysql.CartRepositoryInterface
	ShippingRepositoryInterface mysql.ShippingRepositoryInterface
	ProductRepositoryInterface  mysql.ProductRepositoryInterface
	SettingRepositoryInterface  mysql.SettingRepositoryInterface
}

func NewCartService(
	configWebserver config.Webserver,
	DB *gorm.DB,
	validate *validator.Validate,
	logger *logrus.Logger,
	cartRepositoryInterface mysql.CartRepositoryInterface,
	shippingRepositoryInterface mysql.ShippingRepositoryInterface,
	productRepositoryInterface mysql.ProductRepositoryInterface,
	settingRepositoryInterface mysql.SettingRepositoryInterface) CartServiceInterface {
	return &CartServiceImplementation{
		ConfigWebserver:             configWebserver,
		DB:                          DB,
		Validate:                    validate,
		Logger:                      logger,
		CartRepositoryInterface:     cartRepositoryInterface,
		ShippingRepositoryInterface: shippingRepositoryInterface,
		ProductRepositoryInterface:  productRepositoryInterface,
		SettingRepositoryInterface:  settingRepositoryInterface,
	}
}

func (service *CartServiceImplementation) FindCartByIdUser(requestId string, IdUser string, IdKelurahan int) (addProductToCartResponse response.FindCartByIdUserResponse) {
	carts, _ := service.CartRepositoryInterface.FindCartByIdUser(service.DB, IdUser)
	shippingCost, err := service.SettingRepositoryInterface.FindSettingShippingCost(service.DB)

	exceptions.PanicIfError(err, requestId, service.Logger)
	addProductToCartResponse = response.ToFindCartByIdUserResponse(carts, shippingCost.Value)
	return addProductToCartResponse
}

func (service *CartServiceImplementation) CartPlusQtyProduct(requestId string, updateQtyProductInCartRequest *request.UpdateQtyProductInCartRequest) (updateProductQtyInCartResponse response.UpdateProductQtyInCartResponse) {
	request.ValidateUpdateQtyProductInCartRequest(service.Validate, updateQtyProductInCartRequest, requestId, service.Logger)
	cartProductExist, _ := service.CartRepositoryInterface.FindCartById(service.DB, updateQtyProductInCartRequest.IdCart)
	cartEntity := &entity.Cart{}
	cartEntity.Id = updateQtyProductInCartRequest.IdCart
	cartEntity.Qty = cartProductExist.Qty + 1
	cartResult, err := service.CartRepositoryInterface.UpdateProductInCart(service.DB, updateQtyProductInCartRequest.IdCart, *cartEntity)
	exceptions.PanicIfError(err, requestId, service.Logger)
	updateProductQtyInCartResponse = response.ToUpdateProductQtyInCartResponse(cartResult)
	return updateProductQtyInCartResponse
}

func (service *CartServiceImplementation) CartMinQtyProduct(requestId string, updateQtyProductInCartRequest *request.UpdateQtyProductInCartRequest) (updateProductQtyInCartResponse response.UpdateProductQtyInCartResponse) {
	request.ValidateUpdateQtyProductInCartRequest(service.Validate, updateQtyProductInCartRequest, requestId, service.Logger)
	cartProductExist, _ := service.CartRepositoryInterface.FindCartById(service.DB, updateQtyProductInCartRequest.IdCart)

	if cartProductExist.Qty == 1 {
		err := service.CartRepositoryInterface.DeleteProductInCart(service.DB, cartProductExist.Id)
		exceptions.PanicIfError(err, requestId, service.Logger)
		updateProductQtyInCartResponse = response.ToUpdateProductQtyInCartResponse(entity.Cart{Id: cartProductExist.Id})
		return updateProductQtyInCartResponse
	} else {
		cartEntity := &entity.Cart{}
		cartEntity.Id = updateQtyProductInCartRequest.IdCart
		cartEntity.Qty = cartProductExist.Qty - 1
		cartResult, err := service.CartRepositoryInterface.UpdateProductInCart(service.DB, updateQtyProductInCartRequest.IdCart, *cartEntity)
		exceptions.PanicIfError(err, requestId, service.Logger)
		updateProductQtyInCartResponse = response.ToUpdateProductQtyInCartResponse(cartResult)
		return updateProductQtyInCartResponse
	}
}

func (service *CartServiceImplementation) UpdateQtyProductInCart(requestId string, updateQtyProductInCartRequest *request.UpdateQtyProductInCartRequest) (updateProductQtyInCartResponse response.UpdateProductQtyInCartResponse) {

	request.ValidateUpdateQtyProductInCartRequest(service.Validate, updateQtyProductInCartRequest, requestId, service.Logger)
	cartProductExist, _ := service.CartRepositoryInterface.FindCartById(service.DB, updateQtyProductInCartRequest.IdCart)

	fmt.Println("request id cart = ", updateQtyProductInCartRequest.IdCart)
	fmt.Println("request qty = ", updateQtyProductInCartRequest.Qty)

	if cartProductExist.Id == "" {
		err := errors.New("id cart not found")
		exceptions.PanicIfRecordNotFound(err, requestId, []string{"id cart not found"}, service.Logger)
	}

	if updateQtyProductInCartRequest.Qty == 0 {
		err := service.CartRepositoryInterface.DeleteProductInCart(service.DB, cartProductExist.Id)
		exceptions.PanicIfError(err, requestId, service.Logger)
		updateProductQtyInCartResponse = response.ToUpdateProductQtyInCartResponse(entity.Cart{Id: cartProductExist.Id})
		return updateProductQtyInCartResponse
	} else {
		cartEntity := &entity.Cart{}
		cartEntity.Id = updateQtyProductInCartRequest.IdCart
		cartEntity.Qty = updateQtyProductInCartRequest.Qty
		cartResult, err := service.CartRepositoryInterface.UpdateProductInCart(service.DB, updateQtyProductInCartRequest.IdCart, *cartEntity)
		exceptions.PanicIfError(err, requestId, service.Logger)
		updateProductQtyInCartResponse = response.ToUpdateProductQtyInCartResponse(cartResult)
		return updateProductQtyInCartResponse
	}
}

func (service *CartServiceImplementation) AddProductToCart(requestId string, IdUser string, addProductToCartRequest *request.AddProductToCartRequest) (addProductToCartResponse response.AddProductToCartResponse) {
	// Validate request
	request.ValidateAddProductToCartRequest(service.Validate, addProductToCartRequest, requestId, service.Logger)

	// Cek apakah produk yang dimasukkan sudah ada di keranjang
	cartProductExist, _ := service.CartRepositoryInterface.FindProductInCartByIdUser(service.DB, IdUser, addProductToCartRequest.IdProduct)

	// Cek Prduct Stock
	product, _ := service.ProductRepositoryInterface.FindProductById(service.DB, addProductToCartRequest.IdProduct)
	if product.Stock < 1 {
		exceptions.PanicIfBadRequest(errors.New("stock kosong"), requestId, []string{"Mohon maaf stock sedang kosong"}, service.Logger)
	}

	// Produk belum pernah dimasukkan
	if cartProductExist.Id == "" {
		cartEntity := &entity.Cart{}
		cartEntity.Id = utilities.RandomUUID()
		cartEntity.IdUser = IdUser
		cartEntity.IdProduct = addProductToCartRequest.IdProduct
		cartEntity.Qty = cartEntity.Qty + 1
		cartEntity.CreatedAt = time.Now()
		cart, err := service.CartRepositoryInterface.AddProductToCart(service.DB, *cartEntity)
		exceptions.PanicIfError(err, requestId, service.Logger)
		addProductToCartResponse = response.ToAddProductToCartResponse(cart)
		return addProductToCartResponse
	} else {
		// Jika produk sudah ada
		cartEntity := &entity.Cart{}
		cartEntity.Id = cartProductExist.Id
		cartEntity.Qty = cartProductExist.Qty + addProductToCartRequest.Qty

		cart, err := service.CartRepositoryInterface.UpdateProductInCart(service.DB, cartProductExist.Id, *cartEntity)
		exceptions.PanicIfError(err, requestId, service.Logger)

		addProductToCartResponse = response.ToAddProductToCartResponse(cart)
		return addProductToCartResponse
	}
}
