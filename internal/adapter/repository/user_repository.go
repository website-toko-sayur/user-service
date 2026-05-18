package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/domain/model"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

type UserRepositoryInterface interface {
	GetUserByEmail(ctx context.Context, email string) (*entity.UserEntity, error)
	CreateUserAccount(ctx context.Context, req entity.UserEntity) (int64, error)
	UpdateUserVerified(ctx context.Context, userID int64) (*entity.UserEntity, error)
	UpdatePasswordByID(ctx context.Context, req entity.UserEntity) error
	GetUserByID(ctx context.Context, userID int64) (*entity.UserEntity, error)
	UpdateDataUser(ctx context.Context, req entity.UserEntity) error

	// Modul Customers Admin
	GetCustomerAll(ctx context.Context, query entity.QueryStringCustomer) ([]entity.UserEntity, int64, int64, error)
	GetCustomerByID(ctx context.Context, customerID int64) (*entity.UserEntity, error)
	CreateCustomer(ctx context.Context, req entity.UserEntity) (int64, error)
	UpdateCustomer(ctx context.Context, req entity.UserEntity) error
	DeleteCustomer(ctx context.Context, customerID int64) error
}

func NewUserRepository(db *gorm.DB) UserRepositoryInterface {
	return &userRepository{db: db}
}

func (u *userRepository) GetUserByEmail(ctx context.Context, email string) (*entity.UserEntity, error) {
	modelUser := model.User{}

	result := u.db.WithContext(ctx).Where("email = ? AND is_verified = ?", email, true).Preload("Roles").First(&modelUser)
	if result.Error != nil {
		log.Error().
			Err(result.Error).
			Str("source", "internal.adapter.userRepository.GetUserByEmail")
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		log.Info().
			Str("source", "internal.adapter.userRepository.GetUserByEmail").
			Msg("User not found")
		return nil, errors.New("404")
	}

	return &entity.UserEntity{
		ID:         modelUser.ID,
		Name:       modelUser.Name,
		Email:      email,
		Password:   modelUser.Password,
		RoleName:   modelUser.Roles[0].Name,
		Address:    modelUser.Address,
		Lat:        modelUser.Lat,
		Lng:        modelUser.Lng,
		Phone:      modelUser.Phone,
		Photo:      modelUser.Photo,
		IsVerified: modelUser.IsVerified,
	}, nil
}

func (u *userRepository) CreateUserAccount(ctx context.Context, req entity.UserEntity) (int64, error) {
	tx := u.db.WithContext(ctx).Begin()

	if tx.Error != nil {
		return 0, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var roleID int64

	if err := tx.
		Model(&model.Role{}).
		Select("id").
		Where("name = ?", "Customer").
		Scan(&roleID).
		Error; err != nil {

		tx.Rollback()

		log.Error().
			Err(err).
			Str("source", "internal.adapter.userRepository.CreateUserAccount").
			Msg("failed get customer role")

		return 0, err
	}

	modelUser := model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Roles: []model.Role{
			{ID: roleID},
		},
	}

	if err := tx.Create(&modelUser).Error; err != nil {
		tx.Rollback()

		log.Error().
			Err(err).
			Str("source", "internal.adapter.userRepository.CreateUserAccount").
			Msg("failed create user")

		return 0, err
	}

	modelVerify := model.VerificationToken{
		UserID:    modelUser.ID,
		Token:     req.Token,
		TokenType: "email_verification",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	if err := tx.Create(&modelVerify).Error; err != nil {
		tx.Rollback()

		log.Error().
			Err(err).
			Str("source", "internal.adapter.userRepository.CreateUserAccount").
			Msg("failed create verification token")

		return 0, err
	}

	if err := tx.Commit().Error; err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userRepository.CreateUserAccount").
			Msg("failed commit transaction")

		return 0, err
	}

	return modelUser.ID, nil
}

