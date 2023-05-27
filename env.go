package main

type Env struct {
	S Storage
}

func getEnv() *Env {
	r := NewRedisCli("my-redis:6379", "", 0)

	return &Env{S: r}
}
