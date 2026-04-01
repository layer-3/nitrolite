package stress

import (
	"fmt"
	"sort"
	"time"
)

// ComputeReport aggregates stress test results into a report.
func ComputeReport(method string, totalReqs, connections int, results []Result, totalTime time.Duration) Report {
	report := Report{
		Method:         method,
		TotalReqs:      totalReqs,
		Connections:    connections,
		TotalTime:      totalTime,
		ErrorBreakdown: make(map[string]int),
	}

	durations := make([]time.Duration, 0, len(results))

	for _, r := range results {
		if r.Err != nil {
			report.Failed++
			report.ErrorBreakdown[r.Err.Error()]++
		} else {
			report.Successful++
			durations = append(durations, r.Duration)
		}
	}

	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })

	if len(durations) > 0 {
		report.MinLatency = durations[0]
		report.MaxLatency = durations[len(durations)-1]

		var total time.Duration
		for _, d := range durations {
			total += d
		}
		report.AvgLatency = total / time.Duration(len(durations))
		report.MedianLatency = durationPercentile(durations, 50)
		report.P95Latency = durationPercentile(durations, 95)
		report.P99Latency = durationPercentile(durations, 99)
		if connections > 0 {
			// Pool-based: theoretical throughput = connections / avg latency.
			report.RequestsPerSec = float64(connections) / report.AvgLatency.Seconds()
		} else {
			// Custom orchestration (storm): measured throughput = successful / wall time.
			report.RequestsPerSec = float64(report.Successful) / totalTime.Seconds()
		}
	}

	return report
}

func durationPercentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p / 100.0)
	return sorted[idx]
}

// PrintReport prints a formatted stress test report to stdout.
func PrintReport(report Report) {
	fmt.Println()
	fmt.Println("Stress Test Report")
	fmt.Println("==================")
	fmt.Printf("Method:          %s\n", report.Method)
	fmt.Printf("Total Requests:  %d\n", report.TotalReqs)
	fmt.Printf("Connections:     %d\n", report.Connections)
	fmt.Printf("Duration:        %s\n", report.TotalTime.Round(time.Millisecond))

	fmt.Println()
	fmt.Println("Results")
	fmt.Println("-------")
	successPct := float64(report.Successful) / float64(report.TotalReqs) * 100
	failPct := float64(report.Failed) / float64(report.TotalReqs) * 100
	fmt.Printf("Successful:      %d (%.1f%%)\n", report.Successful, successPct)
	fmt.Printf("Failed:          %d (%.1f%%)\n", report.Failed, failPct)
	fmt.Printf("Requests/sec:    %.2f\n", report.RequestsPerSec)

	fmt.Println()
	fmt.Println("Latency")
	fmt.Println("-------")
	fmt.Printf("Min:             %s\n", report.MinLatency.Round(time.Microsecond))
	fmt.Printf("Max:             %s\n", report.MaxLatency.Round(time.Microsecond))
	fmt.Printf("Average:         %s\n", report.AvgLatency.Round(time.Microsecond))
	fmt.Printf("Median (p50):    %s\n", report.MedianLatency.Round(time.Microsecond))
	fmt.Printf("P95:             %s\n", report.P95Latency.Round(time.Microsecond))
	fmt.Printf("P99:             %s\n", report.P99Latency.Round(time.Microsecond))

	if len(report.ErrorBreakdown) > 0 {
		fmt.Println()
		fmt.Println("Errors")
		fmt.Println("------")
		for errMsg, count := range report.ErrorBreakdown {
			fmt.Printf("  %-60s %d\n", errMsg, count)
		}
	}

	fmt.Println()
}
