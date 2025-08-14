package redis

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
    "github.com/gocasters/rankr/domain/contributor"
)

type ContributorCache struct {
    adapter *Adapter
}

func NewContributorCache(adapter *Adapter) contributor.CacheRepository {
    return &ContributorCache{adapter: adapter}
}

func (c *ContributorCache) GetByID(ctx context.Context, id string) (*contributor.Contributor, error) {
    key := fmt.Sprintf("contributor:%s", id)
    data, err := c.adapter.Client.Get(ctx, key).Result()
    if err != nil {
        if err == redis.Nil {
            return nil, redis.Nil
        }
        return nil, err
    }
    var contrib contributor.Contributor
    err = json.Unmarshal([]byte(data), &contrib)
    return &contrib, err
}

func (c *ContributorCache) SetByID(ctx context.Context, contrib *contributor.Contributor) error {
    key := fmt.Sprintf("contributor:%s", contrib.ID)
    data, err := json.Marshal(contrib)
    if err != nil {
        return err
    }
    return c.adapter.Client.Set(ctx, key, data, 24*time.Hour).Err()
}

func (c *ContributorCache) DeleteByID(ctx context.Context, id string) error {
    key := fmt.Sprintf("contributor:%s", id)
    return c.adapter.Client.Del(ctx, key).Err()
}

func (c *ContributorCache) Clear(ctx context.Context) error {
    // Get all keys matching the pattern
    keys, err := c.adapter.Client.Keys(ctx, "contributor:*").Result()
    if err != nil {
        return err
    }
    
    // Delete all keys
    if len(keys) > 0 {
        return c.adapter.Client.Del(ctx, keys...).Err()
    }
    
    return nil
}
