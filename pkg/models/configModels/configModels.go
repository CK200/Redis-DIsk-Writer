package models

type Config struct {
	InstanceName string `json:"instanceName"`
	Database     struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		Dbname   string `json:"dbname"`
	} `json:"database"`

	RedisQueue struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		Dbname   string `json:"dbname"`
	} `json:"redisQueue"`

	RedisCache struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		Dbname   string `json:"dbname"`
	} `json:"redisCache"`

	Application struct {
		LogPath                    string `json:"logPath"`
		Workers                    string `json:"workers"`
		RAMthresholdPercent        int    `json:"ramWriterThresholdPercent"`
		RAMRepushthresholdPercent  int    `json:"ramRepushThresholdPercent"`
		RepushLogPath              string `json:"repushLogPath"`
		MonitorSleepTimeSeconds    int    `json:"monitorSleepTimeSeconds"`
		RequeueSleepTimeSeconds    int    `json:"requeueSleepTimeSeconds"`
		EnableRespush              bool   `json:"enableRepush"`
		RepushWaitSeconds          int    `json:"repushWaitSeconds"`
		RedisConnectionTimeSeconds int    `json:"redisConnectionTimeSeconds"`
		MaxRepushProcesses         int    `json:"maxRepushProcesses"`
		SlackWebhook               string `json:"slackWebhook"`
	} `json:"application"`
}
