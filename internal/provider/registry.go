package provider

import (
	"fmt"
	"sync"
)

// registry 全局 provider 注册表
var (
	registry = make(map[string]Provider)
	mu       sync.RWMutex
)

// Register 注册一个 provider
func Register(p Provider) {
	if p == nil {
		panic("provider: cannot register nil provider")
	}
	if p.Name() == "" {
		panic("provider: cannot register provider with empty name")
	}
	mu.Lock()
	registry[p.Name()] = p
	mu.Unlock()
}

// Get 根据名称获取 provider
func Get(name string) (Provider, bool) {
	mu.RLock()
	defer mu.RUnlock()
	p, ok := registry[name]
	return p, ok
}

// List 返回所有已注册的 provider 名称
func List() []string {
	mu.RLock()
	defer mu.RUnlock()
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

// GetOrDefault 根据名称获取 provider，name 为空或 "auto" 时返回第一个已注册的
func GetOrDefault(name string) (Provider, error) {
	if name == "auto" || name == "" {
		mu.RLock()
		defer mu.RUnlock()
		for _, p := range registry {
			return p, nil
		}
		return nil, fmt.Errorf("no provider registered")
	}
	p, ok := Get(name)
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
	return p, nil
}

// Clear 清除注册表（仅用于测试）
func Clear() {
	mu.Lock()
	registry = make(map[string]Provider)
	mu.Unlock()
}