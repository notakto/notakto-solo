package concurrency

import (
	"context"
)

func (g *RedisGuard) Release(ctx context.Context, key string) error {
	lockKey := g.prefix + key

	lua := `
if redis.call("GET", KEYS[1]) then
	return redis.call("DEL", KEYS[1])
end
return 0
`
	return g.rdb.Eval(ctx, lua, []string{lockKey}).Err()
}
