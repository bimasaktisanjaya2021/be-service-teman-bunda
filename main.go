package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tensuqiuwulu/be-service-teman-bunda/config"
	"github.com/tensuqiuwulu/be-service-teman-bunda/controllers"
	"github.com/tensuqiuwulu/be-service-teman-bunda/exceptions"
	"github.com/tensuqiuwulu/be-service-teman-bunda/repository/mysql"
	"github.com/tensuqiuwulu/be-service-teman-bunda/routes"
	"github.com/tensuqiuwulu/be-service-teman-bunda/services"
	"github.com/tensuqiuwulu/be-service-teman-bunda/utilities"
)

func main() {

	appConfig := config.GetConfig()

	mysqlDBConnection := mysql.NewDatabaseConnection(&appConfig.Database)

	// Timezone
	location, err := time.LoadLocation(appConfig.Timezone.Timezone)
	time.Local = location
	fmt.Println("Location:", location, err)

	// App
	fmt.Println("Server App : ", string(config.GetConfig().Application.Server))

	// Logger
	logrusLogger := utilities.NewLogger(appConfig.Log)

	// Validate
	validate := validator.New()

	e := echo.New()

	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Skipper:      nil,
		ErrorMessage: "Request Timeout",
		Timeout:      10 * time.Second,
	}))
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		DisableStackAll:   true,
		DisablePrintStack: true,
	}))
	e.HTTPErrorHandler = exceptions.ErrorHandler
	e.Use(middleware.RequestID())
	// e.IPExtractor = echo.ExtractIPDirect()

	e.Use(middleware.CORS())

	// Otp Manager
	otpManagerRepository := mysql.NewOtpManagerRepository(&appConfig.Database)

	// User Address Repository
	userShippingAddressRepository := mysql.NewUserShippingAddressRepository(&appConfig.Database)

	// Setting Repository
	settingsRepository := mysql.NewSettingRepository(&appConfig.Database)

	// Provinsi Repository
	provinsiRepository := mysql.NewProvinsiRepository(&appConfig.Database)

	// Kabupaten Repository
	kabupatenRepository := mysql.NewKabupatenRepository(&appConfig.Database)

	// Kecamatan Repository
	kecamatanRepository := mysql.NewKecamatanRepository(&appConfig.Database)

	// Kelurahan Repository
	kelurahanRepository := mysql.NewKelurahanRepository(&appConfig.Database)

	// Family Repository
	familyRepository := mysql.NewFamilyRepository(&appConfig.Database)

	// Family Members Repository
	familyMembersRepository := mysql.NewFamilyMembersRepository(&appConfig.Database)

	// Shipping Repository
	shippingRepository := mysql.NewShippingRepository(&appConfig.Database)

	// Cart Repository
	cartRepository := mysql.NewCartRepository(&appConfig.Database)

	// Order Repository
	orderRepository := mysql.NewOrderRepository(&appConfig.Database)

	// Balance Point Repository
	balancePointRepository := mysql.NewBalancePointRepository(&appConfig.Database)

	// Balance Point Tx
	balancePointTxRepository := mysql.NewBalancePointTxRepository(&appConfig.Database)

	// User Repository
	userRepository := mysql.NewUserRepository(&appConfig.Database)

	// Product Repository
	productRepository := mysql.NewProductRepository(&appConfig.Database)

	// Order Item Repository
	orderItemRepository := mysql.NewOrderItemRepository(&appConfig.Database)

	// Bank Transfer Repository
	bankTransferRepository := mysql.NewBankTransferRepository(&appConfig.Database)

	// Bank Va Repository
	bankVaRepository := mysql.NewBankVaRepository(&appConfig.Database)

	// Payment Log Repository
	paymentLogRepository := mysql.NewPaymentLogRepository(&appConfig.Database)

	// Product Stock History Repository
	productStockHistoryRepository := mysql.NewProductStockHistoryRepository(&appConfig.Database)

	// User Level Member Repository
	userLevelMemberRepository := mysql.NewUserLevelMemberRepository(&appConfig.Database)

	// Setting Repository
	bannerRepository := mysql.NewBannerRepository(&appConfig.Database)

	// Product Brand Repository
	productBrandRepository := mysql.NewProductBrandRepository(&appConfig.Database)

	// Setting Service
	settingService := services.NewSettingService(
		appConfig.Webserver,
		mysqlDBConnection,
		logrusLogger,
		settingsRepository)

	// User Address Service
	userShippingAddressService := services.NewUserShippingAddressService(
		appConfig.Webserver,
		mysqlDBConnection,
		validate,
		logrusLogger,
		userShippingAddressRepository)

	// Provinsi Service
	provinsiService := services.NewProvinsiService(
		appConfig.Webserver,
		mysqlDBConnection,
		logrusLogger,
		provinsiRepository)

	// Kabupaten Service
	kabupatenService := services.NewKabupatenService(
		appConfig.Webserver,
		mysqlDBConnection,
		logrusLogger,
		kabupatenRepository)

	// Kecamatan
	kecamatanService := services.NewKecamatanService(
		appConfig.Webserver,
		mysqlDBConnection,
		logrusLogger,
		kecamatanRepository)

	// Kelurahan Service
	kelurahanService := services.NewKelurahanService(
		appConfig.Webserver,
		mysqlDBConnection,
		logrusLogger,
		kelurahanRepository)

	// Shipping Service
	shippingService := services.NewShippingService(
		appConfig.Webserver,
		mysqlDBConnection,
		validate,
		logrusLogger,
		shippingRepository)

	// Cart Service
	cartService := services.NewCartService(
		appConfig.Webserver,
		mysqlDBConnection,
		validate,
		logrusLogger,
		cartRepository,
		shippingRepository,
		productRepository,
		settingsRepository,
	)

	// Balance Point Service
	balancePointService := services.NewBalancePointService(
		appConfig.Webserver,
		mysqlDBConnection,
		logrusLogger,
		balancePointRepository,
		settingsRepository,
		orderRepository)

	// Balance Point Tx Service
	balancePointTxService := services.NewBalancePointTxService(
		appConfig.Webserver,
		mysqlDBConnection,
		logrusLogger,
		balancePointTxRepository,
		balancePointRepository)

	// User Service
	userService := services.NewUserService(
		appConfig.Webserver,
		mysqlDBConnection,
		appConfig.Jwt,
		validate,
		logrusLogger,
		appConfig.Email,
		userRepository,
		provinsiRepository,
		familyRepository,
		familyMembersRepository,
		balancePointRepository,
		balancePointTxRepository,
		userShippingAddressRepository)

	// Auth Service
	authService := services.NewAuthService(
		appConfig.Webserver,
		mysqlDBConnection,
		appConfig.Jwt,
		validate,
		logrusLogger,
		userRepository,
		settingsRepository,
		otpManagerRepository)

	// Product Service
	productService := services.NewProductService(
		appConfig.Webserver,
		mysqlDBConnection,
		logrusLogger,
		productRepository,
		appConfig.Payment)

	// Order Service
	orderService := services.NewOrderService(
		appConfig.Webserver,
		mysqlDBConnection,
		appConfig.Jwt,
		validate,
		logrusLogger,
		appConfig.Payment,
		appConfig.Telegram,
		orderRepository,
		cartRepository,
		userRepository,
		orderItemRepository,
		paymentLogRepository,
		bankTransferRepository,
		bankVaRepository,
		productRepository,
		productStockHistoryRepository,
		balancePointRepository,
		balancePointTxRepository,
		userLevelMemberRepository,
		settingsRepository)

	// Payment Channel Service
	paymentChannelService := services.NewPaymentChannelService(
		appConfig.Webserver,
		mysqlDBConnection,
		logrusLogger,
		bankVaRepository,
		bankTransferRepository)

	// Banner Service
	bannerService := services.NewBannerService(
		appConfig.Webserver,
		mysqlDBConnection,
		logrusLogger,
		bannerRepository)

	// Product Brand Service
	productBrandService := services.NewProductBrandService(
		appConfig.Webserver,
		mysqlDBConnection,
		logrusLogger,
		productBrandRepository)

	// Payment Service
	paymentService := services.NewPaymentService(
		appConfig.Webserver,
		mysqlDBConnection,
		validate,
		logrusLogger,
		appConfig.Payment,
		orderRepository,
		orderItemRepository,
		productRepository,
		productStockHistoryRepository,
		paymentLogRepository)

	// Setting Controller
	settingController := controllers.NewSettingController(appConfig.Webserver, settingService)
	routes.SettingRoute(e, appConfig.Webserver, appConfig.Jwt, settingController)

	// User Address Controller
	userShippingAddressController := controllers.NewUserShippingAddressController(appConfig.Webserver, userShippingAddressService)

	// Provinsi Controller
	provinsiController := controllers.NewProvinsiController(appConfig.Webserver, provinsiService)
	routes.ProvinsiRoute(e, appConfig.Webserver, appConfig.Jwt, provinsiController)

	// Kabupaten Controller
	kabupatenController := controllers.NewKabupatenController(appConfig.Webserver, kabupatenService)
	routes.KabupatenRoute(e, appConfig.Webserver, kabupatenController)

	// Kecamatan Controller
	kecamatanController := controllers.NewKecamatanController(appConfig.Webserver, kecamatanService)
	routes.KecamatanRoute(e, appConfig.Webserver, kecamatanController)

	// Kelurahan Controller
	kelurahanController := controllers.NewKelurahanController(appConfig.Webserver, kelurahanService)
	routes.KelurahanRoute(e, appConfig.Webserver, kelurahanController)

	// Shipping Controller
	shippingController := controllers.NewShippingController(appConfig.Webserver, shippingService)
	routes.ShippingRoute(e, appConfig.Webserver, appConfig.Jwt, shippingController)

	// Cart Controller
	cartController := controllers.NewCartController(appConfig.Webserver, cartService)
	routes.CartRoute(e, appConfig.Webserver, appConfig.Jwt, cartController)

	// Balance Point Controller
	balancePointController := controllers.NewBalancePointController(appConfig.Webserver, logrusLogger, balancePointService)
	routes.BalancePointRoute(e, appConfig.Webserver, appConfig.Jwt, balancePointController)

	// Balance Point Tx Controller
	balancePointTxController := controllers.NewBalancePointTxController(appConfig.Webserver, logrusLogger, balancePointTxService)
	routes.BalancePointTxRoute(e, appConfig.Webserver, appConfig.Jwt, balancePointTxController)

	// User Controller
	userController := controllers.NewUserController(appConfig.Webserver, logrusLogger, userService)
	routes.UserRoute(e, appConfig.Webserver, appConfig.Jwt, userController, userShippingAddressController)
	routes.VerifyEmailRoute(e, appConfig.Webserver, appConfig.Jwt, userController)

	// Auth Controller
	authController := controllers.NewAuthController(appConfig.Webserver, logrusLogger, authService)
	routes.AuthRoute(e, appConfig.Webserver, appConfig.Jwt, authController)

	// Product Controller
	productController := controllers.NewProductController(appConfig.Webserver, productService)
	routes.ProductRoute(e, appConfig.Webserver, appConfig.Jwt, productController)

	// Order Controller
	orderController := controllers.NewOrderController(appConfig.Webserver, logrusLogger, orderService)
	routes.OrderRoute(e, appConfig.Webserver, appConfig.Jwt, orderController)

	// Payment Channel Controller
	paymentChannelController := controllers.NewPaymentChannelController(appConfig.Webserver, logrusLogger, paymentChannelService)
	routes.PaymentChannelRoute(e, appConfig.Webserver, appConfig.Jwt, paymentChannelController)

	// Payment Controller
	paymentController := controllers.NewPaymentController(appConfig.Webserver, logrusLogger, paymentService)
	routes.PaymentRoute(e, appConfig.Webserver, appConfig.Jwt, paymentController)

	// Banner Controller
	bannerController := controllers.NewBannerController(appConfig.Webserver, bannerService)
	routes.BannerRoute(e, appConfig.Webserver, appConfig.Jwt, bannerController)

	// Product Brand Controller
	productBrandController := controllers.NewProductBrandController(appConfig.Webserver, productBrandService)
	routes.ProductBrandRoute(e, appConfig.Webserver, appConfig.Jwt, productBrandController)

	// Main Controller
	mainController := controllers.NewMainController(appConfig.Webserver)
	routes.MainRoute(e, appConfig.Webserver, mainController)

	// Careful shutdown
	go func() {
		if err := e.Start(":" + strconv.Itoa(int(appConfig.Webserver.Port))); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	fmt.Println("Running cleanup tasks...")

	// Your cleanup tasks go here
	// mysql database
	mysql.Close(mysqlDBConnection)
	fmt.Println("Echo was successful shutdown.")
}
