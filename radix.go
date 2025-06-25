package nova

import (
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"hash/fnv"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	// 每个方法的分片数量
	methodShardCount = 256
	// 缓存大小
	cacheSize = 10000
	// 预分配容量
	preAllocSize = 4
	// 热点缓存大小
	hotCacheSize = 2000
	// 缓存预热阈值
	hotCacheThreshold = 50
	// 缓存分片数
	cacheShardCount = 16
	// 缓存清理间隔
	cacheCleanupInterval = 5 * time.Minute
	// 节点池大小
	nodePoolSize = 1000
	// 参数map池大小
	paramsPoolSize = 1000
	// 并发度限制
	maxConcurrent = 10000
)

// RadixNode 表示 Radix 树中的一个节点
type RadixNode struct {
	path      string                 // 当前节点的路径
	children  map[string]*RadixNode  // 子节点
	handlers  map[string]HandlerFunc // HTTP方法到处理函数的映射
	params    []string               // 参数名列表
	wildcard  bool                   // 是否为通配符节点
	paramName string                 // 参数名称
	// 节点级别的读写锁
	mu sync.RWMutex
}

// 分片缓存
type shardedCache struct {
	shards []*cacheShard
}

// 缓存分片
type cacheShard struct {
	items unsafe.Pointer // *map[string]*cacheItem
	count uint64
	mu    sync.RWMutex
}

// 分片结构
type shard struct {
	tree  *RadixNode
	cache *lru.Cache
	// 热点计数器
	hotCounter uint64
}

// 参数对象池
var paramsPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]string, 8) // 预分配容量，减少扩容
	},
}

// 多级Context池
var ctxPools = [...]sync.Pool{
	{New: func() interface{} { return &Context{Params: make(map[string]string, 4)} }},
	{New: func() interface{} { return &Context{Params: make(map[string]string, 8)} }},
	{New: func() interface{} { return &Context{Params: make(map[string]string, 16)} }},
}

func getContext(paramCount int) *Context {
	switch {
	case paramCount <= 4:
		return ctxPools[0].Get().(*Context)
	case paramCount <= 8:
		return ctxPools[1].Get().(*Context)
	case paramCount <= 16:
		return ctxPools[2].Get().(*Context)
	default:
		return &Context{Params: make(map[string]string, paramCount)}
	}
}

func putContext(ctx *Context) {
	for k := range ctx.Params {
		delete(ctx.Params, k)
	}
	l := len(ctx.Params)
	switch {
	case l <= 4:
		ctxPools[0].Put(ctx)
	case l <= 8:
		ctxPools[1].Put(ctx)
	case l <= 16:
		ctxPools[2].Put(ctx)
	}
}

// SIMD-like批量分支匹配
func simdMatch(staticBranches []string, target string) int {
	for i := 0; i < len(staticBranches); i += 4 {
		if i+3 < len(staticBranches) {
			if staticBranches[i] == target {
				return i
			}
			if staticBranches[i+1] == target {
				return i + 1
			}
			if staticBranches[i+2] == target {
				return i + 2
			}
			if staticBranches[i+3] == target {
				return i + 3
			}
		} else {
			for j := i; j < len(staticBranches); j++ {
				if staticBranches[j] == target {
					return j
				}
			}
		}
	}
	return -1
}

// RadixNode查找逻辑集成SIMD-like批量分支匹配
func (n *RadixNode) findChildSIMD(part string) *RadixNode {
	staticKeys := make([]string, 0, len(n.children))
	staticNodes := make([]*RadixNode, 0, len(n.children))
	for k, v := range n.children {
		if !v.wildcard {
			staticKeys = append(staticKeys, k)
			staticNodes = append(staticNodes, v)
		}
	}
	idx := simdMatch(staticKeys, part)
	if idx >= 0 {
		return staticNodes[idx]
	}
	// 参数分支
	for _, v := range n.children {
		if v.wildcard {
			return v
		}
	}
	return nil
}

