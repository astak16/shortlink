package main

func main() {

	a := App{}
	a.Initialize(getEnv())
	a.Run(":8080")
}

// func connectRedis() {
// 	client := redis.NewClient(&redis.Options{
// 		Addr:     "my-redis:6379",
// 		Password: "", // 没有密码
// 		DB:       0,  // 默认数据库
// 	})

// 	ctx := context.Background()

// 	// 测试连接是否成功
// 	pong, err := client.Ping(ctx).Result()
// 	fmt.Println(pong, err)

// 	err = client.Set(ctx, "foo", "bar", 0).Err()
// 	if err != nil {
// 		panic(err)
// 	}

// 	// 获取一个键的值
// 	val, err := client.Get(ctx, "foo").Result()
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("foo", val)
// }
