package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/sirupsen/logrus"
	"github.com/tensuqiuwulu/be-service-teman-bunda/config"
	"github.com/tensuqiuwulu/be-service-teman-bunda/exceptions"
	"github.com/tensuqiuwulu/be-service-teman-bunda/models/entity"
	"github.com/tensuqiuwulu/be-service-teman-bunda/models/http/request"
	"github.com/tensuqiuwulu/be-service-teman-bunda/models/http/response"
	modelService "github.com/tensuqiuwulu/be-service-teman-bunda/models/service"
	"github.com/tensuqiuwulu/be-service-teman-bunda/repository/mysql"
	"github.com/tensuqiuwulu/be-service-teman-bunda/utilities"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
)

type OrderServiceInterface interface {
	CreateOrder(requestId string, idUser string, orderRequest *request.CreateOrderRequest) (orderResponse response.CreateOrderResponse)
	UpdateStatusOrder(requestId string, orderRequest *request.CallBackIpaymuRequest) (orderResponse response.UpdateOrderStatusResponse)
	FindOrderByUser(requestId string, idUser string, orderStatus string) (orderResponses []response.FindOrderByUserResponse)
	FindOrderById(requestId string, idOrder string) (orderResponse response.FindOrderByIdOrderResponse)
	CancelOrderById(requestId string, idOrder string) error
	CompleteOrderById(requestId string, idOrder string) error
	OrderCheckPayment(requestId string, idOrder string) (orderCheckPaymentResponse response.OrderCheckPayment)
}

type OrderServiceImplementation struct {
	ConfigurationWebserver                 config.Webserver
	DB                                     *gorm.DB
	ConfigJwt                              config.Jwt
	Validate                               *validator.Validate
	Logger                                 *logrus.Logger
	ConfigPayment                          config.Payment
	ConfigTelegram                         config.Telegram
	OrderRepositoryInterface               mysql.OrderRepositoryInterface
	CartRepositoryInterface                mysql.CartRepositoryInterface
	UserRepositoryInterface                mysql.UserRepositoryInterface
	OrderItemRepositoryInterface           mysql.OrderItemRepositoryInterface
	PaymentLogRepositoryInterface          mysql.PaymentLogRepositoryInterface
	BankTransferRepositoryInterface        mysql.BankTransferRepositoryInterface
	BankVaRepositoryInterface              mysql.BankVaRepositoryInterface
	ProductRepositoryInterface             mysql.ProductRepositoryInterface
	ProductStockHistoryRepositoryInterface mysql.ProductStockHistoryRepositoryInterface
	BalancePointRepositoryInterface        mysql.BalancePointRepositoryInterface
	BalancePointTxRepositoryInterface      mysql.BalancePointTxRepositoryInterface
	UserLevelRepositoryInterface           mysql.UserLevelMemberRepositoryInterface
	SettingRepositoryInterface             mysql.SettingRepositoryInterface
}

func NewOrderService(
	configurationWebserver config.Webserver,
	DB *gorm.DB,
	configJwt config.Jwt,
	validate *validator.Validate,
	logger *logrus.Logger,
	configPayment config.Payment,
	configTelegram config.Telegram,
	orderRepositoryInterface mysql.OrderRepositoryInterface,
	cartRepositoryInterface mysql.CartRepositoryInterface,
	userRepositoryInterface mysql.UserRepositoryInterface,
	orderItemRepositoryInterface mysql.OrderItemRepositoryInterface,
	paymentLogRepositoryInterface mysql.PaymentLogRepositoryInterface,
	bankTransferRepositoryInterface mysql.BankTransferRepositoryInterface,
	bankVaRepositoryInterface mysql.BankVaRepositoryInterface,
	productRepositoryInterface mysql.ProductRepositoryInterface,
	productStockHistoryRepositoryInterface mysql.ProductStockHistoryRepositoryInterface,
	balancePointRepositoryInterface mysql.BalancePointRepositoryInterface,
	balancePointTxRepositoryInterface mysql.BalancePointTxRepositoryInterface,
	userLevelMemberRepositoryInterface mysql.UserLevelMemberRepositoryInterface,
	settingRepositoryInterface mysql.SettingRepositoryInterface) OrderServiceInterface {
	return &OrderServiceImplementation{
		ConfigurationWebserver:                 configurationWebserver,
		DB:                                     DB,
		ConfigJwt:                              configJwt,
		Validate:                               validate,
		Logger:                                 logger,
		ConfigPayment:                          configPayment,
		ConfigTelegram:                         configTelegram,
		OrderRepositoryInterface:               orderRepositoryInterface,
		CartRepositoryInterface:                cartRepositoryInterface,
		UserRepositoryInterface:                userRepositoryInterface,
		OrderItemRepositoryInterface:           orderItemRepositoryInterface,
		PaymentLogRepositoryInterface:          paymentLogRepositoryInterface,
		BankTransferRepositoryInterface:        bankTransferRepositoryInterface,
		BankVaRepositoryInterface:              bankVaRepositoryInterface,
		ProductRepositoryInterface:             productRepositoryInterface,
		ProductStockHistoryRepositoryInterface: productStockHistoryRepositoryInterface,
		BalancePointRepositoryInterface:        balancePointRepositoryInterface,
		BalancePointTxRepositoryInterface:      balancePointTxRepositoryInterface,
		UserLevelRepositoryInterface:           userLevelMemberRepositoryInterface,
		SettingRepositoryInterface:             settingRepositoryInterface,
	}
}

func (service *OrderServiceImplementation) SendTelegram(numberOrder string, mssg string) {

	url, _ := url.Parse("https://api.telegram.org/bot" + service.ConfigTelegram.BotToken + "/sendMessage?chat_id=" + service.ConfigTelegram.ChatId + "&text=" + mssg + " " + numberOrder + "")

	req := &http.Request{
		Method: "POST",
		URL:    url,
		Header: map[string][]string{
			"Content-Type": {"application/json"},
		},
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatalf("An Error Occured %v", err)
		exceptions.PanicIfError(err, "", service.Logger)
	}
	defer resp.Body.Close()
}

