package main

import (
	"bufio"
	"fmt"
	"github.com/felixge/ardrone/client"
	"github.com/go-chi/chi"
	"log"
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
		"#": {duration: 3 * time.Second, state: ardrone.State{Fly: true}},                      // Take off
		"*": {duration: 1 * time.Second, state: ardrone.State{Fly: false}},                     // Land
		"1": {duration: 2 * time.Second, state: ardrone.State{Fly: true, Yaw: -rotationSpeed}}, // Counterclockwise
		"3": {duration: 2 * time.Second, state: ardrone.State{Fly: true, Yaw: rotationSpeed}},  // Clockwise
		"2": {duration: 2 * time.Second, state: ardrone.State{Fly: true, Pitch: movSpeed}},     // Forward
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
		fmt.Println(d)
		if d != "" {
			if c, ok := dtmfCommands[d]; ok {
				client.ApplyFor(c.duration, c.state)
				feedback = "Done"
			} else {
				feedback = "Unknown Command"
			}

		}
		w.Header().Set("Content-type", "text/xml")
		w.Write([]byte(fmt.Sprintf(responseTemplate, feedback)))
	})
	http.Handle("/", r)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	var t string
	for true {
		scanner.Scan()
		t = scanner.Text()

		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading from input: ", err)
		}
		if t == "q" {
			client.Land()
			if err := srv.Shutdown(nil); err != nil {
				panic(err)
			}
			break
		}

		if c, ok := dtmfCommands[t]; ok {
			client.ApplyFor(c.duration, c.state)
		}
	}
}
