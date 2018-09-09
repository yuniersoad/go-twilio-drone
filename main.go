package main

import (
	"bufio"
	"fmt"
	"github.com/felixge/ardrone/client"
	"os"
	"time"
)

func main() {
	client, err := ardrone.Connect(ardrone.DefaultConfig())
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	var t string
	for true {
		scanner.Scan()
		t = scanner.Text()

		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading from input: ", err)
		}

		// scanner.Text() strips new lines
		// so in case of just a new line
		// it's actually an empty string
		if t == "q" {
			client.Land()
			break
		}

		if t == "#" {
			client.Takeoff()
			time.Sleep(3 * time.Second)
		}

		if t == "*" {
			client.Land()
		}

		if t == "1" {
			client.Counterclockwise(0.3)
			time.Sleep(1 * time.Second)
			client.Hover()
			time.Sleep(1 * time.Second)
		}

		if t == "3" {
			client.Clockwise(0.3)
			time.Sleep(1 * time.Second)
			client.Hover()
			time.Sleep(1 * time.Second)
		}

		if t == "2" {
			client.Forward(0.4)
			time.Sleep(1 * time.Second)
			client.Hover()
			time.Sleep(1 * time.Second)
		}
	}
	client.Land()
}
