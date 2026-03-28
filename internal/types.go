package internal

type Video struct {
	ID         int64    `json:"id"`
	Title      string   `json:"title"`
	Category   string   `json:"category"`
	Tags       []string `json:"tags"`
	Duration   int      `json:"duration"`
	UploadTime int64    `json:"upload_time"`
	Heat       float64  `json:"heat"`
}

type User struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Age         int      `json:"age"`
	Gender      string   `json:"gender"`
	Preferences []string `json:"preferences"`
}

type Behavior struct {
	UserID    int64  `json:"user_id"`
	VideoID   int64  `json:"video_id"`
	Action    string `json:"action"`
	WatchTime int    `json:"watch_time"`
	TimeStamp int64  `json:"timestamp"`
}
