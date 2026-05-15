#!/bin/bash

# UE Traffic Simulator - Shell Script Version
# Sends ICMP packets at configured rates to simulate UE traffic

set -e

# Configuration
PAYLOAD_BYTES=1000
PACKET_BITS=$((($PAYLOAD_BYTES + 20 + 8) * 8))  # payload + IP(20) + ICMP(8)

# UE configurations: interface, targets (space-separated), rate (Mbps)
declare -A UE_IFACE=(
    [ue0]="ueTun0"
    [ue1]="ueTun1"
    [ue2]="ueTun2"
    [ue3]="ueTun3"
    [ue4]="ueTun4"
)

declare -A UE_TARGETS=(
    [ue0]="1.1.1.1 8.8.8.8"
    [ue1]="8.8.8.8"
    [ue2]="1.1.1.1"
    [ue3]="8.8.8.8"
    [ue4]="1.1.1.1"
)

declare -A UE_RATE=(
    [ue0]="0.1"
    [ue1]="0.2"
    [ue2]="1.0"
    [ue3]="0.1"
    [ue4]="0.1"
)

# Track ping process IDs so we can stop them safely.
PING_PIDS=()
STOP_REQUESTED=0

# Function to calculate interval in seconds for a given rate
# interval = packet_bits / (rate_Mbps * 1e6)
calc_interval() {
    local rate=$1
    awk -v bits=$PACKET_BITS -v rate=$rate 'BEGIN{printf "%.6f", bits / (rate * 1e6)}'
}

# Function to start ping for a single target
start_ping_target() {
    local ue=$1
    local iface=$2
    local target=$3
    local interval=$4

    echo "[$(date '+%H:%M:%S')] [$ue -> $target] Starting ping with interval ${interval}s"
    
    # Run ping with specified interval and payload size
    ping -I "$iface" -s "$PAYLOAD_BYTES" -i "$interval" "$target" >/dev/null 2>&1 &
    PING_PIDS+=("$!")
}

# Function to display configuration
print_config() {
    echo ""
    echo "UE Traffic Simulator Configuration:"
    echo "+-------+------------+----------+-------+"
    echo "| UE    | Interface  | Targets  | Rate  |"
    echo "+-------+------------+----------+-------+"
    
    for ue in ue0 ue1 ue2 ue3 ue4; do
        local iface=${UE_IFACE[$ue]}
        local targets=${UE_TARGETS[$ue]}
        local rate=${UE_RATE[$ue]}
        local num_targets=$(echo $targets | wc -w)
        
        printf "| %-5s | %-10s | %-8d | %-5.1f |\n" \
            "$ue" "$iface" "$num_targets" "$rate"
    done
    
    echo "+-------+------------+----------+-------+"
}

# Main function
main() {
    echo "========================================="
    echo "UE Traffic Simulator (Shell Script)"
    echo "========================================="
    echo "Payload: ${PAYLOAD_BYTES} bytes"
    echo "Packet size (IP+ICMP+payload): $PACKET_BITS bits"
    echo ""
    
    print_config
    echo ""
    echo "Starting ping processes..."
    echo "Monitor traffic with: sudo tshark -i ueTun0 -a duration:30 -T fields -e frame.len | awk '{sum+=\$1} END{printf \"%d bytes, %.3f Mbps\\n\", sum, (sum*8/30)/1e6}'"
    echo ""
    
    # Start ping processes for each UE
    for ue in ue0 ue1 ue2 ue3 ue4; do
        local iface=${UE_IFACE[$ue]}
        local targets=${UE_TARGETS[$ue]}
        local rate=${UE_RATE[$ue]}
        local num_targets=$(echo $targets | wc -w)
        local per_target_rate=$(awk -v r=$rate -v n=$num_targets 'BEGIN{printf "%.6f", r/n}')
        local interval=$(calc_interval $per_target_rate)
        
        echo "[$ue] Starting:"
        echo "  - Interface: $iface"
        echo "  - Targets: $targets"
        echo "  - Total rate: ${rate} Mbps, Per-target: ${per_target_rate} Mbps"
        echo "  - Interval: ${interval}s"
        echo ""
        
        # Start a ping process for each target
        for target in $targets; do
            start_ping_target "$ue" "$iface" "$target" "$interval"
        done
    done
    
    echo "All ping processes started. Press Ctrl+C to stop."
    
    # Keep the script running
    while true; do
        sleep 1
    done
}

# Cleanup on exit
cleanup() {
    if [ "$STOP_REQUESTED" -eq 1 ]; then
        return
    fi
    STOP_REQUESTED=1

    echo ""
    echo "Shutting down..."

    for pid in "${PING_PIDS[@]}"; do
        kill "$pid" 2>/dev/null || true
    done

    for pid in "${PING_PIDS[@]}"; do
        wait "$pid" 2>/dev/null || true
    done

    echo "Done."
}

trap cleanup INT TERM EXIT

# Run main
main
