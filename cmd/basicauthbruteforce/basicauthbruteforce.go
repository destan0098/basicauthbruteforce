// main.go

package main

import (
	"fmt"
	"github.com/TwiN/go-color"
	BasicAuthBruteForce "github.com/destan0098/basicauthbruteforce/pkg"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var rate int
var randomagent bool
var url, username, password string
var start time.Time
var delay int

func errorpars(err error) {
	if err != nil {
		log.Panic(err.Error())
	}
}

func main() {
	start = time.Now()
	runtime.GOMAXPROCS(1)
	app := &cli.App{
		Flags: []cli.Flag{
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
		},
		Action: func(cCtx *cli.Context) error {
			switch {
			case cCtx.String("url") == "":
				fmt.Println(color.Colorize(color.Red, "[-] Please Enter URL with -d"))
			case cCtx.String("username") == "":
				fmt.Println(color.Colorize(color.Red, "[-] Please Enter Username Wordlist Address with -u"))
			case cCtx.String("password") == "":
				fmt.Println(color.Colorize(color.Red, "[-] Please Enter Password Wordlist Address with -p"))
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

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
	//var j int

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
		fmt.Printf(color.Colorize(color.Green, "[+] Find Username: %s And Password : %s\n"), res.user, res.pass)
	}
	elapsed := time.Since(start)
	fmt.Printf("page took %s", elapsed)
}

var j int
var Useragent string

func workerRoutine(jobs <-chan string, results chan<- struct{ user, pass string }, wg *sync.WaitGroup) {
	defer wg.Done()
	client := BasicAuthBruteForce.NewClient()

	for job := range jobs {
		j++
		//	fmt.Println(j)
		if j == 10 {
			Useragent = BasicAuthBruteForce.Useragent(randomagent)
			fmt.Printf(color.Colorize(color.Green, "-*- User Agent Changed to: \n %s -*- \n"), Useragent)
			j = 0
		}
		userpass := strings.Split(job, ":")
		user, pass := userpass[0], userpass[1]
		err := client.SetHeader(url, Useragent, user, pass)
		if err {
			results <- struct{ user, pass string }{user, pass}
		}
		if delay != 0 {
			time.Sleep(time.Duration(delay) * time.Second)
		}

	}
}
