package helpers

import (
	"context"
	paginator "github.com/ZhanibekTau/go-sdk/pkg/database/gorm/pagination"
	"gorm.io/gorm"
	"time"
)

func NewGormPaginatedHelper[E interface{}](client *gorm.DB) *GormPaginatedHelper[E] {
	return &GormPaginatedHelper[E]{
		client:     client,
		perPage:    30,
		maxPerPage: 1000,
		timeout:    10,
	}
}

// GormPaginatedHelper - Вспомогательный хелпер для постраничного чтения данных
type GormPaginatedHelper[E interface{}] struct {
	client     *gorm.DB
	perPage    int
	maxPerPage int
	timeout    time.Duration
	model      E
}

func (h *GormPaginatedHelper[E]) SetTimeout(timeout time.Duration) *GormPaginatedHelper[E] {
	h.timeout = timeout

	return h
}

func (h *GormPaginatedHelper[E]) SetPerPage(perPage int) *GormPaginatedHelper[E] {
	h.perPage = perPage

	return h
}

func (h *GormPaginatedHelper[E]) Paginated(page int, callback func(client *gorm.DB) *gorm.DB) (*paginator.Paginated[E], error) {
	var structure paginator.Paginated[E]
	var err error
	paging := paginator.Paging{}
	paging.Page = page
	paging.Limit = h.perPage
	paging.MaxLimit = h.maxPerPage
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout*time.Second)
	defer cancel()

	structure.Pagination, err = paginator.Pages(&paginator.Param{
		DB:     callback(h.client).WithContext(ctx),
		Paging: &paging,
	}, &structure.Items)

	if err != nil {
		return nil, err
	}

	structure.Pagination.To = structure.Pagination.From + len(structure.Items)

	if len(structure.Items) == 0 {
		structure.Pagination.From = 0
	}

	structure.Pagination.From += 1

	return &structure, nil
}
