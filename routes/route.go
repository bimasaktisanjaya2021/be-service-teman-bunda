package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/tensuqiuwulu/be-service-teman-bunda/config"
	"github.com/tensuqiuwulu/be-service-teman-bunda/controllers"
	authMiddlerware "github.com/tensuqiuwulu/be-service-teman-bunda/middleware"
)

// Provinsi Route
func ProvinsiRoute(e *echo.Echo, configWebserver config.Webserver, configJWT config.Jwt, provinsiControllerInterface controllers.ProvinsiControllerInterface) {
	group := e.Group("api/v1")
	group.GET("/provinsi", provinsiControllerInterface.FindAllProvinsi)
}

// Kabupaten Route
func KabupatenRoute(e *echo.Echo, configWebserver config.Webserver, kabupatenControllerInterface controllers.KabupatenControllerInterface) {
	group := e.Group("api/v1")
	group.GET("/kabupaten", kabupatenControllerInterface.FindAllKabupatenByIdProvinsi)
}

// Kabupaten Route
func KecamatanRoute(e *echo.Echo, configWebserver config.Webserver, kecamatanControllerInterface controllers.KecamatanControllerInterface) {
	group := e.Group("api/v1")
	group.GET("/kecamatan", kecamatanControllerInterface.FindAllKecamatanByIdKabupaten)
}

// Kelurahan Route
func KelurahanRoute(e *echo.Echo, configWebserver config.Webserver, kelurahanControllerInterface controllers.KelurahanControllerInterface) {
	group := e.Group("api/v1")
	group.GET("/kelurahan", kelurahanControllerInterface.FindAllKelurahanByIdKecamatan)
}

// User Route
func UserRoute(e *echo.Echo, configWebserver config.Webserver, configurationJWT config.Jwt, userControllerInterface controllers.UserControllerInterface, userShippingAddressControllerInterface controllers.UserShippingAddressControllerInterface) {
	group := e.Group("api/v1")
	group.POST("/user/create", userControllerInterface.CreateUser)
	group.GET("/user/referal", userControllerInterface.FindUserByReferal)
	group.GET("/user", userControllerInterface.FindUserById, authMiddlerware.Authentication(configurationJWT))
	group.PUT("/user/update", userControllerInterface.UpdateUser, authMiddlerware.Authentication(configurationJWT))
	group.PUT("/user/update/tokendevice", userControllerInterface.UpdateUserTokenDevice, authMiddlerware.Authentication(configurationJWT))
	group.POST("/password/request/code", userControllerInterface.PasswordCodeRequest)
	group.POST("/password/code/verify", userControllerInterface.PasswordResetCodeVerify)
	group.POST("/user/update/password", userControllerInterface.UpdateUserPassword)
	group.GET("/user/shipping/address", userShippingAddressControllerInterface.FindUserShippingAddress, authMiddlerware.Authentication(configurationJWT))
	group.POST("/user/shipping/address", userShippingAddressControllerInterface.CreateUserShippingAddress, authMiddlerware.Authentication(configurationJWT))
	group.DELETE("/user/shipping/address", userShippingAddressControllerInterface.DeleteUserShippingAddress, authMiddlerware.Authentication(configurationJWT))
	group.PUT("/user/delete/account", userControllerInterface.DeleteAccount, authMiddlerware.Authentication(configurationJWT))
}

func VerifyEmailRoute(e *echo.Echo, configWebserver config.Webserver, configurationJWT config.Jwt, userControllerInterface controllers.UserControllerInterface) {
	e.GET("/user/email/verify", userControllerInterface.UpdateStatusActiveUser)
}

// Auth Route
func AuthRoute(e *echo.Echo, configWebserver config.Webserver, configurationJWT config.Jwt, authControllerInterface controllers.AuthControllerInterface) {
	group := e.Group("api/v1")
	group.POST("/auth/login", authControllerInterface.Login)
	group.POST("/auth/new-token", authControllerInterface.NewToken)
	group.POST("/auth/verify/otp", authControllerInterface.VerifyOtp)
	// group.POST("/auth/send/otp/whatsapp", authControllerInterface.SendOtpWhatsapp)
	group.POST("/auth/sendotp/sms", authControllerInterface.SendOtpBySms)
	group.POST("/auth/sendotp/email", authControllerInterface.SendOtpByEmail)
}

