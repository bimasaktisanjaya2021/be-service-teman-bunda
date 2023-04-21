package services

import (
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
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

type UserServiceInterface interface {
	CreateUser(requestId string, userRequest *request.CreateUserRequest) (userResponse response.CreateUserResponse)
	FindUserByReferal(requestId string, referalCode string) (userResponse response.FindUserByReferalResponse)
	FindUserById(requestId string, id string) (userResponse response.FindUserByIdResponse)
	UpdateUser(requestId string, idUser string, userRequest *request.UpdateUserRequest) error
	UpdateUserTokenDevice(requestId string, idUser string, userRequest *request.UpdateUserTokenDeviceRequest) error
	UpdateStatusActiveUser(requestId string, accessToken string) error
	PasswordCodeRequest(requestId string, passwordRequest *request.PasswordCodeRequest) error
	PasswordResetCodeVerify(requestId string, passwordResetCodeVerifyRequest *request.PasswordResetCodeVerifyRequest) error
	UpdateUserPassword(requestId string, updateUserPasswordRequest *request.UpdateUserPasswordRequest) error
	DeleteAccount(requestId string, idUser string)
}

type UserServiceImplementation struct {
	ConfigurationWebserver                 config.Webserver
	DB                                     *gorm.DB
	ConfigJwt                              config.Jwt
	Validate                               *validator.Validate
	Logger                                 *logrus.Logger
	ConfigEmail                            config.Email
	UserRepositoryInterface                mysql.UserRepositoryInterface
	ProvinsiRepositoryInterface            mysql.ProvinsiRepositoryInterface
	FamilyRepositoryInterface              mysql.FamilyRepositoryInterface
	FamilyMembersRepositoryInterface       mysql.FamilyMembersRepositoryInterface
	BalancePointRepositoryInterface        mysql.BalancePointRepositoryInterface
	BalancePointTxRepositoryInterface      mysql.BalancePointTxRepositoryInterface
	UserShippingAddressRepositoryInterface mysql.UserShippingAddressRepositoryInterface
}

func NewUserService(
	configurationWebserver config.Webserver,
	DB *gorm.DB,
	configJwt config.Jwt,
	validate *validator.Validate,
	logger *logrus.Logger,
	configEmail config.Email,
	userRepositoryInterface mysql.UserRepositoryInterface,
	provinsiRepositoryInterface mysql.ProvinsiRepositoryInterface,
	familyRepositoryInterface mysql.FamilyRepositoryInterface,
	familyMembersRepositoryInterface mysql.FamilyMembersRepositoryInterface,
	balancePointRepositoryInterface mysql.BalancePointRepositoryInterface,
	balancePointTxRepositoryInterface mysql.BalancePointTxRepositoryInterface,
	userShippingAddressRepositoryInterface mysql.UserShippingAddressRepositoryInterface) UserServiceInterface {
	return &UserServiceImplementation{
		ConfigurationWebserver:                 configurationWebserver,
		DB:                                     DB,
		ConfigJwt:                              configJwt,
		Validate:                               validate,
		Logger:                                 logger,
		ConfigEmail:                            configEmail,
		UserRepositoryInterface:                userRepositoryInterface,
		ProvinsiRepositoryInterface:            provinsiRepositoryInterface,
		FamilyRepositoryInterface:              familyRepositoryInterface,
		FamilyMembersRepositoryInterface:       familyMembersRepositoryInterface,
		BalancePointRepositoryInterface:        balancePointRepositoryInterface,
		BalancePointTxRepositoryInterface:      balancePointTxRepositoryInterface,
		UserShippingAddressRepositoryInterface: userShippingAddressRepositoryInterface,
	}
}

func (service *UserServiceImplementation) DeleteAccount(requestId string, idUser string) {

	user, _ := service.UserRepositoryInterface.FindUserById(service.DB, idUser)

	userEntity := &entity.User{}
	userEntity.IsDelete = 1
	userEntity.PasswordResetCode = ""
	errUpdateUser := service.UserRepositoryInterface.DeleteAccount(service.DB, user.Id, *userEntity)
	exceptions.PanicIfError(errUpdateUser, requestId, service.Logger)
}

func (service *UserServiceImplementation) UpdateUserTokenDevice(requestId string, idUser string, updateUserTokenDeviceRequest *request.UpdateUserTokenDeviceRequest) error {
	// Validate request
	request.ValidateUpdateUserTokenDeviceRequest(service.Validate, updateUserTokenDeviceRequest, requestId, service.Logger)

	userEntity := &entity.User{}
	userEntity.TokenDevice = updateUserTokenDeviceRequest.TokenDevice

	err := service.UserRepositoryInterface.UpdateUserTokenDevice(service.DB, idUser, *userEntity)

	return err
}

func (service *UserServiceImplementation) PasswordResetCodeVerify(requestId string, passwordResetCodeVerifyRequest *request.PasswordResetCodeVerifyRequest) error {
	// Validate request
	request.ValidatePasswordResetCodeVerifyRequest(service.Validate, passwordResetCodeVerifyRequest, requestId, service.Logger)

	user, _ := service.UserRepositoryInterface.FindUserByEmail(service.DB, passwordResetCodeVerifyRequest.Email)

	if user.PasswordResetCode == passwordResetCodeVerifyRequest.Code {
		return nil
	} else {
		err := errors.New("email and code not match")
		exceptions.PanicIfBadRequest(err, requestId, []string{"email and code not match"}, service.Logger)
		return err
	}
}

func (service *UserServiceImplementation) VerifyFormToken(requestId, token string) {
	tokenParse, err := jwt.ParseWithClaims(token, &modelService.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(service.ConfigJwt.FormToken), nil
	})

	if !tokenParse.Valid {
		exceptions.PanicIfUnauthorized(err, requestId, []string{"invalid token"}, service.Logger)
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			exceptions.PanicIfUnauthorized(err, requestId, []string{"invalid token"}, service.Logger)
		} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			exceptions.PanicIfUnauthorized(err, requestId, []string{"invalid token"}, service.Logger)
		} else {
			exceptions.PanicIfUnauthorized(err, requestId, []string{"invalid token"}, service.Logger)
		}
	}
}

