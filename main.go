package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Checks for Ctrl+C and anyother termination signals to exit.
func setupCloseHandler(startTime time.Time) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\n\n[!] Ctrl+C detected! Exitting...")
		elapsed := time.Since(startTime)
		fmt.Printf("[!] Scan interrupted. Time passed: %v\n", elapsed)
		fmt.Println("[!] Exiting...")
		fmt.Println("[!] Bye!")
		os.Exit(0)
	}()
}

// Validates the user input and returns the target IP, ports, timing level, and berserk mode status.
func Validate() (string, string, int, bool) {
	ip := flag.String("ip", "", "To enter the target IP.")
	p := flag.String("p", "", "To enter the target port (e.g. 80, 1-100, 80,443).")
	t := flag.Int("T", 5, "Timing template (1-10). 1 is sneaky, 5 is standard, 10 is extreme.")
	berserk := flag.Bool("berserk", false, "Unleash the horde. Warning: Might crash your OS network stack.")
	flag.Parse()

	// Making sure all stuff needed are there
	if *ip == "" || *p == "" {
		fmt.Println("Usage error: -ip and -p flags are strictly required.")
		fmt.Println("Use -h for help menu.")
		os.Exit(1)
	}

	// Validating the IP address structurally.
	parsedIP := net.ParseIP(*ip)
	if parsedIP == nil {
		fmt.Println("What is this IP man, it's not valid structurally. \nCheck again and enter a valid one. \nSkill issue.... \nExited the UI..")
		os.Exit(1)
	}

	if parsedIP.IsLoopback() {
		fmt.Println("Loopback IP detected. Mmmmm, you are scanning yourself? That's nice, have fun!")
	}

	if parsedIP.IsMulticast() {
		fmt.Println("That is a Multicast address. You can't TCP port scan a multicast group. \nExited the UI..")
		os.Exit(1)
	}

	ipv4 := parsedIP.To4()
	if ipv4 != nil {
		// Block 0.x.x.x (Including 0.0.0.0 and 0.0.0.1)
		if ipv4[0] == 0 {
			fmt.Println("Nice try. 0.x.x.x addresses are reserved for default routes and local network identification. \nYou can't scan that. \nExited the UI..")
			os.Exit(1)
		}
		// Block Broadcast addresses
		if ipv4[0] == 255 {
			fmt.Println("That's a broadcast address. We are port scanning a target, not yelling at the whole subnet. \nExited the UI..")
			os.Exit(1)
		}
	}

	if *p == "0" {
		fmt.Println("Is this your Exam score? Port 0 is reserved. \nExited the UI..")
		os.Exit(1)
	}
	return *ip, *p, *t, *berserk
}

// Separates the port input into a list of ports and returns it as integers.
func portSepration(targetPort string) []int {
	var portList []int
	groups := strings.Split(targetPort, ",")
	for _, group := range groups {
		if strings.Contains(group, "-") {
			ports := strings.Split(group, "-")
			start, _ := strconv.Atoi(ports[0])
			end, _ := strconv.Atoi(ports[1])
			if end < start {
				start, end = end, start
			}
			if end > 65535 {
				end = 65535
			}
			if start < 1 {
				start = 1
			}
			for i := start; i <= end; i++ {
				portList = append(portList, i)
			}
		} else {
			p, _ := strconv.Atoi(group)
			if p >= 1 && p <= 65535 {
				portList = append(portList, p)
			}
		}
	}
	return portList
}