func (u *userRepository) UpdateUserVerified(ctx context.Context, userID int64) (*entity.UserEntity, error) {
	modelUser := model.User{}

	if err := u.db.WithContext(ctx).Where("id = ?", userID).Preload("Roles").First(&modelUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Error().
				Err(err).
				Str("source", "internal.adapter.userRepository.UpdateUserVerified")
			return nil, err
		}
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userRepository.UpdateUserVerified")
		return nil, err
	}

	modelUser.IsVerified = true
	if err := u.db.WithContext(ctx).Save(&modelUser).Error; err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userRepository.UpdateUserVerified")
		return nil, err
	}

	return &entity.UserEntity{
		ID:         userID,
		Name:       modelUser.Name,
		Email:      modelUser.Email,
		RoleName:   modelUser.Roles[0].Name,
		Address:    modelUser.Address,
		Lat:        modelUser.Lat,
		Lng:        modelUser.Lng,
		Phone:      modelUser.Phone,
		Photo:      modelUser.Photo,
		IsVerified: modelUser.IsVerified,
	}, nil
}

func (u *userRepository) UpdatePasswordByID(ctx context.Context, req entity.UserEntity) error {
	modelUser := model.User{}

	if err := u.db.WithContext(ctx).Where("id =?", req.ID).First(&modelUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Error().
				Err(err).
				Str("source", "internal.adapter.userRepository.UpdatePasswordByID")
			return err
		}
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userRepository.UpdatePasswordByID")
		return err
	}

	modelUser.Password = req.Password
	if err := u.db.WithContext(ctx).Save(&modelUser).Error; err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userRepository.UpdatePasswordByID")
		return err
	}

	return nil
}

func (u *userRepository) GetUserByID(ctx context.Context, userID int64) (*entity.UserEntity, error) {
	modelUser := model.User{}

	result := u.db.WithContext(ctx).Where("id =? AND is_verified = true", userID).Preload("Roles").First(&modelUser)
	if result.Error != nil {
		log.Error().
			Err(result.Error).
			Str("source", "internal.adapter.userRepository.GetUserByID")
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		log.Info().
			Str("source", "internal.adapter.userRepository.GetUserByID").
			Msg("User not found")
		return nil, errors.New("404")
	}

	return &entity.UserEntity{
		ID:       modelUser.ID,
		Email:    modelUser.Email,
		Name:     modelUser.Name,
		RoleName: modelUser.Roles[0].Name,
		Lat:      modelUser.Lat,
		Lng:      modelUser.Lng,
		Address:  modelUser.Address,
		Phone:    modelUser.Phone,
		Photo:    modelUser.Photo,
	}, nil
}

func (u *userRepository) UpdateDataUser(ctx context.Context, req entity.UserEntity) error {
	modelUser := model.User{
		Name:    req.Name,
		Email:   req.Email,
		Address: req.Address,
		Phone:   req.Phone,
		Photo:   req.Photo,
	}

	result := u.db.WithContext(ctx).Where("id = ? AND is_verified = true", req.ID).First(&modelUser)
	if result.Error != nil {
		log.Error().
			Err(result.Error).
			Str("source", "internal.adapter.userRepository.UpdateDataUser")
		return result.Error
	}

	if result.RowsAffected == 0 {
		log.Info().
			Str("source", "internal.adapter.userRepository.UpdateDataUser").
			Msg("User not found")
		return errors.New("404")
	}

	modelUser.Lat = req.Lat
	modelUser.Lng = req.Lng
	modelUser.Address = req.Address
	modelUser.Phone = req.Phone

	if err := u.db.WithContext(ctx).UpdateColumns(&modelUser).Error; err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userRepository.UpdateDataUser")
		return err
	}

	return nil
}

func (u *userRepository) GetCustomerAll(ctx context.Context, query entity.QueryStringCustomer) ([]entity.UserEntity, int64, int64, error) {
	modelUsers := []model.User{}
	var countData int64

	order := fmt.Sprintf("%s %s", query.OrderBy, query.OrderType)
	offset := (query.Page - 1) * query.Limit

	sqlMain := u.db.WithContext(ctx).Preload("Roles", "name = ?", "Customer").
		Where("name ILIKE ? OR email ILIKE ? OR phone ILIKE ?", "%"+query.Search+"%", "%"+query.Search+"%", "%"+query.Search+"%")

	if err := sqlMain.Model(&modelUsers).Count(&countData).Error; err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userRepository.GetCustomerAll")
		return nil, 0, 0, err
	}

	totalPage := int(math.Ceil(float64(countData) / float64(query.Limit)))

	if err := sqlMain.Order(order).Limit(int(query.Limit)).Offset(int(offset)).Find(&modelUsers).Error; err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userRepository.GetCustomerAll")
		return nil, 0, 0, err
	}

	if len(modelUsers) < 1 {
		err := errors.New("404")
		log.Info().
			Str("source", "internal.adapter.userRepository.GetCustomerAll").
			Msg("No customer found")
		return nil, 0, 0, err
	}

	respEntities := []entity.UserEntity{}
	for _, val := range modelUsers {
		roleName := ""
		for _, role := range val.Roles {
			roleName = role.Name
		}
		respEntities = append(respEntities, entity.UserEntity{
			ID:       val.ID,
			Name:     val.Name,
			Email:    val.Email,
			RoleName: roleName,
			Phone:    val.Email,
			Photo:    val.Photo,
		})
	}
	return respEntities, countData, int64(totalPage), nil
}

