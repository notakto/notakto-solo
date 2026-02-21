package lua

import "github.com/redis/go-redis/v9"

// Unlock deletes a key only if its value matches the provided nonce.
// Prevents a request from releasing a lock it no longer owns.
// KEYS[1] = lock key, ARGV[1] = expected nonce value
var Unlock = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`)
