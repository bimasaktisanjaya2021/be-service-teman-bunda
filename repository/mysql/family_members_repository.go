package mysql

import (
	"github.com/tensuqiuwulu/be-service-teman-bunda/config"
	"github.com/tensuqiuwulu/be-service-teman-bunda/models/entity"
	"gorm.io/gorm"
)

type FamilyMembersRepositoryInterface interface {
	CreateFamilyMembers(DB *gorm.DB, user entity.FamilyMembers) (entity.FamilyMembers, error)
	UpdateFamilyMembers(DB *gorm.DB, idFamilyMembers string, familyMembers entity.FamilyMembers) (entity.FamilyMembers, error)
}

type FamilyMembersRepositoryImplementation struct {
	configurationDatabase *config.Database
}

func NewFamilyMembersRepository(configDatabase *config.Database) FamilyMembersRepositoryInterface {
	return &FamilyMembersRepositoryImplementation{
		configurationDatabase: configDatabase,
	}
}

func (repository *FamilyMembersRepositoryImplementation) CreateFamilyMembers(DB *gorm.DB, familyMembers entity.FamilyMembers) (entity.FamilyMembers, error) {
	results := DB.Create(familyMembers)
	return familyMembers, results.Error
}

func (repository *FamilyMembersRepositoryImplementation) UpdateFamilyMembers(DB *gorm.DB, idFamilyMembers string, familyMembers entity.FamilyMembers) (entity.FamilyMembers, error) {
	result := DB.
		Model(entity.FamilyMembers{}).
		Where("id = ?", idFamilyMembers).
		Updates(entity.FamilyMembers{
			FullName: familyMembers.FullName,
			Phone:    familyMembers.Phone,
			Email:    familyMembers.Email,
			// IdProvinsi:  familyMembers.IdProvinsi,
			// IdKabupaten: familyMembers.IdKabupaten,
			// IdKecamatan: familyMembers.IdKecamatan,
			// IdKelurahan: familyMembers.IdKelurahan,
			// Address:     familyMembers.Address,
		})
	return familyMembers, result.Error
}
