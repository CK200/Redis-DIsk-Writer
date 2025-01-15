package services

import (
	"fmt"
	"log"
	"main/logs"
	database "main/pkg/database/redis"
	"main/pkg/globals"
	"main/pkg/util"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/mem"
)

func StartService(parameters []string) {
	redisWaitSeconds := globals.ApplicationConfig.Application.RedisConnectionTimeSeconds

	defer func() {

		if r := recover(); r != nil {
			// fmt.Println("Recovered:", r)
			fmt.Printf("Recovered: %v\n", r)
			util.SendSlackAlert(fmt.Sprintf("Writer:queueMonitoring script paniced::%v", r))
			buf := make([]byte, 1024)
			n := runtime.Stack(buf, false)
			log.Println("***ERROR***")
			if len(buf) > 0 {
				fmt.Printf("Error : %v\n", buf[:n])
			}

		}

	}()

	for {
		if globals.Shutdown {
			fmt.Println("Marking writer service as done")
			globals.ApplicationWaitGroupServices.Done()
			return
		}

		if !database.CheckConnection() {
			util.SendSlackAlert(fmt.Sprintf("Writer:Cannot establish connection to redis Host:[%s] | Port:[%s].", globals.ApplicationConfig.RedisQueue.Host, globals.ApplicationConfig.RedisQueue.Port))
			fmt.Printf("Respush::Connection issue with redis. Will retry after %d seconds", redisWaitSeconds)
			fmt.Printf("Repush::Sleeping for %d Seconds...\n", redisWaitSeconds)
			fmt.Println("Trying to Re-establish connection to redis.")
			database.EstablishRedisQueueConnecion()
			time.Sleep(time.Duration(redisWaitSeconds) * time.Second)
			continue
		}

		v, _ := mem.VirtualMemory()
		fmt.Println("Writer::Used: ", v.UsedPercent, " | conf used : ", globals.ApplicationConfig.Application.RAMthresholdPercent)
		if v.UsedPercent > float64(globals.ApplicationConfig.Application.RAMthresholdPercent) {
			util.SendSlackAlert(fmt.Sprintf("Writer:RAM usage is above %v", globals.ApplicationConfig.Application.RAMthresholdPercent))

			for _, queue := range parameters {
				for i := 0; i < globals.Workers; i++ {
					globals.ApplicationWaitGroupServices.Add(1)
					go WriteService(queue)
				}
			}
		}

		fmt.Printf("Sleeping for %d seconds...\n", globals.ApplicationConfig.Application.MonitorSleepTimeSeconds)
		time.Sleep(time.Duration(globals.ApplicationConfig.Application.MonitorSleepTimeSeconds) * time.Second)
	}

}

func WriteService(queue string) {
	fmt.Println("New writer for queue: ", queue)
	for {
		counter := database.CustomLlen(queue)
		if counter == 0 || globals.Shutdown {
			fmt.Println("Queue length: ", counter)
			fmt.Println("shutting writer for queue: ", queue)
			globals.ApplicationWaitGroupServices.Done()
			return
		} else {
			for counter > 0 && !globals.Shutdown {
				v, _ := mem.VirtualMemory()
				if v.UsedPercent > float64(globals.ApplicationConfig.Application.RAMthresholdPercent) {
					result, err := database.CustomBLpop(queue)
					if err != nil {
						util.SendSlackAlert(fmt.Sprintf("Writer: closing goroutine due to error with redis push. Host:[%s] Port:[%s]. Error:%s", globals.ApplicationConfig.Database.Host, globals.ApplicationConfig.Database.Port, err.Error()))
						fmt.Printf("Sleeping for %d seconds...\n", globals.ApplicationConfig.Application.RedisConnectionTimeSeconds)
						time.Sleep(time.Duration(globals.ApplicationConfig.Application.RedisConnectionTimeSeconds) * time.Second)
						// set counter to zero to stop this routine
						counter = 0
						break
					}

					if len(result) < 2 {
						fmt.Println("Error: Blpop skipped due to error or insufficient values.")
					} else {
						logs.InfoLog("QUEUE[%v]|MESSAGE[%v]", result[0], result[1])
					}
					counter--
				}
			}
		}
	}
}
