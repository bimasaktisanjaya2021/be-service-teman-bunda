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

type ProductControllerInterface interface {
	FindAllProducts(c echo.Context) error
	FindProductsBySearch(c echo.Context) error
	FindProductById(c echo.Context) error
	FindProductByIdCategory(c echo.Context) error
	FindProductByIdSubCategory(c echo.Context) error
	FindProductByIdBrand(c echo.Context) error
}

type ProductControllerImplementation struct {
	ConfigWebserver         config.Webserver
	Logger                  *logrus.Logger
	ProductServiceInterface services.ProductServiceInterface
}

func NewProductController(configWebserver config.Webserver, productServiceInterface services.ProductServiceInterface) ProductControllerInterface {
	return &ProductControllerImplementation{
		ConfigWebserver:         configWebserver,
		ProductServiceInterface: productServiceInterface,
	}
}

func (controller *ProductControllerImplementation) FindAllProducts(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	page, _ := strconv.Atoi(c.QueryParam("page"))
	productResponses := controller.ProductServiceInterface.FindAllProducts(requestId, limit, page)
	responses := response.Response{Code: 200, Mssg: "success", Data: productResponses, Error: []string{}}
	return c.JSON(http.StatusOK, responses)
}

func (controller *ProductControllerImplementation) FindProductsBySearch(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	product := c.QueryParam("product")
	productResponses := controller.ProductServiceInterface.FindProductsBySearch(requestId, product)
	responses := response.Response{Code: 200, Mssg: "Success", Data: productResponses, Error: []string{}}
	return c.JSON(http.StatusOK, responses)
}

func (controller *ProductControllerImplementation) FindProductById(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	id := c.QueryParam("id_product")
	productResponses := controller.ProductServiceInterface.FindProductById(requestId, id)
	responses := response.Response{Code: 200, Mssg: "Success", Data: productResponses, Error: []string{}}
	return c.JSON(http.StatusOK, responses)
}

func (controller *ProductControllerImplementation) FindProductByIdCategory(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	id := c.QueryParam("id_category")
	productResponses := controller.ProductServiceInterface.FindProductByIdCategory(requestId, id)
	responses := response.Response{Code: 200, Mssg: "Success", Data: productResponses, Error: []string{}}
	return c.JSON(http.StatusOK, responses)
}

func (controller *ProductControllerImplementation) FindProductByIdSubCategory(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	id := c.QueryParam("id_sub_category")
	productResponses := controller.ProductServiceInterface.FindProductByIdSubCategory(requestId, id)
	responses := response.Response{Code: 200, Mssg: "Success", Data: productResponses, Error: []string{}}
	return c.JSON(http.StatusOK, responses)
}

func (controller *ProductControllerImplementation) FindProductByIdBrand(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	id := c.QueryParam("id_brand")
	productResponses := controller.ProductServiceInterface.FindProductByIdBrand(requestId, id)
	responses := response.Response{Code: 200, Mssg: "Success", Data: productResponses, Error: []string{}}
	return c.JSON(http.StatusOK, responses)
}
