package mysql

import (
	"github.com/tensuqiuwulu/be-service-teman-bunda/config"
	"github.com/tensuqiuwulu/be-service-teman-bunda/models/entity"
	"gorm.io/gorm"
)

type SettingRepositoryInterface interface {
	FindSettingsByName(DB *gorm.DB, settingName string) (entity.Settings, error)
	FindSettingShippingCost(db *gorm.DB) (entity.Settings, error)
	FindSettingVerApp(db *gorm.DB, os string) (entity.Settings, error)
	FindNewVersionApp(db *gorm.DB, os int) ([]entity.Settings, error)
}

type SettingRepositoryImplementation struct {
	configurationDatabase *config.Database
}

func NewSettingRepository(configDatabase *config.Database) SettingRepositoryInterface {
	return &SettingRepositoryImplementation{
		configurationDatabase: configDatabase,
	}
}

func (repository *SettingRepositoryImplementation) FindNewVersionApp(db *gorm.DB, os int) ([]entity.Settings, error) {
	var settings []entity.Settings
	results := db.Where("value = ?", os).Find(&settings)
	return settings, results.Error
}

func (repository *SettingRepositoryImplementation) FindSettingsByName(DB *gorm.DB, settingName string) (entity.Settings, error) {
	var settings entity.Settings
	results := DB.Where("settings_name = ?", settingName).First(&settings)
	return settings, results.Error
}

func (repository *SettingRepositoryImplementation) FindSettingShippingCost(DB *gorm.DB) (entity.Settings, error) {
	var settings entity.Settings
	results := DB.Where("settings_name = ?", "shipping_cost").First(&settings)
	return settings, results.Error
}

func (repository *SettingRepositoryImplementation) FindSettingVerApp(DB *gorm.DB, os string) (entity.Settings, error) {
	var settings entity.Settings
	results := DB.Where("settings_name = ?", os).First(&settings)
	return settings, results.Error
}
