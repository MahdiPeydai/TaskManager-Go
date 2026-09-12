package dto

type PaginationRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type PaginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}
