package lists

import (
	"errors"
	"log"

	"github.com/json-iterator/go"

	"github.com/mohuishou/scuplus-go/cache"
)

// Get 获取缓存,如果缓存不存在则新建缓存
func Get(k string) ([]byte, error) {
	v, err := cache.Do("GET", k)
	if err != nil {
		// 不存在缓存
		return nil, err
	}
	b, ok := v.([]byte)
	if !ok {
		return nil, errors.New("not")
	}
	return b, nil
}

// Set set
func Set(k string, data map[string]interface{}, expireTime int) error {
	v, err := jsoniter.Marshal(&data)
	if err != nil {
		return err
	}
	_, err = cache.Do("SET", k, v)
	if err != nil {
		log.Println("set cache token err:", err)
	}
	// 设置过期时间
	_, err = cache.Do("Expire", k, expireTime)
	if err != nil {
		log.Println("set cache code err:", err)
	}
	return err
}