func (service *UserServiceImplementation) PasswordCodeRequest(requestId string, passwordRequest *request.PasswordCodeRequest) error {

	user, _ := service.UserRepositoryInterface.FindUserByEmail(service.DB, passwordRequest.Email)

	if user.Id == " " {
		exceptions.PanicIfRecordNotFound(errors.New("email not found"), requestId, []string{"Email not registered"}, service.Logger)
	}

	userEntity := &entity.User{}
	userEntity.PasswordResetCode = utilities.GenerateRandomCode()

	_, errUpdateUser := service.UserRepositoryInterface.UpdatePasswordResetCodeUser(service.DB, user.Id, *userEntity)
	exceptions.PanicIfError(errUpdateUser, requestId, service.Logger)

	templateData := modelService.BodyCodeEmail{
		Code:     userEntity.PasswordResetCode,
		FullName: user.FamilyMembers.FullName,
	}
	to := user.FamilyMembers.Email
	runtime.GOMAXPROCS(1)
	go service.SendEmailPasswordResetCode(to, templateData)

	return nil
}

// Update status aktif user by email confirmation
func (service *UserServiceImplementation) UpdateStatusActiveUser(requestId string, accessToken string) error {
	tokenParse, err := jwt.ParseWithClaims(accessToken, &modelService.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(service.ConfigJwt.VerifyKey), nil
	})

	if !tokenParse.Valid {
		exceptions.PanicIfUnauthorized(err, requestId, []string{"invalid token"}, service.Logger)
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			exceptions.PanicIfUnauthorized(err, requestId, []string{"invalid token"}, service.Logger)
		} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			exceptions.PanicIfUnauthorized(err, requestId, []string{"invalid token"}, service.Logger)
		} else {
			exceptions.PanicIfUnauthorized(err, requestId, []string{"invalid token"}, service.Logger)
		}
	}

	if claims, ok := tokenParse.Claims.(*modelService.TokenClaims); ok && tokenParse.Valid {
		user, _ := service.UserRepositoryInterface.FindUserById(service.DB, claims.Id)
		userEntity := &entity.User{}
		userEntity.IsActive = 1
		userEntity.PasswordResetCode = ""
		userEntity.VerificationDate = null.NewTime(time.Now(), true)

		_, errUpdateUser := service.UserRepositoryInterface.UpdateStatusActiveUser(service.DB, user.Id, *userEntity)
		exceptions.PanicIfError(errUpdateUser, requestId, service.Logger)

		return nil
	} else {
		err := errors.New("no claims")
		exceptions.PanicIfBadRequest(err, requestId, []string{"no claims"}, service.Logger)
		return nil
	}
}