// findRoute 在 Radix 树中查找路由
func (n *RadixNode) findRoute(method, path string) (HandlerFunc, map[string]string, bool) {
	// 分割路径
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return nil, nil, false
	}

	// 从对象池获取参数map
	params := paramsPool.Get().(map[string]string)
	defer paramsPool.Put(params)
	// 清空参数map
	for k := range params {
		delete(params, k)
	}

	// 从根节点开始遍历
	current := n
	for i, part := range parts {
		if part == "" {
			continue
		}

		// 使用节点级别的读锁
		current.mu.RLock()
		child, exists := current.children[part]
		if !exists {
			// 尝试匹配参数节点
			for _, child := range current.children {
				if child.wildcard {
					params[child.paramName] = part
					exists = true
					break
				}
			}
			current.mu.RUnlock()
			if !exists {
				return nil, nil, false
			}
		} else {
			current.mu.RUnlock()
		}

		// 如果是最后一个部分，返回处理函数
		if i == len(parts)-1 {
			child.mu.RLock()
			handler, exists := child.handlers[method]
			child.mu.RUnlock()
			if !exists {
				return nil, nil, false
			}
			return handler, params, true
		}

		current = child
	}

	return nil, nil, false
}

// 路由缓存项
type cacheItem struct {
	handler HandlerFunc
	params  map[string]string
	// 访问计数
	accessCount uint64
	// 最后访问时间
	lastAccess int64
}

// 无锁热点缓存
type hotCache struct {
	// 使用分片缓存
	shards *shardedCache
	// 使用原子计数器
	count uint64
}

// Router 路由管理器
type Router struct {
	methodShards map[string][]*shard // 按HTTP方法分片
	hotCache     *hotCache           // 无锁热点缓存
	// 热点路由计数器
	hotCounters sync.Map
	// 节点对象池
	nodePool *sync.Pool
	// 参数map对象池
	paramsPool *sync.Pool
	// 缓存清理定时器
	cleanupTicker *time.Ticker
	// 停止信号
	stopChan chan struct{}
}

// NewRouter 创建新的路由管理器
func NewRouter() *Router {
	// 创建分片缓存
	shards := make([]*cacheShard, cacheShardCount)
	for i := range shards {
		shards[i] = &cacheShard{
			items: unsafe.Pointer(&map[string]*cacheItem{}),
		}
	}

	// 创建节点对象池
	nodePool := &sync.Pool{
		New: func() interface{} {
			return &RadixNode{
				children: make(map[string]*RadixNode, preAllocSize),
				handlers: make(map[string]HandlerFunc),
			}
		},
	}

	// 创建参数map对象池
	paramsPool := &sync.Pool{
		New: func() interface{} {
			return make(map[string]string, preAllocSize)
		},
	}

	// 预分配对象池
	for i := 0; i < nodePoolSize; i++ {
		nodePool.Put(nodePool.New())
	}
	for i := 0; i < paramsPoolSize; i++ {
		paramsPool.Put(paramsPool.New())
	}

	router := &Router{
		methodShards: make(map[string][]*shard),
		hotCache: &hotCache{
			shards: &shardedCache{shards: shards},
		},
		nodePool:      nodePool,
		paramsPool:    paramsPool,
		cleanupTicker: time.NewTicker(cacheCleanupInterval),
		stopChan:      make(chan struct{}),
	}

	// 启动缓存清理协程
	go router.cleanupCache()

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}
	for _, m := range methods {
		shards := make([]*shard, methodShardCount)
		for i := range shards {
			cache, err := lru.New(cacheSize / methodShardCount)
			if err != nil {
				panic(fmt.Sprintf("failed to create shard cache: %v", err))
			}
			shards[i] = &shard{
				tree:  router.getNode(),
				cache: cache,
			}
		}
		router.methodShards[m] = shards
	}

	return router
}

// 获取节点对象
func (r *Router) getNode() *RadixNode {
	return r.nodePool.Get().(*RadixNode)
}

// 回收节点对象
func (r *Router) putNode(node *RadixNode) {
	// 清空节点数据
	node.path = ""
	node.children = make(map[string]*RadixNode, preAllocSize)
	node.handlers = make(map[string]HandlerFunc)
	node.params = nil
	node.wildcard = false
	node.paramName = ""
	r.nodePool.Put(node)
}

// 获取参数map
func (r *Router) getParams() map[string]string {
	return r.paramsPool.Get().(map[string]string)
}