// THE ENGINEEEEE
func engine(targetIP string, ports []int, workers int, rateLimit time.Duration, isBerserk bool) {
	var wg sync.WaitGroup // Employees
	jobs := make(chan int, len(ports))

	// Adaptive timing for non-Berserk modes.
	var mu sync.Mutex
	adaptiveTimeout := 1000 * time.Millisecond
	minTimeout := 200 * time.Millisecond
	maxTimeout := 3000 * time.Millisecond

	// Only create a ticker if we have a rate limit. Berserk mode ignores rate limits.
	var ticker *time.Ticker
	if rateLimit > 0 {
		ticker = time.NewTicker(rateLimit)
		defer ticker.Stop()
	}

	// Hiring more employees (goroutines) to handle the workload.
	for i := 0; i < workers; i++ {
		go func() {
			for p := range jobs {
				if rateLimit > 0 {
					<-ticker.C
				}

				ip := net.ParseIP(targetIP)
				var address string
				if ip.To4() == nil && ip.To16() != nil {
					address = fmt.Sprintf("[%s]:%d", targetIP, p)
				} else {
					address = fmt.Sprintf("%s:%d", targetIP, p)
				}

				var currentTimeout time.Duration
				if isBerserk {
					currentTimeout = 1000 * time.Millisecond
				} else {
					mu.Lock()
					currentTimeout = adaptiveTimeout
					mu.Unlock()
				}

				startCall := time.Now()
				conn, err := net.DialTimeout("tcp", address, currentTimeout)
				rtt := time.Since(startCall)

				// If the connection is successful, we can calculate the adaptive timeout based on the RTT (Round Time Trip) it's the time the packet takes for Leave -> Server -> Return.
				if err == nil {
					// if not in berserk mode do the calculation
					if !isBerserk {
						calculatedTimeout := (currentTimeout * 8 / 10) + (rtt*2/10)*2

						mu.Lock()
						if calculatedTimeout < minTimeout {
							adaptiveTimeout = minTimeout
						} else if calculatedTimeout > maxTimeout {
							adaptiveTimeout = maxTimeout
						} else {
							adaptiveTimeout = calculatedTimeout
						}
						mu.Unlock()
					}

					// Reading the banner received.
					conn.SetReadDeadline(time.Now().Add(currentTimeout))
					buffer := make([]byte, 1024)
					n, _ := conn.Read(buffer)

					banner := strings.TrimSpace(string(buffer[:n]))
					banner = strings.ReplaceAll(banner, "\n", " ")
					banner = strings.ReplaceAll(banner, "\r", "")

					if len(banner) > 0 {
						if len(banner) > 50 {
							banner = banner[:47] + "..."
						}
						fmt.Printf("[+] PORT %d IS OPEN! \t[ %s ]\n", p, banner)
					} else {
						fmt.Printf("[+] PORT %d IS OPEN!\n", p)
					}
					conn.Close()
				}
				// Sending the Employees back home, accidently forgetting to pay them.
				wg.Done()
			}
		}()
	}

	// Giving the Employees (goroutines) the work (ports to scan).
	for _, port := range ports {
		wg.Add(1)
		jobs <- port
	}
	close(jobs)
	wg.Wait()
	fmt.Println("\nScan process finished.")
}

// Main function
func main() {
	targetIP, targetPort, timingLevel, isBerserk := Validate()
	fmt.Printf("Good, the IP is valid\n")

	var workers int
	var rate time.Duration

	// Setting the number of workers and rate limit based on the timing level or berserk mode.
	if isBerserk {
		workers = 65535
		rate = 0
		fmt.Println("\n[!!!] BERSERK MODE ENGAGED [!!!]")
		fmt.Println("[!!!] Pray for your router... [!!!]")
	} else {
		switch timingLevel {
		case 1:
			workers, rate = 1, 15*time.Second
		case 2:
			workers, rate = 2, 5*time.Second
		case 3:
			workers, rate = 5, 1*time.Second
		case 4:
			workers, rate = 10, 500*time.Millisecond
		case 5:
			workers, rate = 50, 100*time.Millisecond
		case 6:
			workers, rate = 100, 50*time.Millisecond
		case 7:
			workers, rate = 500, 10*time.Millisecond
		case 8:
			workers, rate = 1000, 5*time.Millisecond
		case 9:
			workers, rate = 5000, 1*time.Millisecond
		case 10:
			workers, rate = 10000, 0
		default:
			workers, rate = 50, 100*time.Millisecond
			timingLevel = 5
		}
	}

	serpratedPort := portSepration(targetPort)
	fmt.Printf("Locked the Target: %s\n", targetIP)

	if isBerserk {
		fmt.Printf("Mode: BERSERK (Workers: %d)\n", workers)
	} else {
		fmt.Printf("Mode: -T%d (Workers: %d)\n", timingLevel, workers)
	}

	fmt.Println("Starting the scan...")

	startTime := time.Now()
	setupCloseHandler(startTime)

	engine(targetIP, serpratedPort, workers, rate, isBerserk)

	elapsed := time.Since(startTime)
	fmt.Printf("Scan finished naturally. Total time: %v\n", elapsed)
}
