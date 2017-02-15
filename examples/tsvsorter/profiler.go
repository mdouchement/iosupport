package main

import (
	"net/http"
	"sync"

	"github.com/wblakecaldwell/profiler"
)

var cache = struct {
	sync.RWMutex
	mapping map[string]interface{}
}{
	mapping: make(map[string]interface{}),
}

func ProfilerRun() {
	profiler.AddMemoryProfilingHandlers()
	profiler.RegisterExtraServiceInfoRetriever(mapping)
	profiler.StartProfiling()
	println("Profiler started on http://localhot:6060/profiler/info.html")
	http.ListenAndServe("localhost:6060", nil)
}

func UpdateMapping(key string, value interface{}) {
	cache.Lock()
	cache.mapping[key] = value
	cache.Unlock()
}

func mapping() map[string]interface{} {
	cache.RLock()
	defer cache.RUnlock()

	copy := map[string]interface{}{}
	for k, v := range cache.mapping {
		copy[k] = v
	}
	return copy
}

func String(key string) string {
	if value, ok := mapping()[key].(string); ok {
		return value
	}
	return ""
}

func Bool(key string) bool {
	if value, ok := mapping()[key].(bool); ok {
		return value
	}
	return false
}

func Int(key string) int {
	if value, ok := mapping()[key].(int); ok {
		return value
	}
	return 0
}

func Float64(key string) float64 {
	if value, ok := mapping()[key].(float64); ok {
		return value
	}
	return 0.0
}

func Uint64(key string) uint64 {
	if value, ok := mapping()[key].(uint64); ok {
		return value
	}
	return 0
}