func (service *OrderServiceImplementation) OrderCheckPayment(requestId string, idOrder string) (orderCheckPaymentResponse response.OrderCheckPayment) {
	order, err := service.OrderRepositoryInterface.FindOrderById(service.DB, idOrder)
	exceptions.PanicIfError(err, requestId, service.Logger)

	if order.PaymentMethod == "va" {
		bankVa, _ := service.BankVaRepositoryInterface.FindBankVaByBankCode(service.DB, order.PaymentChannel)
		orderCheckPaymentResponse = response.ToOrderCheckVaPaymentResponse(order, bankVa)
	} else if order.PaymentMethod == "trf" {
		bankTransfer, _ := service.BankTransferRepositoryInterface.FindBankTransferByBankCode(service.DB, order.PaymentChannel)
		orderCheckPaymentResponse = response.ToOrderCheckTransferPaymentResponse(order, bankTransfer)
	} else if order.PaymentMethod == "qris" {
		qris, _ := service.BankVaRepositoryInterface.FindBankVaByBankCode(service.DB, order.PaymentChannel)
		orderCheckPaymentResponse = response.ToOrderCheckVaPaymentResponse(order, qris)
	} else if order.PaymentMethod == "cc" {
		orderCheckPaymentResponse = response.ToOrderCheckCreditCardPaymentResponse(order)
	}

	return orderCheckPaymentResponse
}

func (service *OrderServiceImplementation) FindOrderByUser(requestId string, numberOrder string, orderStatus string) (orderResponses []response.FindOrderByUserResponse) {
	orders, err := service.OrderRepositoryInterface.FindOrderByUser(service.DB, numberOrder, orderStatus)
	exceptions.PanicIfError(err, requestId, service.Logger)
	orderResponses = response.ToFindOrderByUserResponse(orders)
	return orderResponses
}

func (service *OrderServiceImplementation) FindOrderById(requestId string, idOrder string) (orderResponse response.FindOrderByIdOrderResponse) {
	order, err := service.OrderRepositoryInterface.FindOrderById(service.DB, idOrder)
	exceptions.PanicIfError(err, requestId, service.Logger)

	orderItems, err := service.OrderItemRepositoryInterface.FindOrderItemsByIdOrder(service.DB, idOrder)
	exceptions.PanicIfError(err, requestId, service.Logger)

	orderResponse = response.ToFindOrderByIdOrder(order, orderItems)
	return orderResponse
}

func (service *OrderServiceImplementation) CompleteOrderById(requestId string, idOrder string) error {
	order, _ := service.OrderRepositoryInterface.FindOrderById(service.DB, idOrder)
	user, _ := service.UserRepositoryInterface.FindUserById(service.DB, order.IdUser)

	if order.OrderSatus == "Sampai Di Tujuan" {
		tx := service.DB.Begin()

		// Update order status
		orderEntity := &entity.Order{}
		orderEntity.OrderSatus = "Selesai"
		orderEntity.CompletedAt = null.NewTime(time.Now(), true)

		_, err := service.OrderRepositoryInterface.UpdateOrderStatus(tx, order.NumberOrder, *orderEntity)
		exceptions.PanicIfErrorWithRollback(err, requestId, []string{"Error update order"}, service.Logger, tx)

		if order.PaymentByCash != 0 {
			// hitung bonus point dari pembaran dengan uang
			bonusPoint := ((order.PaymentByCash - order.ShippingCost - order.PaymentFee) * user.UserLevelMember.BonusPercentage) / 100

			// Bonus pribadi dari perbelanjaan
			// Get bonus point from order order
			balancePoint, _ := service.BalancePointRepositoryInterface.FindBalancePointByIdUser(service.DB, order.IdUser)

			// Add to point history
			balancePointTxEntity := &entity.BalancePointTx{}
			balancePointTxEntity.Id = utilities.RandomUUID()
			balancePointTxEntity.IdBalancePoint = balancePoint.Id
			balancePointTxEntity.NoOrder = order.NumberOrder
			balancePointTxEntity.TxType = "debit"
			balancePointTxEntity.TxDate = time.Now()
			balancePointTxEntity.TxNominal = bonusPoint
			balancePointTxEntity.LastPointBalance = balancePoint.BalancePoints
			balancePointTxEntity.NewPointBalance = balancePoint.BalancePoints + bonusPoint
			balancePointTxEntity.CreatedDate = time.Now()
			balancePointTxEntity.Description = "Bonus Dari Pembelian"

			_, errCreateBalancePointTx := service.BalancePointTxRepositoryInterface.CreateBalancePointTx(tx, *balancePointTxEntity)
			exceptions.PanicIfErrorWithRollback(errCreateBalancePointTx, requestId, []string{"create balance point tx error"}, service.Logger, tx)
			// end bonus point untuk pribadi dari pembelian

			// update balance point
			balancePointEntity := &entity.BalancePoint{}
			balancePointEntity.BalancePoints = balancePoint.BalancePoints + bonusPoint

			_, errUpdateBalancePoint := service.BalancePointRepositoryInterface.UpdateBalancePoint(tx, balancePoint.IdUser, *balancePointEntity)
			exceptions.PanicIfErrorWithRollback(errUpdateBalancePoint, requestId, []string{"update balance point error"}, service.Logger, tx)

			// bonus point untuk referal
			if user.RegistrationReferalCode != "" {
				userReferal, _ := service.UserRepositoryInterface.FindUserByReferalCode(service.DB, user.RegistrationReferalCode)
				balancePointReferal, _ := service.BalancePointRepositoryInterface.FindBalancePointByIdUser(service.DB, userReferal.Id)

				// Add to point history
				balancePointTxEntityReferal := &entity.BalancePointTx{}
				balancePointTxEntityReferal.Id = utilities.RandomUUID()
				balancePointTxEntityReferal.IdBalancePoint = balancePointReferal.Id
				balancePointTxEntityReferal.NoOrder = order.NumberOrder
				balancePointTxEntityReferal.TxType = "referal"
				balancePointTxEntityReferal.TxDate = time.Now()
				balancePointTxEntityReferal.TxNominal = bonusPoint
				balancePointTxEntityReferal.LastPointBalance = balancePointReferal.BalancePoints
				balancePointTxEntityReferal.NewPointBalance = balancePointReferal.BalancePoints + bonusPoint
				balancePointTxEntityReferal.CreatedDate = time.Now()

				_, errCreateBalancePointTx := service.BalancePointTxRepositoryInterface.CreateBalancePointTx(tx, *balancePointTxEntityReferal)
				exceptions.PanicIfErrorWithRollback(errCreateBalancePointTx, requestId, []string{"create balance point tx error"}, service.Logger, tx)

				// update balance point
				balancePointEntityReferal := &entity.BalancePoint{}
				balancePointEntityReferal.BalancePoints = balancePointReferal.BalancePoints + bonusPoint

				_, errUpdateBalancePoint := service.BalancePointRepositoryInterface.UpdateBalancePoint(tx, balancePointReferal.IdUser, *balancePointEntityReferal)
				exceptions.PanicIfErrorWithRollback(errUpdateBalancePoint, requestId, []string{"update balance point error"}, service.Logger, tx)
			}
		}

		commit := tx.Commit()
		exceptions.PanicIfError(commit.Error, requestId, service.Logger)

		return err
	} else {
		return errors.New("sudah selesai")
	}
}