func (u *userRepository) GetCustomerByID(ctx context.Context, customerID int64) (*entity.UserEntity, error) {
	modelUser := model.User{}

	if err := u.db.WithContext(ctx).Where("id = ?", customerID).Preload("Roles").First(&modelUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Info().
				Str("source", "internal.adapter.userRepository.GetCustomerByID").
				Msg("User not found")
			return nil, err
		}
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userRepository.GetCustomerByID")

		return nil, err
	}

	roleID := 0
	for _, role := range modelUser.Roles {
		roleID = int(role.ID)
	}

	return &entity.UserEntity{
		ID:      customerID,
		Name:    modelUser.Name,
		Email:   modelUser.Email,
		RoleID:  int64(roleID),
		Address: modelUser.Address,
		Lat:     modelUser.Lat,
		Lng:     modelUser.Lng,
		Phone:   modelUser.Phone,
		Photo:   modelUser.Photo,
	}, nil
}

func (u *userRepository) CreateCustomer(ctx context.Context, req entity.UserEntity) (int64, error) {
	modelRole := model.Role{}

	if err := u.db.WithContext(ctx).Where("id =?", req.RoleID).First(&modelRole).Error; err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userRepository.CreateCustomer")
		return 0, err
	}

	modelUser := model.User{
		Name:       req.Name,
		Email:      req.Email,
		Password:   req.Password,
		Address:    req.Address,
		Lat:        req.Lat,
		Lng:        req.Lng,
		Phone:      req.Phone,
		Photo:      req.Photo,
		Roles:      []model.Role{modelRole},
		IsVerified: true,
	}

	if err := u.db.WithContext(ctx).Create(&modelUser).Error; err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userRepository.CreateCustomer")
		return 0, err
	}

	return modelUser.ID, nil
}

func (u *userRepository) UpdateCustomer(ctx context.Context, req entity.UserEntity) error {
	modelRole := model.Role{}
	if err := u.db.WithContext(ctx).Where("id =?", req.RoleID).First(&modelRole).Error; err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userRepository.UpdateCustomer")
		return err
	}

	modelUser := model.User{}
	if err := u.db.WithContext(ctx).Where("id =?", req.ID).First(&modelUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Info().
				Str("source", "internal.adapter.userRepository.UpdateCustomer").
				Msg("No customer found")
			return err
		}
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userRepository.UpdateCustomer")
		return err
	}

	modelUser.Name = req.Name
	modelUser.Email = req.Email
	modelUser.Phone = req.Phone
	modelUser.Roles = []model.Role{modelRole}
	if req.Address != "" {
		modelUser.Address = req.Address
	}

	if req.Lat != "" {
		modelUser.Lat = req.Lat
	}

	if req.Lng != "" {
		modelUser.Lng = req.Lng
	}
	if req.Photo != "" {
		modelUser.Lat = req.Lat
	}

	if req.Password != "" {
		modelUser.Password = req.Password
	}

	if err := u.db.WithContext(ctx).Save(&modelUser).Error; err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userRepository.UpdateCustomer")
		return err
	}

	return nil
}

func (u *userRepository) DeleteCustomer(ctx context.Context, customerID int64) error {
	modelUser := model.User{}
	if err := u.db.WithContext(ctx).Where("id =?", customerID).First(&modelUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Info().
				Str("source", "internal.adapter.userRepository.DeleteCustomer").
				Msg("No customer found")
			return err
		}
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userRepository.DeleteCustomer")
		return err
	}

	if err := u.db.WithContext(ctx).Delete(&modelUser).Error; err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userRepository.DeleteCustomer")
		return err
	}
	return nil
}
