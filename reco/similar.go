package reco

import (
	"container/heap"
	"video-recommend/index"
)

type UserSim struct {
	UserID       int64
	Similarity   float64
	CommonVideos int
}

type MaxHeap []UserSim

func (h MaxHeap) Len() int           { return len(h) }
func (h MaxHeap) Less(i, j int) bool { return h[i].Similarity > h[j].Similarity }
func (h MaxHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *MaxHeap) Push(x interface{}) {
	*h = append(*h, x.(UserSim))
}

func (h *MaxHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func FindSimilarUsers(userID int64, k int) []UserSim {
	targetVideos := index.UserToVideos[userID]
	if len(targetVideos) == 0 {
		return nil
	}

	candidates := make(map[int64]int)
	for vid := range targetVideos {
		for otherUser := range index.VideoToUsers[vid] {
			if otherUser != userID {
				candidates[otherUser]++
			}
		}
	}

	//Jaccard相似度
	maxHeap := &MaxHeap{}
	heap.Init(maxHeap)

	for otherUser, common := range candidates {
		otherVideos := index.UserToVideos[otherUser]
		if len(otherVideos) == 0 {
			continue
		}

		union := len(targetVideos) + len(otherVideos) - common
		similarity := float64(common) / float64(union)

		heap.Push(maxHeap, UserSim{
			UserID:       otherUser,
			Similarity:   similarity,
			CommonVideos: common,
		})

		if maxHeap.Len() > k {
			heap.Pop(maxHeap)
		}

	}

	result := make([]UserSim, maxHeap.Len())
	for i := len(result) - 1; i >= 0; i-- {
		result[i] = heap.Pop(maxHeap).(UserSim)
	}

	return result
}
