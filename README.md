# UE Simulator

A simple CLI tool to simulate multiple UEs sending continuous ping traffic at specified rates, with statistics display.

## Configuration

Edit the `ueConfigs` map in `main.go` to configure UEs:

```go
var ueConfigs = map[string]struct {
	ip   string
	rate float64 // Mbps
}{
	"ue1": {"1.1.1.1", 1.0},
	"ue2": {"8.8.8.8", 0.5},
}
```

- `ip`: Target IP address for ping
- `rate`: Traffic rate in Mbps (mega bits per second)

## Usage

1. Enter the UE namespace (as per free-ran-ue Makefile):
   ```
   make ns-ue
   ```

2. Run the simulator:
   ```
   ./ue_simulator
   ```

The program will continuously send ping packets from each configured UE at the specified rates. Statistics are displayed in a table every 5 seconds, showing ping destination, rate, and success percentage. Press Ctrl+C to stop.

## Building

```
go build -o ue_simulator main.go
```