// Balance Point Route
func BalancePointRoute(e *echo.Echo, configWebserver config.Webserver, configurationJWT config.Jwt, balancePointControllerInterface controllers.BalancePointControllerInterface) {
	group := e.Group("api/v1")
	group.GET("/balance_point", balancePointControllerInterface.FindBalancePointByIdUser, authMiddlerware.Authentication(configurationJWT))
	group.GET("/balance_point/check/amount", balancePointControllerInterface.BalancePointCheckAmount, authMiddlerware.Authentication(configurationJWT))
	group.GET("/balance_point/check/order_tx", balancePointControllerInterface.BalancePointCheckOrderTx, authMiddlerware.Authentication(configurationJWT))
}

// Balance Point Tx
func BalancePointTxRoute(e *echo.Echo, configWebserver config.Webserver, configurationJWT config.Jwt, balancePointTxControllerInterface controllers.BalancePointTxControllerInterface) {
	group := e.Group("api/v1")
	group.GET("/balance_point_tx", balancePointTxControllerInterface.FindBalancePointTxByIdBalancePoint, authMiddlerware.Authentication(configurationJWT))
}

// Product Route
func ProductRoute(e *echo.Echo, configWebserver config.Webserver, configurationJWT config.Jwt, productControllerInterface controllers.ProductControllerInterface) {
	group := e.Group("api/v1")
	group.GET("/products", productControllerInterface.FindAllProducts, authMiddlerware.Authentication(configurationJWT))
	group.GET("/products/search", productControllerInterface.FindProductsBySearch, authMiddlerware.Authentication(configurationJWT))
	group.GET("/product/id", productControllerInterface.FindProductById, authMiddlerware.Authentication(configurationJWT))
	group.GET("/products/category", productControllerInterface.FindProductByIdCategory, authMiddlerware.Authentication(configurationJWT))
	group.GET("/products/sub_category", productControllerInterface.FindProductByIdSubCategory, authMiddlerware.Authentication(configurationJWT))
	group.GET("/products/brand", productControllerInterface.FindProductByIdBrand, authMiddlerware.Authentication(configurationJWT))
	group.GET("/products/notoken", productControllerInterface.FindAllProducts)
	group.GET("/products/notoken/search", productControllerInterface.FindProductsBySearch)
	group.GET("/product/notoken/id", productControllerInterface.FindProductById)
	group.GET("/products/notoken/category", productControllerInterface.FindProductByIdCategory)
	group.GET("/products/notoken/sub_category", productControllerInterface.FindProductByIdSubCategory)
	group.GET("/products/notoken/brand", productControllerInterface.FindProductByIdBrand)
}

// Cart Route
func CartRoute(e *echo.Echo, configWebserver config.Webserver, configurationJWT config.Jwt, cartControllerInterface controllers.CartControllerInterface) {
	group := e.Group("api/v1")
	group.GET("/cart", cartControllerInterface.FindCartByIdUser, authMiddlerware.Authentication(configurationJWT))
	group.POST("/cart", cartControllerInterface.AddProductToCart, authMiddlerware.Authentication(configurationJWT))
	group.PUT("/cart/plus_qty", cartControllerInterface.CartPlusQtyProduct, authMiddlerware.Authentication(configurationJWT))
	group.PUT("/cart/min_qty", cartControllerInterface.CartMinQtyProduct, authMiddlerware.Authentication(configurationJWT))
	group.PUT("/cart/update_qty", cartControllerInterface.UpdateQtyProductInCart, authMiddlerware.Authentication(configurationJWT))
}

// Shipping Cost Route
func ShippingRoute(e *echo.Echo, configWebserver config.Webserver, configurationJWT config.Jwt, shippingControllerInterface controllers.ShippingControllerInterface) {
	group := e.Group("api/v1")
	group.GET("/shipping/cost", shippingControllerInterface.GetShippingCostByIdKelurahan, authMiddlerware.Authentication(configurationJWT))
}

