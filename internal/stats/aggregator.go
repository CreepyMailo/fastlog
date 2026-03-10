package stats

import (
	"sort"
	"sync"
)

// Item представляет элемент с ключом и значением для сортировки
type Item struct {
	Key   string
	Count int
}

// Aggregator собирает статистику из обработанных логов
type Aggregator struct {
	mu         sync.RWMutex
	ipCounts   map[string]int
	urlCounts  map[string]int
	totalLines int
	errorLines int
}

// NewAggregator создает новый агрегатор
func NewAggregator() *Aggregator {
	return &Aggregator{
		ipCounts:  make(map[string]int),
		urlCounts: make(map[string]int),
	}
}

// AddIP добавляет IP в статистику
func (a *Aggregator) AddIP(ip string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ipCounts[ip]++
}

// AddURL добавляет URL в статистику
func (a *Aggregator) AddURL(url string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.urlCounts[url]++
}

// AddLine увеличивает счетчик обработанных строк
func (a *Aggregator) AddLine() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.totalLines++
}

// AddErrorLine увеличивает счетчик ошибочных строк
func (a *Aggregator) AddErrorLine() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.errorLines++
}

// GetTotalLines возвращает общее количество обработанных строк
func (a *Aggregator) GetTotalLines() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.totalLines
}

// GetErrorLines возвращает количество строк с ошибками
func (a *Aggregator) GetErrorLines() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.errorLines
}

// GetTopIPs возвращает топ N IP-адресов
func (a *Aggregator) GetTopIPs(n int) []Item {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return getTopN(a.ipCounts, n)
}

// GetTopURLs возвращает топ N URL
func (a *Aggregator) GetTopURLs(n int) []Item {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return getTopN(a.urlCounts, n)
}

// getTopN возвращает топ N элементов из мапы
func getTopN(counts map[string]int, n int) []Item {
	if n <= 0 {
		return []Item{}
	}

	items := make([]Item, 0, len(counts))
	for key, count := range counts {
		items = append(items, Item{Key: key, Count: count})
	}

	// Сортируем по убыванию
	sort.Slice(items, func(i, j int) bool {
		return items[i].Count > items[j].Count
	})

	// Обрезаем до N элементов
	if len(items) > n {
		items = items[:n]
	}

	return items
}
