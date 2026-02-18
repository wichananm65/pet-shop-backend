package banner

// BannerItem is the public DTO returned by the banner API.
type BannerItem struct {
	BannerID  int     `json:"bannerID"`
	BannerImg *string `json:"bannerImg,omitempty"`
	Link      *string `json:"link,omitempty"`
	Alt       *string `json:"alt,omitempty"`
}
