package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	mp "github.com/mackerelio/go-mackerel-plugin"
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
		Prefix:      "latency",
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

func (m RedisPlugin) calculateCapacity(pub, sub *redis.Client, stat map[string]float64) error {
	subscribe, err := sub.Subscribe(m.ChannelName)
	if err != nil {
		logger.Errorf("Failed to subscribe. %s", err)
		return err
	}
	defer subscribe.Close()
	if _, err := pub.Publish(m.ChannelName, m.Message).Result(); err != nil {
		logger.Errorf("Failed to publish. %s", err)
		return err
	}
	start := time.Now()
	if _, err := subscribe.ReceiveMessage(); err != nil {
		log.Fatal(err)
	}
	duration := time.Now().Sub(start)

	stat["latency"] = float64(duration) / float64(time.Microsecond)
	return nil
}

// FetchMetrics interface for mackerelplugin
func (m RedisPlugin) FetchMetrics() (map[string]float64, error) {
	pubClient := redis.NewClient(&m.PubRedisOpt)
	subClient := redis.NewClient(&m.SubRedisOpt)

	stat := make(map[string]float64)

	if _, ok := stat["latency"]; !ok {
		stat["latency"] = 0
	}

	if err := m.calculateCapacity(pubClient, subClient, stat); err != nil {
		logger.Infof("Failed to calculate capacity. (The cause may be that AWS Elasticache Redis has no `CONFIG` command.) Skip these metrics. %s", err)
	}

	return stat, nil
}

// GraphDefinition interface for mackerelplugin
func (m RedisPlugin) GraphDefinition() map[string](mp.Graphs) {
	labelPrefix := strings.Title(m.Prefix)

	var graphdef = map[string](mp.Graphs){
		(m.Prefix + ".latency"): mp.Graphs{
			Label: (labelPrefix + " Latency"),
			Unit:  "integer",
			Metrics: [](mp.Metrics){
				mp.Metrics{Name: "latency", Label: "Latency", Diff: false},
			},
		},
	}

	return graphdef
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

	if *optTempfile != "" {
		helper.Tempfile = *optTempfile
	} else {
		helper.Tempfile = fmt.Sprintf("/tmp/mackerel-plugin-redis-%s-%s", redisParam.PubRedisOpt.Addr, redisParam.SubRedisOpt.Addr)
	}

	if os.Getenv("MACKEREL_AGENT_PLUGIN_META") != "" {
		helper.OutputDefinitions()
	} else {
		helper.OutputValues()
	}
}
