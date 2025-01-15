package database

import (
	"context"
	"fmt"
	"main/logs"
	"main/pkg/globals"
	"time"

	// "wh_dequeuer/pkg/globals"

	"github.com/redis/go-redis/v9"
)

var RedisQueueClient *redis.Client
var RedisCacheClient *redis.Client

var Ctx = context.Background()

func CheckConnection() bool {
	if err := RedisQueueClient.Ping(Ctx).Err(); err != nil {
		return false
	}
	return true
}

func EstablishRedisQueueConnecion() {
	fmt.Println("Establishing Redis Queue Connection")
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v", globals.ApplicationConfig.RedisQueue.Host+":"+globals.ApplicationConfig.RedisQueue.Port),
		Username: globals.ApplicationConfig.RedisQueue.Username,
		Password: globals.ApplicationConfig.RedisQueue.Password, // no password set
		DB:       0,                                             // use default DB
	})

	RedisQueueClient = rdb
	fmt.Println("Redis Queue Connection Estblished....")
}

func EstablishRedisCacheConnecion() {
	fmt.Println("Establishing Redis Cache Connection")
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v", globals.ApplicationConfig.RedisCache.Host+":"+globals.ApplicationConfig.RedisCache.Port),
		Username: globals.ApplicationConfig.RedisCache.Username,
		Password: globals.ApplicationConfig.RedisCache.Password, // no password set
		DB:       0,                                             // use default DB

	})
	RedisCacheClient = rdb
	fmt.Println("Redis Cache Estblished....")
}

func CustomBLpop(listName string) ([]string, error) {
	// fmt.Println("In custom pop function")
	if RedisQueueClient == nil {
		EstablishRedisQueueConnecion()
	}
	poppedString, err := RedisQueueClient.BLPop(Ctx, time.Second*1, listName).Result()
	// fmt.Println("poppedString::%v", poppedString)
	// logs.ErrorLog("poppedString::%v", poppedString)
	if err != nil {
		// fmt.Println("err", err)
		// fmt.Println("errString", err.Error())
		return nil, err
	}

	return poppedString, nil
}

func CustomLlen(listName string) int {

	if RedisQueueClient == nil {
		EstablishRedisQueueConnecion()
	}
	lenght, err := RedisQueueClient.LLen(Ctx, listName).Result()
	if err != nil {
		logs.ErrorLog("Error to perform llen operation: %v", err)
		return 0
	}

	return int(lenght)
}

func CustomRpush(listName string, stringParam string) error {
	// fmt.Println("Performing rpush, QUEUE[%v], String[%v]", listName, stringParam)
	if RedisQueueClient == nil {
		EstablishRedisQueueConnecion()
	}
	_, err := RedisQueueClient.RPush(Ctx, listName, stringParam).Result()
	if err != nil {
		logs.ErrorLog("Error to perform rpush, QUEUE[%v], String[%v]", listName, stringParam)
		logs.ErrorLog("Error in CustomRpush function, %v", err)
		return err
	}
	return nil
}

func CustomSetKey(key string, value string, expiration time.Duration) {
	if RedisCacheClient == nil {
		EstablishRedisCacheConnecion()
	}
	err := RedisCacheClient.Set(Ctx, key, value, expiration).Err()
	if err != nil {
		logs.ErrorLog("Error in CustomSetKey function, %v", err)

	}

}

func CustomHgetAll(hashKey string) map[string]string {
	fmt.Printf("HashKey is :: %v", hashKey)
	if RedisCacheClient == nil {
		EstablishRedisCacheConnecion()
	}
	hashMap, err := RedisCacheClient.HGetAll(Ctx, hashKey).Result()
	if err != nil {
		logs.ErrorLog("Error in CustomHGetAll function, %v", err)
		return nil
	}
	fmt.Printf("Map for client formed :: %v", hashMap)
	return hashMap
}

func CustomHGet(hashKey string, key string) string {
	fmt.Printf("CustomHGet :: HashKey::%v :: key:: %v", hashKey, key)
	if RedisCacheClient == nil {
		EstablishRedisCacheConnecion()
	}
	value, err := RedisCacheClient.HGet(Ctx, hashKey, key).Result()

	if err != nil {
		logs.ErrorLog("Error in CustomHGet function, %v", err)
		return ""
	}
	return value
}

func CustomHSet(hashKey string, key string, field string) error {
	fmt.Printf("CustomHSet :: HashKey::%v :: key:: %v :: field :: %v", hashKey, key, field)
	if RedisCacheClient == nil {
		EstablishRedisCacheConnecion()
	}
	_, err := RedisCacheClient.HSet(Ctx, hashKey, key, field).Result()

	if err != nil {
		logs.ErrorLog("Error in CustomHSet function, %v", err)
		return err
	}
	return nil
}
