package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

// Configuration section - Modify these values as needed
var ueConfigs = map[string]struct {
	iface string
	ip    string
	rate  float64 // Mbps
}{
	"ue0": {"ueTun0", "1.1.1.1", 1.0},
	"ue1": {"ueTun1", "8.8.8.8", 0.5},
	"ue2": {"ueTun2", "1.1.1.1", 2.0},
	"ue3": {"ueTun3", "8.8.8.8", 1.5},
	"ue4": {"ueTun4", "1.1.1.1", 1.2},
}
const packetSizeBits = 84 * 8 // 672 bits (20 IP + 8 ICMP + 56 data)

type Stats struct {
	mu      sync.Mutex
	sent    int
	success int
}

var stats = make(map[string]*Stats)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		fmt.Println("UE Traffic Simulator")
		fmt.Println("Usage: ./ue_simulator")
		fmt.Println("Simulates multiple UEs sending continuous ping traffic at specified rates.")
		fmt.Println("Modify ueConfigs in the source code to change configurations.")
		fmt.Println("Press Ctrl+C to stop.")
		return
	}

	// Initialize stats
	for ueName := range ueConfigs {
		stats[ueName] = &Stats{}
	}

	// Start statistics display goroutine
	go func() {
		for {
			time.Sleep(5 * time.Second)
			printTable()
		}
	}()

	var wg sync.WaitGroup

	for ueName, config := range ueConfigs {
		wg.Add(1)
		go func(name, iface, ip string, rate float64) {
			defer wg.Done()
			sendPing(name, iface, ip, rate)
		}(ueName, config.iface, config.ip, config.rate)
	}

	// Wait indefinitely since pings are continuous
	wg.Wait()
	fmt.Println("All UEs have stopped.")
}

func sendPing(ueName, iface, ip string, rate float64) {
	// Calculate interval between pings to maintain the specified rate
	intervalSeconds := packetSizeBits / (rate * 1e6)
	interval := time.Duration(intervalSeconds*1e9) * time.Nanosecond

	fmt.Printf("[%s] Starting continuous ping on %s to %s at %.1f Mbps\n", ueName, iface, ip, rate)

	for {
		cmd := exec.Command("ping", "-I", iface, ip, "-c", "1")
		// Suppress output to avoid clutter
		cmd.Stdout = nil
		cmd.Stderr = nil

		err := cmd.Run()
		stats[ueName].mu.Lock()
		stats[ueName].sent++
		if err == nil {
			stats[ueName].success++
		}
		stats[ueName].mu.Unlock()

		time.Sleep(interval)
	}
}

func printTable() {
	fmt.Println("\nUE Traffic Statistics:")
	fmt.Println("+-------+------------+------------+-------+------------+")
	fmt.Println("| UE    | Interface  | IP         | Rate  | Success %  |")
	fmt.Println("+-------+------------+------------+-------+------------+")

	for ueName, config := range ueConfigs {
		stat := stats[ueName]
		stat.mu.Lock()
		sent := stat.sent
		success := stat.success
		stat.mu.Unlock()

		var successRate float64
		if sent > 0 {
			successRate = float64(success) / float64(sent) * 100
		}

		fmt.Printf("| %-5s | %-10s | %-10s | %-5.1f | %-10.1f |\n", ueName, config.iface, config.ip, config.rate, successRate)
	}

	fmt.Println("+-------+------------+------------+-------+------------+")
}
