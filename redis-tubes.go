package main

import (
  "flag"
  "./proxy"
  "gopkg.in/redis.v3"
)


var connid = uint64(0)
var localAddr = flag.String("l", "localhost:9999", "local address")
var verbose = flag.Bool("v", false, "display server actions")
var redis_client *redis.Client


type RedisPorts struct {
  key string
  redis_client *redis.Client
}

func (rp *RedisPorts) GetRandomPort() (string, error) {
  member, err := rp.redis_client.SRandMember(rp.key).Result()
  if err != nil {
    return "", err
  }
  return member, nil
}

func main() {
  flag.Parse()
  redis_client = redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "", // no password set
    DB:       0,  // use default DB
  })

  s := &proxy.Server{
    LocalAddr: *localAddr,
    Verbose:    *verbose,
    Ports: &RedisPorts{key: *localAddr, redis_client: redis_client},
  }
  s.Run()
}
