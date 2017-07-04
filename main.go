package main

import (
	"fmt"
	"github.com/dags-/bugbot/bot"
	"flag"
	"bufio"
	"os"
	"math/rand"
	"time"
	"sync"
)

const bugs = "https://bugbot.dags.me/common-errors.json"

func main() {
	token := flag.String("token", "", "Auth token")
	errs := flag.String("errors", bugs, "Common errors url")
	flag.Parse()
	
	if *token == "" {
		fmt.Println("No token provided")
		return
	}

	go handleStop()

	bot.Start(*token, *errs)
}

func handleStop() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if scanner.Text() == "stop" {
			fmt.Println("Stopping...")
			os.Exit(0)
			break
		}
	}
}

func test() {
	done := make(chan interface{})
	defer close(done)

	t0 := task(done)
	t1 := task(done)
	t3 := task(done)
	t4 := task(done)
	t5 := task(done)
	t6 := task(done)

	out := merge(done, t0, t1, t3, t4, t5, t6)
	fmt.Println(<- out)
}

func task(done chan interface{}) <- chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		t := rand.Intn(10)
		dur := time.Duration(t) * time.Second
		time.Sleep(dur)

		select {
		case out <- fmt.Sprint(dur):
		case <- done:
			fmt.Println("task Task closed", dur)
			return
		}
	}()

	return out
}

func merge(done chan interface{}, in ...<- chan string) (<- chan string) {
	var wg sync.WaitGroup
	out := make(chan string)

	output := func(c <- chan string) {
		defer wg.Done()
		for n := range c {
			select {
			case out <- n:
			case <- done:
				fmt.Println("merge Task closed")
				return
			}
		}
	}

	wg.Add(len(in))
	for _, i := range in {
		go output(i)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}