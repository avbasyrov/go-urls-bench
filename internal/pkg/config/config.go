package config

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"time"
)

type Params struct {
	RequestTimeout        time.Duration `yaml:"request-timeout"`
	ThrottleMultiplier    float64       `yaml:"throttle-multiplier"`
	ThrottleMinimalMargin time.Duration `yaml:"throttle-minimal-margin"`
	QueryCacheTtl         time.Duration `yaml:"query-cache-ttl"`
	MaxResponseTime       time.Duration `yaml:"max-response-time"`
}

func Parse() (*Params, error) {
	p := &Params{}

	yamlFile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, p)
	if err != nil {
		return nil, err
	}

	return p, nil
}
