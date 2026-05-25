package redis

import (
	"context"
)

func (r *Redis) GetTopNRequests(n int) ([]string, error) {
	ctx := context.Background()
	requests, err := r.client.LRange(ctx, "requests", 0, int64(n-1)).Result()
	if err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *Redis) AddRequest(request string) error {
	ctx := context.Background()
	return r.client.LPush(ctx, "requests", request).Err()
}
