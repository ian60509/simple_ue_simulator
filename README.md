# UE Simulator

A simple CLI tool to simulate multiple UEs sending continuous ping traffic at specified rates, with statistics display.

## Configuration

Edit the `ueConfigs` map in `main.go` to configure UEs:

```go
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
```

- `iface`: Tunnel interface name to bind the ping to
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

The program will continuously send ping packets from each configured UE at the specified rates. Statistics are displayed in a table every 5 seconds, showing interface, ping destination, rate, and success percentage. Press Ctrl+C to stop.

## Building

```
go build -o ue_simulator main.go
```