package main

import (
	"fmt"
	"github.com/rycus86/podlike/config"
	"github.com/rycus86/podlike/engine"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func run(components []*engine.Component) {
	var (
		exitChan      = make(chan engine.ComponentExited, len(components))
		signalChan    = make(chan os.Signal, 1)
		configuration = config.Parse()
	)

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for _, component := range components {
		err := component.Start(configuration)
		if err != nil {
			fmt.Println("Failed to start", component.Name, ":", err)

			done(components)
			return
		}

		current := component

		go func() {
			current.WaitFor(exitChan)
		}()
	}

	for {
		select {
		case exit := <-exitChan:
			fmt.Print("Exited: ", exit.Component.Name)
			if exit.Error != nil {
				fmt.Println(" Error:", exit.Error)
			} else {
				fmt.Println(" Status:", exit.StatusCode)
			}

			done(components)
			return

		case <-signalChan:
			fmt.Println("Exiting...")

			done(components)
			return

		}
	}
}

func done(components []*engine.Component) {
	var wg sync.WaitGroup

	wg.Add(len(components))

	for _, comp := range components {
		item := comp

		go func() {
			item.Stop()
			wg.Done()
		}()
	}

	wg.Wait()
}

func main() {
	client, err := engine.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	components, err := client.GetComponents()
	if err != nil {
		panic(err)
	}

	if len(components) == 0 {
		panic("No components found")
	}

	run(components)
}
