package database

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ZhanibekTau/go-sdk/pkg/config"
	"github.com/ZhanibekTau/go-sdk/pkg/logger"
	"github.com/redis/go-redis/v9"
	"time"
)

// NewRedisHelper - Новый Хелпер для работы с редисом
func NewRedisHelper[E interface{}](redisClient *redis.Client) *RedisHelper[E] {
	return &RedisHelper[E]{
		redisClient: redisClient,
	}
}

// RedisHelper - Хелпер для работы с редисом
type RedisHelper[E interface{}] struct {
	redisClient *redis.Client
	appInfo     *config.AppInfo
	result      E
}

// SetRequestData - установить Доп данные для запроса (используется для логирования)
func (redisHelper *RedisHelper[E]) SetRequestData(appInfo *config.AppInfo) *RedisHelper[E] {
	redisHelper.appInfo = appInfo

	return redisHelper
}

// GetByModel Возвращает значение по ключу
func (redisHelper *RedisHelper[E]) GetByModel(key string) (*E, error) {
	ctx := context.Background()
	val, err := redisHelper.redisClient.Get(ctx, key).Result()

	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	if val == "" {
		return nil, nil
	}

	unMarshErr := json.Unmarshal([]byte(val), &redisHelper.result)

	if unMarshErr != nil {
		return nil, unMarshErr
	}

	if redisHelper.appInfo != nil {
		logger.FormattedLogWithAppInfo(redisHelper.appInfo, "GOT DATA FROM CACHE: "+val)
	} else {
		println("GOT DATA FROM CACHE: " + val)
	}

	return &redisHelper.result, nil
}

// SetByModel Записывает значение по ключу
func (redisHelper *RedisHelper[E]) SetByModel(key string, model *E, ttl time.Duration) error {
	jsonModel, err := json.Marshal(model)

	if err != nil {

		return err
	}

	ctx := context.Background()
	err = redisHelper.redisClient.Set(ctx, key, jsonModel, ttl).Err()

	if err != nil {
		return err
	}

	if redisHelper.appInfo != nil {
		logger.FormattedLogWithAppInfo(redisHelper.appInfo, "SET DATA TO CACHE: "+string(jsonModel))
	} else {
		println("SET DATA TO CACHE: " + string(jsonModel))
	}

	return nil
}

// GetString Возвращает значение по ключу
func (redisHelper *RedisHelper[E]) GetString(key string) (string, error) {
	ctx := context.Background()
	val, err := redisHelper.redisClient.Get(ctx, key).Result()

	if err != nil && !errors.Is(err, redis.Nil) {
		return "", err
	}

	if val == "" {
		return "", nil
	}

	if redisHelper.appInfo != nil {
		logger.FormattedLogWithAppInfo(redisHelper.appInfo, "GOT DATA FROM CACHE: "+val)
	} else {
		println("GOT DATA FROM CACHE: " + val)
	}

	return val, nil
}

// SetString Записывает значение по ключу
func (redisHelper *RedisHelper[E]) SetString(key string, string string, ttl time.Duration) error {
	ctx := context.Background()
	err := redisHelper.redisClient.Set(ctx, key, string, ttl).Err()

	if err != nil {
		return err
	}

	if redisHelper.appInfo != nil {
		logger.FormattedLogWithAppInfo(redisHelper.appInfo, "SET DATA TO CACHE: "+string)
	} else {
		println("SET DATA TO CACHE: " + string)
	}

	return nil
}

// GetArrayOfPointerModels Возвращает list по ключу
func (redisHelper *RedisHelper[E]) GetArrayOfPointerModels(key string) ([]*E, error) {
	resultArr := make([]*E, 0)
	ctx := context.Background()
	val, err := redisHelper.redisClient.Get(ctx, key).Result()

	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	if val == "" {
		return nil, nil
	}

	unMarshErr := json.Unmarshal([]byte(val), &resultArr)

	if unMarshErr != nil {
		return nil, unMarshErr
	}

	if redisHelper.appInfo != nil {
		logger.FormattedLogWithAppInfo(redisHelper.appInfo, "GOT LIST FROM CACHE: "+val)
	} else {
		println("GOT LIST FROM CACHE: " + val)
	}

	if len(resultArr) == 0 {
		return nil, nil
	}

	return resultArr, nil
}

// SetArrayOfPointerModels Записывает list по ключу
func (redisHelper *RedisHelper[E]) SetArrayOfPointerModels(key string, models []*E, ttl time.Duration) error {
	if len(models) == 0 {
		return nil
	}

	str, err := json.Marshal(models)

	if err != nil {

		return err
	}

	ctx := context.Background()

	rErr := redisHelper.redisClient.Set(ctx, key, str, ttl).Err()

	if rErr != nil {
		return rErr
	}

	if redisHelper.appInfo != nil {
		logger.FormattedLogWithAppInfo(redisHelper.appInfo, "SET LIST TO CACHE: "+string(str))
	} else {
		println("SET DATA TO CACHE: " + string(str))
	}

	return nil
}

// GetPointerArrayOfModels Возвращает list по ключу
func (redisHelper *RedisHelper[E]) GetPointerArrayOfModels(key string) (*[]E, error) {
	resultArr := make([]E, 0)
	ctx := context.Background()
	val, err := redisHelper.redisClient.Get(ctx, key).Result()

	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	if val == "" {
		return nil, nil
	}

	unMarshErr := json.Unmarshal([]byte(val), &resultArr)

	if unMarshErr != nil {
		return nil, unMarshErr
	}

	if redisHelper.appInfo != nil {
		logger.FormattedLogWithAppInfo(redisHelper.appInfo, "GOT LIST FROM CACHE: "+val)
	} else {
		println("GOT LIST FROM CACHE: " + val)
	}

	if len(resultArr) == 0 {
		return nil, nil
	}

	return &resultArr, nil
}

// SetPointerArrayOfModels Записывает list по ключу
func (redisHelper *RedisHelper[E]) SetPointerArrayOfModels(key string, models *[]E, ttl time.Duration) error {
	if len(*models) == 0 {
		return nil
	}

	str, err := json.Marshal(models)

	if err != nil {

		return err
	}

	ctx := context.Background()

	rErr := redisHelper.redisClient.Set(ctx, key, str, ttl).Err()

	if rErr != nil {
		return rErr
	}

	if redisHelper.appInfo != nil {
		logger.FormattedLogWithAppInfo(redisHelper.appInfo, "SET LIST TO CACHE: "+string(str))
	} else {
		println("SET DATA TO CACHE: " + string(str))
	}

	return nil
}
