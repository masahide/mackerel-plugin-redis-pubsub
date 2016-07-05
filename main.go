package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	mp "github.com/mackerelio/go-mackerel-plugin-helper"
	"github.com/mackerelio/mackerel-agent/logging"
	redis "gopkg.in/redis.v4"
)

var (
	logger = logging.GetLogger("metrics.plugin.redis")
	//workerNum       = 1
	redisParam = RedisPlugin{
		PubRedisOpt: redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		},
		SubRedisOpt: redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		},
		ChannelName: "test",
		Message:     "Publish message",
		Prefix:      "redis.pubsub.latency",
	}
)

// RedisPlugin mackerel plugin for Redis
type RedisPlugin struct {
	PubRedisOpt redis.Options
	SubRedisOpt redis.Options
	ChannelName string
	Message     string
	Prefix      string
	Tempfile    string
}

// FetchMetrics interface for mackerelplugin
func (m RedisPlugin) FetchMetrics() (map[string]interface{}, error) {

	pubClient := redis.NewClient(&m.PubRedisOpt)
	subClient := redis.NewClient(&m.SubRedisOpt)

	subscribe, err := subClient.Subscribe(m.ChannelName)
	if err != nil {
		logger.Errorf("Failed to subscribe. %s", err)
		return nil, err
	}
	defer subscribe.Close()
	if _, err := pubClient.Publish(m.ChannelName, m.Message).Result(); err != nil {
		logger.Errorf("Failed to publish. %s", err)
		return nil, err
	}
	start := time.Now()
	if _, err := subscribe.ReceiveMessage(); err != nil {
		logger.Infof("Failed to calculate capacity. (The cause may be that AWS Elasticache Redis has no `CONFIG` command.) Skip these metrics. %s", err)
		return nil, err
	}
	duration := time.Now().Sub(start)

	return map[string]interface{}{m.metricName(): float64(duration) / float64(time.Microsecond)}, nil

}

func (m RedisPlugin) metricName() string {
	return strings.Replace(strings.Replace(m.SubRedisOpt.Addr, ".", "_", -1), ":", "-", -1)
}

// GraphDefinition interface for mackerelplugin
func (m RedisPlugin) GraphDefinition() map[string](mp.Graphs) {
	labelPrefix := strings.Title(m.Prefix)
	return map[string](mp.Graphs){
		m.Prefix: mp.Graphs{
			Label: labelPrefix,
			Unit:  "float",
			Metrics: [](mp.Metrics){
				mp.Metrics{Name: m.metricName(), Label: m.SubRedisOpt.Addr},
			},
		},
	}
}

func main() {

	redisParam.Prefix = *flag.String("metric-key-prefix", redisParam.Prefix, "Metric key prefix")
	redisParam.Message = *flag.String("msg", redisParam.Message, "publish message")
	redisParam.PubRedisOpt.Password = *flag.String("pubpassword", redisParam.PubRedisOpt.Password, "redis pub password (default:\"\")")
	redisParam.PubRedisOpt.Addr = *flag.String("pubaddr", redisParam.PubRedisOpt.Addr, "redis pub address ")
	redisParam.PubRedisOpt.DB = *flag.Int("pubdb", redisParam.PubRedisOpt.DB, "redis pub db number (default: 0)")
	redisParam.SubRedisOpt.Password = *flag.String("subpassword", redisParam.SubRedisOpt.Password, "redis sub password (default:\"\")")
	redisParam.SubRedisOpt.Addr = *flag.String("subaddr", redisParam.SubRedisOpt.Addr, "redis sub address ")
	redisParam.SubRedisOpt.DB = *flag.Int("subdb", redisParam.SubRedisOpt.DB, "redis sub db number (default: 0)")
	redisParam.ChannelName = *flag.String("n", redisParam.ChannelName, "channel name")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	flag.Parse()

	helper := mp.NewMackerelPlugin(redisParam)

	helper.Tempfile = *optTempfile
	if helper.Tempfile == "" {
		helper.Tempfile = fmt.Sprintf("/tmp/mackerel-plugin-redis-%s-%s", redisParam.PubRedisOpt.Addr, redisParam.SubRedisOpt.Addr)
	}
	helper.Run()
}
