package main

import (
	"errors"
	"fmt"
	"github.com/rycus86/podlike/pkg/component"
	"github.com/rycus86/podlike/pkg/config"
	"github.com/rycus86/podlike/pkg/controller"
	"github.com/rycus86/podlike/pkg/flags"
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

func runInit(components []*component.Component, configuration *config.Configuration) int64 {
	if len(components) == 0 {
		return 0
	}

	var (
		exitChan   = make(chan component.ExitEvent, 1)
		signalChan = make(chan os.Signal, 1)
	)

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	healthcheck.Initialize("$init$", healthcheck.StateStarting)

	for _, c := range components {
		current := c

		go func() {
			if err := current.Start(configuration); err != nil {
				exitChan <- component.ExitEvent{
					Component: current,
					Error:     err,
				}

				return
			}

			current.WaitFor(exitChan)

			current.Stop()
		}()

	waitLoop:
		for {
			select {
			case exit := <-exitChan:
				var exitCode int64 = 130

				fmt.Print("Exited: ", exit.Component.Name)
				if exit.Error != nil {
					fmt.Println(" Error:", exit.Error)
				} else {
					fmt.Println(" Status:", exit.StatusCode)

					if exit.StatusCode == 0 {
						// this is OK and expected
						break waitLoop

					} else {
						exitCode = exit.StatusCode

					}
				}

				done(components)
				return exitCode

			case s := <-signalChan:
				fmt.Printf("Exiting [%s] ...\n", s.String())

				done(components)

				if s == syscall.SIGINT {
					return 131
				} else {
					return 132
				}

			}
		}
	}

	healthcheck.SetState("$init$", healthcheck.StateHealthy)

	return 0
}

func main() {
	configuration := flags.Parse()

	hcServer, err := healthcheck.Serve()
	if err != nil {
		panic(fmt.Sprintf("failed to serve the health check information : %s", err.Error()))
	}
	defer hcServer.Close()

	cli, err := controller.NewClient()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize the controller client : %s", err.Error()))
	}
	defer cli.Close()

	go cli.WatchHealthcheckEvents()

	initComponents, err := cli.GetInitComponents()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize the init components : %s", err.Error()))
	}

	components, err := cli.GetComponents()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize the components : %s", err.Error()))
	}

	if len(components) == 0 {
		panic("no components found")
	}

	if exitCode := runInit(initComponents, configuration); exitCode == 0 {
		// only run the actual components when
		// all the init components have successfully finished
		run(components, configuration)

	} else {
		os.Exit(int(exitCode % 0xff))

	}
}
