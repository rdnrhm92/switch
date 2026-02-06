import React, {forwardRef, useEffect, useImperativeHandle, useRef, useState} from 'react';
import {Button, Card, Drawer, Form, Input, List, Select, Tabs} from 'antd';
import {PlusOutlined} from '@ant-design/icons';
import CodeMirror from '@uiw/react-codemirror';
import { json, jsonParseLinter } from '@codemirror/lang-json';
import { linter, lintGutter } from '@codemirror/lint';
import { yaml as yamlLang } from '@codemirror/lang-yaml';
import YAML from 'yaml';
import jsYaml from 'js-yaml';
import {FormattedMessage, useIntl} from "@@/exports";

// 编辑器组件
const ConfigEditor: React.FC<{
  value?: Record<string, any>;
  onChange?: (value: Record<string, any>) => void;
  submitTrigger?: number;
}> = ({ value, onChange, submitTrigger }) => {
  const intl = useIntl();
  const [usage, setUsage] = useState<'json' | 'yaml'>('json');
  const [text, setText] = useState<string>('{}');
  const [errorMessage, setErrorMessage] = useState('');
  const initialTrigger = useRef(submitTrigger);

  useEffect(() => {
    console.log('ConfigEditor useEffect triggered, value:', value);
    if (value) {
      try {
        const newText = usage === 'json' ? JSON.stringify(value, null, 2) : jsYaml.dump(value);
        console.log(' new text', newText)
        if (newText !== text) {
          setText(newText);
        }
        setErrorMessage('');
      } catch (e) {
        setErrorMessage(intl.formatMessage({ id: 'pages.env.driver.editor.warning' }));
        setText(usage === 'json' ? '{}' : '');
      }
    } else {
      setText(usage === 'json' ? '{}' : '');
    }
  }, [value, usage]);

  const handleTabChange = (newUsage: 'json' | 'yaml') => {
    if (usage === newUsage) return;

    setErrorMessage('');
    try {
      let parsedObject: any;
      if (usage === 'json') {
        parsedObject = text.trim() ? JSON.parse(text) : {};
      } else {
        parsedObject = jsYaml.load(text);
        if (parsedObject === null || typeof parsedObject !== 'object') parsedObject = {};
      }

      const newText = newUsage === 'json' ? JSON.stringify(parsedObject, null, 2) : jsYaml.dump(parsedObject);
      setText(newText);
      setUsage(newUsage);
    } catch (e: any) {
      setErrorMessage(`${intl.formatMessage({ id: 'pages.env.driver.editor.convertError' })}: ${e.message}`);
    }
  };

  const handleTextChange = (newText: string) => {
    setText(newText);
    setErrorMessage('');
  };

  useEffect(() => {
    if (submitTrigger === undefined || submitTrigger === initialTrigger.current) return;
    initialTrigger.current = submitTrigger;

    console.log('submitTrigger触发，当前text内容:', text);
    try {
      let parsedValue;
      if (usage === 'json') {
        parsedValue = text.trim() ? JSON.parse(text) : {};
      } else {
        parsedValue = jsYaml.load(text);
      }
      console.log('提交的时候解析后的parsedValue:', parsedValue)
      onChange?.(parsedValue);
      setErrorMessage('');
    } catch (e: any) {
      setErrorMessage(`${intl.formatMessage({ id: 'pages.env.driver.editor.formatError' })}: ${e.message}`);
    }
  }, [submitTrigger]);

  return (
    <>
      <Tabs
        activeKey={usage}
        onChange={(key) => handleTabChange(key as 'json' | 'yaml')}
        size="small"
        type="card"
        style={{ marginBottom: '8px' }}
        items={[{ label: 'JSON', key: 'json' }, { label: 'YAML', key: 'yaml' }]}
      />
      <CodeMirror
        key={usage}
        value={text}
        height="300px"
        extensions={usage === 'json'
          ? [json(), linter(jsonParseLinter(), { delay: 1000 }), lintGutter()]
          : [
            yamlLang(),
            linter(view => {
              const diagnostics: any[] = [];
              const docString = view.state.doc.toString();
              if (!docString.trim()) return diagnostics;
              try {
                YAML.parse(docString);
              } catch (e: any) {
                if (e.name === 'YAMLParseError' && e.linePos) {
                  const errorPos = e.linePos[0];
                  if (errorPos.line > view.state.doc.lines) return diagnostics;
                  const line = view.state.doc.line(errorPos.line);
                  const from = line.from + errorPos.col - 1;
                  const diagnosticFrom = Math.min(from, line.to);
                  diagnostics.push({
                    from: diagnosticFrom,
                    to: line.to,
                    severity: 'error',
                    message: e.message,
                  });
                }
              }
              return diagnostics;
            }, { delay: 1000 }),
            lintGutter(),
          ]}
        onChange={handleTextChange}
        theme="dark"
      />
      {errorMessage && <div style={{ color: "red", marginTop: '8px' }}>{errorMessage}</div>}
    </>
  );
};