// Order Route
func OrderRoute(e *echo.Echo, configWebserver config.Webserver, configurationJWT config.Jwt, orderControllerInterface controllers.OrderControllerInterface) {
	group := e.Group("api/v1")
	group.POST("/order/create", orderControllerInterface.CreateOrder, authMiddlerware.Authentication(configurationJWT))
	group.POST("/order/update", orderControllerInterface.UpdateStatusOrder)
	group.GET("/order", orderControllerInterface.FindOrderByUser, authMiddlerware.Authentication(configurationJWT))
	group.GET("/order/detail/id", orderControllerInterface.FindOrderById, authMiddlerware.Authentication(configurationJWT))
	group.PUT("/order/cancel/id", orderControllerInterface.CancelOrderById, authMiddlerware.Authentication(configurationJWT))
	group.PUT("/order/complete/id", orderControllerInterface.CompleteOrderById, authMiddlerware.Authentication(configurationJWT))
	group.GET("/order/payment/check", orderControllerInterface.OrderCheckPayment, authMiddlerware.Authentication(configurationJWT))
}

// List Payment
func PaymentChannelRoute(e *echo.Echo, configWebserver config.Webserver, configurationJWT config.Jwt, paymentChannelControllerInterface controllers.PaymentChannelControllerInterface) {
	group := e.Group("api/v1")
	group.GET("/payment/list", paymentChannelControllerInterface.FindListPaymentChannel, authMiddlerware.Authentication(configurationJWT))

	group2 := e.Group("api/v2")
	group2.GET("/payment/list", paymentChannelControllerInterface.FindListPaymentChannel, authMiddlerware.Authentication(configurationJWT))
}

// List Payment
func PaymentRoute(e *echo.Echo, configWebserver config.Webserver, configurationJWT config.Jwt, paymentControllerInterface controllers.PaymentControllerInterface) {
	group := e.Group("api/v1")
	group.POST("/payment/status", paymentControllerInterface.PaymentStatus, authMiddlerware.Authentication(configurationJWT))
	e.GET("thank-you", paymentControllerInterface.PaymentCreditCardSuccess)
	e.GET("credit-card-cancel-order", paymentControllerInterface.PaymentCreditCardCancel)
}

// List Payment
func BannerRoute(e *echo.Echo, configWebserver config.Webserver, configurationJWT config.Jwt, bannerControllerInterface controllers.BannerControllerInterface) {
	group := e.Group("api/v1")
	group.GET("/banner", bannerControllerInterface.FindAllBanner, authMiddlerware.Authentication(configurationJWT))
	group.GET("/banner/notoken", bannerControllerInterface.FindAllBanner)
}

// Product Brand
func ProductBrandRoute(e *echo.Echo, configWebserver config.Webserver, configurationJWT config.Jwt, productBrandControllerInterface controllers.ProductBrandControllerInterface) {
	group := e.Group("api/v1")
	group.GET("/productbrand", productBrandControllerInterface.FindAllProductBrand, authMiddlerware.Authentication(configurationJWT))
}

// Setting
func SettingRoute(e *echo.Echo, configWebserver config.Webserver, configurationJWT config.Jwt, settingControllerInterface controllers.SettingControllerInterface) {
	group := e.Group("api/v1")
	group.GET("/setting/shippingcost", settingControllerInterface.FindSettingShippingCost, authMiddlerware.Authentication(configurationJWT))
	group.GET("/setting/verapp/android", settingControllerInterface.FindSettingVerAppAndroid)
	group.GET("/setting/verapp/ios", settingControllerInterface.FindSettingVerAppIos)
	group.GET("/setting/version", settingControllerInterface.FindNewVersionApp)

	gropv2 := e.Group("api/v2")
	gropv2.GET("/setting/version", settingControllerInterface.FindNewVersionApp2)
}

// Main Route
func MainRoute(e *echo.Echo, configWebserver config.Webserver, mainControllerInterface controllers.MainControllerInterface) {
	e.GET("/", mainControllerInterface.Main)
}