func (service *UserServiceImplementation) UpdateUserPassword(requestId string, updateUserPasswordRequest *request.UpdateUserPasswordRequest) error {
	service.VerifyFormToken(requestId, updateUserPasswordRequest.FormToken)

	// Validate request
	request.ValidateUpdateUserPasswordRequest(service.Validate, updateUserPasswordRequest, requestId, service.Logger)

	user, _ := service.UserRepositoryInterface.FindUserByPhone(service.DB, updateUserPasswordRequest.Credential)

	password := updateUserPasswordRequest.Password
	bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	exceptions.PanicIfBadRequest(err, requestId, []string{"Error Generate Password"}, service.Logger)

	userEntity := &entity.User{}
	userEntity.Password = string(bcryptPassword)
	userEntity.PasswordResetCode = ""
	_, errUpdateUser := service.UserRepositoryInterface.UpdateUserPassword(service.DB, user.Id, *userEntity)
	exceptions.PanicIfError(errUpdateUser, requestId, service.Logger)

	return nil
}

func (service *UserServiceImplementation) UpdateUser(requestId string, idUser string, userRequest *request.UpdateUserRequest) error {

	// Validate request
	request.ValidateUpdateUserRequest(service.Validate, userRequest, requestId, service.Logger)

	user, _ := service.UserRepositoryInterface.FindUserById(service.DB, idUser)

	if userRequest.Username != user.Username {
		// Check username if exsict
		checkUsername, _ := service.UserRepositoryInterface.FindUserByUsername(service.DB, userRequest.Username)
		if checkUsername.Id != "" {
			err := errors.New("username already exist")
			exceptions.PanicIfRecordAlreadyExists(err, requestId, []string{"Username already exist"}, service.Logger)
		}
	}

	if userRequest.Email != user.FamilyMembers.Email {
		// Check email if exist
		checkEmail, _ := service.UserRepositoryInterface.FindUserByEmail(service.DB, userRequest.Email)
		if checkEmail.Id != "" {
			err := errors.New("email already exist")
			exceptions.PanicIfRecordAlreadyExists(err, requestId, []string{"Email already exist"}, service.Logger)
		}
	}

	if userRequest.Phone != user.FamilyMembers.Phone {
		// Check phone if exsict
		checkPhone, _ := service.UserRepositoryInterface.FindUserByPhone(service.DB, userRequest.Phone)
		if checkPhone.Id != "" {
			err := errors.New("phone already exist")
			exceptions.PanicIfRecordAlreadyExists(err, requestId, []string{"Phone already exist"}, service.Logger)
		}
	}

	tx := service.DB.Begin()

	// Generate Password
	password := userRequest.Password
	bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	exceptions.PanicIfBadRequest(err, requestId, []string{"Error Generate Password"}, service.Logger)

	// Update family members profile
	familyMembersEntity := &entity.FamilyMembers{}
	familyMembersEntity.FullName = userRequest.FullName
	familyMembersEntity.Email = userRequest.Email
	familyMembersEntity.Phone = userRequest.Phone
	familyMembers, errUpdateFamilyMembers := service.FamilyMembersRepositoryInterface.UpdateFamilyMembers(tx, user.IdFamilyMembers, *familyMembersEntity)
	exceptions.PanicIfErrorWithRollback(errUpdateFamilyMembers, requestId, []string{"Error update family members"}, service.Logger, tx)

	// Update user profile
	userEntity := &entity.User{}
	userEntity.IdFamilyMembers = familyMembers.Id
	userEntity.Username = userRequest.Username
	if userRequest.Password != "" {
		userEntity.Password = string(bcryptPassword)
	}
	_, errUpdateUser := service.UserRepositoryInterface.UpdateUser(tx, idUser, *userEntity)
	exceptions.PanicIfErrorWithRollback(errUpdateUser, requestId, []string{"Error update user"}, service.Logger, tx)

	commit := tx.Commit()
	exceptions.PanicIfError(commit.Error, requestId, service.Logger)
	return nil
}

