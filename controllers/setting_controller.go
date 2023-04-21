package controllers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/tensuqiuwulu/be-service-teman-bunda/config"
	"github.com/tensuqiuwulu/be-service-teman-bunda/models/http/response"
	"github.com/tensuqiuwulu/be-service-teman-bunda/services"
)

type SettingControllerInterface interface {
	FindSettingShippingCost(c echo.Context) error
	FindSettingVerAppAndroid(c echo.Context) error
	FindSettingVerAppIos(c echo.Context) error
	FindNewVersionApp(c echo.Context) error
	FindNewVersionApp2(c echo.Context) error
}

type SettingControllerImplementation struct {
	ConfigWebserver         config.Webserver
	Logger                  *logrus.Logger
	SettingServiceInterface services.SettingServiceInterface
}

func NewSettingController(configWebserver config.Webserver, settingServiceInterface services.SettingServiceInterface) SettingControllerInterface {
	return &SettingControllerImplementation{
		ConfigWebserver:         configWebserver,
		SettingServiceInterface: settingServiceInterface,
	}
}

func (controller *SettingControllerImplementation) FindNewVersionApp(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	OS, _ := strconv.Atoi(c.QueryParam("os"))
	verApp := controller.SettingServiceInterface.FindNewVersionApp(requestId, OS)
	responses := response.Response{Code: 200, Mssg: "success", Data: verApp, Error: []string{}}
	return c.JSON(http.StatusOK, responses)
}

func (controller *SettingControllerImplementation) FindNewVersionApp2(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	OS, _ := strconv.Atoi(c.QueryParam("os"))
	verApp := controller.SettingServiceInterface.FindNewVersionApp2(requestId, OS)
	responses := response.Response{Code: 200, Mssg: "success", Data: verApp, Error: []string{}}
	return c.JSON(http.StatusOK, responses)
}

func (controller *SettingControllerImplementation) FindSettingShippingCost(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	shippingCost := controller.SettingServiceInterface.FindSettingShippingCost(requestId)
	responses := response.Response{Code: 200, Mssg: "success", Data: shippingCost, Error: []string{}}
	return c.JSON(http.StatusOK, responses)
}

func (controller *SettingControllerImplementation) FindSettingVerAppAndroid(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	verApp := controller.SettingServiceInterface.FindSettingVerAppAndroid(requestId)
	responses := response.Response{Code: 200, Mssg: "success", Data: verApp, Error: []string{}}
	return c.JSON(http.StatusOK, responses)
}

func (controller *SettingControllerImplementation) FindSettingVerAppIos(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	verApp := controller.SettingServiceInterface.FindSettingVerAppIos(requestId)
	responses := response.Response{Code: 200, Mssg: "success", Data: verApp, Error: []string{}}
	return c.JSON(http.StatusOK, responses)
}