func (service *OrderServiceImplementation) CancelOrderById(requestId string, idOrder string) error {
	// get data order
	order, _ := service.OrderRepositoryInterface.FindOrderById(service.DB, idOrder)

	if order.OrderSatus == "Menunggu Pembayaran" {

		tx := service.DB.Begin()

		if order.PaymentByPoint != 0 {
			// get data balance point
			balancePoint, _ := service.BalancePointRepositoryInterface.FindBalancePointByIdUser(service.DB, order.IdUser)

			balancePointEntity := &entity.BalancePoint{}
			balancePointEntity.BalancePoints = balancePoint.BalancePoints + order.PaymentByPoint

			_, errUpdateBalancePoint := service.BalancePointRepositoryInterface.UpdateBalancePoint(tx, balancePoint.IdUser, *balancePointEntity)
			exceptions.PanicIfErrorWithRollback(errUpdateBalancePoint, requestId, []string{"update balance point error"}, service.Logger, tx)

			// Add to point history
			balancePointTxEntity := &entity.BalancePointTx{}
			balancePointTxEntity.Id = utilities.RandomUUID()
			balancePointTxEntity.IdBalancePoint = balancePoint.Id
			balancePointTxEntity.NoOrder = order.NumberOrder
			balancePointTxEntity.TxType = "debit"
			balancePointTxEntity.TxDate = time.Now()
			balancePointTxEntity.TxNominal = order.PaymentByPoint
			balancePointTxEntity.LastPointBalance = balancePoint.BalancePoints
			balancePointTxEntity.NewPointBalance = balancePoint.BalancePoints + order.PaymentByPoint
			balancePointTxEntity.CreatedDate = time.Now()
			balancePointTxEntity.Description = "Pengembalian Point"

			_, errCreateBalancePointTx := service.BalancePointTxRepositoryInterface.CreateBalancePointTx(tx, *balancePointTxEntity)
			exceptions.PanicIfErrorWithRollback(errCreateBalancePointTx, requestId, []string{"create balance point tx error"}, service.Logger, tx)
		}

		orderEntity := &entity.Order{}
		orderEntity.OrderSatus = "Dibatalkan"
		orderEntity.CanceledAt = null.NewTime(time.Now(), true)

		_, err := service.OrderRepositoryInterface.UpdateOrderStatus(tx, order.NumberOrder, *orderEntity)
		exceptions.PanicIfErrorWithRollback(err, requestId, []string{"Error update order"}, service.Logger, tx)

		commit := tx.Commit()
		exceptions.PanicIfError(commit.Error, requestId, service.Logger)

		return err

	} else {
		return errors.New("sudah dibatalkan")
	}
}

