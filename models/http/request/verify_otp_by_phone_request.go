package request

import (
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/tensuqiuwulu/be-service-teman-bunda/exceptions"
)

type VerifyOtpRequest struct {
	Credential string `json:"credential" form:"credential" validate:"required"`
	OtpCode    string `json:"otp_code" form:"otp_code" validate:"required"`
}

func ReadFromVerifyOtpRequestBody(c echo.Context, requestId string, logger *logrus.Logger) (verifyOtp *VerifyOtpRequest) {
	verifyOtpRequest := new(VerifyOtpRequest)
	if err := c.Bind(verifyOtpRequest); err != nil {
		exceptions.PanicIfError(err, requestId, logger)
	}
	verifyOtp = verifyOtpRequest
	return verifyOtp
}

func ValidateVerifyOtpByPhoneRequest(validate *validator.Validate, verifyOtpByPhoneRequest *VerifyOtpRequest, requestId string, logger *logrus.Logger) {
	var errorStrings []string
	var errorString string
	err := validate.Struct(verifyOtpByPhoneRequest)
	if err != nil {
		for _, errorValidation := range err.(validator.ValidationErrors) {
			errorString = errorValidation.Field() + " is " + errorValidation.Tag()
			errorStrings = append(errorStrings, errorString)
		}
		// panic(exception.NewError(400, errorStrings))
		exceptions.PanicIfBadRequest(err, requestId, errorStrings, logger)
	}
}
