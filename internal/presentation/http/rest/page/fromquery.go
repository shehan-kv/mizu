package page

import (
	"mizu/internal/presentation/http/rest/query"
	"net/http"
)

func FromQuery(r *http.Request) (PagingParams, error) {
	limit, err := query.ExtractInt(r, "limit", 15)
	if err != nil {
		return PagingParams{}, err
	}
	page, err := query.ExtractInt(r, "page", 1)
	if err != nil {
		return PagingParams{}, err
	}
	return PagingParams{
		Limit:  limit,
		Offset: (page - 1) * limit,
		Page:   page,
	}, nil
}
