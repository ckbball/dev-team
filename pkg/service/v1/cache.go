package v1

import (
  "github.com/go-redis/cache/v7"
)

type cache interface {
  AddEntry()
  GetEntry()
}

type redisCache struct{}

func NewRedisCache() *redisCache {
  return &redisCache{}
}
