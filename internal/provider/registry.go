package provider

import (
	"fmt"
	"sync"
)

// registry 全局 provider 注册表
var (
	registry = make(map[string]Provider)
	order    []string
	nextAuto uint64
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
	if _, exists := registry[p.Name()]; !exists {
		order = append(order, p.Name())
	}
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
	names := make([]string, len(order))
	copy(names, order)
	return names
}

// Candidates returns providers in the order used for one request.
// Auto requests start at a round-robin position and then try the remaining providers.
func Candidates(name string) ([]Provider, error) {
	mu.Lock()
	defer mu.Unlock()

	if name != "auto" && name != "" {
		p, ok := registry[name]
		if !ok {
			return nil, fmt.Errorf("unknown provider: %s", name)
		}
		return []Provider{p}, nil
	}
	if len(order) == 0 {
		return nil, fmt.Errorf("no provider registered")
	}

	start := int(nextAuto % uint64(len(order)))
	nextAuto++
	providers := make([]Provider, 0, len(order))
	for i := 0; i < len(order); i++ {
		providers = append(providers, registry[order[(start+i)%len(order)]])
	}
	return providers, nil
}

// GetOrDefault returns the first candidate and is kept for callers that only need one provider.
func GetOrDefault(name string) (Provider, error) {
	providers, err := Candidates(name)
	if err != nil {
		return nil, err
	}
	return providers[0], nil
}

// Clear 清除注册表（仅用于测试）
func Clear() {
	mu.Lock()
	registry = make(map[string]Provider)
	order = nil
	nextAuto = 0
	mu.Unlock()
}
