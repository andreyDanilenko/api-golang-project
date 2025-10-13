package models

type Link struct {
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
	CreatedAt   int64  `json:"created_at"`
}
