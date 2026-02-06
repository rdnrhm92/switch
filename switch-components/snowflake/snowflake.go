package snowflake

import (
	"sync"

	"github.com/bwmarrin/snowflake"
)

// Generator 雪花算法生成器
type Generator struct {
	node *snowflake.Node
	mu   sync.Mutex
}

var (
	defaultGenerator *Generator
	once             sync.Once
)

// New create a instance
func NewGenerator(nodeID int64) (*Generator, error) {
	node, err := snowflake.NewNode(nodeID)
	if err != nil {
		return nil, err
	}

	return &Generator{
		node: node,
	}, nil
}

// GetDefault get default instance
func GetDefault() *Generator {
	once.Do(func() {
		node, err := snowflake.NewNode(1)
		if err != nil {
			panic(err)
		}
		defaultGenerator = &Generator{
			node: node,
		}
	})
	return defaultGenerator
}

// NextID 生成下一个ID
func (g *Generator) NextID() int64 {
	return g.node.Generate().Int64()
}

// NextIDString 生成下一个ID(字符串)
func (g *Generator) NextIDString() string {
	return g.node.Generate().String()
}
