package data

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"video-recommend/data/categories"
)

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
	Action    string `json:"action"` // click/like/share/complete
	WatchTime int    `json:"watch_time"`
	TimeStamp int64  `json:"timestamp"`
}

func GenerateVideos() {
	const testCount = 500000

	os.MkdirAll("data/videos", 0755)

	const workerCount = 10
	taskChan := make(chan int, testCount)

	go func() {
		for i := 0; i < testCount; i++ {
			taskChan <- i
		}
		close(taskChan)
	}()

	fmt.Println("正在生成", testCount, "条测试数据...")

	var wg sync.WaitGroup
	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func(workID int) {
			defer wg.Done()

			outputFile := fmt.Sprintf("data/videos/testData%d.jsonl", workID)
			file, err := os.Create(outputFile)
			if err != nil {
				panic(err)
			}
			defer file.Close()
			encoder := json.NewEncoder(file)

			for id := range taskChan {
				video := generateDateVideos(id)
				encoder.Encode(video)
			}
		}(w)
	}

	wg.Wait()

	fmt.Println("测试数据生成完成！")
}

func GenerateUsers() {
	const testCount = 100000

	os.MkdirAll("data/users", 0755)

	const workerCount = 5

	taskChan := make(chan int, testCount)

	go func() {
		for i := 0; i < testCount; i++ {
			taskChan <- i
		}
		close(taskChan)
	}()

	fmt.Println("正在生成", testCount, "条用户测试数据...")

	var wg sync.WaitGroup
	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func(workID int) {
			defer wg.Done()

			outputFile := fmt.Sprintf("data/users/testData%d.jsonl", workID)
			file, err := os.Create(outputFile)
			if err != nil {
				panic(err)
			}
			defer file.Close()
			encoder := json.NewEncoder(file)

			for id := range taskChan {
				user := generateDateUsers(id)
				encoder.Encode(user)
			}
		}(w)
	}

	wg.Wait()

	fmt.Println("用户测试数据生成完成！")
}

func GenerateBehaviors() {
	os.MkdirAll("data/behaviors", 0755)

	const testCount = 1000000
	const workerCount = 10

	perWorker := testCount / workerCount

	fmt.Println("正在生成", testCount, "条用户行为测试数据...")

	var wg sync.WaitGroup
	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func(workID int) {
			defer wg.Done()

			outputFile := fmt.Sprintf("data/behaviors/testData%d.jsonl", workID)
			file, err := os.Create(outputFile)
			if err != nil {
				panic(err)
			}
			defer file.Close()
			encoder := json.NewEncoder(file)

			for i := 0; i < perWorker; i++ {
				behavior := generateOneBehavior()
				encoder.Encode(behavior)
			}
		}(w)
	}

	wg.Wait()

	fmt.Println("用户行为测试数据生成完成！")
}

func generateDateVideos(id int) Video {
	categories := randomCategory()
	video := Video{
		ID:         int64(id),
		Title:      randomTitle(),
		Category:   categories,
		Tags:       randomTags(categories),
		Duration:   randomDuration(),
		UploadTime: randomUploadTime(),
		Heat:       50.0 + rand.Float64()*50.0, // 50 ~ 100
	}
	return video
}

func generateDateUsers(id int) User {
	user := User{
		ID:          int64(id),
		Name:        randomName(),
		Age:         rand.Intn(80) + 12,
		Gender:      randomGender(),
		Preferences: randomPreferences(),
	}

	return user
}

func generateOneBehavior() Behavior {
	user := users[rand.Intn(len(users))]
	video := selectVideo(user)

	action, watchTime := behaviorFunnel(video)

	behavior := Behavior{
		UserID:    user.ID,
		VideoID:   video.ID,
		Action:    action,
		WatchTime: watchTime,
		TimeStamp: randomBehaviorTime(),
	}

	return behavior
}

func randomTitle() string {
	return randomPrefix() + randomMiddle() + randomsuffix()
}

func randomPrefix() string {
	prefix := categories.Prefixes[rand.Intn(len(categories.Prefixes))]
	return prefix
}

func randomMiddle() string {
	middle := categories.Middles[rand.Intn(len(categories.Middles))]
	return middle
}

