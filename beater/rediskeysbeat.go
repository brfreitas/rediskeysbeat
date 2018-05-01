package beater

import (
	"fmt"
	"strconv"
	"time"

	"github.com/brfreitas/rediskeysbeat/config"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/gomodule/redigo/redis"
)

type Rediskeysbeat struct {
	done      chan struct{}
	config    config.Config
	client    beat.Client
	redisPool *redis.Pool
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Rediskeysbeat{
		done:   make(chan struct{}),
		config: c,
	}
	return bt, nil
}

func (bt *Rediskeysbeat) setup(b *beat.Beat) error {
	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}
	//bt.events = b.Publisher.Connect()
	//bt.done = make(chan struct{})

	// Set up redis pool
	redisPool := redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial(bt.config.Network, bt.config.Host+":"+strconv.Itoa(bt.config.Port))
		if err != nil {
			return nil, err
		}

		return c, err
	}, bt.config.MaxConn)

	bt.redisPool = redisPool
	if bt.config.Auth.Required {
		c := bt.redisPool.Get()
		defer c.Close()

		authed, err := c.Do("AUTH", bt.config.Auth.RequiredPass)
		if err != nil {
			return err
		} else {
			logp.Debug("rediskeybeat", "AUTH %v", authed)
		}
	}

	return nil
}

func (bt *Rediskeysbeat) Run(b *beat.Beat) error {
	logp.Info("rediskeysbeat is running! Hit CTRL-C to stop it.")
	var err error

	err = bt.setup(b)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(bt.config.Period)
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}
		keys, err := bt.getKeys(bt.config.KeyPattern)
		if err != nil {
			return err
		}

		for i := 0; i < len(keys); i++ {
			fields, err := bt.getValues(keys[i])
			if err != nil {
				return err
			}
			event := beat.Event{
				Timestamp: time.Now(),
				Fields: common.MapStr{
					"type":          bt.config.KeyEntity,
					"key":           keys[i],
					"fields-values": fields,
				},
			}
			bt.client.Publish(event)
			logp.Info("Event sent")
		}
	}
}

func (bt *Rediskeysbeat) getKeys(pattern string) ([]string, error) {

	conn := bt.redisPool.Get()
	defer conn.Close()

	if bt.config.Auth.Required {
		_, err := conn.Do("AUTH", bt.config.Auth.RequiredPass)
		if err != nil {
			return nil, err
		}
	}

	iter := 0
	keys := []string{}
	for {
		arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", pattern, "COUNT", 100000))
		if err != nil {
			return keys, fmt.Errorf("error retrieving '%s' keys", pattern)
		}
		iter, _ = redis.Int(arr[0], nil)
		k, _ := redis.Strings(arr[1], nil)
		keys = append(keys, k...)
		if iter == 0 {
			break
		}
	}

	return keys, nil
}

func (bt *Rediskeysbeat) getValues(key string) ([]string, error) {

	conn := bt.redisPool.Get()
	defer conn.Close()

	if bt.config.Auth.Required {
		_, err := conn.Do("AUTH", bt.config.Auth.RequiredPass)
		if err != nil {
			return nil, err
		}
	}

	values, err := redis.Values(conn.Do("HGETALL", key))
	if err != nil {
		logp.Debug("rediskeybeat", "HGETALL", err)
	}
	result := []string{}
	for i := 0; i < len(values); i += 2 {
		key, _ := redis.String(values[i], nil)
		value, _ := redis.String(values[i+1], nil)

		result = append(result, fmt.Sprintf("%s:%s", key, value))
	}
	return result, nil
}

func (bt *Rediskeysbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
