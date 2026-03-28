package index

import (
	"fmt"
	"time"
	"video-recommend/internal"
)

var (
	VideoToUsers map[int64]map[int64]bool
	UserToVideos map[int64]map[int64]bool
	VideoHeat    map[int64]float64
)

func BuildAll() {
	fmt.Println("[1] 程序启动:", time.Now().Format("15:04:05"))
	buildVideoToUsers()
	fmt.Println("[1] 程序启动:", time.Now().Format("15:04:05"))
	buildUserToVideos()
	fmt.Println("[1] 程序启动:", time.Now().Format("15:04:05"))
	buildVideoHeat()
}

func buildVideoToUsers() {
	VideoToUsers = make(map[int64]map[int64]bool, len(internal.VideoIndex))

	for userID, behaviors := range internal.BehaviorIndex {
		for _, b := range behaviors {
			if b.Action == "watch" || b.Action == "kanwan" || b.Action == "like" {
				if VideoToUsers[b.VideoID] == nil {
					VideoToUsers[b.VideoID] = make(map[int64]bool)
				}
				VideoToUsers[b.VideoID][userID] = true
			}
		}
	}
}

func buildUserToVideos() {
	UserToVideos = make(map[int64]map[int64]bool, len(internal.UserIndex))

	for userID, behaviors := range internal.BehaviorIndex {
		UserToVideos[userID] = make(map[int64]bool)
		for _, b := range behaviors {
			if b.Action == "watch" || b.Action == "kanwan" {
				UserToVideos[userID][b.VideoID] = true
			}
		}
	}
}

func buildVideoHeat() {
	VideoHeat = make(map[int64]float64, len(internal.VideoIndex))

	for vid, video := range internal.VideoIndex {
		heat := video.Heat
		if users, ok := VideoToUsers[vid]; ok {
			heat += float64(len(users)) * 2
		}
		VideoHeat[vid] = heat
	}
}