func (service *OrderServiceImplementation) UpdateStatusOrder(requestId string, paymentRequestCallback *request.CallBackIpaymuRequest) (orderResponse response.UpdateOrderStatusResponse) {
	// validate request

	request.ValidateCallBackIpaymuRequest(service.Validate, paymentRequestCallback, requestId, service.Logger)

	order, _ := service.OrderRepositoryInterface.FindOrderByNumberOrder(service.DB, paymentRequestCallback.ReferenceId)
	if order.Id == "" {
		err := errors.New("order not found")
		exceptions.PanicIfRecordNotFound(err, requestId, []string{"order not found"}, service.Logger)
	}

	// Cek payment status ke ipaymu
	var ipaymu_va = string(service.ConfigPayment.IpaymuVa)
	var ipaymu_key = string(service.ConfigPayment.IpaymuKey)

	url, _ := url.Parse(string(service.ConfigPayment.IpaymuTranscationUrl))
	postBody, _ := json.Marshal(map[string]interface{}{
		"transactionId": order.TrxId,
	})

	bodyHash := sha256.Sum256([]byte(postBody))
	bodyHashToString := hex.EncodeToString(bodyHash[:])
	stringToSign := "POST:" + ipaymu_va + ":" + strings.ToLower(string(bodyHashToString)) + ":" + ipaymu_key

	h := hmac.New(sha256.New, []byte(ipaymu_key))
	h.Write([]byte(stringToSign))
	signature := hex.EncodeToString(h.Sum(nil))

	reqBody := ioutil.NopCloser(strings.NewReader(string(postBody)))

	req := &http.Request{
		Method: "POST",
		URL:    url,
		Header: map[string][]string{
			"Content-Type": {"application/json"},
			"va":           {ipaymu_va},
			"signature":    {signature},
		},
		Body: reqBody,
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}
	defer resp.Body.Close()

	var dataPaymentStatus modelService.PaymentStatusResponse

	if err := json.NewDecoder(resp.Body).Decode(&dataPaymentStatus); err != nil {
		fmt.Println(err)
	}

	if order.PaymentStatus == "Sudah Dibayar" {
		orderResponse = response.ToUpdateOrderStatusResponse(order)
		return orderResponse
	} else {
		if order.OrderSatus == "Menunggu Pembayaran" {
			tx := service.DB.Begin()

			if dataPaymentStatus.Data.Status == 1 || dataPaymentStatus.Data.Status == 6 {
				// Update status order
				orderEntity := &entity.Order{}
				orderEntity.OrderSatus = "Menunggu Konfirmasi"
				orderEntity.PaymentStatus = "Sudah Dibayar"
				orderEntity.PaymentSuccessAt = null.NewTime(time.Now(), true)

				orderResult, err := service.OrderRepositoryInterface.UpdateOrderStatus(tx, paymentRequestCallback.ReferenceId, *orderEntity)
				exceptions.PanicIfErrorWithRollback(err, requestId, []string{"Error update order"}, service.Logger, tx)

				// Create response log
				paymentLogEntity := &entity.PaymentLog{}
				paymentLogEntity.Id = utilities.RandomUUID()
				paymentLogEntity.IdOrder = order.Id
				paymentLogEntity.NumberOrder = order.NumberOrder
				paymentLogEntity.TypeLog = "Respon Success Ipaymu"
				paymentLogEntity.PaymentMethod = order.PaymentMethod
				paymentLogEntity.PaymentChannel = order.PaymentChannel
				paymentLogEntity.Log = fmt.Sprintf("%+v\n", paymentRequestCallback)
				paymentLogEntity.CreatedAt = time.Now()

				// s := fmt.Sprintf("%+v\n", paymentRequestCallback)
				// fmt.Println(s)

				_, errCreateLog := service.PaymentLogRepositoryInterface.CreatePaymentLog(tx, *paymentLogEntity)
				exceptions.PanicIfErrorWithRollback(errCreateLog, requestId, []string{"Error create log"}, service.Logger, tx)

				//update product stock
				orderItems, _ := service.OrderItemRepositoryInterface.FindOrderItemsByIdOrder(service.DB, order.Id)
				for _, orderItem := range orderItems {
					productEntity := &entity.Product{}
					productEntityStockHistory := &entity.ProductStockHistory{}
					product, errFindProduct := service.ProductRepositoryInterface.FindProductById(tx, orderItem.IdProduct)
					exceptions.PanicIfErrorWithRollback(errFindProduct, requestId, []string{"product not found"}, service.Logger, tx)

					productEntityStockHistory.IdProduct = orderItem.IdProduct
					productEntityStockHistory.TxDate = time.Now()
					productEntityStockHistory.StockOpname = product.Stock
					productEntityStockHistory.StockOutQty = orderItem.Qty
					productEntityStockHistory.StockFinal = product.Stock - orderItem.Qty
					productEntityStockHistory.Description = "Pembelian " + order.NumberOrder
					productEntityStockHistory.CreatedAt = time.Now()
					_, errAddProductStockHistory := service.ProductStockHistoryRepositoryInterface.AddProductStockHistory(tx, *productEntityStockHistory)
					exceptions.PanicIfErrorWithRollback(errAddProductStockHistory, requestId, []string{"add stock history error"}, service.Logger, tx)

					productEntity.Stock = product.Stock - orderItem.Qty
					_, errUpdateProductStock := service.ProductRepositoryInterface.UpdateProductStock(tx, orderItem.IdProduct, *productEntity)
					exceptions.PanicIfErrorWithRollback(errUpdateProductStock, requestId, []string{"update stock error"}, service.Logger, tx)
				}

				commit := tx.Commit()
				exceptions.PanicIfError(commit.Error, requestId, service.Logger)

				runtime.GOMAXPROCS(1)
				// Send notif telegram
				go service.SendTelegram(order.NumberOrder, "Pembayaran Sukses (VA/QRIS)")

				// Send push notification
				user, _ := service.UserRepositoryInterface.FindUserById(service.DB, order.IdUser)
				go utilities.SendPushNotification(user.TokenDevice, &modelService.NotificationData{Title: "Pembayaran Berhasil", Body: "Selamat Pembayaran Anda Sudah Dikonfirmasi"})

				orderResponse = response.ToUpdateOrderStatusResponse(orderResult)
				return orderResponse
			} else if paymentRequestCallback.StatusCode == -2 {
				orderEntity := &entity.Order{}
				orderEntity.OrderSatus = "Dibatalkan"
				orderEntity.PaymentStatus = "Belum Dibayar"
				orderEntity.PaymentSuccessAt = null.NewTime(time.Now(), true)

				orderResult, err := service.OrderRepositoryInterface.UpdateOrderStatus(tx, paymentRequestCallback.ReferenceId, *orderEntity)
				exceptions.PanicIfErrorWithRollback(err, requestId, []string{"Error update order"}, service.Logger, tx)

				// Create response log
				paymentLogEntity := &entity.PaymentLog{}
				paymentLogEntity.Id = utilities.RandomUUID()
				paymentLogEntity.IdOrder = order.Id
				paymentLogEntity.NumberOrder = order.NumberOrder
				paymentLogEntity.TypeLog = "Expired"
				paymentLogEntity.PaymentMethod = order.PaymentMethod
				paymentLogEntity.PaymentChannel = order.PaymentChannel
				paymentLogEntity.Log = fmt.Sprintf("%+v\n", paymentRequestCallback)
				paymentLogEntity.CreatedAt = time.Now()

				// s := fmt.Sprintf("%+v\n", paymentRequestCallback)
				// fmt.Println(s)

				_, errCreateLog := service.PaymentLogRepositoryInterface.CreatePaymentLog(tx, *paymentLogEntity)
				exceptions.PanicIfErrorWithRollback(errCreateLog, requestId, []string{"Error create log"}, service.Logger, tx)

				commit := tx.Commit()
				exceptions.PanicIfError(commit.Error, requestId, service.Logger)
				orderResponse = response.ToUpdateOrderStatusResponse(orderResult)
				return orderResponse
			} else {
				fmt.Println("kode Status ipaymu = ", dataPaymentStatus.Data.Status)
				exceptions.PanicIfError(errors.New("hahaah"), requestId, service.Logger)
				return
			}
		} else {
			orderResponse = response.ToUpdateOrderStatusResponse(order)
			return orderResponse
		}
	}
}

func (service *OrderServiceImplementation) GenerateNumberOrder() (numberOrder string) {
	now := time.Now()
	orderEntity := &entity.Order{}
	for {
		rand.Seed(time.Now().Unix())
		charSet := "0123456789"
		var output strings.Builder
		length := 7

		for i := 0; i < length; i++ {
			random := rand.Intn(len(charSet))
			randomChar := charSet[random]
			output.WriteString(string(randomChar))
		}

		orderEntity.NumberOrder = "ORDER/" + now.Format("20060102") + "/" + output.String()

		// Check referal code if exist
		checkNumberOrder, _ := service.OrderRepositoryInterface.FindOrderByNumberOrder(service.DB, orderEntity.NumberOrder)
		if checkNumberOrder.Id == "" {
			break
		}
	}
	return orderEntity.NumberOrder
}

