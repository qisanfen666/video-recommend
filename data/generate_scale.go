package data

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 与 JSONL 分片文件名 testData0.. 一致；修改时需同步 internal/loader.go
const (
	VideoDataShards    = 10
	UserDataShards     = 5
	BehaviorDataShards = 10
)

const (
	maxVideoCount    = 300_000
	maxUserCount     = 50_000
	maxBehaviorCount = 2_000_000
)

// PhaseStat 单次数据阶段的性能快照（课设展示用）
type PhaseStat struct {
	Name        string  `json:"name"`
	Label       string  `json:"label"`
	Count       int64   `json:"count"`
	DurationMs  int64   `json:"duration_ms"`
	ItemsPerSec float64 `json:"items_per_sec"`
}

// DiskBreakdown data 目录下各 JSONL 体积
type DiskBreakdown struct {
	Videos    int64 `json:"videos"`
	Users     int64 `json:"users"`
	Behaviors int64 `json:"behaviors"`
	Total     int64 `json:"total"`
}

// GenReport 生成阶段汇总（不含 internal 重新加载）
type GenReport struct {
	Phases              []PhaseStat   `json:"phases"`
	DiskBytes           DiskBreakdown `json:"disk_bytes"`
	TotalGenerationMs   int64         `json:"total_generation_ms"`
	VideoCount          int           `json:"video_count"`
	UserCount           int           `json:"user_count"`
	BehaviorCount       int           `json:"behavior_count"`
}

// ReloadStat 从磁盘加载到内存索引的耗时
type ReloadStat struct {
	LoadMs  int64 `json:"load_ms"`
	IndexMs int64 `json:"index_ms"`
	InitMs  int64 `json:"init_ms"`
	TotalMs int64 `json:"total_ms"`
}

// ValidateScale 防止误操作撑爆内存或耗时过长
func ValidateScale(videos, users, behaviors int) error {
	if videos < 1 || videos > maxVideoCount {
		return fmt.Errorf("video_count 须在 1～%d 之间", maxVideoCount)
	}
	if users < 1 || users > maxUserCount {
		return fmt.Errorf("user_count 须在 1～%d 之间", maxUserCount)
	}
	if behaviors < 1 || behaviors > maxBehaviorCount {
		return fmt.Errorf("behavior_count 须在 1～%d 之间", maxBehaviorCount)
	}
	return nil
}

func throughput(count int64, d time.Duration) float64 {
	if d <= 0 {
		return 0
	}
	sec := d.Seconds()
	if sec < 1e-9 {
		return 0
	}
	return float64(count) / sec
}

func jsonlDirBytes(dir string) (int64, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}
	var n int64
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".jsonl") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		n += info.Size()
	}
	return n, nil
}

func measureDisk() (DiskBreakdown, error) {
	var b DiskBreakdown
	var err error
	b.Videos, err = jsonlDirBytes(filepath.Clean("data/videos"))
	if err != nil {
		return b, fmt.Errorf("统计 videos 目录: %w", err)
	}
	b.Users, err = jsonlDirBytes(filepath.Clean("data/users"))
	if err != nil {
		return b, fmt.Errorf("统计 users 目录: %w", err)
	}
	b.Behaviors, err = jsonlDirBytes(filepath.Clean("data/behaviors"))
	if err != nil {
		return b, fmt.Errorf("统计 behaviors 目录: %w", err)
	}
	b.Total = b.Videos + b.Users + b.Behaviors
	return b, nil
}

// GenerateDataset 按规模写 JSONL、构建生成器内存中的分类索引，并返回各阶段耗时与落盘体积。
func GenerateDataset(videoN, userN, behaviorN int) (GenReport, error) {
	if err := ValidateScale(videoN, userN, behaviorN); err != nil {
		return GenReport{}, err
	}
	var phases []PhaseStat
	var totalGen time.Duration

	p1, d1, err := generateVideosWrite(videoN)
	if err != nil {
		return GenReport{}, err
	}
	phases = append(phases, p1)
	totalGen += d1

	p2, d2, err := generateUsersWrite(userN)
	if err != nil {
		return GenReport{}, err
	}
	phases = append(phases, p2)
	totalGen += d2

	tLoad := time.Now()
	LoadData()
	dl := time.Since(tLoad)
	phases = append(phases, PhaseStat{
		Name:        "load_seed_slices",
		Label:       "加载视频/用户到内存（供行为生成与按类目抽样）",
		Count:       int64(videoN + userN),
		DurationMs:  dl.Milliseconds(),
		ItemsPerSec: throughput(int64(videoN+userN), dl),
	})
	totalGen += dl

	p3, d3, err := generateBehaviorsWrite(behaviorN)
	if err != nil {
		return GenReport{}, err
	}
	phases = append(phases, p3)
	totalGen += d3

	disk, err := measureDisk()
	if err != nil {
		return GenReport{}, err
	}

	return GenReport{
		Phases:            phases,
		DiskBytes:         disk,
		TotalGenerationMs: totalGen.Milliseconds(),
		VideoCount:        videoN,
		UserCount:         userN,
		BehaviorCount:     behaviorN,
	}, nil
}
