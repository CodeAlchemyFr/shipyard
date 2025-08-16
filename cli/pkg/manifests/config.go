package manifests

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config represents the paas.yaml configuration file
type Config struct {
	App       AppConfig       `yaml:"app"`
	Build     BuildConfig     `yaml:"build,omitempty"`
	Resources ResourcesConfig `yaml:"resources,omitempty"`
	Scaling   ScalingConfig   `yaml:"scaling,omitempty"`
	Health    HealthConfig    `yaml:"health,omitempty"`
	Env       map[string]string `yaml:"env,omitempty"`
	Secrets   map[string]string `yaml:"secrets,omitempty"`
	Addons    []string        `yaml:"addons,omitempty"`
	Domains   []string        `yaml:"domains,omitempty"`
}

type AppConfig struct {
	Name  string `yaml:"name"`
	Image string `yaml:"image"`
	Port  int    `yaml:"port,omitempty"`
}

type BuildConfig struct {
	Dockerfile string `yaml:"dockerfile,omitempty"`
	Context    string `yaml:"context,omitempty"`
}

type ResourcesConfig struct {
	CPU    string `yaml:"cpu,omitempty"`
	Memory string `yaml:"memory,omitempty"`
}

type ScalingConfig struct {
	Min       int `yaml:"min,omitempty"`
	Max       int `yaml:"max,omitempty"`
	TargetCPU int `yaml:"target_cpu,omitempty"`
}

type HealthConfig struct {
	Liveness  ProbeConfig `yaml:"liveness,omitempty"`
	Readiness ProbeConfig `yaml:"readiness,omitempty"`
}

type ProbeConfig struct {
	Path                string `yaml:"path,omitempty"`
	Port                int    `yaml:"port,omitempty"`
	InitialDelaySeconds int    `yaml:"initialDelaySeconds,omitempty"`
	PeriodSeconds       int    `yaml:"periodSeconds,omitempty"`
}

type ServiceConfig struct {
	Type         string `yaml:"type,omitempty"`
	ExternalPort int    `yaml:"externalPort,omitempty"`
}

// LoadConfig loads and parses the paas.yaml configuration file
func LoadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filename, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", filename, err)
	}

	// Set defaults
	if config.App.Port == 0 {
		config.App.Port = 3000
	}
	if config.Resources.CPU == "" {
		config.Resources.CPU = "100m"
	}
	if config.Resources.Memory == "" {
		config.Resources.Memory = "128Mi"
	}
	if config.Scaling.Min == 0 {
		config.Scaling.Min = 1
	}
	if config.Scaling.Max == 0 {
		config.Scaling.Max = 10
	}
	if config.Scaling.TargetCPU == 0 {
		config.Scaling.TargetCPU = 70
	}

	return &config, nil
}