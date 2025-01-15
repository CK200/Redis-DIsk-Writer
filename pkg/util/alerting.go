package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"main/pkg/globals"
	"net/http"
)

type Slack struct {
	Text string `json:"text"`
}

func SendSlackAlert(message string) {
	instance := globals.ApplicationConfig.InstanceName
	slackHook := globals.ApplicationConfig.Application.SlackWebhook
	todo := Slack{(instance + message)}
	jsonReq, _ := json.Marshal(todo)
	resp, err := http.Post(slackHook, "application/json; charset=utf-8", bytes.NewBuffer(jsonReq))
	if err != nil {
		fmt.Println("SLACK ERROR in queueOnDisk application ::", err)
	}

	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)

	bodyString := string(bodyBytes)
	fmt.Println(bodyString)
}
