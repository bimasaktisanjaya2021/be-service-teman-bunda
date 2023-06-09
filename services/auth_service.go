package services

import (
	"errors"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
	"github.com/tensuqiuwulu/be-service-teman-bunda/config"
	"github.com/tensuqiuwulu/be-service-teman-bunda/exceptions"
	"github.com/tensuqiuwulu/be-service-teman-bunda/models/entity"
	"github.com/tensuqiuwulu/be-service-teman-bunda/models/http/request"
	"github.com/tensuqiuwulu/be-service-teman-bunda/models/http/response"
	modelService "github.com/tensuqiuwulu/be-service-teman-bunda/models/service"
	"github.com/tensuqiuwulu/be-service-teman-bunda/repository/mysql"
	"github.com/tensuqiuwulu/be-service-teman-bunda/utilities"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
)

type AuthServiceInterface interface {
	Login(requestId string, authRequest *request.AuthRequest) (authResponse interface{})
	NewToken(requestId string, refreshToken string) (token string)
	GenerateToken(user modelService.User) (token string, err error)
	GenerateRefreshToken(user modelService.User) (token string, err error)
	VerifyOtp(requestId string, verifyOtpRequest *request.VerifyOtpRequest) (string, error)
	SendOtpBySms(requestId string, sendOtpBySmsRequest *request.SendOtpBySmsRequest) error
	SendOtpByEmail(requestId string, sendOtpByEmail *request.SendOtpByEmailRequest) error
}

type AuthServiceImplementation struct {
	ConfigurationWebserver        config.Webserver
	DB                            *gorm.DB
	ConfigJwt                     config.Jwt
	Validate                      *validator.Validate
	Logger                        *logrus.Logger
	UserRepositoryInterface       mysql.UserRepositoryInterface
	SettingRepositoryInterface    mysql.SettingRepositoryInterface
	OtpManagerRepositoryInterface mysql.OtpManagerRepositoryInterface
}

func NewAuthService(
	configurationWebserver config.Webserver,
	DB *gorm.DB,
	configJwt config.Jwt,
	validate *validator.Validate,
	logger *logrus.Logger,
	userRepositoryInterface mysql.UserRepositoryInterface,
	settingRepositoryInterface mysql.SettingRepositoryInterface,
	otpManagerRepositoryInterface mysql.OtpManagerRepositoryInterface) AuthServiceInterface {
	return &AuthServiceImplementation{
		ConfigurationWebserver:        configurationWebserver,
		DB:                            DB,
		ConfigJwt:                     configJwt,
		Validate:                      validate,
		Logger:                        logger,
		UserRepositoryInterface:       userRepositoryInterface,
		SettingRepositoryInterface:    settingRepositoryInterface,
		OtpManagerRepositoryInterface: otpManagerRepositoryInterface,
	}
}

func (service *AuthServiceImplementation) SendOtpByEmail(requestId string, sendOtpByEmailRequest *request.SendOtpByEmailRequest) error {
	request.ValidateSendOtpByEmailRequest(service.Validate, sendOtpByEmailRequest, requestId, service.Logger)

	emailLowerCase := strings.ToLower(sendOtpByEmailRequest.Email)
	user, _ := service.UserRepositoryInterface.FindUserByEmail(service.DB, emailLowerCase)
	if user.Id == "" {
		exceptions.PanicIfRecordNotFound(errors.New("record not found"), requestId, []string{"user not found"}, service.Logger)
	}

	otpCode := utilities.GenerateRandomCode()
	bcryptOtpCode, err := bcrypt.GenerateFromPassword([]byte(otpCode), bcrypt.DefaultCost)
	exceptions.PanicIfBadRequest(err, requestId, []string{"Error Generate otp code"}, service.Logger)

	userEntity := &entity.User{}
	userEntity.OtpCode = string(bcryptOtpCode)
	userEntity.OtpCodeExpiredDueDate = null.NewTime(time.Now().Add(time.Minute*5), true)
	errUpdateOtpCodeUser := service.UserRepositoryInterface.UpdateOtpCodeUser(service.DB, user.Id, *userEntity)
	exceptions.PanicIfError(errUpdateOtpCodeUser, requestId, service.Logger)

	dataEmail := modelService.BodyCodeEmail{
		Code:     otpCode,
		FullName: user.FamilyMembers.FullName,
	}

	template := "./template/verifikasi_code_password.html"
	subject := "Kode OTP Teman Bunda"
	go utilities.SendEmail(user.FamilyMembers.Email, subject, dataEmail, template)
	return nil
}

