mackerel-plugin-redis-pubsub
============================

Redis custom metrics plugin for mackerel.io agent.

## Synopsis

```shell
mackerel-plugin-redis-pubsub [-pubaddr=<pub hostname:port>] [-pubpassword=<pub password>] [-pubdb=<pub db>] [-subaddr=<sub hostname:port>] [-subpassword=<sub password>] [-subdb=<sub db>] [-metric-key-prefix=<prefix>]
```

```
Usage of mackerel-plugin-redis-pubsub:
  -metric-key-prefix string
        Metric key prefix (default "latency")
  -msg string
        publish message (default "Publish message")
  -n string
        channel name (default "test")
  -pubaddr string
        redis pub address  (default "localhost:6379")
  -pubdb int
        redis pub db number (default: 0)
  -pubpassword string
        redis pub password (default:"")
  -subaddr string
        redis sub address  (default "localhost:6379")
  -subdb int
        redis sub db number (default: 0)
  -subpassword string
        redis sub password (default:"")
  -tempfile string
        Temp file name
```

## Example of mackerel-agent.conf

```
[plugin.metrics.redis-pubsub]
command = "/path/to/mackerel-plugin-redis-pubsub -pubaddr=127.0.0.1:6379 -subaddr=127.0.0.1:6379
```


## References

- http://redis.io/commands/INFO

