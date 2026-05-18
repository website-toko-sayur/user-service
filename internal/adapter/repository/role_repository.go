package repository

import (
	"context"
	"errors"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/domain/model"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type roleRepository struct {
	db *gorm.DB
}

type RoleRepositoryInterface interface {
	GetAll(ctx context.Context, search string) ([]entity.RoleEntity, error)
	GetByID(ctx context.Context, id int64) (*entity.RoleEntity, error)
	Create(ctx context.Context, req entity.RoleEntity) error
	Delete(ctx context.Context, id int64) error
	Update(ctx context.Context, req entity.RoleEntity) error
}

func NewRoleRepository(db *gorm.DB) RoleRepositoryInterface {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(ctx context.Context, req entity.RoleEntity) error {
	modelRole := model.Role{
		Name: req.Name,
	}

	if err := r.db.Create(&modelRole).Error; err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.roleRepository.Create")
		return err
	}

	return nil
}

func (r *roleRepository) Delete(ctx context.Context, id int64) error {
	modelRole := model.Role{}

	if err := r.db.Where("id = ?", id).Preload("Users").First(&modelRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Info().
				Str("source", "internal.adapter.roleRepository.Delete").
				Msg("Role not found")
			return err
		}
		log.Error().
			Err(err).
			Str("source", "internal.adapter.roleRepository.Delete")
		return err
	}

	if len(modelRole.Users) > 0 {
		err := errors.New("400")
		log.Info().
			Str("source", "internal.adapter.roleRepository.Delete").
			Msg("Role is associated with users")
		return err
	}

	if err := r.db.Delete(&modelRole).Error; err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.roleRepository.Delete")
		return err
	}

	return nil
}

func (r *roleRepository) GetAll(ctx context.Context, search string) ([]entity.RoleEntity, error) {
	modelRoles := []model.Role{}

	if err := r.db.Where("name ILIKE ?", "%"+search+"%").Find(&modelRoles).Error; err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.roleRepository.GetAll")
		return nil, err
	}

	if len(modelRoles) == 0 {
		err := errors.New("404")
		log.Info().
			Str("source", "internal.adapter.roleRepository.GetAll").
			Msg("Role is associated with users")
		return nil, err
	}

	entityRole := []entity.RoleEntity{}
	for _, modelRole := range modelRoles {
		entityRole = append(entityRole, entity.RoleEntity{
			ID:   modelRole.ID,
			Name: modelRole.Name,
		})
	}

	return entityRole, nil
}

func (r *roleRepository) GetByID(ctx context.Context, id int64) (*entity.RoleEntity, error) {
	modelRole := model.Role{}

	if err := r.db.Where("id = ?", id).First(&modelRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Info().
				Str("source", "internal.adapter.roleRepository.GetByID").
				Msg("Role not found")
			return nil, err
		}
		log.Error().
			Err(err).
			Str("source", "internal.adapter.roleRepository.GetByID")
		return nil, err
	}

	return &entity.RoleEntity{
		ID:   modelRole.ID,
		Name: modelRole.Name,
	}, nil
}

func (r *roleRepository) Update(ctx context.Context, req entity.RoleEntity) error {
	modelRole := model.Role{
		Name: req.Name,
	}

	if err := r.db.Where("id = ?", req.ID).First(&modelRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Info().
				Str("source", "internal.adapter.roleRepository.Update").
				Msg("Role not found")
			return err
		}
		log.Error().
			Err(err).
			Str("source", "internal.adapter.roleRepository.Update")
		return err
	}

	if err := r.db.Save(modelRole).Error; err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.roleRepository.Update")
		return err
	}

	return nil
}
