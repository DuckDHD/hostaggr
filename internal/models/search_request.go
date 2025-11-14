package models

// SearchRequest represents the incoming search query
type SearchRequest struct {
	City    string
	CheckIn string
	Nights  int
	Adults  int
}
