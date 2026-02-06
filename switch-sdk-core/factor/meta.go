package factor

// SwitchFactor 系统内嵌的一些开关规则，在switch-sdk-go中给出了具体的实现，业务侧可以使用Register覆盖switch-sdk-go中的内容
type SwitchFactor struct {
	Factor      string `json:"factor"`
	Description string `json:"description"`
	JsonSchema  string `json:"json_schema"`
}

var Custom_Whole = &SwitchFactor{
	Factor:      "Custom_Whole",
	Description: "自定义map匹配,规则map中的k跟v必须在上下文中都找到",
	JsonSchema: `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "array",
		"items": {
			"type": "object",
			"properties": {
				"context_key": {
					"type": "string",
					"minLength": 1,
					"description": "用于从上下文中取值的键"
				},
				"context_val": {
					"type": "array",
					"items": {"type": "string", "minLength": 1},
					"minItems": 1,
					"description": "用于比较的值"
				}
			},
			"required": ["context_key", "context_val"],
			"additionalProperties": false
		},
		"minItems": 1,
		"description": "自定义全量匹配配置，所有条件都必须满足"
	}`,
}

var Custom_Arbitrarily = &SwitchFactor{
	Factor:      "Custom_Arbitrarily",
	Description: "自定义map匹配,规则map中的k跟v有任意一个可以上下文中找到",
	JsonSchema: `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "array",
		"items": {
			"type": "object",
			"properties": {
				"context_key": {
					"type": "string",
					"minLength": 1,
					"description": "用于从上下文中取值的键"
				},
				"context_val": {
					"type": "array",
					"items": {"type": "string", "minLength": 1},
					"minItems": 1,
					"description": "用于比较的值"
				}
			},
			"required": ["context_key", "context_val"],
			"additionalProperties": false
		},
		"minItems": 1,
		"description": "自定义任意匹配配置，满足任意一个条件即可"
	}`,
}

var UserNick = &SwitchFactor{
	Factor:      "UserNick",
	Description: "用户昵称精确匹配",
	JsonSchema: `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"context_key": {
				"type": "string",
				"minLength": 1,
				"description": "用于从上下文中取值的键"
			},
			"context_val": {
				"type": "array",
				"items": {"type": "string", "minLength": 1},
				"minItems": 1,
				"description": "用于比较的值"
			},
			"isBlack": {
				"type": "boolean",
				"description": "是否是黑名单"
			}
		},
		"required": ["context_key", "context_val", "isBlack"],
		"additionalProperties": false
	}`,
}

var TelNum = &SwitchFactor{
	Factor:      "TelNum",
	Description: "用户手机号精确匹配",
	JsonSchema: `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"context_key": {
				"type": "string",
				"minLength": 1,
				"description": "用于从上下文中取值的键"
			},
			"context_val": {
				"type": "array",
				"items": {
					"type": "string",
					"pattern": "^1[3-9]\\d{9}$",
					"description": "手机号格式"
				},
				"minItems": 1,
				"description": "手机号列表"
			}
		},
		"required": ["context_key", "context_val"],
		"additionalProperties": false
	}`,
}

var Single = &SwitchFactor{
	Factor:      "Single",
	Description: "单一开关",
	JsonSchema: `{
		"$schema": "http://json-schema.org/draft-07/schema#",
   		"type": "object",
   		"properties": {
      		"enabled": {
			"type": "boolean",
			"description": "开关是否启用",
         	"default": false
      		}
   		},
   		"required": ["enabled"],
   		"additionalProperties": false
	}`,
}

var TimeRange = &SwitchFactor{
	Factor:      "TimeRange",
	Description: "时间范围匹配",
	JsonSchema: `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"start_time": {
				"type": "string",
				"pattern": "^[0-9]+$",
				"description": "开始时间 (Unix时间戳秒)"
			},
			"end_time": {
				"type": "string",
				"pattern": "^[0-9]+$",
				"description": "结束时间 (Unix时间戳秒)"
			}
		},
		"required": ["start_time", "end_time"],
		"additionalProperties": false
	}`,
}

var UserId = &SwitchFactor{
	Factor:      "UserId",
	Description: "用户ID匹配",
	JsonSchema: `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"context_key": {
				"type": "string",
				"minLength": 1,
				"description": "用于从上下文中取值的键"
			},
			"context_val": {
				"type": "array",
				"items": {
					"type": "string",
					"pattern": "^[0-9]+$",
					"description": "用户ID，必须是数字"
				},
				"minItems": 1,
				"description": "用户ID列表"
			}
		},
		"required": ["context_key", "context_val"],
		"additionalProperties": false
	}`,
}

var Location = &SwitchFactor{
	Factor:      "Location",
	Description: "区域匹配",
	JsonSchema: `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"context_key": {
				"type": "string",
				"minLength": 1,
				"description": "用于从上下文中取值的键"
			},
			"context_val": {
				"type": "array",
				"items": {"type": "string", "minLength": 1},
				"minItems": 1,
				"description": "区域代码列表"
			}
		},
		"required": ["context_key", "context_val"],
		"additionalProperties": false
	}`,
}

var UserName = &SwitchFactor{
	Factor:      "UserName",
	Description: "用户名匹配",
	JsonSchema: `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"context_key": {
				"type": "string",
				"minLength": 1,
				"description": "用于从上下文中取值的键"
			},
			"context_val": {
				"type": "array",
				"items": {"type": "string", "minLength": 1},
				"minItems": 1,
				"description": "用户名列表"
			}
		},
		"required": ["context_key", "context_val"],
		"additionalProperties": false
	}`,
}

var IP = &SwitchFactor{
	Factor:      "IP",
	Description: "用户IP匹配",
	JsonSchema: `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"context_key": {
				"type": "string",
				"minLength": 1,
				"description": "用于从上下文中取值的键"
			},
			"context_val": {
				"type": "array",
				"items": {
					"type": "string",
					"anyOf": [
						{
							"format": "ipv4",
							"description": "IPv4地址"
						},
						{
							"format": "ipv6",
							"description": "IPv6地址"
						},
						{
							"pattern": "^([0-9]{1,3}\\.){3}[0-9]{1,3}/[0-9]{1,2}$",
							"description": "IPv4 CIDR格式"
						}
					]
				},
				"minItems": 1,
				"description": "IP地址列表"
			},
			"isBlack": {
				"type": "boolean",
				"description": "是否是黑名单"
			}
		},
		"required": ["context_key", "context_val", "isBlack"],
		"additionalProperties": false
	}`,
}
