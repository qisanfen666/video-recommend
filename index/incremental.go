package index

import "video-recommend/internal"

func ensureVideoHeat() {
	if VideoHeat == nil {
		VideoHeat = make(map[int64]float64)
	}
}

// ApplyBehavior 将单条行为增量合并进倒排结构（与 builder 中规则一致）
func ApplyBehavior(b internal.Behavior) {
	if VideoToUsers == nil {
		VideoToUsers = make(map[int64]map[int64]bool)
	}
	if UserToVideos == nil {
		UserToVideos = make(map[int64]map[int64]bool)
	}
	ensureVideoHeat()

	if b.Action == "watch" || b.Action == "kanwan" || b.Action == "like" {
		if VideoToUsers[b.VideoID] == nil {
			VideoToUsers[b.VideoID] = make(map[int64]bool)
		}
		VideoToUsers[b.VideoID][b.UserID] = true
	}
	if b.Action == "watch" || b.Action == "kanwan" {
		if UserToVideos[b.UserID] == nil {
			UserToVideos[b.UserID] = make(map[int64]bool)
		}
		UserToVideos[b.UserID][b.VideoID] = true
	}
	recomputeVideoHeatEntry(b.VideoID)
}

func recomputeVideoHeatEntry(vid int64) {
	v, ok := internal.VideoIndex[vid]
	if !ok {
		return
	}
	heat := v.Heat
	if users, ok := VideoToUsers[vid]; ok {
		heat += float64(len(users)) * 2
	}
	VideoHeat[vid] = heat
}
