package mysql

import (
	"github.com/tensuqiuwulu/be-service-teman-bunda/config"
	"github.com/tensuqiuwulu/be-service-teman-bunda/models/entity"
	"gorm.io/gorm"
)

type BankTransferRepositoryInterface interface {
	FindBankTransferByBankCode(DB *gorm.DB, bankCode string) (entity.BankTransfer, error)
	FindAllBankTransfer(Db *gorm.DB) ([]entity.BankTransfer, error)
}

type BankTransferRepositoryImplementation struct {
	configurationDatabase *config.Database
}

func NewBankTransferRepository(configDatabase *config.Database) BankTransferRepositoryInterface {
	return &BankTransferRepositoryImplementation{
		configurationDatabase: configDatabase,
	}
}

func (repository *BankTransferRepositoryImplementation) FindBankTransferByBankCode(DB *gorm.DB, bankCode string) (entity.BankTransfer, error) {
	var bankTransfer entity.BankTransfer
	results := DB.Where("bank_code = ?", bankCode).Where("is_active = ?", "1").First(&bankTransfer)
	return bankTransfer, results.Error
}

func (repository *BankTransferRepositoryImplementation) FindAllBankTransfer(DB *gorm.DB) ([]entity.BankTransfer, error) {
	var bankTransfers []entity.BankTransfer
	results := DB.Where("is_active = ?", "1").Find(&bankTransfers)
	return bankTransfers, results.Error
}
