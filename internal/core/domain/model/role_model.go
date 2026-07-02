package model

import (
	"time"

	"gorm.io/gorm"
)

type Role struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"type:varchar(255);unique;not null"`
	CreatedAt time.Time `gorm:"type:timestamp;default:current_timestamp"`
	UpdatedAt *time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Users     []User         `gorm:"many2many:user_role"` // merujuk table pivot, user_role -> penghubung antara user & role
}

/*
 buat partial unique index, agar unique index hanya berlaku untuk data yang belum dihapus

 ALTER TABLE roles
     DROP CONSTRAINT IF EXISTS uni_roles_name;

 CREATE UNIQUE INDEX uni_roles_name
     ON roles (name)
     WHERE deleted_at IS NULL;
*/
