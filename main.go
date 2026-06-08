package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"sync"

	h0neytr4p "github.com/t3chn0m4g3/h0neytr4p/pkg"
)

func PrintBanner() {
	fmt.Println(`                                                                
 /$$        /$$$$$$                                  /$$               /$$   /$$          
| $$       /$$$_  $$                                | $$              | $$  | $$          
| $$$$$$$ | $$$$\ $$ /$$$$$$$   /$$$$$$  /$$   /$$ /$$$$$$    /$$$$$$ | $$  | $$  /$$$$$$ 
| $$__  $$| $$ $$ $$| $$__  $$ /$$__  $$| $$  | $$|_  $$_/   /$$__  $$| $$$$$$$$ /$$__  $$
| $$  \ $$| $$\ $$$$| $$  \ $$| $$$$$$$$| $$  | $$  | $$    | $$  \__/|_____  $$| $$  \ $$
| $$  | $$| $$ \ $$$| $$  | $$| $$_____/| $$  | $$  | $$ /$$| $$            | $$| $$  | $$
| $$  | $$|  $$$$$$/| $$  | $$|  $$$$$$$|  $$$$$$$  |  $$$$/| $$            | $$| $$$$$$$/
|__/  |__/ \______/ |__/  |__/ \_______/ \____  $$   \___/  |__/            |__/| $$____/ 
                                        /$$  | $$                               | $$      
                                       |  $$$$$$/                [ v0.4 ]      | $$      
                                        \______/                                |__/      
Built by a Red team, with <3
Built by zer0p1k4chu & g0dsky (https://github.com/pbssubhash/h0neytr4p)
Adjusted for T-Pot by t3chn0m4g3 (https://github.com/t3chn0m4g3/h0neytr4p)
	`)
}

func main() {
	PrintBanner()
	var wg sync.WaitGroup
	trapsFolder := flag.String("traps", "Default", "Traps folder - It's a string.")
	logFile := flag.String("log", "Default", "Log file - It's a string.")
	catchall := flag.Bool("catchall", true, "Catch all or only trap based payloads.")
	payload := flag.String("payload", "Default", "Payload folder - It's a string.")
	cert := flag.String("cert", "Default", "Certificate File")
	key := flag.String("key", "Default", "Certificate File")
	verbose := flag.Bool("verbose", true, "Use -verbose=false for disabling streaming output; by default it's true.")
	wildcard := flag.Bool("wildcard", false, "Load all traps on ports 80 and 443.")
	help := flag.Bool("help", false, "Print Help")
	flag.Parse()

	if *help || *trapsFolder == "Default" || *logFile == "Default" || *payload == "Default" {
		fmt.Println("Wrong Arguments.. Exiting Now")
		flag.PrintDefaults()
		os.Exit(1)
	}
	fmt.Printf("[ Traps folder            ] -> [ %-30s]\n", *trapsFolder)
	fmt.Printf("[ Logfile                 ] -> [ %-30s]\n", *logFile)
	fmt.Printf("[ Payloads folder         ] -> [ %-30s]\n", *payload)
	fmt.Printf("[ Catch all payloads      ] -> [ %-30t]\n", *catchall)
	fmt.Printf("[ Payload multipart limit ] -> [ %-30d]\n", h0neytr4p.MaxMultipartSize)
	fmt.Printf("[ Payload other limit     ] -> [ %-30d]\n", h0neytr4p.MaxJSONFormSize)
	fmt.Printf("[ Wildcard mode           ] -> [ %-30t]\n", *wildcard)
	fmt.Println()

	trapConfig, err := h0neytr4p.ParseTraps(*trapsFolder)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing traps: %v\n", err)
		os.Exit(1)
	}
	if err := h0neytr4p.InitLogFile(*logFile, *verbose); err != nil {
		fmt.Fprintf(os.Stderr, "Error configuring log file: %v\n", err)
		os.Exit(1)
	}
	defer h0neytr4p.CloseLogFile()
	if err := h0neytr4p.InitPayloadFolder(*payload); err != nil {
		fmt.Fprintf(os.Stderr, "Error configuring payload folder: %v\n", err)
		os.Exit(1)
	}

	predefinedPorts := []string{"443", "80"}
	filteredTraps := make(map[string][]h0neytr4p.Trap)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			if sig == os.Interrupt {
				fmt.Println("Interrupt received. Gracefully exiting the program.")
				os.Exit(1)
			}
		}
	}()

	for _, trap := range trapConfig {
		if *wildcard {
			for _, port := range predefinedPorts {
				filteredTraps[port] = append(filteredTraps[port], trap)
			}
		} else {
			port := trap.Basicinfo.Port
			filteredTraps[port] = append(filteredTraps[port], trap)
		}
	}

	portsToLoad := predefinedPorts
	if !*wildcard {
		portsToLoad = make([]string, 0, len(filteredTraps))
		for port := range filteredTraps {
			portsToLoad = append(portsToLoad, port)
		}
		sort.Strings(portsToLoad)
	}

	errCh := make(chan error, len(portsToLoad))
	for _, port := range portsToLoad {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			if err := h0neytr4p.StartHandler(p, filteredTraps[p], *cert, *key, *catchall); err != nil {
				errCh <- fmt.Errorf("handler on port %s failed: %w", p, err)
			}
		}(port)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
