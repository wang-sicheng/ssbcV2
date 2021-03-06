package redis

import (
	"context"
	"github.com/cloudflare/cfssl/log"
	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var rdb = redis.NewClient(&redis.Options{
	Addr:     "127.0.0.1:6379",
	Password: "",
})

//初始化
//func init() {
//	rdb = redis.NewClient(&redis.Options{
//		Addr:     "localhost:6380",
//		Password: "", // no password set
//		DB:       0,  // use default DB
//	})
//}

//set
func SetIntoRedis(key string, value string) error {
	err := rdb.Set(ctx, key, value, 0).Err()
	if err != nil {
		panic(err)
	}
	return err
}

//get
func GetFromRedis(key string) (string, error) {
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		log.Errorf("the key:%s does not exist\n", key)
		return "", nil
	} else if err != nil {
		panic(err)
	} else {
		return val, nil
	}
}

// list push
func PushToList(key string, value string) error {
	err := rdb.RPush(ctx, key, value).Err()
	if err != nil {
		log.Errorf("event push to list error: %s", err)
		return err
	}
	return nil
}

func ExampleClient() {
	err := rdb.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		panic(err)
	}
	val, err := rdb.Get(ctx, "key").Result()
	if err != nil {
		panic(err)
	}
	log.Info("key", val)

	val2, err := rdb.Get(ctx, "key2").Result()
	if err == redis.Nil {
		log.Info("key2 does not exist")
	} else if err != nil {
		panic(err)
	} else {
		log.Info("key2", val2)
	}
	// Output: key value
	// key2 does not exist
}
