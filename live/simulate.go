package live

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"time"
	"video-recommend/index"
	"video-recommend/internal"
)

// SimulateRequest 动态注入行为（内存 + 可选落盘）
type SimulateRequest struct {
	Count         int     `json:"count"`
	HeatDecay     float64 `json:"heat_decay"`
	Persist       bool    `json:"persist"`
	Seed          int64   `json:"seed"`
	TimestampBase int64   `json:"timestamp_base"`
}

// SimulateReport 用于前端展示「动态」效果
type SimulateReport struct {
	Requested       int     `json:"requested"`
	Applied         int     `json:"applied"`
	HeatDecayFactor float64 `json:"heat_decay_factor"`
	TimeWindow      struct {
		Min int64 `json:"min_ts"`
		Max int64 `json:"max_ts"`
	} `json:"time_window"`
	HotTop5Before []HotSnap            `json:"hot_top5_before"`
	HotTop5After  []HotSnap            `json:"hot_top5_after"`
	SampleNew     []internal.Behavior  `json:"sample_new_behaviors"`
	UserPrefShift []PrefShift          `json:"user_preference_shifts"`
	DurationMs    int64                `json:"duration_ms"`
}

type HotSnap struct {
	VideoID  int64   `json:"video_id"`
	Title    string  `json:"title"`
	Heat     float64 `json:"heat"`
	Category string  `json:"category"`
}

type PrefShift struct {
	UserID     int64    `json:"user_id"`
	Name       string   `json:"name"`
	Before     []string `json:"preferences_before"`
	After      []string `json:"preferences_after"`
	Behaviors  int      `json:"behavior_count_after"`
	NewInBatch int      `json:"new_behaviors_in_batch"`
}

func snapHotTop5() []HotSnap {
	items := internal.GetTop20()
	n := 5
	if len(items) < n {
		n = len(items)
	}
	out := make([]HotSnap, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, HotSnap{
			VideoID:  items[i].VideoID,
			Title:    items[i].Title,
			Heat:     items[i].Heat,
			Category: items[i].Category,
		})
	}
	return out
}

func maxStoredBehaviorTS() int64 {
	var m int64
	for _, bs := range internal.BehaviorIndex {
		for _, b := range bs {
			if b.TimeStamp > m {
				m = b.TimeStamp
			}
		}
	}
	return m
}

func heatBump(action string) float64 {
	switch action {
	case "click":
		return 0.15
	case "watch":
		return 1.2
	case "kanwan":
		return 4.0
	case "like":
		return 2.5
	case "share":
		return 1.0
	default:
		return 0
	}
}

func synthAction(rng *rand.Rand, video internal.Video) (string, int) {
	chance := rng.Float64()
	if chance < 0.38 {
		return "click", 0
	}
	if video.Duration <= 0 {
		return "click", 0
	}
	watchTime := rng.Intn(video.Duration)
	action := "watch"
	if video.Duration-watchTime < 10 {
		action = "kanwan"
	} else if chance < 0.68 {
		action = "like"
	} else if chance < 0.78 {
		action = "share"
	}
	return action, watchTime
}

func buildVideoCategoryIndex() (byCat map[string][]int64, all []int64) {
	byCat = make(map[string][]int64)
	for id, v := range internal.VideoIndex {
		all = append(all, id)
		byCat[v.Category] = append(byCat[v.Category], id)
	}
	return byCat, all
}

func pickVideo(rng *rand.Rand, u internal.User, byCat map[string][]int64, all []int64) (internal.Video, bool) {
	if len(all) == 0 {
		return internal.Video{}, false
	}
	if rng.Float64() < 0.65 && len(u.Preferences) > 0 {
		pref := u.Preferences[rng.Intn(len(u.Preferences))]
		cands := byCat[pref]
		if len(cands) > 0 {
			vid := cands[rng.Intn(len(cands))]
			v, ok := internal.VideoIndex[vid]
			return v, ok
		}
	}
	vid := all[rng.Intn(len(all))]
	v, ok := internal.VideoIndex[vid]
	return v, ok
}

