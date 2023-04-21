package request

import (
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/tensuqiuwulu/be-service-teman-bunda/exceptions"
)

type CreateOrderRequest struct {
	TotalBill      float64 `json:"total_bill" form:"total_bill" validate:"required"`
	PaymentByPoint float64 `json:"payment_by_point" form:"payment_by_point"`
	// PaymentByCash  float64 `json:"payment_by_cash" form:"payment_by_cash"`
	Address        string  `json:"address" form:"address" validate:"required"`
	CourierNote    string  `json:"courier_note" form:"courier_note"`
	ShippingCost   float64 `json:"shipping_cost" form:"shipping_cost"`
	PaymentMethod  string  `json:"payment_method" form:"payment_method" validate:"required"`
	PaymentChannel string  `json:"payment_channel" form:"payment_channel" validate:"required"`
	PaymentFee     float64 `json:"payment_fee" form:"payment_fee"`
}

func ReadFromCreateOrderRequestBody(c echo.Context, requestId string, logger *logrus.Logger) (createOrder *CreateOrderRequest) {
	createOrderRequest := new(CreateOrderRequest)
	if err := c.Bind(createOrderRequest); err != nil {
		exceptions.PanicIfError(err, requestId, logger)
	}
	createOrder = createOrderRequest
	return createOrder
}

func ValidateCreateOrderRequest(validate *validator.Validate, createOrder *CreateOrderRequest, requestId string, logger *logrus.Logger) {
	var errorStrings []string
	var errorString string
	err := validate.Struct(createOrder)
	if err != nil {
		for _, errorValidation := range err.(validator.ValidationErrors) {
			errorString = errorValidation.Field() + " is " + errorValidation.Tag()
			errorStrings = append(errorStrings, errorString)
		}
		// panic(exception.NewError(400, errorStrings))
		exceptions.PanicIfBadRequest(err, requestId, errorStrings, logger)
	}
}
