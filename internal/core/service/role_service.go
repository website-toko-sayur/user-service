package service

import (
	"context"
	"user-service/internal/adapter/repository"
	"user-service/internal/core/domain/entity"
)

type roleService struct {
	repo repository.RoleRepositoryInterface
}

type RoleServiceInterface interface {
	GetAll(ctx context.Context, search string) ([]entity.RoleEntity, error)
	GetByID(ctx context.Context, id int64) (*entity.RoleEntity, error)
	Create(ctx context.Context, req entity.RoleEntity) error
	Delete(ctx context.Context, id int64) error
	Update(ctx context.Context, req entity.RoleEntity) error
}

func NewRoleService(repo repository.RoleRepositoryInterface) RoleServiceInterface {
	return &roleService{repo: repo}
}

func (r *roleService) Create(ctx context.Context, req entity.RoleEntity) error {
	return r.repo.Create(ctx, req)
}

func (r *roleService) Delete(ctx context.Context, id int64) error {
	return r.repo.Delete(ctx, id)
}

func (r *roleService) GetAll(ctx context.Context, search string) ([]entity.RoleEntity, error) {
	return r.repo.GetAll(ctx, search)
}

func (r *roleService) GetByID(ctx context.Context, id int64) (*entity.RoleEntity, error) {
	return r.repo.GetByID(ctx, id)
}

func (r *roleService) Update(ctx context.Context, req entity.RoleEntity) error {
	return r.repo.Update(ctx, req)
}
