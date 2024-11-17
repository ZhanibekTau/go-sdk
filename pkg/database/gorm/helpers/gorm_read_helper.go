package helpers

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"time"
)

func NewGormReadHelper[E interface{}](client *gorm.DB) *GormPaginatedHelper[E] {
	return &GormPaginatedHelper[E]{
		client:  client,
		timeout: 10,
	}
}

// GormReadHelper - Вспомогательный хелпер для чтения данных
type GormReadHelper[E interface{}] struct {
	client  *gorm.DB
	timeout time.Duration
	model   E
}

// GetByCondition Возвращает одну модель по запросу
func (h *GormReadHelper[E]) GetByCondition(callback func(client *gorm.DB) *gorm.DB) (*E, error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout*time.Second)
	defer cancel()

	var model E
	result := callback(h.client).WithContext(ctx).First(&model)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &model, nil
}

// GetById Возвращает одну модель по ID
func (h *GormReadHelper[E]) GetById(id int) (*E, error) {
	result, err := h.GetByCondition(func(client *gorm.DB) *gorm.DB {

		return client.Where("id = ?", id)
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