// SimulateLiveBehaviors 生成并应用一批新行为（依赖 internal + index，避免与 index 循环引用）
func SimulateLiveBehaviors(req SimulateRequest) (SimulateReport, error) {
	const maxBatch = 50_000
	if req.Count < 1 || req.Count > maxBatch {
		return SimulateReport{}, fmt.Errorf("count 须在 1～%d 之间", maxBatch)
	}
	if req.HeatDecay > 0 && req.HeatDecay < 1 {
		// ok
	} else {
		req.HeatDecay = 1
	}

	start := time.Now()
	rep := SimulateReport{
		Requested:       req.Count,
		HeatDecayFactor:   req.HeatDecay,
	}
	rep.HotTop5Before = snapHotTop5()

	rng := rand.New(rand.NewSource(req.Seed))
	if req.Seed == 0 {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	tsBase := req.TimestampBase
	if tsBase <= 0 {
		tsBase = maxStoredBehaviorTS()
		if tsBase <= 0 {
			tsBase = time.Now().Unix()
		}
	}

	byCat, allIDs := buildVideoCategoryIndex()
	if len(allIDs) == 0 || len(internal.UserIndex) == 0 {
		return SimulateReport{}, fmt.Errorf("无视频或用户，无法模拟")
	}

	userIDs := make([]int64, 0, len(internal.UserIndex))
	for id := range internal.UserIndex {
		userIDs = append(userIDs, id)
	}

	previewSet := make(map[int64]struct{})
	beforePrefs := make(map[int64][]string)
	newPerUser := make(map[int64]int)

	behaviors := make([]internal.Behavior, 0, req.Count)
	tsMin := int64(1<<62 - 1)
	tsMax := int64(0)
	cursor := tsBase

	for i := 0; i < req.Count; i++ {
		u := internal.UserIndex[userIDs[rng.Intn(len(userIDs))]]
		v, ok := pickVideo(rng, u, byCat, allIDs)
		if !ok {
			continue
		}
		act, wt := synthAction(rng, v)
		cursor += 1 + rng.Int63n(90)
		b := internal.Behavior{
			UserID:    u.ID,
			VideoID:   v.ID,
			Action:    act,
			WatchTime: wt,
			TimeStamp: cursor,
		}
		behaviors = append(behaviors, b)
		if len(previewSet) < 8 {
			if _, seen := previewSet[u.ID]; !seen {
				pu := internal.UserIndex[u.ID]
				beforePrefs[u.ID] = append([]string(nil), pu.Preferences...)
				previewSet[u.ID] = struct{}{}
			}
		}
		if b.TimeStamp < tsMin {
			tsMin = b.TimeStamp
		}
		if b.TimeStamp > tsMax {
			tsMax = b.TimeStamp
		}
	}

	if req.HeatDecay < 1 {
		for id := range internal.VideoIndex {
			v := internal.VideoIndex[id]
			v.Heat *= req.HeatDecay
			internal.VideoIndex[id] = v
		}
		internal.RebuildHotHeap()
	}

	var persistBuf []byte
	if req.Persist {
		persistBuf = make([]byte, 0, len(behaviors)*64)
	}

	for _, b := range behaviors {
		internal.BehaviorIndex[b.UserID] = append(internal.BehaviorIndex[b.UserID], b)
		newPerUser[b.UserID]++
		index.ApplyBehavior(b)

		v, ok := internal.VideoIndex[b.VideoID]
		if ok {
			delta := heatBump(b.Action)
			if delta > 0 {
				v.Heat += delta
				internal.VideoIndex[b.VideoID] = v
				internal.TouchVideoHeat(v)
			}
		}
		refreshUserPreferences(b.UserID)

		if req.Persist {
			line, err := json.Marshal(b)
			if err == nil {
				persistBuf = append(persistBuf, line...)
				persistBuf = append(persistBuf, '\n')
			}
		}
	}

	if req.Persist && len(persistBuf) > 0 {
		_ = os.MkdirAll("data/behaviors", 0755)
		f, err := os.OpenFile("data/behaviors/stream.jsonl", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			_, _ = f.Write(persistBuf)
			_ = f.Close()
		}
	}

	rep.Applied = len(behaviors)
	rep.TimeWindow.Min = tsMin
	rep.TimeWindow.Max = tsMax
	if len(behaviors) > 0 {
		rep.SampleNew = behaviors
		if len(behaviors) > 5 {
			rep.SampleNew = behaviors[:5]
		}
	}
	rep.HotTop5After = snapHotTop5()

	shifts := make([]PrefShift, 0, len(previewSet))
	for uid := range previewSet {
		u := internal.UserIndex[uid]
		shifts = append(shifts, PrefShift{
			UserID:     uid,
			Name:       u.Name,
			Before:     beforePrefs[uid],
			After:      append([]string(nil), u.Preferences...),
			Behaviors:  len(internal.BehaviorIndex[uid]),
			NewInBatch: newPerUser[uid],
		})
	}
	sort.Slice(shifts, func(i, j int) bool { return shifts[i].UserID < shifts[j].UserID })
	rep.UserPrefShift = shifts

	rep.DurationMs = time.Since(start).Milliseconds()
	return rep, nil
}

func refreshUserPreferences(userID int64) {
	u, ok := internal.UserIndex[userID]
	if !ok {
		return
	}
	counts := make(map[string]int)
	for _, b := range internal.BehaviorIndex[userID] {
		if b.Action != "watch" && b.Action != "kanwan" && b.Action != "like" {
			continue
		}
		v, vok := internal.VideoIndex[b.VideoID]
		if !vok {
			continue
		}
		w := 1
		switch b.Action {
		case "kanwan":
			w = 4
		case "like":
			w = 2
		}
		counts[v.Category] += w
	}
	u.Preferences = topNPreferenceKeys(counts, 3)
	internal.UserIndex[userID] = u
}

func topNPreferenceKeys(m map[string]int, n int) []string {
	type kv struct {
		k string
		v int
	}
	s := make([]kv, 0, len(m))
	for k, v := range m {
		s = append(s, kv{k, v})
	}
	sort.Slice(s, func(i, j int) bool {
		if s[i].v != s[j].v {
			return s[i].v > s[j].v
		}
		return s[i].k < s[j].k
	})
	out := make([]string, 0, n)
	for i := 0; i < n && i < len(s); i++ {
		out = append(out, s[i].k)
	}
	return out
}
