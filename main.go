package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Configuration section - Modify these values as needed
var ueConfigs = map[string]struct {
	iface string
	ips   []string
	rate  float64 // Mbps
}{
	"ue0": {"ueTun0", []string{"1.1.1.1", "8.8.8.8"}, 1.0},
	"ue1": {"ueTun1", []string{"8.8.8.8"}, 0.5},
	"ue2": {"ueTun2", []string{"1.1.1.1"}, 2.0},
	"ue3": {"ueTun3", []string{"8.8.8.8"}, 1.5},
	"ue4": {"ueTun4", []string{"1.1.1.1"}, 1.2},
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
		go func(name, iface string, ips []string, rate float64) {
			defer wg.Done()
			sendPing(name, iface, ips, rate)
		}(ueName, config.iface, config.ips, config.rate)
	}

	// Wait indefinitely since pings are continuous
	wg.Wait()
	fmt.Println("All UEs have stopped.")
}

func sendPing(ueName, iface string, ips []string, rate float64) {
	// Calculate interval between pings to maintain the specified rate
	n := len(ips)
	perTargetRate := rate
	if n > 0 {
		perTargetRate = rate / float64(n)
	}

	intervalSeconds := packetSizeBits / (perTargetRate * 1e6)
	interval := time.Duration(intervalSeconds*1e9) * time.Nanosecond

	fmt.Printf("[%s] Starting continuous ping on %s to %v at %.1f Mbps (total)\n", ueName, iface, ips, rate)

	for {
		for _, ip := range ips {
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

		ipsJoined := strings.Join(config.ips, ",")
		fmt.Printf("| %-5s | %-10s | %-20s | %-5.1f | %-10.1f |\n", ueName, config.iface, ipsJoined, config.rate, successRate)
	}

	fmt.Println("+-------+------------+------------+-------+------------+")
}
