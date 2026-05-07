package internal

import (
	"container/heap"
	"fmt"
	"sort"
	"time"
)

type Item struct {
	VideoID  int64
	Title    string
	Heat     float64
	Category string
}

type MinHeap []Item

func (h MinHeap) Len() int           { return len(h) }
func (h MinHeap) Less(i, j int) bool { return h[i].Heat < h[j].Heat }
func (h MinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *MinHeap) Push(x interface{}) {
	*h = append(*h, x.(Item))
}

func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

var (
	hotHeap *MinHeap
	hotMap  map[int64]struct{} // 视频是否在 Top20 候选堆中（堆内元素按 VideoID 对齐）
)

func findHeapIndexByVideoID(videoID int64) int {
	for i, item := range *hotHeap {
		if item.VideoID == videoID {
			return i
		}
	}
	return -1
}

// RebuildHotHeap 根据当前 VideoIndex 全量重建小顶堆（用于衰减后或与增量逻辑配合）
func RebuildHotHeap() {
	hotHeap = &MinHeap{}
	heap.Init(hotHeap)
	hotMap = make(map[int64]struct{}, 500000)
	for _, video := range VideoIndex {
		update(video)
	}
}

func Init() {
	fmt.Println("[i1] 程序启动:", time.Now().Format("15:04:05"))
	RebuildHotHeap()
}

func update(video Video) {
	if _, ok := hotMap[video.ID]; ok {
		idx := findHeapIndexByVideoID(video.ID)
		if idx >= 0 {
			(*hotHeap)[idx].Heat = video.Heat
			(*hotHeap)[idx].Title = video.Title
			(*hotHeap)[idx].Category = video.Category
			heap.Fix(hotHeap, idx)
		}
		return
	}

	if hotHeap.Len() < 20 {
		item := Item{
			VideoID:  video.ID,
			Title:    video.Title,
			Heat:     video.Heat,
			Category: video.Category,
		}
		heap.Push(hotHeap, item)
		hotMap[video.ID] = struct{}{}
		return
	}

	if video.Heat > (*hotHeap)[0].Heat {
		old := heap.Pop(hotHeap).(Item)
		delete(hotMap, old.VideoID)

		item := Item{
			VideoID:  video.ID,
			Title:    video.Title,
			Heat:     video.Heat,
			Category: video.Category,
		}
		heap.Push(hotHeap, item)
		hotMap[video.ID] = struct{}{}
	}
}

func GetTop20() []Item {
	result := make([]Item, hotHeap.Len())
	copy(result, *hotHeap)

	sort.Slice(result, func(i, j int) bool {
		return result[i].Heat > result[j].Heat
	})

	return result
}
