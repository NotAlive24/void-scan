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
func Validate() (string, string, int, bool) {
	ip := flag.String("ip", "", "To enter the target IP.")
	p := flag.String("p", "", "To enter the target port (e.g. 80, 1-100, 80,443).")
	t := flag.Int("T", 5, "Timing template (1-10). 1 is sneaky, 5 is standard, 10 is extreme.")
	berserk := flag.Bool("berserk", false, "Unleash the horde. Warning: Might crash your OS network stack.")
	flag.Parse()
	return *ip, *p, *t, *berserk
}
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
			portList = append(portList, p)
		}
	}
	return portList
}
func engine(targetIP string, ports []int, workers int, rateLimit time.Duration) {
	var wg sync.WaitGroup
	jobs := make(chan int, len(ports))
	var ticker *time.Ticker
	if rateLimit > 0 {
		ticker = time.NewTicker(rateLimit)
		defer ticker.Stop()
	}
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
				conn, err := net.DialTimeout("tcp", address, 1*time.Second)
				if err == nil {
					conn.SetReadDeadline(time.Now().Add(1 * time.Second))
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
				wg.Done()
			}
		}()
	}
	for _, port := range ports {
		wg.Add(1)
		jobs <- port
	}
	close(jobs)
	wg.Wait()
	fmt.Println("\nScan process finished.")
}
func main() {
	targetIP, targetPort, timingLevel, isBerserk := Validate()
	fmt.Printf("Good, the IP is valid\n")
	var workers int
	var rate time.Duration

	if isBerserk {
		workers = 65535
		rate = 0
		fmt.Println("\n[!!!] BERSERK MODE ENGAGED [!!!]")
		fmt.Println("[!!!] Pray for your router... [!!!]\n")
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
	engine(targetIP, serpratedPort, workers, rate)
	elapsed := time.Since(startTime)
	fmt.Printf("Scan finished naturally. Total time: %v\n", elapsed)
}