func (service *UserServiceImplementation) CreateUser(requestId string, userRequest *request.CreateUserRequest) (userResponse response.CreateUserResponse) {

	// Validate request
	request.ValidateCreateUserRequest(service.Validate, userRequest, requestId, service.Logger)

	service.VerifyFormToken(requestId, userRequest.FormToken)

	emailLowerCase := strings.ToLower(userRequest.Email)
	// Check email if exsict
	checkEmail, _ := service.UserRepositoryInterface.FindUserByEmail(service.DB, emailLowerCase)
	if checkEmail.Id != "" {
		err := errors.New("email already exist")
		exceptions.PanicIfRecordAlreadyExists(err, requestId, []string{"Email sudah digunakan"}, service.Logger)
	}

	phone := strings.Replace(userRequest.Phone, "-", "", -1)
	phoneFinal := strings.Replace(phone, "+62", "0", -1)

	// Check phone if exsict
	checkPhone, _ := service.UserRepositoryInterface.FindUserByPhone(service.DB, phoneFinal)
	if checkPhone.Id != "" {
		err := errors.New("phone already exist")
		exceptions.PanicIfRecordAlreadyExists(err, requestId, []string{"Phone sudah digunakan"}, service.Logger)
	}

	// Begin Transcation
	tx := service.DB.Begin()
	exceptions.PanicIfError(tx.Error, requestId, service.Logger)

	// Generate Password
	password := strings.ReplaceAll(userRequest.Password, " ", "")
	bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	exceptions.PanicIfBadRequest(err, requestId, []string{"Error Generate Password"}, service.Logger)

	// Generate referal code
	referalCode := service.GenerateReferalCode(userRequest.FullName)

	// Create family profile
	familyEntity := &entity.Family{}
	familyEntity.Id = utilities.RandomUUID()
	family, err := service.FamilyRepositoryInterface.CreateFamily(tx, *familyEntity)
	exceptions.PanicIfErrorWithRollback(err, requestId, []string{"Error create family"}, service.Logger, tx)

	// Create family members profile
	familyMembersEntity := &entity.FamilyMembers{}
	familyMembersEntity.Id = utilities.RandomUUID()
	familyMembersEntity.IdFamily = familyEntity.Id
	familyMembersEntity.FullName = userRequest.FullName
	familyMembersEntity.Email = emailLowerCase
	familyMembersEntity.Phone = phoneFinal
	familyMembers, err := service.FamilyMembersRepositoryInterface.CreateFamilyMembers(tx, *familyMembersEntity)
	exceptions.PanicIfErrorWithRollback(err, requestId, []string{"Error create family members"}, service.Logger, tx)

	// Crate user profile
	userEntity := &entity.User{}
	userEntity.Id = utilities.RandomUUID()
	userEntity.IdFamilyMembers = familyMembers.Id
	userEntity.IdLevelMember = 1
	userEntity.Password = string(bcryptPassword)
	userEntity.IsActive = 1
	userEntity.VerificationDate = null.NewTime(time.Now(), true)
	if userRequest.RegistrationReferalCode == "" {
		// dafault kode referal jika inputan kosong
		userEntity.RegistrationReferalCode = "0X0ROQIBA"
	} else {
		userEntity.RegistrationReferalCode = strings.ToUpper(userRequest.RegistrationReferalCode)
	}

	userEntity.CreatedDate = time.Now()
	userEntity.VerificationDueDate = time.Now().Add(time.Hour * 24)
	userEntity.ReferalCode = referalCode
	userEntity.RefreshToken = ""

	user, err := service.UserRepositoryInterface.CreateUser(tx, *userEntity)
	exceptions.PanicIfErrorWithRollback(err, requestId, []string{"Error insert user"}, service.Logger, tx)

	// Create user balance points
	balancePointEntity := &entity.BalancePoint{}
	balancePointEntity.Id = utilities.RandomUUID()
	balancePointEntity.IdUser = userEntity.Id
	balancePointEntity.CreatedDate = time.Now()
	balancePoint, err := service.BalancePointRepositoryInterface.CreateBalancePoint(tx, *balancePointEntity)
	exceptions.PanicIfErrorWithRollback(err, requestId, []string{"Error insert balance point"}, service.Logger, tx)

	// Kode untuk bonus point registrasi referal
	if userRequest.RegistrationReferalCode != "" {
		// dapat bonus point jika menggunakan kode referal

		balancePointEntity := &entity.BalancePoint{}
		balancePointEntity.BalancePoints = 10000

		_, errUpdateBalancePoint := service.BalancePointRepositoryInterface.UpdateBalancePoint(tx, balancePoint.IdUser, *balancePointEntity)
		exceptions.PanicIfErrorWithRollback(errUpdateBalancePoint, requestId, []string{"update balance point error"}, service.Logger, tx)

		// Add to point history
		balancePointTxEntity := &entity.BalancePointTx{}
		balancePointTxEntity.Id = utilities.RandomUUID()
		balancePointTxEntity.IdBalancePoint = balancePoint.Id
		balancePointTxEntity.TxType = "debit"
		balancePointTxEntity.TxDate = time.Now()
		balancePointTxEntity.TxNominal = balancePointEntity.BalancePoints
		balancePointTxEntity.LastPointBalance = 0
		balancePointTxEntity.NewPointBalance = balancePointEntity.BalancePoints
		balancePointTxEntity.CreatedDate = time.Now()
		balancePointTxEntity.Description = "Bonus Registrasi"

		_, errCreateBalancePointTx := service.BalancePointTxRepositoryInterface.CreateBalancePointTx(tx, *balancePointTxEntity)
		exceptions.PanicIfErrorWithRollback(errCreateBalancePointTx, requestId, []string{"create balance point tx error"}, service.Logger, tx)
	}
	// end of promo registration code

	commit := tx.Commit()
	exceptions.PanicIfError(commit.Error, requestId, service.Logger)
	userResponse = response.ToUserCreateUserResponse(user, family, familyMembers, balancePoint)

	return userResponse
}

