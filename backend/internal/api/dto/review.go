package dto

type ReviewListQuery struct {
	Tab        string `form:"tab"`
	CursorTime string `form:"cursor_time"`
	CursorID   uint64 `form:"cursor_id"`
	PageSize   int    `form:"page_size"`
}
