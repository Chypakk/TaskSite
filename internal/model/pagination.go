package model

type PaginationQuery struct {
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
	Status string `json:"status,omitempty"`
	Sort   string `json:"sort,omitempty"`
}

func DefaultPagination() PaginationQuery {
    return PaginationQuery{
        Page:  1,
        Limit: 20,
        Sort:  "created_at:desc",
    }
}

func (p *PaginationQuery) Validate() {
    if p.Page < 1 {
        p.Page = 1
    }
    if p.Limit < 1 {
        p.Limit = 1
    }
    if p.Limit > 20 {
        p.Limit = 20
    }
}

func (p *PaginationQuery) Offset() int {
    return (p.Page - 1) * p.Limit
}

type PaginationMeta struct {
    Page     int  `json:"page"`
    Limit    int  `json:"limit"`
    Total    int  `json:"total"`
    Pages    int  `json:"pages"`
    HasNext  bool `json:"has_next"`
    HasPrev  bool `json:"has_prev"`
}

func NewPaginationMeta(page, limit, total int) PaginationMeta {
    pages := total / limit
    if total%limit > 0 {
        pages++
    }
    return PaginationMeta{
        Page:    page,
        Limit:   limit,
        Total:   total,
        Pages:   pages,
        HasNext: page < pages,
        HasPrev: page > 1,
    }
}