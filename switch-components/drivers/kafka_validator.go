package drivers

import (
	"context"
	"fmt"
	"time"
)

//kafka的通用校验
//1.kafka的broker列表是否可以连通
//2.kafka是否可以连接(ACL认证等)以及校验kafka集群的状态
//3.topic是否存在以及topic的健康状态,分区是否正常等

// checkTopicExists 检查Topic是否存在的通用逻辑
func checkTopicExists(brokers []string, topic string, security *SecurityConfig, connectTimeout, validateTimeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), validateTimeout)
	defer cancel()

	//Kafka协议层面的检查（带认证配置）用于判断集群状态以及认证信息是否正确
	//不区分scram ssl plain三种认证方式的差异性统一进行所有节点全部认证
	dialer, err := createDialer(security, connectTimeout)
	if err != nil {
		return fmt.Errorf("failed to create dialer with auth config: %w", err)
	}

	for _, broker := range brokers {

		//TCP连接测试 在校验acl的过程中连接不到集群还是失败的
		conn, err := dialer.DialContext(ctx, "tcp", broker)
		if err != nil {
			return fmt.Errorf("failed to connect to broker %s: %w", broker, err)
		}

		// 获取指定Topic的分区信息,没有topic则获取集群所有分区信息
		partitions, err := conn.ReadPartitions(topic)
		conn.Close()

		if err != nil {
			return fmt.Errorf("failed to read partitions for topic %s on broker %s: %w", topic, broker, err)
		}

		if len(partitions) == 0 {
			return fmt.Errorf("topic %s exists but has no partitions", topic)
		}

		// 检查分区健康状态
		// 判断leader节点的可用性
		for _, partition := range partitions {
			if partition.Leader.ID < 0 {
				return fmt.Errorf("topic %s partition %d has no leader", topic, partition.ID)
			}
		}

	}

	return nil
}

// checkProducerTopicExists 检查Producer的Topic是否存在
func checkProducerTopicExists(config *KafkaProducerConfig) error {
	connectTimeout := config.getConnectTimeout()
	validateTimeout := config.getValidateTimeout()
	return checkTopicExists(config.Brokers, config.Topic, config.Security, connectTimeout, validateTimeout)
}

// checkConsumerTopicExists 检查Consumer的Topic是否存在
func checkConsumerTopicExists(config *KafkaConsumerConfig) error {
	connectTimeout := config.getConnectTimeout()
	validateTimeout := config.getValidateTimeout()
	return checkTopicExists(config.Brokers, config.Topic, config.Security, connectTimeout, validateTimeout)
}
