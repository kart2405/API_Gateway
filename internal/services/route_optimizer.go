package services

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kart2405/API_Gateway/internal/config"
)

// RouteNode represents a node in the prefix tree
type RouteNode struct {
	children map[string]*RouteNode
	route    *RouteConfig
	isEnd    bool
}

// RouteOptimizer provides optimized route matching using hash-map and prefix-tree
type RouteOptimizer struct {
	hashMap     map[string]*RouteConfig
	prefixTree  *RouteNode
	mutex       sync.RWMutex
	lastUpdated time.Time
}

// NewRouteOptimizer creates a new route optimizer instance
func NewRouteOptimizer() *RouteOptimizer {
	return &RouteOptimizer{
		hashMap: make(map[string]*RouteConfig),
		prefixTree: &RouteNode{
			children: make(map[string]*RouteNode),
		},
	}
}

// BuildOptimizedRoutes builds both hash-map and prefix-tree structures
func (ro *RouteOptimizer) BuildOptimizedRoutes(routes []RouteConfig) {
	ro.mutex.Lock()
	defer ro.mutex.Unlock()

	// Clear existing structures
	ro.hashMap = make(map[string]*RouteConfig)
	ro.prefixTree = &RouteNode{
		children: make(map[string]*RouteNode),
	}

	// Build hash-map for exact matches (O(1) lookup)
	for i := range routes {
		route := &routes[i]
		ro.hashMap[route.ServiceName] = route
	}

	// Build prefix tree for pattern matching
	for i := range routes {
		route := &routes[i]
		ro.insertIntoPrefixTree(route.ServiceName, route)
	}

	ro.lastUpdated = time.Now()
}

// insertIntoPrefixTree inserts a route into the prefix tree
func (ro *RouteOptimizer) insertIntoPrefixTree(serviceName string, route *RouteConfig) {
	node := ro.prefixTree

	// Split service name into segments for better matching
	segments := strings.Split(serviceName, "-")

	for _, segment := range segments {
		if node.children[segment] == nil {
			node.children[segment] = &RouteNode{
				children: make(map[string]*RouteNode),
			}
		}
		node = node.children[segment]
	}

	node.route = route
	node.isEnd = true
}

// FindRouteByHashMap performs O(1) lookup using hash-map
func (ro *RouteOptimizer) FindRouteByHashMap(serviceName string) (*RouteConfig, bool) {
	ro.mutex.RLock()
	defer ro.mutex.RUnlock()

	route, exists := ro.hashMap[serviceName]
	return route, exists
}

// FindRouteByPrefixTree performs pattern matching using prefix tree
func (ro *RouteOptimizer) FindRouteByPrefixTree(serviceName string) (*RouteConfig, bool) {
	ro.mutex.RLock()
	defer ro.mutex.RUnlock()

	segments := strings.Split(serviceName, "-")
	node := ro.prefixTree

	for _, segment := range segments {
		if node.children[segment] == nil {
			return nil, false
		}
		node = node.children[segment]
	}

	if node.isEnd && node.route != nil {
		return node.route, true
	}

	return nil, false
}

// FindRouteOptimized uses the most efficient method to find a route
func (ro *RouteOptimizer) FindRouteOptimized(serviceName string) (*RouteConfig, bool) {
	// Try hash-map first (O(1) lookup)
	if route, exists := ro.FindRouteByHashMap(serviceName); exists {
		return route, true
	}

	// Fallback to prefix tree for pattern matching
	return ro.FindRouteByPrefixTree(serviceName)
}

// GetRouteStats returns statistics about the optimized route structures
func (ro *RouteOptimizer) GetRouteStats() map[string]interface{} {
	ro.mutex.RLock()
	defer ro.mutex.RUnlock()

	return map[string]interface{}{
		"hash_map_size":    len(ro.hashMap),
		"prefix_tree_size": ro.countPrefixTreeNodes(ro.prefixTree),
		"last_updated":     ro.lastUpdated,
	}
}

// countPrefixTreeNodes recursively counts nodes in the prefix tree
func (ro *RouteOptimizer) countPrefixTreeNodes(node *RouteNode) int {
	if node == nil {
		return 0
	}

	count := 1
	for _, child := range node.children {
		count += ro.countPrefixTreeNodes(child)
	}

	return count
}

// BenchmarkRouteLookup performs performance benchmarking
func (ro *RouteOptimizer) BenchmarkRouteLookup(serviceNames []string) map[string]float64 {
	results := make(map[string]float64)

	// Benchmark hash-map lookup
	start := time.Now()
	for _, name := range serviceNames {
		ro.FindRouteByHashMap(name)
	}
	hashMapTime := time.Since(start).Microseconds()
	results["hash_map_microseconds"] = float64(hashMapTime)

	// Benchmark prefix tree lookup
	start = time.Now()
	for _, name := range serviceNames {
		ro.FindRouteByPrefixTree(name)
	}
	prefixTreeTime := time.Since(start).Microseconds()
	results["prefix_tree_microseconds"] = float64(prefixTreeTime)

	// Benchmark optimized lookup
	start = time.Now()
	for _, name := range serviceNames {
		ro.FindRouteOptimized(name)
	}
	optimizedTime := time.Since(start).Microseconds()
	results["optimized_microseconds"] = float64(optimizedTime)

	// Calculate improvement percentage
	if hashMapTime > 0 {
		improvement := ((float64(hashMapTime) - float64(optimizedTime)) / float64(hashMapTime)) * 100
		results["improvement_percentage"] = improvement
	}

	return results
}

// Global route optimizer instance
var GlobalRouteOptimizer = NewRouteOptimizer()

// InitializeRouteOptimizer initializes the global route optimizer with current routes
func InitializeRouteOptimizer() error {
	var routes []RouteConfig
	if err := config.DB.Find(&routes).Error; err != nil {
		return fmt.Errorf("failed to load routes for optimization: %w", err)
	}

	GlobalRouteOptimizer.BuildOptimizedRoutes(routes)
	return nil
}
