package statistics

import "sync"

var statisticsCache map[string]int64

var lock sync.Mutex

func init() {
	statisticsCache = map[string]int64{
		"inSuccess":  0,
		"outSuccess": 0,
		"inFailed":   0,
		"outFailed":  0,
	}
}
func IncIn() {
	lock.Lock()
	defer lock.Unlock()
	statisticsCache["inSuccess"] = statisticsCache["inSuccess"] + 1
}
func DecIn() {
	lock.Lock()
	if statisticsCache["inSuccess"]-1 > 0 {
		statisticsCache["inSuccess"] = statisticsCache["inSuccess"] - 1
	}
}
func IncOut() {
	lock.Lock()
	defer lock.Unlock()
	statisticsCache["outSuccess"] = statisticsCache["outSuccess"] + 1
}
func DecOut() {
	lock.Lock()
	defer lock.Unlock()

	if statisticsCache["outSuccess"]-1 > 0 {
		statisticsCache["outSuccess"] = statisticsCache["outSuccess"] - 1
	}
}
func IncInFailed() {
	lock.Lock()
	defer lock.Unlock()
	statisticsCache["inFailed"] = statisticsCache["inFailed"] + 1
}

func IncOutFailed() {
	lock.Lock()
	defer lock.Unlock()
	statisticsCache["outFailed"] = statisticsCache["outFailed"] + 1
}

func AllStatistics() *map[string]int64 {
	return &statisticsCache
}
