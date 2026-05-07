package internal

// TouchVideoHeat 在视频基础热度变更后刷新 Top20 小顶堆（供 live 等包调用，避免 internal 依赖 index）
func TouchVideoHeat(v Video) {
	update(v)
}
