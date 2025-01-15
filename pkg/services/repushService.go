package services

import (
	"bufio"
	"compress/gzip"
	"fmt"
	database "main/pkg/database/redis"
	"main/pkg/globals"
	"main/pkg/util"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/shirou/gopsutil/mem"
)

func RepushService() {
	fmt.Println("In repush service")
	repushWaitSeconds := globals.ApplicationConfig.Application.RepushWaitSeconds
	redisWaitSeconds := globals.ApplicationConfig.Application.RedisConnectionTimeSeconds
	maxProc := globals.ApplicationConfig.Application.MaxRepushProcesses
	fmt.Println("Max Repush Proc: ", maxProc)
	for {
		if globals.Shutdown {
			globals.ApplicationWaitGroupServices.Done()
			fmt.Println("Marking repush service as done")
			return
		}

		// Restart if memory full
		v, _ := mem.VirtualMemory()
		fmt.Println("Repush::Used: ", v.UsedPercent, " | Conf used : ", globals.ApplicationConfig.Application.RAMRepushthresholdPercent)
		if v.UsedPercent >= float64(globals.ApplicationConfig.Application.RAMRepushthresholdPercent) {
			fmt.Println("Memory not sufficient skipping REPUSH")
			fmt.Printf("Sleeping for %d Seconds...\n", repushWaitSeconds)
			time.Sleep(time.Duration(repushWaitSeconds) * time.Second)
			continue
		}

		if !database.CheckConnection() {
			util.SendSlackAlert(fmt.Sprintf("Repush:Cannot establish connection to redis Host:[%s] | Port:[%s].", globals.ApplicationConfig.RedisQueue.Host, globals.ApplicationConfig.RedisQueue.Port))
			fmt.Printf("Respush::Connection issue with redis. Will retry after %d seconds", redisWaitSeconds)
			fmt.Printf("Repush::Sleeping for %d Seconds...\n", redisWaitSeconds)
			fmt.Println("Trying to Re-establish connection to redis.")
			database.EstablishRedisQueueConnecion()
			time.Sleep(time.Duration(redisWaitSeconds) * time.Second)
			continue
		}

		path := globals.ApplicationConfig.Application.RepushLogPath

		// Slice to store file paths
		var files []string

		// Walk through the directory
		err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Check if it is a file (not a directory)
			if !info.IsDir() {
				if !strings.Contains(filePath, "_processed") && !strings.HasSuffix(filePath, ".log") && !strings.Contains(filePath, "_pending") {
					files = append(files, filePath)
				}
			}

			return nil
		})

		if err != nil {
			fmt.Printf("Error walking the provided file path %q: %v\n", path, err)
			globals.ApplicationWaitGroupServices.Done()
			return
		}

		// for
		fmt.Println("FILES :", files)
		for _, fileName := range files {
			if globals.FileCounter < maxProc {
				globals.ApplicationWaitGroupServices.Add(1)
				go FileParseAndRepush(fileName)
				globals.GlobalMutex.Lock()
				globals.FileCounter++
				globals.GlobalMutex.Unlock()
			}
		}

		fmt.Printf("Sleeping for %d seconds...\n", globals.ApplicationConfig.Application.RequeueSleepTimeSeconds)
		time.Sleep(time.Duration(globals.ApplicationConfig.Application.RequeueSleepTimeSeconds) * time.Second)
	}

}

func FileParseAndRepush(filename string) {
	defer globals.ApplicationWaitGroupServices.Done()

	// Return if memory full
	v, _ := mem.VirtualMemory()
	if v.UsedPercent >= float64(globals.ApplicationConfig.Application.RAMRepushthresholdPercent) {
		fmt.Println("Memory not sufficient skipping REPUSH")
		return
	}

	defer func() {
		globals.GlobalMutex.Lock()
		globals.FileCounter--
		globals.GlobalMutex.Unlock()
	}()
	fmt.Printf("Processing file: %s\n", filename)

	// Open the gzip file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", filename, err)
		return
	}
	defer file.Close()

	// Create a gzip reader
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		fmt.Printf("Error creating gzip reader for file %s: %v\n", filename, err)
		return
	}
	defer gzipReader.Close()

	// Create a buffered reader to read line by line
	scanner := bufio.NewScanner(gzipReader)
	var lineNumber = 1
	for scanner.Scan() {
		logLine := scanner.Text()

		// Process each line here
		// fmt.Println("Processing line:", logLine)
		/////
		// Define the regex patterns
		queuePattern := `QUEUE\[(.*?)\]`
		messagePattern := `MESSAGE\[(.*)\]$`

		// Compile the regex patterns
		queueRegex := regexp.MustCompile(queuePattern)
		messageRegex := regexp.MustCompile(messagePattern)

		// Extract the QUEUE value
		queueMatch := queueRegex.FindStringSubmatch(logLine)
		queueValue := ""
		if len(queueMatch) > 1 {
			queueValue = queueMatch[1]
		}

		// Extract the MESSAGE value
		messageMatch := messageRegex.FindStringSubmatch(logLine)
		messageValue := ""
		if len(messageMatch) > 1 {
			messageValue = messageMatch[1]
		}

		// Print the extracted values
		// fmt.Printf("QUEUE: %s\n", queueValue)
		// fmt.Printf("MESSAGE: %s\n", messageValue)

		// Performing Rpush
		if queueValue != "" && messageValue != "" {
			err := database.CustomRpush(queueValue, messageValue)
			if err != nil {
				// Rename pending if any issues with the push as _pending
				newFileName := strings.TrimSuffix(filename, ".gz") + "_pending.gz"
				rerr := os.Rename(filename, newFileName)
				if rerr != nil {
					fmt.Printf("Error renaming file %s to %s: %v\n", filename, newFileName, rerr)
				}
				fmt.Printf("Renamed file to: %s\n", newFileName)
				fmt.Printf("Repush Failed: Unable push to Redis. Log file %s at Line Number %d. Renamed file %s to %s. Host:[%s] Port:[%s]. Redis Error: %s", filename, lineNumber, filename, newFileName, globals.ApplicationConfig.Database.Host, globals.ApplicationConfig.Database.Port, err.Error())
				fmt.Printf("Repush Failed: Cannot establish connection to redis Host:%s | Port:%s.", globals.ApplicationConfig.RedisQueue.Host, globals.ApplicationConfig.RedisQueue.Port)

				util.SendSlackAlert(fmt.Sprintf("Repush:Unable push to Redis. Log file %s at Line Number %d. Renamed file %s to %s. Host:[%s] Port:[%s]. Redis Error: %s", filename, lineNumber, filename, newFileName, globals.ApplicationConfig.Database.Host, globals.ApplicationConfig.Database.Port, err.Error()))
				util.SendSlackAlert(fmt.Sprintf("Repush:Cannot establish connection to redis Host:%s | Port:%s.", globals.ApplicationConfig.RedisQueue.Host, globals.ApplicationConfig.RedisQueue.Port))
				return
			}
		}

		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
		util.SendSlackAlert(fmt.Sprintf("Repush:Unable to read queue log file %s: %s", filename, err.Error()))
		return
	}

	// // Rename the file by appending "_processed" as suffix
	newFileName := strings.TrimSuffix(filename, ".gz") + "_processed.gz"
	err = os.Rename(filename, newFileName)
	if err != nil {
		fmt.Printf("Error renaming file %s to %s: %v\n", filename, newFileName, err)
		return
	}
	fmt.Printf("Renamed file to: %s\n", newFileName)
}