func (service *UserServiceImplementation) GenerateReferalCode(fullName string) string {
	var splitNickname []string
	var referalName string
	var referalCode string

	splitFullName := strings.Split(fullName, " ")

	if len(splitFullName) >= 2 {
		splitNickname = strings.Split(string(splitFullName[len(splitFullName)-2]), "")
	} else {
		splitNickname = strings.Split(fullName, "")
	}

	referalName = splitNickname[0] + splitNickname[1] + splitNickname[2]

	// Check if referal code exist
	for {
		rand.Seed(time.Now().UTC().UnixNano())
		generateCode := 100 + rand.Intn(999-100)
		referalCode = strings.ToUpper(referalName) + strconv.Itoa(generateCode)

		result, _ := service.UserRepositoryInterface.FindUserByReferalCode(service.DB, referalCode)
		if result.Id == "" {
			break
		}
	}

	return referalCode
}

func (service *UserServiceImplementation) FindUserByReferal(requestId string, referal string) (userResponse response.FindUserByReferalResponse) {
	user, err := service.UserRepositoryInterface.FindUserByReferal(service.DB, referal)
	exceptions.PanicIfError(err, requestId, service.Logger)
	if len(user.Id) == 0 {
		exceptions.PanicIfRecordNotFound(errors.New("referal code not found"), requestId, []string{"referal code not found"}, service.Logger)
	}
	userResponse = response.ToUserFindByReferalResponse(user)
	return userResponse
}

func (service *UserServiceImplementation) FindUserById(requestId string, id string) (userResponse response.FindUserByIdResponse) {
	user, _ := service.UserRepositoryInterface.FindUserById(service.DB, id)
	if user.Id == "" {
		err := errors.New("user not found")
		exceptions.PanicIfRecordNotFound(err, requestId, []string{"Not Found"}, service.Logger)
	}
	userCount, _ := service.UserRepositoryInterface.CountUserByRegistrationReferal(service.DB, user.ReferalCode)
	userResponse = response.ToUserFindByIdResponse(user, userCount)
	return userResponse
}

func (service *UserServiceImplementation) SendEmailVerification(to string, data interface{}) {
	var err error
	template := "./template/verifikasi_email.html"
	subject := "Verifikasi Email Teman Bunda"
	err = utilities.SendEmail(to, subject, data, template)
	if err == nil {
		fmt.Println("send email '" + subject + "' success")
	} else {
		fmt.Println(err)
	}
}

func (service *UserServiceImplementation) SendEmailPasswordResetCode(to string, data interface{}) {
	var err error
	template := "./template/verifikasi_code_password.html"
	subject := "Permintaan Reset Password"
	err = utilities.SendEmail(to, subject, data, template)
	if err == nil {
		fmt.Println("send email '" + subject + "' success")
	} else {
		fmt.Println(err)
	}
}

func (service *UserServiceImplementation) GenerateTokenVerify(user modelService.User) (token string, err error) {
	// Create the Claims
	claims := modelService.TokenClaims{
		Id:       user.Id,
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			// ExpiresAt: time.Now().Add(time.Minute * time.Duration(service.ConfigJwt.Tokenexpiredtime)).Unix(),
			Issuer: "ayaka",
		},
	}

	tokenWithClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = tokenWithClaims.SignedString([]byte(service.ConfigJwt.VerifyKey))
	if err != nil {
		return "", err
	}
	return token, err
}