// 配置示例展示组件
const ConfigExample: React.FC<{
  driverType?: string;
  usage?: string;
}> = ({ driverType, usage }) => {
  const [exampleUsage, setExampleUsage] = useState<'json' | 'yaml'>('json');


  const kafkaProducerDemo = (exampleUsage: string)=>{
    // Kafka 生产者配置
    if (exampleUsage === 'json') {
      return `{
  // broker列表(必填)
  "brokers": ["localhost:9092", "localhost:9093"],
  // kafka主题(必填) 必须提前创建不支持自动创建 驱动启动时会验证主题有效性
  "topic": "example-topic",

  // 确认机制(非必填) all-所有副本确认 one-leader确认 none-不需要确认
  "requiredAcks": "all",
  // 等待broker返回确认的最大超时时间(非必填)
  "timeout": "30s",
  // 当消息发送失败时(网络问题 broker不可用等),客户端会自动重试retries次(非必填)
  "retries": 3,
  // 写失败后最小重试间隔(非必填)
  "retryBackoffMin": "100ms",
  // 写失败后最大重试间隔(非必填)
  "retryBackoffMax": "1s",
  // 消息压缩算法(非必填) gzip snappy lz4 zstd
  "compression": "snappy",

  // broker连接拨号器的连接超时时间(非必填)
  "connectTimeout": "10s",
  // broker测试连通性时的超时时间(非必填)
  "validateTimeout": "10s",

  // 时间超过batchTimeout触发写(非必填)
  "batchTimeout": "1s",
  // 消息大小满足batchBytes触发写(非必填)
  "batchBytes": 1048576,
  // 消息数量满足batchSize触发写(非必填)
  "batchSize": 50,

  // 安全配置(非必填)
  "security": {
    // sasl验证(非必填)
    "sasl": {
      // 是否开启
      "enabled": false,
      // 认证机制 PLAIN SCRAM-SHA-256 SCRAM-SHA-512等
      "mechanism": "PLAIN",
      // 用户名
      "username": "your-username",
      // 密码
      "password": "your-password"
    },
    // tls验证(非必填)
    "tls": {
      // 是否开启
      "enabled": false,
      // CA证书地址 用于验证kafka服务端的有效性(系统路径)
      "caFile": "/path/to/ca.pem",
      // 客户端证书地址(系统路径)
      "certFile": "/path/to/cert.pem",
      // 客户端密钥地址(系统路径)
      "keyFile": "/path/to/key.pem",
      // 是否跳过证书验证
      "insecureSkipVerify": false
    }
  }
}\n\n`
    } else {
      return `
# broker列表(必填)
brokers:
  - localhost:9092
  - localhost:9093
# kafka主题(必填) 必须提前创建不支持自动创建 驱动启动时会验证主题有效性
topic: "example-topic"

# 确认机制(非必填) all-所有副本确认 one-leader确认 none-不需要确认
requiredAcks: "all"
# 等待broker返回确认的最大超时时间(非必填)
timeout: "30s"
# 当消息发送失败时(网络问题 broker不可用等),客户端会自动重试retries次(非必填)
retries: 3
# 写失败后最小重试间隔(非必填)
retryBackoffMin: "100ms"
# 写失败后最大重试间隔(非必填)
retryBackoffMax: "1s"
# 消息压缩算法(非必填) gzip snappy lz4 zstd
compression: "snappy"

# broker连接拨号器的连接超时时间(非必填)
connectTimeout: "10s"
# broker测试连通性时的超时时间(非必填)
validateTimeout: "10s"

# 时间超过batchTimeout触发写(非必填)
batchTimeout: "1s"
# 消息大小满足batchBytes触发写(非必填)
batchBytes: 1048576
# 消息数量满足batchSize触发写(非必填)
batchSize: 50

# 安全配置(非必填)
security:
  # sasl验证(非必填)
  sasl:
    # 是否开启
    enabled: false
    # 认证机制 PLAIN SCRAM-SHA-256 SCRAM-SHA-512等
    mechanism: "PLAIN"
    # 用户名
    username: "your-username"
    # 密码
    password: "your-password"
  # tls验证(非必填)
  tls:
    # 是否开启
    enabled: false
    # CA证书地址 用于验证kafka服务端的有效性(系统路径)
    caFile: "/path/to/ca.pem"
    # 客户端证书地址(系统路径)
    certFile: "/path/to/cert.pem"
    # 客户端密钥地址(系统路径)
    keyFile: "/path/to/key.pem"
    # 是否跳过证书验证
    insecureSkipVerify: false
    \n\n`
    }
  }

  const kafkaConsumerDemo = (exampleUsage: string)=>{
    // Kafka 消费者配置
    if (exampleUsage === 'json') {
      return `{
  // broker列表(必填)
  "brokers": ["localhost:9092", "localhost:9093"],
  // kafka主题(必填) 必须提前创建不支持自动创建 驱动启动时会验证主题有效性
  "topic": "example-topic",

  // 消费者组ID(非必填)(没有特殊需求此处请设置为空,交由框架去生成消费者组ID)
  "groupId": "",
  // 偏移量重置策略(非必填) earliest(从最早消息开始消费) latest(从最新消息开始消费)
  "autoOffsetReset": "latest",
  // 自动提交(非必填)
  "enableAutoCommit": true,
  // 手动提交间隔(非必填)(可以适当设置大一点,开关消息重复消费并无影响)
  "autoCommitInterval": "10s",

  // 连接超时配置(非必填)
  "connectTimeout": "10s",
  // 验证连接的连通性超时配置(非必填)
  "validateTimeout": "10s",

  // 读消息超时配置(非必填)
  "readTimeout": "10s",
  // 提交偏移量超时配置(非必填)
  "commitTimeout": "10s",

  // 安全配置(非必填)
  "security": {
    // sasl验证(非必填)
    "sasl": {
      // 是否开启
      "enabled": false,
      // 认证机制 PLAIN SCRAM-SHA-256 SCRAM-SHA-512等
      "mechanism": "PLAIN",
      // 用户名
      "username": "your-username",
      // 密码
      "password": "your-password"
    },
    // tls验证(非必填)
    "tls": {
      // 是否开启
      "enabled": false,
      // CA证书地址 用于验证kafka服务端的有效性(系统路径)
      "caFile": "/path/to/ca.pem",
      // 客户端证书地址(系统路径)
      "certFile": "/path/to/cert.pem",
      // 客户端密钥地址(系统路径)
      "keyFile": "/path/to/key.pem",
      // 是否跳过证书验证
      "insecureSkipVerify": false
    }
  },
  // 重试配置(非必填)
  "retry": {
    // 失败超出count次数将重启 成功则重置
    "count": 1,
    // 重启间隔时间
    "backoff": "3s"
  }
}\n\n`
    } else {
      return `
# broker列表(必填)
brokers:
  - localhost:9092
  - localhost:9093
# kafka主题(必填) 必须提前创建不支持自动创建 驱动启动时会验证主题有效性
topic: "example-topic"

# 消费者组ID(非必填)(没有特殊需求此处请设置为空,交由框架去生成消费者组ID)
groupId: ""
# 偏移量重置策略(非必填) earliest(从最早消息开始消费) latest(从最新消息开始消费)
autoOffsetReset: "latest"
# 自动提交(非必填)
enableAutoCommit: true
# 手动提交间隔(非必填)(可以适当设置大一点,开关消息重复消费并无影响)
autoCommitInterval: "1s"

# 连接超时配置(非必填)
connectTimeout: "10s"
# 验证连接的连通性超时配置(非必填)
validateTimeout: "10s"

# 读消息超时配置(非必填)
readTimeout: "10s"
# 提交偏移量超时配置(非必填)
commitTimeout: "10s"

# 安全配置(非必填)
security:
  # sasl验证(非必填)
  sasl:
    # 是否开启
    enabled: false
    # 认证机制 PLAIN SCRAM-SHA-256 SCRAM-SHA-512等
    mechanism: "PLAIN"
    # 用户名
    username: "your-username"
    # 密码
    password: "your-password"
  # tls验证(非必填)
  tls:
    # 是否开启
    enabled: false
    # CA证书地址 用于验证kafka服务端的有效性(系统路径)
    caFile: "/path/to/ca.pem"
    # 客户端证书地址(系统路径)
    certFile: "/path/to/cert.pem"
    # 客户端密钥地址(系统路径)
    keyFile: "/path/to/key.pem"
    # 是否跳过证书验证
    insecureSkipVerify: false
# 重试配置(非必填)
retry:
  # 失败超出count次数将重启 成功则重置
  count: 1
  # 重启间隔时间
  backoff: "3s"
    \n\n`
    }
  }

  const webhookProducerDemo = (exampleUsage: string)=>{
    // webhook 生产者配置
    if (exampleUsage === 'json') {
      return `{
  // 客户端端口(非必填)  默认20002
  "port": "20002",

  // 是否忽略错误(非必填) 当设置为true时 即使有错误 本次通知也将意味是成功
  "ignoreExceptions": false,
  // 服务端发送请求的最大超时时间(非必填)
  "timeOut": "30s",

  // 重试配置(非必填)
  "retry": {
    // 失败超出count次数将重启 成功则重置
    "count": 3,
    // 重启间隔时间
    "backoff": "1s"
  },

  // 黑名单配置(非必填)(只能配置IP 不能包含协议头或者端口以及路径)
  "blacklistIPs": [
    "192.168.1.100",
    "10.0.0.50"
  ],

  // 安全配置(非必填) 客户端跟服务端通信时将使用此配置做安全验证
  "security": {
    // HMAC-SHA256签名密钥
    "secret": "your-webhook-secret"
  }
}\n\n`
    } else {
      return `
# 客户端端口(非必填)  默认20002
port: "20002"

# 是否忽略错误(非必填) 当设置为true时 即使有错误 本次通知也将意味是成功
ignoreExceptions: false
# 服务端发送请求的最大超时时间(非必填)
timeOut: "30s"

# 重试配置(非必填)
retry:
  # 失败超出count次数将重启 成功则重置
  count: 3
  # 重启间隔时间
  backoff: "1s"

# 黑名单配置(非必填)(只能配置IP 不能包含协议头或者端口以及路径)
blacklistIPs:
  - "192.168.1.100"
  - "10.0.0.50"

# 安全配置(非必填) 客户端跟服务端通信时将使用此配置做安全验证
security:
  # HMAC-SHA256签名密钥
  secret: "your-webhook-secret"
  \n\n`
    }
  }

  const webhookConsumerDemo = (exampleUsage: string)=> {
      // Webhook 消费者配置
      if (exampleUsage === 'json') {
        return `{
  // 客户端端口(非必填但请跟webhook生产者配置保持一致 默认20002)
  "port": "20002",

  // 黑名单配置(非必填)(只能配置IP 不能包含协议头或者端口以及路径)
  "blacklistIPs": [
    "192.168.1.100",
    "10.0.0.50"
  ],

  // 重试配置(非必填)
  "retry": {
    // 失败超出count次数将重启 成功则重置
    "count": 3,
    // 重启间隔时间
    "backoff": "1s"
  },

  // 安全配置(非必填) 客户端跟服务端通信时将使用此配置做安全验证
  "security": {
    // HMAC-SHA256签名密钥
    "secret": "your-webhook-secret"
  }
}\n\n`
      } else {
        return `
# 客户端端口(非必填但请跟webhook生产者配置保持一致 默认20002)
port: "20002"

# 黑名单配置(非必填)(必须配置IP不能包含协议头或者端口以及路径)
blacklistIPs:
  - "192.168.1.100"
  - "10.0.0.50"

# 重试配置(非必填)
retry:
  # 失败超出count次数将重启 成功则重置
  count: 3
  # 重启间隔时间
  backoff: "1s"

# 安全配置(非必填) 客户端跟服务端通信时将使用此配置做安全验证
security:
  # HMAC-SHA256签名密钥
  secret: "your-webhook-secret"
  \n\n`
      }
  }

  const pollingProducerDemo = (exampleUsage: string)=>{
    // Polling 生产者配置
    if (exampleUsage === 'json') {
      return `{
  // 服务端端口(非必填) 默认10002
  "port": "10002",

  // 单次请求(非必填) 服务端Hold时间
  "long_poll_timeout": "60s",

  // 服务端读超时时间(非必填)
  "server_read_timeout": "90s",
  // 服务端写超时时间(非必填)
  "server_write_timeout": "70s",
  // 服务端最大空闲时间(非必填)
  "server_idle_timeout": "120s",

  // 安全配置(非必填)
  "security": {
    // 服务端支持的token列表
    "valid_tokens": ["token1", "token2", "token3"],
    // 证书地址(系统路径)
    "cert_file": "/path/to/server.crt",
    // 密钥地址(系统路径)
    "key_file": "/path/to/server.key"
  },

  // 重试配置(非必填)
  "retry": {
    // 失败超出count次数将重启 成功则重置
    "count": 1,
    // 重启间隔时间
    "backoff": "3s"
  }
}
\n\n`
    } else {
      return `
# 服务端端口(非必填) 默认10002
port: "10002"

# 单次请求(非必填) 服务端Hold时间
long_poll_timeout: "60s"

# 服务端读超时时间(非必填)
server_read_timeout: "90s"
# 服务端写超时时间(非必填)
server_write_timeout: "70s"
# 服务端最大空闲时间(非必填)
server_idle_timeout: "120s"

# 安全配置(非必填)
security:
  # 服务端支持的token列表
  valid_tokens:
    - "token1"
    - "token2"
    - "token3"
  # 证书地址(系统路径)
  cert_file: "/path/to/server.crt"
  # 密钥地址(系统路径)
  key_file: "/path/to/server.key"

# 重试配置(非必填)
retry:
  # 失败超出count次数将重启 成功则重置
  count: 1
  # 重启间隔时间
  backoff: "3s"
  \n\n`
    }
  }

  const pollingConsumerDemo = (exampleUsage: string)=> {
    // polling 消费者配置
    if (exampleUsage === 'json') {
      return `{
  // 服务端地址(必填) 仅支持http,https协议头,支持ip/域名+port,不支持自定义路径,支持空路径或根路径 注意端口要跟生产者保持一致
  "url": "http://localhost:10002",

  // 正常情况下的长轮询间隔(非必填) 防止消息缺漏 尽量配小
  "poll_interval": "30s",
  // 单次请求超时时间(非必填) 避免因为网络问题客户端占用资源,建议配置比服务端的long_poll_timeout大
  "request_timeout": "30s",

  // HTTP自定义请求头配置(非必填)
  "headers": {
    "customer_header": "switch-polling-client"
  },
  // 用户代理(非必填)
  "user_agent": "switch-polling-client/1.0",

  // 安全配置(非必填)
  "security": {
    // token
    "token": "your-bearer-token",
    // 是否开启https
    "insecure_skip_verify": false
  },

  // 是否忽略消息处理异常(非必填) true时处理失败不中断轮询
  "ignore_exceptions": false,

  // 重试配置(非必填)
  "retry": {
  // 失败超出count次数将重启 成功则重置
    "count": 1,
  // 重启间隔时间
  "backoff": "3s"
  }
}\n\n`} else {
      return `
# 服务端地址(必填) 仅支持http,https协议头,支持ip/域名+port,不支持自定义路径,支持空路径或根路径 注意端口要跟生产者保持一致
url: "http://localhost:10002"

# 正常情况下的长轮询间隔(非必填) 防止消息缺漏 尽量配小
poll_interval: "30s"
# 单次请求超时时间(非必填) 避免因为网络问题客户端占用资源,建议配置比服务端的long_poll_timeout大
request_timeout: "30s"

# HTTP自定义请求头配置(非必填)
headers:
  customer_header: "switch-polling-client"

# 用户代理(非必填)
user_agent: "switch-polling-client/1.0"

# 安全配置(非必填)
security:
  # token
  bearer_token: "your-bearer-token"
  # 是否开启https
  insecure_skip_verify: false

# 是否忽略消息处理异常(非必填) true时处理失败不中断轮询
ignore_exceptions: false

# 重试配置(非必填)
retry:
  # 失败超出count次数将重启 成功则重置
  count: 1
  # 重启间隔时间
  backoff: "3s"
\n\n`
    }
  }


  const getExampleConfigText = () => {
  // 获取配置demo
    if (driverType === 'kafka' && usage === 'consumer') {
      return kafkaConsumerDemo(exampleUsage)
    }

    if (driverType === 'kafka' && usage === 'producer') {
      return kafkaProducerDemo(exampleUsage)
    }

    if (driverType === 'webhook' && usage === 'consumer') {
      return webhookConsumerDemo(exampleUsage)
    }

    if (driverType === 'webhook' && usage === 'producer') {
      return webhookProducerDemo(exampleUsage)
    }

    if (driverType === 'polling' && usage === 'consumer') {
      return pollingConsumerDemo(exampleUsage)
    }

    if (driverType === 'polling' && usage === 'producer') {
      return pollingProducerDemo(exampleUsage)
    }
    return exampleUsage === 'json' ? '{}' : '';
  };

  const exampleText = getExampleConfigText();

  return (
    <div style={{ marginTop: '16px' }}>
      <div style={{ marginBottom: '8px', fontWeight: 'bold' }}>配置示例：</div>
      <Tabs
        activeKey={exampleUsage}
        onChange={(key) => setExampleUsage(key as 'json' | 'yaml')}
        size="small"
        type="card"
        style={{ marginBottom: '8px' }}
        items={[{ label: 'JSON', key: 'json' }, { label: 'YAML', key: 'yaml' }]}
      />
      <CodeMirror
        key={exampleUsage}
        value={exampleText}
        height="350px"
        extensions={exampleUsage === 'json' ? [json()] : [yamlLang()]}
        editable={false}
        theme="dark"
      />
    </div>
  );
};



