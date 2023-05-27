package main

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/pilu/go-base62"
)

const (
	URLIDKEY           = "next.url.id"
	ShortlinkKey       = "shortlink:%s:url"
	URLHashKey         = "urlhash:%s:url"
	ShortlinkDetailKey = "shortlink:%s:detail"
)

type RedisCli struct {
	Cli *redis.Client
}

type URLDetail struct {
	URL                 string        `json:"url"`
	CreatedAt           string        `json:"created_at"`
	ExpirationInMinutes time.Duration `json:"expiration_in_minutes"`
}

func NewRedisCli(addr string, passwd string, db int) *RedisCli {
	c := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: passwd,
		DB:       db,
	})

	if _, err := c.Ping().Result(); err != nil {
		panic(err)
	}
	return &RedisCli{Cli: c}
}

func (r *RedisCli) Shorten(url string, exp int64) (string, error) {
	h := toSha1(url)

	d, err := r.Cli.Get(fmt.Sprintf(URLHashKey, h)).Result()
	if err == redis.Nil {
		// 不存在

	} else if err != nil {
		return "", err
	} else {
		if d == "{}" {
		} else {
			return d, nil
		}
	}
	err = r.Cli.Incr(URLIDKEY).Err()
	if err != nil {
		return "", err
	}

	id, err := r.Cli.Get(URLIDKEY).Int64()
	if err != nil {
		return "", err
	}
	eid := base62.Encode(int(id))

	err = r.Cli.Set(fmt.Sprintf(ShortlinkKey, eid), url, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	err = r.Cli.Set(fmt.Sprintf(URLHashKey, eid), url, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	detail, err := json.Marshal(&URLDetail{
		URL:                 url,
		CreatedAt:           time.Now().String(),
		ExpirationInMinutes: time.Duration(exp),
	})
	if err != nil {
		return "", err
	}

	err = r.Cli.Set(fmt.Sprintf(ShortlinkDetailKey, eid), detail, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	return eid, nil
}

func (r *RedisCli) ShortlinkInfo(eid string) (interface{}, error) {
	d, err := r.Cli.Get(fmt.Sprintf(ShortlinkDetailKey, eid)).Result()
	if err == redis.Nil {
		return "", StatusError{404, errors.New("Unknown short URL")}
	} else if err != nil {
		return "", err
	} else {
		return d, nil
	}
}

func (r *RedisCli) Unshorten(eid string) (string, error) {
	url, err := r.Cli.Get(fmt.Sprintf(ShortlinkKey, eid)).Result()
	if err == redis.Nil {
		return "", StatusError{404, err}
	} else if err != nil {
		return "", err
	} else {
		return url, nil
	}
}

// ToSha1 ...
func toSha1(str string) string {
	var (
		sha = sha1.New()
	)
	return string(sha.Sum([]byte(str)))
}
