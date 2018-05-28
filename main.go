package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rycus86/podlike/config"
	"github.com/rycus86/podlike/engine"
	"github.com/rycus86/podlike/healthcheck"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	shouldExit bool
)

func run(components []*engine.Component) {
	var (
		exitChan      = make(chan engine.ComponentExited, len(components))
		signalChan    = make(chan os.Signal, 1)
		configuration = config.Parse()

		wg sync.WaitGroup
	)

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(len(components))

	for _, component := range components {
		current := component

		go func() {
			err := handleStart(current, configuration)

			wg.Done()

			if err != nil {
				exitChan <- engine.ComponentExited{
					Component: current,
					Error:     err,
				}

				return
			}

			current.WaitFor(exitChan)
		}()
	}

	wg.Wait()

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

		case s := <-signalChan:
			fmt.Printf("Exiting [%s] ...\n", s.String())

			done(components)
			return

		}
	}
}

func handleStart(current *engine.Component, configuration *config.Configuration) error {
	dependencies, err := current.GetDependencies()
	if err != nil {
		return errors.New(fmt.Sprintf(
			"Failed to get the dependencies for %s: %s", current.Name, err))
	}

	for _, dependency := range dependencies {
		healthcheck.WaitUntilReady(dependency.Name, dependency.NeedsHealthyState)
	}

	err = current.Start(configuration)
	if err != nil {
		return errors.New(fmt.Sprintf(
			"Failed to start %s: %s", current.Name, err))
	}

	if shouldExit {
		go current.Stop()
	}

	return nil
}

func done(components []*engine.Component) {
	shouldExit = true

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

func checkAndExecuteHealthcheckIfNecessary() {
	configuration := config.Parse()

	if configuration.IsHealthcheck {
		if healthcheck.Check() {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
}

func main() {
	checkAndExecuteHealthcheckIfNecessary()

	hcServer, err := healthcheck.Serve()
	if err != nil {
		panic(err)
	}
	defer hcServer.Close()

	client, err := engine.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	go client.WatchHealthcheckEvents()

	components, err := client.GetComponents()
	if err != nil {
		panic(err)
	}

	if len(components) == 0 {
		panic("No components found")
	}

	run(components)
}
