package types

import (
	"fmt"
	"github.com/spf13/viper"
	"kubesphere.io/openpitrix-jobs/pkg/s3"
	"time"
)

const (
	// DefaultConfigurationName is the default name of configuration
	defaultConfigurationName = "kubesphere"

	// DefaultConfigurationPath the default location of the configuration file
	defaultConfigurationPath = "/etc/kubesphere"
)

// Config defines everything needed for apiserver to deal with external services
type Config struct {
	S3Options           *S3Options           `json:"s3,omitempty" yaml:"s3,omitempty" mapstructure:"s3"`
	OpenPitrixOptions   *OpenPitrixOptions   `json:"openpitrix,omitempty" yaml:"openpitrix,omitempty" mapstructure:"openpitrix"`
	MultiClusterOptions *MultiClusterOptions `json:"multicluster,omitempty" yaml:"multicluster,omitempty" mapstructure:"multicluster"`
	MySql               *MySqlOptions        `json:"mysql,omitempty" yaml:"mysql,omitempty"`
}

type MySqlOptions struct {
	Host     string `json:"host" yaml:"host"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

// Options contains configuration to access a s3 service
type S3Options struct {
	Endpoint        string `json:"endpoint,omitempty" yaml:"endpoint"`
	Region          string `json:"region,omitempty" yaml:"region"`
	DisableSSL      bool   `json:"disableSSL" yaml:"disableSSL"`
	ForcePathStyle  bool   `json:"forcePathStyle" yaml:"forcePathStyle"`
	AccessKeyID     string `json:"accessKeyID,omitempty" yaml:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey,omitempty" yaml:"secretAccessKey"`
	SessionToken    string `json:"sessionToken,omitempty" yaml:"sessionToken"`
	Bucket          string `json:"bucket,omitempty" yaml:"bucket"`
}

//// NewS3Options creates a default disabled Options(empty endpoint)
//func NewS3Options() *s3Options {
//	return &s3Options{
//		Endpoint:        "",
//		Region:          "us-east-1",
//		DisableSSL:      true,
//		ForcePathStyle:  true,
//		AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
//		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
//		SessionToken:    "",
//		Bucket:          "s2i-binaries",
//	}
//}

type MultiClusterOptions struct {
	// Enable
	Enable           bool `json:"enable"`
	EnableFederation bool `json:"enableFederation,omitempty"`

	// ProxyPublishService is the service name of multicluster component tower.
	//   If this field provided, apiserver going to use the ingress.ip of this service.
	// This field will be used when generating agent deployment yaml for joining clusters.
	ProxyPublishService string `json:"proxyPublishService,omitempty"`

	// ProxyPublishAddress is the public address of tower for all cluster agents.
	//   This field takes precedence over field ProxyPublishService.
	// If both field ProxyPublishService and ProxyPublishAddress are empty, apiserver will
	// return 404 Not Found for all cluster agent yaml requests.
	ProxyPublishAddress string `json:"proxyPublishAddress,omitempty"`

	// AgentImage is the image used when generating deployment for all cluster agents.
	AgentImage string `json:"agentImage,omitempty"`

	// ClusterControllerResyncSecond is the resync period used by cluster controller.
	ClusterControllerResyncSecond time.Duration `json:"clusterControllerResyncSecond,omitempty" yaml:"clusterControllerResyncSecond"`
}

//// NewOptions() returns a default nil options
//func NewOptions() *MultiClusterOptions {
//	return &MultiClusterOptions{
//		Enable:                        false,
//		EnableFederation:              false,
//		ProxyPublishAddress:           "",
//		ProxyPublishService:           "",
//		AgentImage:                    "kubesphere/tower:v1.0",
//		ClusterControllerResyncSecond: DefaultResyncPeriod,
//	}
//}

type OpenPitrixOptions struct {
	S3Options *s3.Options `json:"s3,omitempty" yaml:"s3,omitempty" mapstructure:"s3"`
}

// TryLoadFromDisk loads configuration from default location after server startup
// return nil error if configuration file not exists
func TryLoadFromDisk() (*Config, error) {
	viper.SetConfigName(defaultConfigurationName)
	viper.AddConfigPath(defaultConfigurationPath)

	// Load from current working directory, only used for debugging
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, err
		} else {
			return nil, fmt.Errorf("error parsing configuration file %s", err)
		}
	}

	conf := &Config{}

	if err := viper.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, nil
}
