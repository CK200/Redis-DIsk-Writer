package main

import (
	"fmt"
	"main/config"
	"main/pkg/globals"
	"main/pkg/services"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	fmt.Println("Starting queueOnDisk script....")
	config.SetUpApplication()
}

func main() {
	parameters := os.Args[1:]
	fmt.Println("in main startingservices...")
	fmt.Println("parameters:", parameters)

	if len(parameters) < 1 {
		fmt.Println("Invalid number of arguements... Please provide queue names")
		return
	} else {
		globals.ApplicationWaitGroupServices.Add(1)
		fmt.Println("INITIATING WRITER SERVICE")
		go services.StartService(parameters)

		if globals.ApplicationConfig.Application.EnableRespush {
			fmt.Println("INITIATING REPSUH SERVICE")
			globals.ApplicationWaitGroupServices.Add(1)
			go services.RepushService()
		} else {
			fmt.Println("REPUSH SERVICE DISABLED")
		}
	}

	serviceShutDown := make(chan os.Signal, 1)
	signal.Notify(serviceShutDown, os.Interrupt, syscall.SIGTERM)

	<-serviceShutDown
	fmt.Println("Stopping services .....")
	globals.Shutdown = true
	fmt.Println("Waiting for services...")
	globals.ApplicationWaitGroupServices.Wait()
	fmt.Println("Wait over exting code...")
}
