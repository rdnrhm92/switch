package resp

import (
	"fmt"
	"regexp"
	"strings"
)

type Message string

// 消息根据{}做匹配
var pattern = `\{.*?\}`
var re = regexp.MustCompile(pattern)

// MessageBuilder 消息构建器
type MessageBuilder struct {
	template string
	params   []interface{}
	prefix   string
	suffix   string
}

// MessageTemplate 消息模板
type MessageTemplate struct {
	Template string `json:"template"`
	Category string `json:"category"`
}

// NewMessageBuilder 创建消息构建器
func NewMessageBuilder() *MessageBuilder {
	return &MessageBuilder{
		params: make([]interface{}, 0),
	}
}

// WithTemplate WithTemplate
func (mb *MessageBuilder) WithTemplate(template string) *MessageBuilder {
	mb.template = template
	return mb
}

// WithParam WithParam
func (mb *MessageBuilder) WithParam(value ...interface{}) *MessageBuilder {
	mb.params = append(mb.params, value...)
	return mb
}

// WithPrefix 前缀
func (mb *MessageBuilder) WithPrefix(prefix string) *MessageBuilder {
	mb.prefix = prefix
	return mb
}

// WithSuffix 后缀
func (mb *MessageBuilder) WithSuffix(suffix string) *MessageBuilder {
	mb.suffix = suffix
	return mb
}

// Build 构建消息
func (mb *MessageBuilder) Build() Message {
	message := mb.template

	placeholders := re.FindAllString(message, -1)

	numToReplace := len(placeholders)
	if len(mb.params) < numToReplace {
		numToReplace = len(mb.params)
	}

	var replacements []string
	for i := 0; i < numToReplace; i++ {
		replacements = append(replacements, placeholders[i])
		replacements = append(replacements, fmt.Sprint(mb.params[i]))
	}

	replacer := strings.NewReplacer(replacements...)
	message = replacer.Replace(message)

	if mb.prefix != "" {
		message = mb.prefix + message
	}
	if mb.suffix != "" {
		message = message + mb.suffix
	}

	return Message(message)
}

// BuildMessage
func BuildMessage(template string, params ...interface{}) Message {
	return NewMessageBuilder().WithTemplate(template).WithParam(params...).Build()
}

// BuildErrorMessage err message
func BuildErrorMessage(template string, err error) Message {
	return BuildMessage(template, err.Error())
}
