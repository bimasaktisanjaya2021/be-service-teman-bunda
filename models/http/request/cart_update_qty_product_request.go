package request

import (
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/tensuqiuwulu/be-service-teman-bunda/exceptions"
)

type UpdateQtyProductInCartRequest struct {
	IdCart string `json:"id_cart" form:"id_cart" validate:"required"`
	Qty    int    `json:"qty" form:"qty"`
}

func ReadFromUpdateProductInCartRequestBody(c echo.Context, requestId string, logger *logrus.Logger) (updateQtyProductInCart *UpdateQtyProductInCartRequest) {
	updateQtyProductInCartRequest := new(UpdateQtyProductInCartRequest)
	if err := c.Bind(updateQtyProductInCartRequest); err != nil {
		exceptions.PanicIfError(err, requestId, logger)
	}
	updateQtyProductInCart = updateQtyProductInCartRequest
	return updateQtyProductInCart
}

func ValidateUpdateQtyProductInCartRequest(validate *validator.Validate, updateQtyProductInCart *UpdateQtyProductInCartRequest, requestId string, logger *logrus.Logger) {
	var errorStrings []string
	var errorString string
	err := validate.Struct(updateQtyProductInCart)
	if err != nil {
		for _, errorValidation := range err.(validator.ValidationErrors) {
			errorString = errorValidation.Field() + " is " + errorValidation.Tag()
			errorStrings = append(errorStrings, errorString)
		}
		// panic(exception.NewError(400, errorStrings))
		exceptions.PanicIfBadRequest(err, requestId, errorStrings, logger)
	}
}
