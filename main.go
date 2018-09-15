package main

import (
	"bufio"
	"fmt"
	"github.com/felixge/ardrone/client"
	"github.com/go-chi/chi"
	"net/http"
	"os"
	"time"
)

const (
	responseTemplate = `<Response><Say language="en-US">%s</Say><Gather timeout="10" numDigits="1" language="en-US" input="dtmf" finishOnKey=""/><Redirect method="POST">/dron</Redirect></Response>`
	rotationSpeed    = 0.1
	movSpeed         = 0.1
)

type Command struct {
	duration time.Duration
	state    ardrone.State
}

var (
	dtmfCommands = map[string]Command{
		//Rotation
		"1": {duration: 1200 * time.Millisecond, state: ardrone.State{Fly: true, Yaw: -rotationSpeed}}, // Counterclockwise
		"3": {duration: 1200 * time.Millisecond, state: ardrone.State{Fly: true, Yaw: rotationSpeed}},  // Clockwise

		//Move
		"2": {duration: 900 * time.Millisecond, state: ardrone.State{Fly: true, Pitch: movSpeed}},  // Forward
		"8": {duration: 900 * time.Millisecond, state: ardrone.State{Fly: true, Pitch: -movSpeed}}, // Back
		"6": {duration: 900 * time.Millisecond, state: ardrone.State{Fly: true, Roll: movSpeed}},   // Right
		"4": {duration: 900 * time.Millisecond, state: ardrone.State{Fly: true, Roll: -movSpeed}},  // Left

		//Altitude
		"5": {duration: 1000 * time.Millisecond, state: ardrone.State{Fly: true, Vertical: movSpeed}}, // Up
		"0": {duration: 500 * time.Millisecond, state: ardrone.State{Fly: true, Vertical: -movSpeed}}, // Down
	}
)

func main() {
	client, err := ardrone.Connect(ardrone.DefaultConfig())
	if err != nil {
		panic(err)
	}
	fmt.Println("Drone connected")
	go func() {
		for c := range client.Navdata {
			if c.Demo.Battery <= 18 {
				fmt.Printf("WARNING: Low Drone Batterry:  %d %%\n", c.Demo.Battery)
			}
		}
	}()

	srv := &http.Server{Addr: ":8080"}
	r := chi.NewRouter()
	r.Post("/dron", func(w http.ResponseWriter, r *http.Request) {
		feedback := "What are my orders?"

		r.ParseForm()
		d := r.Form.Get("Digits")

		if d != "" {
			fmt.Printf("Drone will %s\n", d)
			if executeDronCommand(client, d) {
				feedback = "Done"
			} else {
				feedback = "Unknown command"
			}
		}
		fmt.Println(feedback)
		w.Header().Set("Content-type", "text/xml")
		w.Write([]byte(fmt.Sprintf(responseTemplate, feedback)))
	})
	http.Handle("/", r)
	go func() {
		fmt.Println("Httpserver: Started")
		if err := srv.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			fmt.Printf("Httpserver: ListenAndServe() error: %s", err)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	var k string
	for true {
		scanner.Scan()
		k = scanner.Text()

		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading from input: ", err)
		}
		if k == "q" {
			client.Land()
			if err := srv.Shutdown(nil); err != nil {
				panic(err)
			}
			break
		}
		_ = executeDronCommand(client, k)
	}
}

func executeDronCommand(client *ardrone.Client, command string) bool {
	if command == "#" {
		client.Takeoff()
		return true
	}
	if command == "*" {
		client.Land()
		return true
	}
	if c, ok := dtmfCommands[command]; ok {
		client.Apply(c.state)
		time.Sleep(c.duration)
		client.Hover()
		return true
	}
	return false
}
