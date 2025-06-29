package internal

import "github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"

func Paginate[T any](items []T, page, limit int) ([]T, *swagger.PaginationInfo) {
	total := len(items)
	totalPages := (total + limit - 1) / limit

	start := (page - 1) * limit
	end := start + limit
	if end > total {
		end = total
	}
	if start > total {
		start = total
	}

	pagination := swagger.NewPaginationInfoBuilder().
		Limit(int64(limit)).
		Page(int64(page)).
		Total(int64(total)).
		TotalPages(int64(totalPages)).
		Build()

	return items[start:end], pagination
}