func (service *AuthServiceImplementation) SendOtpBySms(requestId string, sendOtpBySmsRequest *request.SendOtpBySmsRequest) error {
	request.ValidateSendOtpBySmsRequest(service.Validate, sendOtpBySmsRequest, requestId, service.Logger)

	resultOtp, err := service.OtpManagerRepositoryInterface.FindOtpByPhone(service.DB, sendOtpBySmsRequest.Phone)
	exceptions.PanicIfError(err, requestId, service.Logger)

	if sendOtpBySmsRequest.TypeOtp == 1 {
		user, _ := service.UserRepositoryInterface.FindUserByPhone(service.DB, resultOtp.Phone)
		if len(user.Id) != 0 {
			exceptions.PanicIfBadRequest(errors.New("phone already use"), requestId, []string{"phone already use"}, service.Logger)
		}
	}

	if len(resultOtp.Id) == 0 {
		otpManagerEntity := &entity.OtpManager{}
		otpManagerEntity.Id = utilities.RandomUUID()
		otpManagerEntity.OtpCode = utilities.GenerateRandomCode()
		otpManagerEntity.Phone = sendOtpBySmsRequest.Phone
		otpManagerEntity.PhoneLimit = 5
		otpManagerEntity.IpAddressLimit = 5
		otpManagerEntity.OtpExperiedAt = time.Now().Add(time.Minute * 5)
		otpManagerEntity.CreatedDate = time.Now()

		go utilities.SendSmsOtp(sendOtpBySmsRequest.Phone, otpManagerEntity.OtpCode)

		createOtpErr := service.OtpManagerRepositoryInterface.CreateOtp(service.DB, otpManagerEntity)
		exceptions.PanicIfError(createOtpErr, requestId, service.Logger)

	} else {
		if resultOtp.PhoneLimit <= 0 {
			exceptions.PanicIfBadRequest(errors.New("phone daily limit"), requestId, []string{"phone daily limit"}, service.Logger)
		}

		otpManagerEntity := &entity.OtpManager{}
		otpManagerEntity.OtpCode = utilities.GenerateRandomCode()
		// otpManagerEntity.OtpCode = "123456"
		otpManagerEntity.PhoneLimit = resultOtp.PhoneLimit - 1
		otpManagerEntity.OtpExperiedAt = time.Now().Add(time.Minute * 5)
		otpManagerEntity.UpdatedDate = null.NewTime(time.Now(), true)

		// Send OTP
		go utilities.SendSmsOtp(sendOtpBySmsRequest.Phone, otpManagerEntity.OtpCode)

		// Update OTP
		updateOtpErr := service.OtpManagerRepositoryInterface.UpdateOtp(service.DB, resultOtp.Id, otpManagerEntity)
		exceptions.PanicIfError(updateOtpErr, requestId, service.Logger)
	}
	return nil
}

func (service *AuthServiceImplementation) VerifyOtp(requestId string, verifyOtpRequest *request.VerifyOtpRequest) (string, error) {
	request.ValidateVerifyOtpByPhoneRequest(service.Validate, verifyOtpRequest, requestId, service.Logger)

	// var user entity.User

	otp, _ := service.OtpManagerRepositoryInterface.FindOtpByPhone(service.DB, verifyOtpRequest.Credential)
	if len(otp.Id) == 0 {
		exceptions.PanicIfRecordNotFound(errors.New("otp not found"), requestId, []string{"otp not found"}, service.Logger)
	}

	if verifyOtpRequest.OtpCode != otp.OtpCode {
		exceptions.PanicIfBadRequest(errors.New("otp not match"), requestId, []string{"otp not match"}, service.Logger)
	}

	// var userModelService modelService.User
	// userModelService.Phone = otp.Phone
	token, _ := service.GenerateTokenForm()
	// verifyOtpResponse := response.ToVerifyOtpResponse(token)

	return token, nil

	// Jika user melakukan pendaftaran
	// if user.IsActive == 0 && user.NotVerification != 1 {
	// 	userEntity := &entity.User{}
	// 	userEntity.OtpCode = " "
	// 	userEntity.IsActive = 1
	// 	userEntity.VerificationDate = null.NewTime(time.Now(), true)
	// 	_, err := service.UserRepositoryInterface.UpdateStatusActiveUser(service.DB, user.Id, *userEntity)
	// 	exceptions.PanicIfError(err, requestId, service.Logger)
	// 	return "", nil
	// } else if user.IsActive == 1 {
	// 	userEntity := &entity.User{}
	// 	userEntity.OtpCode = " "
	// 	errUpdateOtpCodeUser := service.UserRepositoryInterface.UpdateOtpCodeUser(service.DB, user.Id, *userEntity)
	// 	exceptions.PanicIfError(errUpdateOtpCodeUser, requestId, service.Logger)
	// 	token, _ := service.GenerateTokenForm()
	// 	return token, nil
	// } else {
	// 	err := errors.New("bad request")
	// 	exceptions.PanicIfBadRequest(err, requestId, []string{"bad request"}, service.Logger)
	// 	return "", nil
	// }
}

