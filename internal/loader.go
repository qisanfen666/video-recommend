package internal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

var (
	VideoIndex    = make(map[int64]Video)
	UserIndex     = make(map[int64]User)
	BehaviorIndex = make(map[int64][]Behavior)
)

func LoadAll() error {
	if err := loadVideos(); err != nil {
		return fmt.Errorf("加载视频数据失败: %w", err)
	}
	if err := loadUsers(); err != nil {
		return fmt.Errorf("加载用户数据失败: %w", err)
	}
	if err := loadBehaviors(); err != nil {
		return fmt.Errorf("加载用户行为数据失败: %w", err)
	}

	fmt.Print("数据加载完成！")

	return nil
}

func loadVideos() error {
	for i := 0; i < 10; i++ {
		fileName := fmt.Sprintf("data/videos/testData%d.jsonl", i)
		file, err := os.Open(fileName)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(file)
		const maxCapacity = 1024 * 1024
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)

		for scanner.Scan() {
			var video Video
			json.Unmarshal(scanner.Bytes(), &video)
			VideoIndex[video.ID] = video
		}

		file.Close()
	}

	return nil
}

func loadUsers() error {
	for i := 0; i < 5; i++ {
		fileName := fmt.Sprintf("data/users/testData%d.jsonl", i)
		file, err := os.Open(fileName)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(file)
		const maxCapacity = 1024 * 1024
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)

		for scanner.Scan() {
			var user User
			json.Unmarshal(scanner.Bytes(), &user)
			UserIndex[user.ID] = user
		}

		file.Close()
	}

	return nil
}

func loadBehaviors() error {
	for i := 0; i < 10; i++ {
		fileName := fmt.Sprintf("data/behaviors/testData%d.jsonl", i)
		file, err := os.Open(fileName)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(file)
		const maxCapacity = 1024 * 1024
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)

		for scanner.Scan() {
			var behavior Behavior
			json.Unmarshal(scanner.Bytes(), &behavior)
			BehaviorIndex[behavior.UserID] = append(BehaviorIndex[behavior.UserID], behavior)
		}

		file.Close()
	}

	return nil
}

func GetVideo(id int64) (Video, bool) {
	video, ok := VideoIndex[id]
	return video, ok
}

func GetUser(id int64) (User, bool) {
	user, ok := UserIndex[id]
	return user, ok
}

func GetBehaviors(id int64) ([]Behavior, bool) {
	behaviors, ok := BehaviorIndex[id]
	return behaviors, ok
}
