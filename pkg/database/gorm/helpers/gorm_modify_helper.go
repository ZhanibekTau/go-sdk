package helpers

import (
	"context"
	"gorm.io/gorm"
	"time"
)

func NewGormModifyHelper[E interface{}](client *gorm.DB) *GormModifyHelper[E] {
	return &GormModifyHelper[E]{
		client:  client,
		timeout: 10,
	}
}

// GormModifyHelper - Вспомогательный хелпер для модификации данных
type GormModifyHelper[E interface{}] struct {
	client  *gorm.DB
	timeout time.Duration
	model   E
}

func (h *GormModifyHelper[E]) SetTimeout(timeout time.Duration) *GormModifyHelper[E] {
	h.timeout = timeout

	return h
}

func (h *GormModifyHelper[E]) Create(model *E) (*E, error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout*time.Second)
	defer cancel()
	result := h.client.WithContext(ctx).Create(model)

	if result.Error != nil {
		return nil, result.Error
	}

	return model, nil
}

func (h *GormModifyHelper[E]) BatchCreate(model *[]E) (*gorm.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout*time.Second)
	defer cancel()
	result := h.client.WithContext(ctx).Create(model)

	if result.Error != nil {
		return nil, result.Error
	}

	return result, nil
}

func (h *GormModifyHelper[E]) Update(model *E) error {
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout*time.Second)
	defer cancel()
	result := h.client.WithContext(ctx).Save(model)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (h *GormModifyHelper[E]) Delete(model *E) error {
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout*time.Second)
	defer cancel()
	result := h.client.WithContext(ctx).Delete(model)

	if result.Error != nil {
		return result.Error
	}

	return nil
}
