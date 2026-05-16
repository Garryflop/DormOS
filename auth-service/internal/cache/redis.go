package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr, password string, db int) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisCache{client: client}
}

func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

type CachedTokenClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	DormID string `json:"dorm_id"`
	Email  string `json:"email"`
}

func (c *RedisCache) SetTokenValidation(ctx context.Context, token string, claims *CachedTokenClaims, ttl time.Duration) error {
	data, err := json.Marshal(claims)
	if err != nil {
		return fmt.Errorf("marshal token claims: %w", err)
	}
	return c.client.Set(ctx, tokenKey(token), data, ttl).Err()
}

func (c *RedisCache) GetTokenValidation(ctx context.Context, token string) (*CachedTokenClaims, error) {
	data, err := c.client.Get(ctx, tokenKey(token)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("get token cache: %w", err)
	}
	var claims CachedTokenClaims
	if err := json.Unmarshal(data, &claims); err != nil {
		return nil, fmt.Errorf("unmarshal token claims: %w", err)
	}
	return &claims, nil
}

func (c *RedisCache) InvalidateToken(ctx context.Context, token string) error {
	return c.client.Del(ctx, tokenKey(token)).Err()
}

type CachedUserRole struct {
	Role   string `json:"role"`
	DormID string `json:"dorm_id"`
}

const userRoleTTL = 5 * time.Minute

func (c *RedisCache) SetUserRole(ctx context.Context, userID string, role *CachedUserRole) error {
	data, err := json.Marshal(role)
	if err != nil {
		return fmt.Errorf("marshal user role: %w", err)
	}
	return c.client.Set(ctx, userRoleKey(userID), data, userRoleTTL).Err()
}

func (c *RedisCache) GetUserRole(ctx context.Context, userID string) (*CachedUserRole, error) {
	data, err := c.client.Get(ctx, userRoleKey(userID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("get user role cache: %w", err)
	}
	var role CachedUserRole
	if err := json.Unmarshal(data, &role); err != nil {
		return nil, fmt.Errorf("unmarshal user role: %w", err)
	}
	return &role, nil
}

func (c *RedisCache) InvalidateUserRole(ctx context.Context, userID string) error {
	return c.client.Del(ctx, userRoleKey(userID)).Err()
}

func tokenKey(token string) string {
	if len(token) > 16 {
		return "auth:token:" + token[:16]
	}
	return "auth:token:" + token
}

func userRoleKey(userID string) string {
	return "auth:role:" + userID
}
