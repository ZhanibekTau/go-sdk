package pagination

type PagerRequest struct {
	Page    int `form:"page"`
	PerPage int `form:"per_page"`
}
