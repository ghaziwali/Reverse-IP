package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var (
	commonPorts = map[int]string{
		21:   "FTP",
		22:   "SSH",
		23:   "Telnet",
		25:   "SMTP",
		53:   "DNS",
		80:   "HTTP",
		110:  "POP3",
		143:  "IMAP",
		443:  "HTTPS",
		465:  "SMTPS",
		587:  "SMTP (Submission)",
		993:  "IMAPS",
		995:  "POP3S",
		3306: "MySQL",
		3389: "RDP",
		5432: "PostgreSQL",
		5900: "VNC",
	}
)

func clearScreen() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func resolveIP(url string, writer *bufio.Writer) {
	// Remove the protocol part from the URL if present
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")

	// Split the URL to get the host part
	host := strings.Split(url, "/")[0]

	// Resolve IP addresses associated with the host
	ips, err := net.LookupIP(host)
	if err != nil {
		writer.WriteString(fmt.Sprintf("Could not resolve IP for %s: %v\n", host, err))
		writer.Flush()
		return
	}

	// Print resolved IP addresses
	for _, ip := range ips {
		writer.WriteString(fmt.Sprintf("IP address for %s: %s\n", host, ip.String()))
	}

	// Scan common ports for services
	writer.WriteString("\nScanning common ports...\n")
	var wg sync.WaitGroup
	for port := range commonPorts {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			service := getServiceName(host, p)
			if service != "" {
				writer.WriteString(fmt.Sprintf("Port %d (%s) open: %s\n", p, service, host))
			}
		}(port)
	}
	wg.Wait()
	writer.WriteString("\n")
	writer.Flush()
}

func getServiceName(host string, port int) string {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return ""
	}
	defer conn.Close()
	return commonPorts[port]
}

func resolveIPsFromFile(filename string) {
	clearScreen()

	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Could not open file %s: %v\n", filename, err)
		return
	}
	defer file.Close()

	outputFile, err := os.Create("results.txt")
	if err != nil {
		fmt.Printf("Could not create results file: %v\n", err)
		return
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)
	consoleWriter := bufio.NewWriter(os.Stdout)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := scanner.Text()
		consoleWriter.WriteString(fmt.Sprintf("Results for %s:\n", url))
		consoleWriter.Flush()
		writer.WriteString(fmt.Sprintf("Results for %s:\n", url))
		resolveIP(url, writer)
		resolveIP(url, consoleWriter)
		consoleWriter.WriteString("--------------------------------------------------\n")
		consoleWriter.Flush()
		writer.WriteString("--------------------------------------------------\n")
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
	}

	writer.Flush()
	clearScreen()
	fmt.Println("Results have been written to results.txt")
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("Choose an option:")
		fmt.Println("1: Resolve IP and scan ports of a single URL")
		fmt.Println("2: Resolve IPs and scan ports from a list of URLs in a text file")
		fmt.Println("Type 'exit' to quit")

		fmt.Print("Enter your choice: ")
		choice, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		choice = strings.TrimSpace(choice)

		if choice == "exit" {
			fmt.Println("Exiting...")
			break
		}

		switch choice {
		case "1":
			clearScreen()
			fmt.Print("Enter URL: ")
			url, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading input: %v\n", err)
				continue
			}
			url = strings.TrimSpace(url)
			resolveIP(url, bufio.NewWriter(os.Stdout))
		case "2":
			fmt.Print("Enter the path to the text file: ")
			filename, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading input: %v\n", err)
				continue
			}
			filename = strings.TrimSpace(filename)
			resolveIPsFromFile(filename)
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}