func randomsuffix() string {
	suffix := categories.Suffixes[rand.Intn(len(categories.Suffixes))]
	return suffix
}

func randomCategory() string {
	category := categories.Categories[rand.Intn(len(categories.Categories))]
	return category
}

func randomTags(category string) []string {
	nums := rand.Intn(5) + 1
	var tags []string

	for i := 0; i < nums; i++ {
		randomIndex := rand.Intn(len(categories.CategoryTags[category]) - 1)
		tags = append(tags, categories.CategoryTags[category][randomIndex])
	}

	return tags
}

func randomDuration() int {
	return rand.Intn(600) + 30 // 30秒到10分钟
}

func randomUploadTime() int64 {
	return rand.Int63n(1672502400) + 1735689600 // 2023-01-01到2026-01-01
}

func randomName() string {
	return randomAdjective() + randomNoun() + strconv.Itoa(rand.Intn(10000)+1)
}

func randomAdjective() string {
	return categories.Adjectives[rand.Intn(len(categories.Adjectives))]
}

func randomNoun() string {
	return categories.Nouns[rand.Intn(len(categories.Nouns))]
}

func randomGender() string {
	return categories.Genders[rand.Intn(len(categories.Genders))]
}

func randomPreferences() []string {
	nums := rand.Intn(3) + 1
	var preferences []string

	for i := 0; i < nums; i++ {
		preferences = append(preferences, categories.Categories[rand.Intn(len(categories.Categories))])
	}

	return preferences
}

func randomBehaviorTime() int64 {
	return rand.Int63n(1672502400) + 1735689600
}

func selectVideo(user User) Video {
	if rand.Float64() < 0.7 && len(user.Preferences) > 0 {
		pref := user.Preferences[rand.Intn(len(user.Preferences))]
		candidates := videoByCat[pref]
		if len(candidates) > 0 {
			return candidates[rand.Intn(len(candidates))]
		}
	}

	//return weightedByHeat(videos)
	return videos[rand.Intn(len(videos))]
}

// func weightedByHeat(videos []Video) Video {
// 	totalHeat := 0.0
// 	for _, video := range videos {
// 		totalHeat += video.Heat
// 	}

// 	r := rand.Float64() * totalHeat
// 	cum := 0.0
// 	for _, video := range videos {
// 		cum += video.Heat
// 		if r < cum {
// 			return video
// 		}
// 	}
// 	return videos[len(videos)-1]
// }

func behaviorFunnel(video Video) (string, int) {
	chance := rand.Float64()
	if chance < 0.4 {
		return "点击", 0
	}

	watchTime := rand.Intn(video.Duration)
	action := "观看"

	if video.Duration-watchTime < 10 {
		action = "看完"
	} else if chance < 0.7 {
		action = "点赞"
	} else if chance < 0.8 {
		action = "转发"
	}

	return action, watchTime
}

var (
	videos     []Video
	users      []User
	videoByCat = map[string][]Video{}
)

func LoadData() {
	fmt.Println("正在加载数据...")
	loadVideos()
	loadUsers()
	buildCatIndex()
	fmt.Println("加载完成")
}

func loadVideos() {
	videos = make([]Video, 0, 500000)
	for i := 0; i < 10; i++ {
		fileName := fmt.Sprintf("data/videos/testData%d.jsonl", i)
		file, err := os.Open(fileName)
		if err != nil {
			panic(err)
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			var video Video
			json.Unmarshal(scanner.Bytes(), &video)
			videos = append(videos, video)
		}

		file.Close()
	}
}

func loadUsers() {
	users = make([]User, 0, 100000)
	for i := 0; i < 5; i++ {
		fileName := fmt.Sprintf("data/users/testData%d.jsonl", i)
		file, err := os.Open(fileName)
		if err != nil {
			panic(err)
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			var user User
			json.Unmarshal(scanner.Bytes(), &user)
			users = append(users, user)
		}
		file.Close()
	}
}

func buildCatIndex() {
	videoByCat = make(map[string][]Video)
	for _, video := range videos {
		videoByCat[video.Category] = append(videoByCat[video.Category], video)
	}
}
