package BasicAuthBruteForce

import (
	"bufio"
	"log"
	"math/rand"
	"os"
	"time"
)

func Useragent(rnd bool) string {
	var randomagent string
	var agent []string
	agentfile := "./internal/UserAgent.txt"
	if rnd {
		file, err := os.Open(agentfile)
		if err != nil {
			log.Fatal(err)
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				log.Println(err)
			}
		}(file)
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			agent = append(agent, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
		rand.Seed(time.Now().UnixNano())
		randIdx := rand.Intn(len(agent))
		randomagent = agent[randIdx]
	} else {
		randomagent = "BasicAuth"
	}

	return randomagent
}
