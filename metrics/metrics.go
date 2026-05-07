package metrics

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Record struct {
	Path       string
	Method     string
	Duration   time.Duration
	StatusCode int
	Timestamp  time.Time
}

type Stats struct {
	Count       int64
	TotalTime   time.Duration
	AvgTime     time.Duration
	MaxTime     time.Duration
	MinTime     time.Duration
	ErrorCount  int64
	SuccessRate float64
}

// 全局存储
var (
	ringBuffer [1000]Record
	ringIndex  int
	ringMutex  sync.RWMutex

	aggStats = make(map[string]*Stats)
	aggMutex sync.RWMutex
)

// routeKey 用于统计聚合：同一 Gin 路由模板下的不同实参（如 /api/videos/1 与 /api/videos/2）
// 会合并为同一条键，例如 /api/videos/:id。未匹配到路由时退回原始 URL 路径。
func routeKey(c *gin.Context) string {
	if p := c.FullPath(); p != "" {
		return p
	}
	return c.Request.URL.Path
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		record := Record{
			Path:       routeKey(c),
			Method:     c.Request.Method,
			Duration:   duration,
			StatusCode: c.Writer.Status(),
			Timestamp:  start,
		}

		ringMutex.Lock()
		ringBuffer[ringIndex] = record
		ringIndex = (ringIndex + 1) % 1000
		ringMutex.Unlock()

		updateStats(record)
	}
}

func updateStats(r Record) {
	aggMutex.Lock()
	defer aggMutex.Unlock()

	stats, ok := aggStats[r.Path]
	if !ok {
		stats = &Stats{
			MinTime: r.Duration,
		}
		aggStats[r.Path] = stats
	}

	stats.Count++
	stats.TotalTime += r.Duration
	stats.AvgTime = stats.TotalTime / time.Duration(stats.Count)

	if r.Duration > stats.MaxTime {
		stats.MaxTime = r.Duration
	}
	if r.Duration < stats.MinTime {
		stats.MinTime = r.Duration
	}

	if r.StatusCode >= 400 {
		stats.ErrorCount++
	}
	stats.SuccessRate = float64(stats.Count-stats.ErrorCount) / float64(stats.Count) * 100
}

// 获取最近n条请求
func GetRecent(n int) []Record {
	ringMutex.RLock()
	defer ringMutex.RUnlock()

	result := make([]Record, 0, n)
	for i := 0; i < n && i < 1000; i++ {
		idx := (ringIndex - 1 - i + 1000) % 1000
		if !ringBuffer[idx].Timestamp.IsZero() {
			result = append(result, ringBuffer[idx])
		}
	}
	return result
}

func GetAllStats() map[string]Stats {
	aggMutex.RLock()
	defer aggMutex.RUnlock()

	result := make(map[string]Stats)
	for k, v := range aggStats {
		result[k] = *v
	}

	return result
}
