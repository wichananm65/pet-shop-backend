package category

// CategoryItem is the public DTO returned by the category API.
// JSON tags follow the camelCase convention used elsewhere in the project.
type CategoryItem struct {
	CategoryID     int     `json:"categoryID"`
	CategoryName   string  `json:"categoryName"`
	CategoryNameTH *string `json:"categoryNameTH,omitempty"`
	CategoryImg    *string `json:"categoryImg,omitempty"`
}
