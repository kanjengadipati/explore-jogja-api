package notification

import (
	"gorm.io/gorm"
)

type Repository interface {
	Create(notif *Notification) error
	FindByUserID(userID uint) ([]Notification, error)
	MarkAsRead(id uint, userID uint) error
	CountUnread(userID uint) (int64, error)
}

type GormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(notif *Notification) error {
	return r.db.Create(notif).Error
}

func (r *GormRepository) FindByUserID(userID uint) ([]Notification, error) {
	var notifs []Notification
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&notifs).Error
	return notifs, err
}

func (r *GormRepository) MarkAsRead(id uint, userID uint) error {
	return r.db.Model(&Notification{}).Where("id = ? AND user_id = ?", id, userID).Update("is_read", true).Error
}

func (r *GormRepository) CountUnread(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&count).Error
	return count, err
}
