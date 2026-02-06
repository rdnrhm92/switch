package drivers

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
)

// createDialer 创建kafka网络连接拨号器
func createDialer(security *SecurityConfig, timeout time.Duration) (*kafka.Dialer, error) {
	dialer := &kafka.Dialer{Timeout: timeout}
	if security == nil {
		return dialer, nil
	}

	//如果配置 & 开启了安全配置则加载
	if security.TLS != nil && security.TLS.Enabled {
		tlsConfig, err := createTLSConfig(security.TLS)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS configuration: %w", err)
		}
		dialer.TLS = tlsConfig
	}

	//账号密码
	if security.SASL != nil && security.SASL.Enabled {
		mechanism, err := createSASLMechanism(security.SASL)
		if err != nil {
			return nil, fmt.Errorf("failed to create SASL mechanism: %w", err)
		}
		dialer.SASLMechanism = mechanism
	}
	return dialer, nil
}

// createSASLMechanism 创建SASL认证机制
func createSASLMechanism(cfg *SASLConfig) (sasl.Mechanism, error) {
	if cfg == nil || !cfg.Enabled {
		return nil, nil
	}

	mechanism := strings.ToUpper(strings.TrimSpace(cfg.Mechanism))
	if mechanism == "" {
		mechanism = "SCRAM-SHA-512"
	}

	switch mechanism {
	case "PLAIN":
		return plain.Mechanism{
			Username: cfg.Username,
			Password: cfg.Password,
		}, nil

	case "SCRAM-SHA-256":
		return scram.Mechanism(scram.SHA256, cfg.Username, cfg.Password)

	case "SCRAM-SHA-512":
		return scram.Mechanism(scram.SHA512, cfg.Username, cfg.Password)

	default:
		return nil, fmt.Errorf("unsupported SASL mechanism: %s (supported: PLAIN, SCRAM-SHA-256, SCRAM-SHA-512)", cfg.Mechanism)
	}
}

// createTLSConfig 创建安全配置
func createTLSConfig(cfg *TLSConfig) (*tls.Config, error) {
	if cfg == nil || !cfg.Enabled {
		return nil, nil
	}
	tlsConfig := &tls.Config{InsecureSkipVerify: cfg.InsecureSkipVerify}
	if cfg.CaFile != "" {
		caCert, err := os.ReadFile(cfg.CaFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}
	if cfg.CertFile != "" && cfg.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client key pair: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	return tlsConfig, nil
}
