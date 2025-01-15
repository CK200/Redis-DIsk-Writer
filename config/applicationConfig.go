package config

import (
	"encoding/json"
	"fmt"
	"main/logs"
	database "main/pkg/database/redis"
	"main/pkg/globals"
	models "main/pkg/models/configModels"
	"os"
	"strconv"
)

func SetUpApplication() {
	fmt.Println("SetUp config details....")
	setupConfig()
	fmt.Println("SettingUp Logs....")
	setUpApplicationLogs()
	setUpApplicationDatabase()
	fmt.Println("DB set up done")
	globals.Shutdown = false
	NumberOfWorkers, err := strconv.Atoi(globals.ApplicationConfig.Application.Workers)
	if err != nil {
		globals.Workers = 2
	} else {
		globals.Workers = NumberOfWorkers
	}
	fmt.Println("Workers: ", globals.Workers)
}

func setupConfig() {

	file, err := os.ReadFile("./config/config.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Unmarshal the JSON data into a Config struct
	var config models.Config

	err = json.Unmarshal(file, &config)
	if err != nil {
		fmt.Println(err)
		return
	}

	globals.ApplicationConfig = &config

}

func setUpApplicationLogs() {
	logs.SetUpQueueLogs()
}

func setUpApplicationDatabase() {
	// databaseMySql.EstablishDbConnection()
	//database.EstablishRedisCacheConnecion()
	database.EstablishRedisQueueConnecion()
}