func (service *OrderServiceImplementation) CreateOrder(requestId string, idUser string, orderRequest *request.CreateOrderRequest) (orderResponse response.CreateOrderResponse) {

	// Validate request
	request.ValidateCreateOrderRequest(service.Validate, orderRequest, requestId, service.Logger)

	// Get data user
	user, _ := service.UserRepositoryInterface.FindUserById(service.DB, idUser)

	tx := service.DB.Begin()
	exceptions.PanicIfError(tx.Error, requestId, service.Logger)

	// Create Order
	orderEntity := &entity.Order{}
	orderEntity.Id = utilities.RandomUUID()
	orderEntity.IdUser = user.Id
	orderEntity.NumberOrder = service.GenerateNumberOrder()
	orderEntity.FullName = user.FamilyMembers.FullName
	orderEntity.Email = user.FamilyMembers.Email
	orderEntity.Address = orderRequest.Address
	orderEntity.Phone = user.FamilyMembers.Phone
	orderEntity.CourierNote = orderRequest.CourierNote
	orderEntity.OrderSatus = "Menunggu Pembayaran"
	orderEntity.OrderedAt = time.Now()
	orderEntity.PaymentMethod = orderRequest.PaymentMethod
	orderEntity.PaymentChannel = orderRequest.PaymentChannel
	orderEntity.PaymentStatus = "Belum Dibayar"
	orderEntity.PaymentByPoint = orderRequest.PaymentByPoint
	orderEntity.PaymentFee = orderRequest.PaymentFee
	if orderRequest.PaymentMethod != "trf" {
		orderEntity.PaymentByCash = (orderRequest.TotalBill + orderRequest.PaymentFee) - orderRequest.PaymentByPoint
	}
	orderEntity.ShippingCost = orderRequest.ShippingCost
	orderEntity.ShippingStatus = "Menunggu"

	// Jika berbelanja menggunakan point
	if orderRequest.PaymentByPoint > 0 {
		// get data balance point
		balancePoint, _ := service.BalancePointRepositoryInterface.FindBalancePointByIdUser(service.DB, orderEntity.IdUser)

		// update balance point
		balancePointEntity := &entity.BalancePoint{}
		balancePointEntity.BalancePoints = balancePoint.BalancePoints - orderEntity.PaymentByPoint

		_, errUpdateBalancePoint := service.BalancePointRepositoryInterface.UpdateBalancePoint(tx, balancePoint.IdUser, *balancePointEntity)
		exceptions.PanicIfErrorWithRollback(errUpdateBalancePoint, requestId, []string{"update balance point error"}, service.Logger, tx)

		// add balance point tx history
		balancePointTxEntity := &entity.BalancePointTx{}
		balancePointTxEntity.Id = utilities.RandomUUID()
		balancePointTxEntity.IdBalancePoint = balancePoint.Id
		balancePointTxEntity.NoOrder = orderEntity.NumberOrder
		balancePointTxEntity.TxType = "credit"
		balancePointTxEntity.TxDate = time.Now()
		balancePointTxEntity.TxNominal = orderEntity.PaymentByPoint
		balancePointTxEntity.LastPointBalance = balancePoint.BalancePoints
		balancePointTxEntity.NewPointBalance = balancePoint.BalancePoints - orderEntity.PaymentByPoint
		balancePointTxEntity.CreatedDate = time.Now()

		_, errCreateBalancePointTx := service.BalancePointTxRepositoryInterface.CreateBalancePointTx(tx, *balancePointTxEntity)
		exceptions.PanicIfErrorWithRollback(errCreateBalancePointTx, requestId, []string{"create balance point tx error"}, service.Logger, tx)
	}

	// Get data cart
	cartItems, _ := service.CartRepositoryInterface.FindCartByIdUser(service.DB, idUser)
	if len(cartItems) == 0 {
		// error jika tidak ada item di cart
		exceptions.PanicIfRecordNotFound(errors.New("data not found"), requestId, []string{"Keranjang Kosong"}, service.Logger)
	}

	// Create order items
	var totalPriceProduct float64
	var orderItems []entity.OrderItem
	var product []string
	var qty []int
	var price []float64
	var totalPrice float64
	var paymentPointForCC float64
	for _, cartItem := range cartItems {
		productResult, err := service.ProductRepositoryInterface.FindProductById(service.DB, cartItem.IdProduct)
		exceptions.PanicIfError(err, requestId, service.Logger)
		if productResult.Stock <= 0 {
			exceptions.PanicIfRecordNotFound(errors.New("stock product is empty"), requestId, []string{"Stock Produk Habis ", productResult.ProductName}, service.Logger)
		}
		orderItemEntity := &entity.OrderItem{}
		orderItemEntity.Id = utilities.RandomUUID()
		orderItemEntity.IdOrder = orderEntity.Id
		orderItemEntity.IdProduct = cartItem.IdProduct
		orderItemEntity.NoSku = cartItem.Product.NoSku
		orderItemEntity.ProductName = cartItem.Product.ProductName
		orderItemEntity.PictureUrl = cartItem.Product.PictureUrl
		orderItemEntity.Description = cartItem.Product.Description
		orderItemEntity.Weight = cartItem.Product.Weight
		orderItemEntity.Volume = cartItem.Product.Volume
		orderItemEntity.Qty = cartItem.Qty
		orderItemEntity.FlagPromo = cartItem.Product.ProductDiscount.FlagPromo
		orderItemEntity.Thumbnail = cartItem.Product.Thumbnail
		orderItemEntity.PriceBeforeDiscount = cartItem.Product.Price
		orderItemEntity.PriceAfterDiscount = cartItem.Product.ProductDiscount.Nominal
		if cartItem.Product.ProductDiscount.FlagPromo == "true" {
			orderItemEntity.Price = cartItem.Product.ProductDiscount.Nominal
			totalPriceProduct = cartItem.Product.ProductDiscount.Nominal
		} else {
			orderItemEntity.Price = cartItem.Product.Price
			totalPriceProduct = cartItem.Product.Price
		}

		orderItemEntity.TotalPrice = totalPriceProduct * (float64(cartItem.Qty))
		orderItemEntity.CreatedAt = time.Now()
		totalPrice = totalPrice + orderItemEntity.TotalPrice
		orderItems = append(orderItems, *orderItemEntity)
		if orderRequest.PaymentMethod == "cc" {
			product = append(product, orderItemEntity.ProductName)
			qty = append(qty, orderItemEntity.Qty)
			price = append(price, orderItemEntity.Price)
			paymentPointForCC = orderRequest.PaymentByPoint * (-1)
		}
	}

	orderEntity.TotalBill = totalPrice + orderRequest.ShippingCost

	fmt.Println("Total Bill = ", totalPrice+orderRequest.ShippingCost-orderRequest.PaymentByPoint)
	fmt.Println("Request Total Bill = ", orderRequest.TotalBill-orderRequest.PaymentByPoint)

	fmt.Println("Total Bill + Fee", totalPrice+orderRequest.ShippingCost+orderRequest.PaymentFee-orderRequest.PaymentByPoint)
	fmt.Println("Request Bill + Fee", (orderRequest.TotalBill+orderRequest.PaymentFee)-orderRequest.PaymentByPoint)

	if ((totalPrice + orderRequest.ShippingCost) - orderRequest.PaymentByPoint) != orderRequest.TotalBill-orderRequest.PaymentByPoint {
		fmt.Print("harga tidak sama, user name = ", user.FamilyMembers.FullName)
		exceptions.PanicIfErrorWithRollback(errors.New("price not same 1"), requestId, []string{"price not same 1"}, service.Logger, tx)
	}

	if (((totalPrice + orderRequest.ShippingCost) + orderRequest.PaymentFee) - orderRequest.PaymentByPoint) != (orderRequest.TotalBill+orderRequest.PaymentFee)-orderRequest.PaymentByPoint {
		fmt.Print("harga tidak sama 2, user name = ", user.FamilyMembers.FullName)
		exceptions.PanicIfErrorWithRollback(errors.New("price not same 2"), requestId, []string{"price not same 2"}, service.Logger, tx)
	}

	errCreateOrderItem := service.OrderItemRepositoryInterface.CreateOrderItems(tx, orderItems)
	exceptions.PanicIfErrorWithRollback(errCreateOrderItem, requestId, []string{"Error create order"}, service.Logger, tx)

	// Pilih metode pembayaran
	switch orderRequest.PaymentMethod {
	// Credit Card
	case "cc":
		var ipaymu_va = string(service.ConfigPayment.IpaymuVa)
		var ipaymu_key = string(service.ConfigPayment.IpaymuKey)

		// tambahkan ongkos kirim
		product = append(product, "Shipping Cost", "Payment Fee", "Payment Point")
		qty = append(qty, 1, 1, 1)
		price = append(price, orderRequest.ShippingCost, orderRequest.PaymentFee, paymentPointForCC)

		url, _ := url.Parse(string(service.ConfigPayment.IpaymuSnapUrl))

		postBody, _ := json.Marshal(map[string]interface{}{
			"product":       product,
			"qty":           qty,
			"price":         price,
			"returnUrl":     string(service.ConfigPayment.IpaymuThankYouPage),
			"cancelUrl":     string(service.ConfigPayment.IpaymuCancelUrl),
			"notifyUrl":     string(service.ConfigPayment.IpaymuCallbackUrl),
			"referenceId":   orderEntity.NumberOrder,
			"buyerName":     user.FamilyMembers.FullName,
			"buyerEmail":    user.FamilyMembers.Email,
			"buyerPhone":    user.FamilyMembers.Phone,
			"paymentMethod": orderRequest.PaymentMethod,
		})

		bodyHash := sha256.Sum256([]byte(postBody))
		bodyHashToString := hex.EncodeToString(bodyHash[:])
		stringToSign := "POST:" + ipaymu_va + ":" + strings.ToLower(string(bodyHashToString)) + ":" + ipaymu_key

		h := hmac.New(sha256.New, []byte(ipaymu_key))
		h.Write([]byte(stringToSign))
		signature := hex.EncodeToString(h.Sum(nil))

		reqBody := io.NopCloser(strings.NewReader(string(postBody)))

		req := &http.Request{
			Method: "POST",
			URL:    url,
			Header: map[string][]string{
				"Content-Type": {"application/json"},
				"va":           {ipaymu_va},
				"signature":    {signature},
			},
			Body: reqBody,
		}

		reqDump, _ := httputil.DumpRequestOut(req, true)
		fmt.Printf("REQUEST:\n%s", string(reqDump))

		resp, err := http.DefaultClient.Do(req)

		if err != nil {
			log.Fatalf("An Error Occured %v", err)
			exceptions.PanicIfError(err, requestId, service.Logger)
		}
		defer resp.Body.Close()

		var dataResponseIpaymu modelService.PaymentCreditCardResponse

		if err := json.NewDecoder(resp.Body).Decode(&dataResponseIpaymu); err != nil {
			fmt.Println(err)
			exceptions.PanicIfError(err, requestId, service.Logger)
		}

		if dataResponseIpaymu.Status != 200 {
			fmt.Println("LOG RESPONSE IPAYMU = ", dataResponseIpaymu)
			exceptions.PanicIfErrorWithRollback(errors.New("error response ipaymu"), requestId, []string{"Error response ipaymu"}, service.Logger, tx)
			return
		} else if dataResponseIpaymu.Status == 200 {
			// make log
			paymentLogEntity := &entity.PaymentLog{}
			paymentLogEntity.Id = utilities.RandomUUID()
			paymentLogEntity.IdOrder = orderEntity.Id
			paymentLogEntity.NumberOrder = orderEntity.NumberOrder
			paymentLogEntity.TypeLog = "Create Trx Ipaymu"
			paymentLogEntity.PaymentMethod = orderRequest.PaymentMethod
			paymentLogEntity.PaymentChannel = orderRequest.PaymentChannel
			paymentLogEntity.Log = fmt.Sprintf("%+v\n", dataResponseIpaymu)
			paymentLogEntity.CreatedAt = time.Now()

			_, err := service.PaymentLogRepositoryInterface.CreatePaymentLog(tx, *paymentLogEntity)
			exceptions.PanicIfErrorWithRollback(err, requestId, []string{"Error create log"}, service.Logger, tx)

			orderEntity.PaymentNo = dataResponseIpaymu.Data.Url
			orderEntity.PaymentName = "Credit Card"
			orderEntity.PaymentDueDate = null.NewTime(time.Now().Add(time.Hour*24), true)
			order, errUpdateOrderPayment := service.OrderRepositoryInterface.CreateOrder(tx, *orderEntity)
			exceptions.PanicIfErrorWithRollback(errUpdateOrderPayment, requestId, []string{"Error update order"}, service.Logger, tx)

			// delete data item in cart
			errDelete := service.CartRepositoryInterface.DeleteAllProductInCartByIdUser(tx, idUser, cartItems)
			exceptions.PanicIfErrorWithRollback(errDelete, requestId, []string{"Error delete in cart"}, service.Logger, tx)

			commit := tx.Commit()
			exceptions.PanicIfError(commit.Error, requestId, service.Logger)

			runtime.GOMAXPROCS(1)
			go service.SendTelegram(orderEntity.NumberOrder, "Ada Orderan Masuk (CC)")

			orderResponse = response.ToCreateOrderCreditCardResponse(order, dataResponseIpaymu)
			return orderResponse
		}

	// VA, QRIS
	case "va", "qris":
		// Send request to ipaymu
		var ipaymu_va = string(service.ConfigPayment.IpaymuVa)
		var ipaymu_key = string(service.ConfigPayment.IpaymuKey)

		url, _ := url.Parse(string(service.ConfigPayment.IpaymuUrl))

		postBody, _ := json.Marshal(map[string]interface{}{
			"name":           user.FamilyMembers.FullName,
			"phone":          user.FamilyMembers.Phone,
			"email":          user.FamilyMembers.Email,
			"amount":         orderEntity.PaymentByCash,
			"notifyUrl":      string(service.ConfigPayment.IpaymuCallbackUrl),
			"expired":        24,
			"expiredType":    "hours",
			"referenceId":    orderEntity.NumberOrder,
			"paymentMethod":  orderRequest.PaymentMethod,
			"paymentChannel": orderRequest.PaymentChannel,
		})

		bodyHash := sha256.Sum256([]byte(postBody))
		bodyHashToString := hex.EncodeToString(bodyHash[:])
		stringToSign := "POST:" + ipaymu_va + ":" + strings.ToLower(string(bodyHashToString)) + ":" + ipaymu_key

		h := hmac.New(sha256.New, []byte(ipaymu_key))
		h.Write([]byte(stringToSign))
		signature := hex.EncodeToString(h.Sum(nil))

		reqBody := ioutil.NopCloser(strings.NewReader(string(postBody)))

		req := &http.Request{
			Method: "POST",
			URL:    url,
			Header: map[string][]string{
				"Content-Type": {"application/json"},
				"va":           {ipaymu_va},
				"signature":    {signature},
			},
			Body: reqBody,
		}

		reqDump, _ := httputil.DumpRequestOut(req, true)
		fmt.Printf("REQUEST:\n%s", string(reqDump))

		resp, err := http.DefaultClient.Do(req)

		if err != nil {
			log.Fatalf("An Error Occured %v", err)
			exceptions.PanicIfError(err, requestId, service.Logger)
		}
		defer resp.Body.Close()

		// get data bank
		bankVa, _ := service.BankVaRepositoryInterface.FindBankVaByBankCode(service.DB, orderRequest.PaymentChannel)

		var dataResponseIpaymu modelService.PaymentResponse

		if err := json.NewDecoder(resp.Body).Decode(&dataResponseIpaymu); err != nil {
			fmt.Println(err)
			exceptions.PanicIfError(err, requestId, service.Logger)
		}
		// ss := fmt.Sprintf("%+v\n", resp.Body)
		// fmt.Println("log = ", ss)

		if dataResponseIpaymu.Status != 200 {
			fmt.Println("LOG RESPONSE IPAYMU = ", dataResponseIpaymu)
			exceptions.PanicIfErrorWithRollback(errors.New("error response ipaymu"), requestId, []string{"Error response ipaymu"}, service.Logger, tx)
			return
		} else if dataResponseIpaymu.Status == 200 {
			// make log
			paymentLogEntity := &entity.PaymentLog{}
			paymentLogEntity.Id = utilities.RandomUUID()
			paymentLogEntity.IdOrder = orderEntity.Id
			paymentLogEntity.NumberOrder = orderEntity.NumberOrder
			paymentLogEntity.TypeLog = "Create Trx Ipaymu"
			paymentLogEntity.PaymentMethod = orderRequest.PaymentMethod
			paymentLogEntity.PaymentChannel = orderRequest.PaymentChannel
			paymentLogEntity.Log = fmt.Sprintf("%+v\n", dataResponseIpaymu)
			paymentLogEntity.CreatedAt = time.Now()

			_, err := service.PaymentLogRepositoryInterface.CreatePaymentLog(tx, *paymentLogEntity)
			exceptions.PanicIfErrorWithRollback(err, requestId, []string{"Error create log"}, service.Logger, tx)

			orderEntity.PaymentNo = dataResponseIpaymu.Data.PaymentNo
			orderEntity.PaymentName = dataResponseIpaymu.Data.PaymentName
			orderEntity.TrxId = dataResponseIpaymu.Data.TransactionId
			paymentDueDate, _ := time.Parse("2006-01-02 15:04:05", dataResponseIpaymu.Data.Expired)
			orderEntity.PaymentDueDate = null.NewTime(paymentDueDate, true)
			order, errUpdateOrderPayment := service.OrderRepositoryInterface.CreateOrder(tx, *orderEntity)
			exceptions.PanicIfErrorWithRollback(errUpdateOrderPayment, requestId, []string{"Error update order"}, service.Logger, tx)

			// delete data item in cart
			errDelete := service.CartRepositoryInterface.DeleteAllProductInCartByIdUser(tx, idUser, cartItems)
			exceptions.PanicIfErrorWithRollback(errDelete, requestId, []string{"Error delete in cart"}, service.Logger, tx)

			commit := tx.Commit()
			exceptions.PanicIfError(commit.Error, requestId, service.Logger)

			runtime.GOMAXPROCS(1)
			go service.SendTelegram(order.NumberOrder, "Ada Orderan Masuk (VA/QRIS)")

			orderResponse = response.ToCreateOrderVaResponse(order, dataResponseIpaymu.Data.TransactionId, dataResponseIpaymu, bankVa)
			return orderResponse
		}

	// TRANSFER
	case "trf":
		// Get data bank by code
		bankTransfer, _ := service.BankTransferRepositoryInterface.FindBankTransferByBankCode(service.DB, orderRequest.PaymentChannel)
		if bankTransfer.Id == "" {
			exceptions.PanicIfErrorWithRollback(errors.New("bank not found"), requestId, []string{"Bank not found"}, service.Logger, tx)
		}

		payment := &modelService.PaymentResponse{}

		// buat 3 nomor acak
		rand.Seed(time.Now().UnixNano())
		min := 111
		max := 299
		rand3Number := rand.Intn(max-min+1) + min

		min2 := 11
		max2 := 99
		rand2Number := rand.Intn(max2-min2+1) + min

		sisaPembagi := math.Mod(orderEntity.TotalBill, 1000)
		if sisaPembagi < 100 {
			payment.Data.Total = orderEntity.TotalBill + float64(rand3Number)
		} else if sisaPembagi >= 100 {
			payment.Data.Total = orderEntity.TotalBill + float64(rand2Number)
		}

		payment.Data.PaymentName = bankTransfer.BankName
		payment.Data.PaymentNo = bankTransfer.NoAccount
		payment.Data.ReferenceId = orderEntity.NumberOrder

		orderEntity.PaymentNo = bankTransfer.NoAccount
		orderEntity.PaymentName = bankTransfer.BankName
		orderEntity.PaymentByCash = payment.Data.Total - orderRequest.PaymentByPoint
		orderEntity.PaymentDueDate = null.NewTime(time.Now().Add(time.Hour*24), true)
		payment.Data.Expired = orderEntity.PaymentDueDate.Time.Format("2006-01-02 15:04:05")

		order, errUpdateOrderPayment := service.OrderRepositoryInterface.CreateOrder(tx, *orderEntity)
		exceptions.PanicIfErrorWithRollback(errUpdateOrderPayment, requestId, []string{"Error update order"}, service.Logger, tx)

		// delete data item in cart
		errDelete := service.CartRepositoryInterface.DeleteAllProductInCartByIdUser(tx, idUser, cartItems)
		exceptions.PanicIfErrorWithRollback(errDelete, requestId, []string{"Error delete in cart"}, service.Logger, tx)

		commit := tx.Commit()
		exceptions.PanicIfError(commit.Error, requestId, service.Logger)

		runtime.GOMAXPROCS(1)
		go service.SendTelegram(payment.Data.ReferenceId, "Ada Orderan Masuk (TRF)")

		orderResponse = response.ToCreateOrderTransferResponse(order, *payment, bankTransfer)
		return orderResponse

	// Cash On Delivery
	case "cod":
		orderEntity.OrderSatus = "Menunggu Konfirmasi"
		order, errUpdateOrderPayment := service.OrderRepositoryInterface.CreateOrder(tx, *orderEntity)
		exceptions.PanicIfErrorWithRollback(errUpdateOrderPayment, requestId, []string{"Error update order"}, service.Logger, tx)

		// delete data item in cart
		errDelete := service.CartRepositoryInterface.DeleteAllProductInCartByIdUser(tx, idUser, cartItems)
		exceptions.PanicIfErrorWithRollback(errDelete, requestId, []string{"Error delete in cart"}, service.Logger, tx)

		commit := tx.Commit()
		exceptions.PanicIfError(commit.Error, requestId, service.Logger)

		runtime.GOMAXPROCS(1)
		go service.SendTelegram(order.NumberOrder, "Ada Orderan Masuk (COD)")

		orderResponse = response.ToCreateOrderCodResponse(order)
		return orderResponse

	// Point
	case "point":
		orderEntity.OrderSatus = "Menunggu Konfirmasi"
		orderEntity.PaymentStatus = "Sudah Dibayar"
		orderEntity.PaymentMethod = orderRequest.PaymentMethod
		orderEntity.PaymentChannel = orderRequest.PaymentChannel
		orderEntity.PaymentSuccessAt = null.NewTime(time.Now(), true)
		order, errUpdateOrderPayment := service.OrderRepositoryInterface.CreateOrder(tx, *orderEntity)
		exceptions.PanicIfErrorWithRollback(errUpdateOrderPayment, requestId, []string{"Error update order"}, service.Logger, tx)

		//update product stock
		// orderItems, _ := service.OrderItemRepositoryInterface.FindOrderItemsByIdOrder(service.DB, orderEntity.Id)
		// log.Println("orderItems = ", orderItems)
		for _, orderItem := range cartItems {
			product, errFindProduct := service.ProductRepositoryInterface.FindProductById(tx, orderItem.IdProduct)
			exceptions.PanicIfErrorWithRollback(errFindProduct, requestId, []string{"product not found"}, service.Logger, tx)

			productEntity := &entity.Product{}
			productEntity.Stock = product.Stock - orderItem.Qty
			_, errUpdateProductStock := service.ProductRepositoryInterface.UpdateProductStock(tx, orderItem.IdProduct, *productEntity)
			exceptions.PanicIfErrorWithRollback(errUpdateProductStock, requestId, []string{"update stock error"}, service.Logger, tx)

			productEntityStockHistory := &entity.ProductStockHistory{}
			productEntityStockHistory.IdProduct = orderItem.IdProduct
			productEntityStockHistory.TxDate = time.Now()
			productEntityStockHistory.StockOpname = product.Stock
			productEntityStockHistory.StockOutQty = orderItem.Qty
			productEntityStockHistory.StockFinal = product.Stock - orderItem.Qty
			productEntityStockHistory.Description = "Pembelian " + order.NumberOrder
			productEntityStockHistory.CreatedAt = time.Now()
			_, errAddProductStockHistory := service.ProductStockHistoryRepositoryInterface.AddProductStockHistory(tx, *productEntityStockHistory)
			exceptions.PanicIfErrorWithRollback(errAddProductStockHistory, requestId, []string{"add stock history error"}, service.Logger, tx)
		}

		// delete da	ta item in cart
		errDelete := service.CartRepositoryInterface.DeleteAllProductInCartByIdUser(tx, idUser, cartItems)
		exceptions.PanicIfErrorWithRollback(errDelete, requestId, []string{"Error delete in cart"}, service.Logger, tx)

		commit := tx.Commit()
		exceptions.PanicIfError(commit.Error, requestId, service.Logger)

		runtime.GOMAXPROCS(1)
		go service.SendTelegram(order.NumberOrder, "Ada Orderan Masuk (Point)")

		orderResponse = response.ToCreateOrderFullPointResponse(order)
		return orderResponse
	default:
		exceptions.PanicIfErrorWithRollback(errors.New("payment method not found"), requestId, []string{"payment method not found"}, service.Logger, tx)
		return
	}
	return
}