// 回收参数map
func (r *Router) putParams(params map[string]string) {
	for k := range params {
		delete(params, k)
	}
	r.paramsPool.Put(params)
}

// 缓存清理
func (r *Router) cleanupCache() {
	for {
		select {
		case <-r.cleanupTicker.C:
			r.hotCache.cleanup()
		case <-r.stopChan:
			r.cleanupTicker.Stop()
			return
		}
	}
}

// 预热缓存
func (r *Router) WarmupCache(paths []string) {
	for _, path := range paths {
		// 预热热点缓存
		r.hotCache.warmup(path)
		// 预热分片缓存
		shardIdx := getShard(path)
		for _, shards := range r.methodShards {
			shards[shardIdx].cache.Add(path, nil)
		}
	}
}

// 分片缓存方法
func (c *shardedCache) get(key string) *cacheItem {
	shardIdx := getShard(key) % cacheShardCount
	shard := c.shards[shardIdx]
	shard.mu.RLock()
	defer shard.mu.RUnlock()
	items := *(*map[string]*cacheItem)(atomic.LoadPointer(&shard.items))
	return items[key]
}

func (c *shardedCache) add(key string, item *cacheItem) {
	shardIdx := getShard(key) % cacheShardCount
	shard := c.shards[shardIdx]
	shard.mu.Lock()
	defer shard.mu.Unlock()

	oldItems := *(*map[string]*cacheItem)(atomic.LoadPointer(&shard.items))
	newItems := make(map[string]*cacheItem, len(oldItems)+1)
	for k, v := range oldItems {
		newItems[k] = v
	}
	newItems[key] = item

	atomic.StorePointer(&shard.items, unsafe.Pointer(&newItems))
	atomic.AddUint64(&shard.count, 1)
}

func (c *shardedCache) cleanup() {
	now := time.Now().UnixNano()
	for _, shard := range c.shards {
		shard.mu.Lock()
		items := *(*map[string]*cacheItem)(atomic.LoadPointer(&shard.items))
		newItems := make(map[string]*cacheItem)

		for k, v := range items {
			if now-atomic.LoadInt64(&v.lastAccess) < int64(cacheCleanupInterval) {
				newItems[k] = v
			}
		}

		atomic.StorePointer(&shard.items, unsafe.Pointer(&newItems))
		atomic.StoreUint64(&shard.count, uint64(len(newItems)))
		shard.mu.Unlock()
	}
}

// 无锁热点缓存方法
func (c *hotCache) get(key string) *cacheItem {
	return c.shards.get(key)
}

func (c *hotCache) add(key string, item *cacheItem) {
	c.shards.add(key, item)
	atomic.AddUint64(&c.count, 1)
}

func (c *hotCache) cleanup() {
	c.shards.cleanup()
}

func (c *hotCache) warmup(key string) {
	// 预热热点缓存
	item := &cacheItem{
		lastAccess: time.Now().UnixNano(),
	}
	c.add(key, item)
}

// 批量操作
func (r *Router) BatchAddRoutes(routes []struct {
	Method  string
	Path    string
	Handler HandlerFunc
}) error {
	// 按方法分组
	groupedRoutes := make(map[string][]struct {
		Path    string
		Handler HandlerFunc
	})
	for _, route := range routes {
		groupedRoutes[route.Method] = append(groupedRoutes[route.Method], struct {
			Path    string
			Handler HandlerFunc
		}{route.Path, route.Handler})
	}

	// 批量添加路由
	for method, methodRoutes := range groupedRoutes {
		// 按分片分组
		shardedRoutes := make(map[int][]struct {
			Path    string
			Handler HandlerFunc
		})
		for _, route := range methodRoutes {
			shardIdx := getShard(route.Path)
			shardedRoutes[shardIdx] = append(shardedRoutes[shardIdx], route)
		}

		// 并发添加路由
		var wg sync.WaitGroup
		errChan := make(chan error, len(shardedRoutes))
		for shardIdx, routes := range shardedRoutes {
			wg.Add(1)
			go func(shardIdx int, routes []struct {
				Path    string
				Handler HandlerFunc
			}) {
				defer wg.Done()
				sh := r.methodShards[method][shardIdx]
				for _, route := range routes {
					if err := sh.tree.addRoute(method, route.Path, route.Handler); err != nil {
						errChan <- err
						return
					}
				}
			}(shardIdx, routes)
		}
		wg.Wait()
		close(errChan)

		// 检查错误
		if err := <-errChan; err != nil {
			return err
		}
	}

	return nil
}

