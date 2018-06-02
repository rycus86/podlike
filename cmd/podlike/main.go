package main

import (
	"errors"
	"fmt"
	"github.com/rycus86/podlike/pkg/component"
	"github.com/rycus86/podlike/pkg/config"
	"github.com/rycus86/podlike/pkg/controller"
	"github.com/rycus86/podlike/pkg/healthcheck"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	shouldExit bool
)

func run(components []*component.Component, configuration *config.Configuration) {
	var (
		exitChan   = make(chan component.ExitEvent, len(components))
		signalChan = make(chan os.Signal, 1)

		wg sync.WaitGroup
	)

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(len(components))

	for _, c := range components {
		current := c

		go func() {
			err := handleStart(current, configuration)

			wg.Done()

			if err != nil {
				exitChan <- component.ExitEvent{
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

func handleStart(current *component.Component, configuration *config.Configuration) error {
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

func done(components []*component.Component) {
	shouldExit = true

	var wg sync.WaitGroup

	wg.Add(len(components))

	for _, c := range components {
		item := c

		go func() {
			item.Stop()
			wg.Done()
		}()
	}

	wg.Wait()
}

func main() {
	configuration := config.Parse()

	hcServer, err := healthcheck.Serve()
	if err != nil {
		panic(err)
	}
	defer hcServer.Close()

	cli, err := controller.NewClient()
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	go cli.WatchHealthcheckEvents()

	components, err := cli.GetComponents()
	if err != nil {
		panic(err)
	}

	if len(components) == 0 {
		panic("No components found")
	}

	run(components, configuration)
}