//定义一个切换tab的接口
export interface TabExchangeHandler {
  setActiveTab: (key: 'producer' | 'consumer') => void;
  isCreate: (create: boolean) => void;
  //获取当前在哪个tab页
  getActiveTab: () => {};
}


// 主配置组件
const DriverConfiguration = forwardRef<TabExchangeHandler, {
  value?: API.Driver[];
  onChange?: (value: API.Driver[]) => void;
  onTabChange?: () => void;
  onSaveClick?: () => void;
}>(({value = [], onChange, onTabChange, onSaveClick}, ref) => {
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [isCreate, setIsCreate] = useState(true);
  const [editingDriver, setEditingDriver] = useState<Partial<API.Driver> | null>(null);
  const [form] = Form.useForm();
  const intl = useIntl();
  const [activeTabKey, setActiveTabKey] = useState<'producer' | 'consumer'>('producer');
  const tempIdCounter = useRef(0);
  const [submitTrigger, setSubmitTrigger] = useState(0);

  const getNewTempId = () => {
    tempIdCounter.current -= 1;
    return tempIdCounter.current;
  };

  const watchedUsage = Form.useWatch('usage', form);
  const watchedDriverType = Form.useWatch('driverType', form);

  useEffect(() => {
    if (drawerVisible && editingDriver) {
      form.setFieldsValue(editingDriver);
    }
  }, [editingDriver, drawerVisible]);

  useImperativeHandle(ref, () => ({
    setActiveTab: (key: 'producer' | 'consumer') => {
      setActiveTabKey(key);
    },
    isCreate: (create: boolean) => {
      setIsCreate(create);
    },
    getActiveTab: () => {
      return activeTabKey;
    }
  }));

  const handleAdd = () => {
    // 默认使用当前激活的 tab 作为驱动的用途
    setEditingDriver({ id: getNewTempId(), usage: activeTabKey, driverConfig: {} });
    setIsCreate(true);
    setDrawerVisible(true);
  };

  const handleEdit = (driver: API.Driver) => {
    console.log('编辑的driver', driver)
    setEditingDriver(driver);
    setDrawerVisible(true);
  };

  const handleDelete = (id: number) => {
    const newDrivers = value.filter(d => d.id !== id);
    onChange?.(newDrivers);
  };

  const onDrawerClose = () => {
    setDrawerVisible(false);
    setEditingDriver(null);
    form.resetFields();
  };

  //保存
  //这部分提交的逻辑下发到二级组件中了
  const handleSave = () => {
    setSubmitTrigger(prev => prev + 1);
    //保存的时候移出错误
    onSaveClick?.();
  };

  const onFormFinish = (formValues: any) => {
    console.log('原始的drivers', [...value])
    console.log('本次提交的表单数据', {...formValues})
    console.log('本次修改的表单数据', {...editingDriver})

    //编辑或者新增后的内容
    const newDriver: API.Driver = {
      ...editingDriver,
      ...formValues
    } as API.Driver;

    //新值替换旧值(id匹配)
    let newDrivers;
    if (value.some(d => d.id === newDriver.id)) {
      newDrivers = value.map(d => d.id === newDriver.id ? newDriver : d);
    } else {
      newDrivers = [...value, newDriver];
    }
    console.log('最终的drivers', newDrivers)
    //依赖外界的表单提交，此处只是把数据向外界暴露
    onChange?.(newDrivers);
    onDrawerClose();
  };

  const renderDriverList = (usage: 'producer' | 'consumer') => {
    return (
      <List
        grid={{ gutter: 16, column: 2 }}
        dataSource={value.filter(d => d.usage === usage)}
        renderItem={driver => (
          <List.Item>
            <Card title={driver.name} size="small">
              <p>类型: {driver.driverType}</p>
              <Button size="small" onClick={() => handleEdit(driver)}>
                <FormattedMessage id="pages.env.searchTable.driver.update" />
              </Button>
              <Button size="small" danger onClick={() => handleDelete(driver.id)} style={{ marginLeft: 8 }}>
                <FormattedMessage id="pages.env.searchTable.driver.delete" />
              </Button>
            </Card>
          </List.Item>
        )}
      />
    );
  };

  return (
    <div>
      <Tabs
        style={{ margin: '0 -45px' }}
        activeKey={activeTabKey}
        onChange={(key) => {
          setActiveTabKey(key as 'producer' | 'consumer');
          onTabChange?.();
        }}
        tabBarExtraContent={
          <Button type="primary" onClick={handleAdd} icon={<PlusOutlined />}>
            <FormattedMessage id="pages.env.searchTable.driver.add"/>
          </Button>
        }
        items={[
          {
            label: intl.formatMessage({ id: 'pages.env.driver.producer' }),
            key: 'producer',
            children: renderDriverList('producer'),
          },
          {
            label: intl.formatMessage({ id: 'pages.env.driver.consumer' }),
            key: 'consumer',
            children: renderDriverList('consumer'),
          },
        ]}
      />

      <Drawer
        title={isCreate ? intl.formatMessage({ id: 'pages.env.searchTable.driver.createDriver' }) : intl.formatMessage({ id: 'pages.env.searchTable.driver.updateDriver' })}
        width={720}
        onClose={onDrawerClose}
        open={drawerVisible}
        styles={{ body: { paddingBottom: 80 } }}
        footer={
          <div style={{ textAlign: 'right' }}>
            <Button onClick={onDrawerClose} style={{ marginRight: 8 }}>
              <FormattedMessage id="pages.system.cancel" />
            </Button>
            <Button onClick={handleSave} type="primary">
              <FormattedMessage id="pages.system.save" />
            </Button>
          </div>
        }
      >
        <Form form={form} layout="vertical" onFinish={onFormFinish}>
          <Form.Item name="name" label={intl.formatMessage({ id: 'pages.env.driver.driverName' })} rules={[{ required: true, message: intl.formatMessage({ id: 'pages.env.driver.driverNameRequired' }) }]}>
            <Input placeholder={intl.formatMessage({ id: 'pages.env.driver.driverNamePlaceholder' })} />
          </Form.Item>
          <Form.Item name="usage" label={intl.formatMessage({ id: 'pages.env.driver.driverRole' })} rules={[{ required: true, message: intl.formatMessage({ id: 'pages.env.driver.driverRoleRequired' }) }]}>
            <Select placeholder={intl.formatMessage({ id: 'pages.env.driver.driverRolePlaceholder' })} disabled>
              <Select.Option value="producer">{intl.formatMessage({ id: 'pages.env.driver.producer' })}</Select.Option>
              <Select.Option value="consumer">{intl.formatMessage({ id: 'pages.env.driver.consumer' })}</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="driverType" label={intl.formatMessage({ id: 'pages.env.driver.driverType' })} rules={[{ required: true, message: intl.formatMessage({ id: 'pages.env.driver.driverTypeRequired' }) }]}>
            <Select placeholder={intl.formatMessage({ id: 'pages.env.driver.driverTypePlaceholder' })}>
              <Select.Option value="kafka">Kafka</Select.Option>
              <Select.Option value="webhook">Webhook</Select.Option>
              <Select.Option value="polling">Polling</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="driverConfig" label={intl.formatMessage({ id: 'pages.env.driver.driverConfig' })}>
            <ConfigEditor
              value={form.getFieldValue('driverConfig')}
              submitTrigger={submitTrigger}
              onChange={(newValue) => {
                console.log('ConfigEditor onChange:', newValue)
                // 不设置到表单，避免Form处理导致的问题
                // form.setFieldsValue({ driverConfig: newValue });
                setTimeout(() => {
                  // 手动构造表单数据并提交
                  const currentFormValues = form.getFieldsValue();
                  const finalFormValues = {
                    ...currentFormValues,
                    driverConfig: newValue
                  };
                  console.log('手动构造的表单数据:', finalFormValues);
                  // 直接调用onFormFinish，绕过Form的处理
                  onFormFinish(finalFormValues);
                }, 0);
              }}
            />
          </Form.Item>
          <ConfigExample
            driverType={watchedDriverType}
            usage={watchedUsage}
          />
        </Form>
      </Drawer>
    </div>
  );
});

export default DriverConfiguration;
