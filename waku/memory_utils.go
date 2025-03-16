package waku

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var (
	testName  string
	iteration int
	phase     string
	mu        sync.Mutex
)

func main() {
	flag.StringVar(&testName, "testName", "FullTestSuite", "Name of the test ")
	flag.IntVar(&iteration, "iteration", 0, "Iteration number")
	flag.StringVar(&phase, "phase", "", "'start' or 'end')")
	flag.Parse()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	heapKB := memStats.HeapAlloc / 1024

	rssKB, err := getRSSKB()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get RSS:", err)
		rssKB = 0
	}

	if err := recordMemoryMetricsCSV(testName, iteration, phase, heapKB, rssKB); err != nil {
		fmt.Fprintln(os.Stderr, "Error recording metrics:", err)
		os.Exit(1)
	}
}

func recordMemoryMetricsCSV(testName string, iter int, phase string, heapKB, rssKB uint64) error {
	mu.Lock()
	defer mu.Unlock()

	f, err := os.OpenFile("memory_metrics.csv", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	stat, err := f.Stat()
	if err != nil {
		return err
	}
	if stat.Size() == 0 {
		header := []string{"TestName", "Iteration", "Phase", "HeapAlloc(KB)", "RSS(KB)", "Timestamp"}
		if err := w.Write(header); err != nil {
			return err
		}
	}

	row := []string{
		testName,
		strconv.Itoa(iter),
		phase,
		strconv.FormatUint(heapKB, 10),
		strconv.FormatUint(rssKB, 10),
		time.Now().Format(time.RFC3339),
	}

	return w.Write(row)
}
