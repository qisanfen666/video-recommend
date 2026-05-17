package domain

type Recommendation struct {
	VideoID  int64   `json:"video_id"`
	Title    string  `json:"title"`
	CoverURL string  `json:"cover_url"`
	Category string  `json:"category"`
	Score    float64 `json:"score"`
	Reason   string  `json:"reason"`
	Rank     int     `json:"rank"`
}

type RecommendRequest struct {
	VideoCount int    `json:"video_count" binding:"min=1,max=100"`
	Context    string `json:"context"`
}

type SimilarUser struct {
	UserID       int64   `json:"user_id"`
	Username     string  `json:"username"`
	Similarity   float64 `json:"similarity"`
	CommonVideos int     `json:"common_videos"`
}

type UserSimilarityRequest struct {
	Limit int `json:"limit" binding:"min=1,max=50"`
}

type RecommendResponse struct {
	Recommendations []Recommendation `json:"recommendations"`
	Algorithm       string           `json:"algorithm"`
	Count           int              `json:"count"`
}

type HotRankRequest struct {
	Limit    int    `form:"limit" binding:"min=1,max=100"`
	Category string `form:"category"`
}

type HotRankResponse struct {
	Ranks     []HotVideo `json:"ranks"`
	UpdatedAt string     `json:"updated_at"`
}