func (service *AuthServiceImplementation) Login(requestId string, authRequest *request.AuthRequest) (authResponse interface{}) {
	var userModelService modelService.User
	var user entity.User

	request.ValidateAuth(service.Validate, authRequest, requestId, service.Logger)

	// jika username tidak ditemukan
	user, _ = service.UserRepositoryInterface.FindUserByPhone(service.DB, authRequest.Credential)
	if user.Id == "" {
		// cek apakah yg di input email
		emailLowerCase := strings.ToLower(authRequest.Credential)
		user, _ = service.UserRepositoryInterface.FindUserByEmail(service.DB, emailLowerCase)
		if user.Id == "" {
			user, _ = service.UserRepositoryInterface.FindUserByUsername(service.DB, authRequest.Credential)
			if len(user.Id) == 0 {
				exceptions.PanicIfRecordNotFound(errors.New("user not found"), requestId, []string{"not found"}, service.Logger)
			}
		}
	}

	if user.IsDelete == 1 {
		exceptions.PanicIfRecordNotFound(errors.New("user not found"), requestId, []string{"not found"}, service.Logger)
	}

	if user.IsActive == 1 {
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(authRequest.Password))
		exceptions.PanicIfBadRequest(err, requestId, []string{"Invalid Credentials"}, service.Logger)

		userModelService.Id = user.Id
		userModelService.Username = user.Username
		userModelService.IdKelurahan = user.FamilyMembers.IdKelurahan

		token, err := service.GenerateToken(userModelService)
		exceptions.PanicIfError(err, requestId, service.Logger)

		refreshToken, err := service.GenerateRefreshToken(userModelService)
		exceptions.PanicIfError(err, requestId, service.Logger)

		_, err = service.UserRepositoryInterface.SaveUserRefreshToken(service.DB, userModelService.Id, refreshToken)
		exceptions.PanicIfError(err, requestId, service.Logger)

		setting, _ := service.SettingRepositoryInterface.FindSettingsByName(service.DB, "ver_app")

		authResponse = response.ToAuthResponse(userModelService.Id, userModelService.Username, token, refreshToken, setting.SettingsTitle)

		return authResponse
	} else {
		exceptions.PanicIfUnauthorized(errors.New("account is not active"), requestId, []string{"not active"}, service.Logger)
		return nil
	}

}

func (service *AuthServiceImplementation) NewToken(requestId string, refreshToken string) (token string) {
	tokenParse, err := jwt.ParseWithClaims(refreshToken, &modelService.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(service.ConfigJwt.Key), nil
	})

	if !tokenParse.Valid {
		exceptions.PanicIfUnauthorized(err, requestId, []string{"invalid token"}, service.Logger)
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			exceptions.PanicIfUnauthorized(err, requestId, []string{"invalid token"}, service.Logger)
		} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			exceptions.PanicIfUnauthorized(err, requestId, []string{"expired token"}, service.Logger)
		} else {
			exceptions.PanicIfError(err, requestId, service.Logger)
		}
	}

	if claims, ok := tokenParse.Claims.(*modelService.TokenClaims); ok && tokenParse.Valid {
		//fmt.Printf("%v %v", claims, ok)
		user, err := service.UserRepositoryInterface.FindUserByUsernameAndRefreshToken(service.DB, claims.Username, refreshToken)
		exceptions.PanicIfRecordNotFound(err, requestId, []string{"User tidak ada"}, service.Logger)

		var userModelService modelService.User
		userModelService.Id = user.Id
		userModelService.Username = user.Username
		// userModelService.CreatedDate = user.CreatedDate
		token, err := service.GenerateRefreshToken(userModelService)
		exceptions.PanicIfError(err, requestId, service.Logger)
		return token
	} else {
		err := errors.New("no claims")
		exceptions.PanicIfBadRequest(err, requestId, []string{"no claims"}, service.Logger)
		return ""
	}
}

func (service *AuthServiceImplementation) GenerateToken(user modelService.User) (token string, err error) {
	// Create the Claims
	claims := modelService.TokenClaims{
		Id:          user.Id,
		Username:    user.Username,
		IdKelurahan: user.IdKelurahan,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * time.Duration(service.ConfigJwt.Tokenexpiredtime)).Unix(),
			Issuer:    "aether",
		},
	}

	tokenWithClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = tokenWithClaims.SignedString([]byte(service.ConfigJwt.Key))
	if err != nil {
		return "", err
	}
	return token, err
}

func (service *AuthServiceImplementation) GenerateRefreshToken(user modelService.User) (token string, err error) {
	// Create the Claims
	claims := modelService.TokenClaims{
		Id:       user.Id,
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().AddDate(0, 0, int(service.ConfigJwt.Refreshtokenexpiredtime)).Unix(),
			Issuer:    "aether",
		},
	}

	tokenWithClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = tokenWithClaims.SignedString([]byte(service.ConfigJwt.Key))
	if err != nil {
		return "", err
	}
	return token, err
}

func (service *AuthServiceImplementation) GenerateTokenForm() (token string, err error) {
	// Create the Claims
	claims := modelService.TokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * time.Duration(service.ConfigJwt.FormTokenexpiredtime)).Unix(),
			Issuer:    "aether",
		},
	}

	tokenWithClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = tokenWithClaims.SignedString([]byte(service.ConfigJwt.FormToken))
	if err != nil {
		return "", err
	}
	return token, err
}
