package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Meta describes pagination metadata.
type Meta struct {
	Page      int   `json:"page"`
	PerPage   int   `json:"perPage"`
	Total     int64 `json:"total"`
	TotalPage int   `json:"totalPage"`
}

// paginationBody is the JSON payload for paginated responses.
type paginationBody struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Meta    Meta   `json:"meta"`
}

// Paginate writes a 200 OK response with pagination metadata.
//
// When TotalPage is zero it is computed from Total and PerPage:
//
//	totalPage = ceil(total / perPage)
//
// Pass data as an empty slice ([]T{}) when there are no rows so the payload
// serializes as [] instead of null.
func Paginate(c *gin.Context, data any, meta Meta) {
	if meta.TotalPage <= 0 && meta.PerPage > 0 {
		meta.TotalPage = int((meta.Total + int64(meta.PerPage) - 1) / int64(meta.PerPage))
	}
	c.JSON(http.StatusOK, paginationBody{
		Success: true,
		Code:    CodeSuccess,
		Message: messageSuccess,
		Data:    data,
		Meta:    meta,
	})
}
