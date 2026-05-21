# void-scan

`void-scan` is an ultra-fast, concurrent TCP network scanner and reconnaissance tool written in Go. Evolving from a basic sequential prober, it features a highly stable Worker Pool architecture, customizable timing templates, real-time service banner grabbing, and a dedicated graceful shutdown engine.

---

## Features

- **Goroutine Worker Pool:** Replaces standard "shotgun" asynchronous architectures with a fixed pool model to strictly manage OS file descriptor limits and prevent local socket exhaustion.
- **Granular Timing Control:** Implements 10 distinct speed levels (`-T1` to `-T10`) using a time-channel metronome logic to adapt seamlessly from stealthy network probing to intensive bandwidth mapping.
- **Berserk Mode:** Overrides all safety thresholds to spin up 65,535 dedicated workers simultaneously, achieving raw unthrottled execution speeds.
- **Signal Interception:** Traps `Ctrl+C` (`SIGINT/SIGTERM`) to gracefully close active network connections, restore terminal states, and return precise elapsed runtime analytics before exiting.
- **Adaptive Timing:** Can now adapt it's scan speed on non-berserk scan according to the targets network latency reducing the number of missed ports. 

---

## Installation

Make sure you have Go installed on your machine:


### Clone the repository
```bash
git clone https://github.com/NotAlive24/void-scan.git
```

### Enter project directory
```bash
cd void-scan
```

### Build the executable binary
```bash
go build -o vscan main.go
```


---

## Usage

### General Syntax
```bash
./vscan -ip [TARGET_IP] -p [PORTS] [SPEED_FLAGS]
```


### Standard balanced scan
### Default: 50 workers, 100ms throttle
```bash
./vscan -ip 192.168.1.1 -p 1-1000
```

### Strict stealth profiling
### 1 worker, 15-second tracking delay
```bash
./vscan -ip 192.168.1.1 -p 22,80,443 -T 1
```

### Aggressive enterprise port mapping
### 500 workers, 10ms throttle
```bash
./vscan -ip 192.168.1.1 -p 1-65535 -T 7
```

### Max out OS capabilities
### 10,000 workers, no throttle
```bash
./vscan -ip 192.168.1.1 -p 1-65535 -T 10
```

### Engage Berserk Mode
### Warning: May trigger local network saturation
```bash
./vscan -ip 192.168.1.1 -p 1-65535 -berserk
```


> **Note:** For `-T10` and `-berserk`, you may hit operating system file descriptor limits.
> If needed, temporarily increase the shell limit:

```bash
ulimit -n 100000
```

---

## 📜 License

Distributed under the MIT License. See `LICENSE` for details.
