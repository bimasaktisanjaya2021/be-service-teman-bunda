package request

import (
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/tensuqiuwulu/be-service-teman-bunda/exceptions"
)

type UpdateUserRequest struct {
	FullName string `json:"full_name" form:"full_name"`
	Email    string `json:"email" form:"email"`
	Phone    string `json:"phone" form:"phone"`
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}

func ReadFromUpdateUserRequestBody(c echo.Context, requestId string, logger *logrus.Logger) (updateUser *UpdateUserRequest) {
	updateUserRequest := new(UpdateUserRequest)
	if err := c.Bind(updateUserRequest); err != nil {
		exceptions.PanicIfError(err, requestId, logger)
	}
	updateUser = updateUserRequest
	return updateUser
}

func ValidateUpdateUserRequest(validate *validator.Validate, updateUser *UpdateUserRequest, requestId string, logger *logrus.Logger) {
	var errorStrings []string
	err := validate.Struct(updateUser)
	var errorString string
	if err != nil {
		for _, errorValidation := range err.(validator.ValidationErrors) {
			errorString = errorValidation.Field() + " is " + errorValidation.Tag()
			errorStrings = append(errorStrings, errorString)
		}
		// panic(exception.NewError(400, errorStrings))
		exceptions.PanicIfBadRequest(err, requestId, errorStrings, logger)
	}
}
