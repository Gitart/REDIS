## Iterating over keys

It's not recommended to use the `KEYS prefix:*` command to get all the keys in a Redis instance, especially in production environments, because it can be a slow and resource-intensive operation that can impact the performance of the Redis instance.

Instead, you can iterate over Redis keys that match some pattern using the [SCANopen in new window](https://redis.io/commands/scan) command:

```
var cursor uint64
for {
	var keys []string
	var err error
	keys, cursor, err = rdb.Scan(ctx, cursor, "prefix:*", 0).Result()
	if err != nil {
		panic(err)
	}

	for _, key := range keys {
		fmt.Println("key", key)
	}

	if cursor == 0 { // no more keys
		break
	}
}
```

go-redis allows to simplify the code above to:

```
iter := rdb.Scan(ctx, 0, "prefix:*", 0).Iterator()
for iter.Next(ctx) {
	fmt.Println("keys", iter.Val())
}
if err := iter.Err(); err != nil {
	panic(err)
}
```

## [#](#sets-and-hashes) Sets and hashes

You can also iterate over set elements:

```
iter := rdb.SScan(ctx, "set-key", 0, "prefix:*", 0).Iterator()
```

And hashes:

```
iter := rdb.HScan(ctx, "hash-key", 0, "prefix:*", 0).Iterator()
iter := rdb.ZScan(ctx, "sorted-hash-key", 0, "prefix:*", 0).Iterator()
```

## [#](#cluster-and-ring) Cluster and Ring

If you are using [Redis Cluster](/guide/cluster.html) or [Redis Ring](/guide/ring.html), you need to scan each cluster node separately:

```
err := rdb.ForEachMaster(ctx, func(ctx context.Context, rdb *redis.Client) error {
	iter := rdb.Scan(ctx, 0, "prefix:*", 0).Iterator()

	...

	return iter.Err()
})
if err != nil {
	panic(err)
}
```

## [#](#delete-keys-without-ttl) Delete keys without TTL

You can also use `SCAN` to delete keys without a TTL:

```
iter := rdb.Scan(ctx, 0, "", 0).Iterator()

for iter.Next(ctx) {
	key := iter.Val()

    d, err := rdb.TTL(ctx, key).Result()
    if err != nil {
        panic(err)
    }

    if d == -1 { // -1 means no TTL
        if err := rdb.Del(ctx, key).Err(); err != nil {
            panic(err)
        }
    }
}

if err := iter.Err(); err != nil {
	panic(err)
}
```

For a more efficient version that uses pipelines see the [exampleopen in new window](https://github.com/redis/go-redis/tree/master/example/del-keys-without-ttl).

## [#](#monitoring-performance) Monitoring Performance

To [monitor go-redis performance](/guide/go-redis-monitoring.html), you can use OpenTelemetry instrumentation that comes with go-redis and Uptrace.

Uptrace is an open source [DataDog competitoropen in new window](https://uptrace.dev/blog/datadog-competitors.html) that supports [OpenTelemetry tracingopen in new window](https://uptrace.dev/opentelemetry/distributed-tracing.html), [OpenTelemetry metricsopen in new window](https://uptrace.dev/opentelemetry/metrics.html), and logs. You can use it to monitor applications and set up automatic alerts to receive notifications via email, Slack, Telegram, and more.