// 优化热点路径检测
func (r *Router) detectHotPath(path string) bool {
	// 使用原子操作更新计数器
	value, _ := r.hotCounters.LoadOrStore(path, uint64(1))
	count := atomic.AddUint64(value.(*uint64), 1)
	return count >= hotCacheThreshold
}

// 关闭路由管理器
func (r *Router) Close() {
	close(r.stopChan)
}

// getShard 获取分片索引
func getShard(path string) int {
	h := fnv.New32()
	h.Write([]byte(path))
	return int(h.Sum32()) % methodShardCount
}

// addRoute 添加路由到 Radix 树
func (n *RadixNode) addRoute(method, path string, handler HandlerFunc) error {
	// 分割路径
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return fmt.Errorf("invalid path: %s", path)
	}

	// 从根节点开始遍历
	current := n
	for i, part := range parts {
		if part == "" {
			continue
		}

		// 检查是否是参数节点
		isParam := strings.HasPrefix(part, ":")
		paramName := ""
		if isParam {
			paramName = part[1:]
		}

		// 使用节点级别的锁
		current.mu.Lock()
		child, exists := current.children[part]
		if !exists {
			child = &RadixNode{
				path:      part,
				children:  make(map[string]*RadixNode),
				handlers:  make(map[string]HandlerFunc),
				wildcard:  isParam,
				paramName: paramName,
			}
			current.children[part] = child
		}
		current.mu.Unlock()

		// 如果是最后一个部分，设置处理函数
		if i == len(parts)-1 {
			child.mu.Lock()
			if _, exists := child.handlers[method]; exists {
				child.mu.Unlock()
				return fmt.Errorf("route already exists: %s %s", method, path)
			}
			child.handlers[method] = handler
			child.mu.Unlock()
		}

		current = child
	}

	return nil
}

// FindRoute 查找路由
func (r *Router) FindRoute(method, path string) (HandlerFunc, map[string]string, bool) {
	// 尝试从无锁热点缓存获取
	cacheKey := method + ":" + path
	if item := r.hotCache.get(cacheKey); item != nil {
		// 使用原子操作更新访问计数和时间
		atomic.AddUint64(&item.accessCount, 1)
		atomic.StoreInt64(&item.lastAccess, time.Now().UnixNano())
		// 复制参数map，避免并发修改
		params := paramsPool.Get().(map[string]string)
		for k, v := range item.params {
			params[k] = v
		}
		return item.handler, params, true
	}

	shards, ok := r.methodShards[method]
	if !ok {
		return nil, nil, false
	}
	shardIdx := getShard(path)
	sh := shards[shardIdx]

	handler, params, found := sh.tree.findRoute(method, path)

	if found {
		// 更新热点计数器
		count := atomic.AddUint64(&sh.hotCounter, 1)
		if count >= hotCacheThreshold {
			if _, loaded := r.hotCounters.LoadOrStore(cacheKey, true); !loaded {
				item := &cacheItem{
					handler:    handler,
					params:     paramsPool.Get().(map[string]string),
					lastAccess: time.Now().UnixNano(),
				}
				for k, v := range params {
					item.params[k] = v
				}
				r.hotCache.add(cacheKey, item)
			}
		}
	}

	return handler, params, found
}

// AddRoute 添加路由
func (r *Router) AddRoute(method, path string, handler HandlerFunc) error {
	shards, ok := r.methodShards[method]
	if !ok {
		return fmt.Errorf("unsupported method: %s", method)
	}
	shardIdx := getShard(path)
	sh := shards[shardIdx]
	return sh.tree.addRoute(method, path, handler)
}

// ServeHTTP 实现 http.Handler 接口
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler, params, ok := r.FindRoute(req.Method, req.URL.Path)
	if !ok {
		http.NotFound(w, req)
		return
	}
	ctx := GetContext(w, req)
	ctx.Params = params
	handler(ctx)
	PutContext(ctx)
}
