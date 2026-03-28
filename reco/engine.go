package reco

import (
	"container/heap"
	"video-recommend/index"
	"video-recommend/internal"
)

type Recommend struct {
	VideoID int64
	Score   float64
	Reason  string
}

type RecoHeap []Recommend

func (h RecoHeap) Len() int           { return len(h) }
func (h RecoHeap) Less(i, j int) bool { return h[i].Score > h[j].Score }
func (h RecoHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *RecoHeap) Push(x interface{}) {
	*h = append(*h, x.(Recommend))
}
func (h *RecoHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func RecommendForUser(userID int64, k int) []Recommend {
	similarUsers := FindSimilarUsers(userID, 50)
	if len(similarUsers) == 0 {
		return getHotRecommendations(k)
	}

	targetWatched := index.UserToVideos[userID]
	candidates := make(map[int64]float64)

	for _, sim := range similarUsers {
		otherVideos := index.UserToVideos[sim.UserID]
		for vid := range otherVideos {
			if targetWatched[vid] {
				continue
			}
			candidates[vid] += sim.Similarity
		}
	}

	recoHeap := &RecoHeap{}
	heap.Init(recoHeap)

	for vid, score := range candidates {
		heap.Push(recoHeap, Recommend{
			VideoID: vid,
			Score:   score,
			Reason:  "reason1",
		})
		if recoHeap.Len() > k {
			heap.Pop(recoHeap)
		}
	}

	result := make([]Recommend, recoHeap.Len())
	for i := len(result) - 1; i >= 0; i-- {
		result[i] = heap.Pop(recoHeap).(Recommend)
	}

	return result
}

func getHotRecommendations(k int) []Recommend {
	hotVideos := internal.GetTop20()

	result := make([]Recommend, 0, k)
	for i := k - 1; i >= 0; i-- {
		result[i] = Recommend{
			VideoID: hotVideos[i].VideoID,
			Score:   hotVideos[i].Heat,
			Reason:  "reason2",
		}
	}

	return result
}
