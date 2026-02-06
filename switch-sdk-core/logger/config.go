package logger

// Config 日志配置
type LoggerConfig struct {
	// 日志级别 ("debug", "info", "warn", "error", "dpanic", "panic", "fatal")
	Level string `json:"level" yaml:"level"`
	// 日志输出目录
	OutputDir string `json:"output_dir" yaml:"output_dir"`
	// 日志文件名格式
	FileNameFormat string `json:"file_name_format" yaml:"file_name_format"`
	// 单个日志文件的最大大小（MB）
	MaxSize int `json:"max_size" yaml:"max_size"`
	// 保留的最大旧文件数量
	MaxBackups int `json:"max_backups" yaml:"max_backups"`
	// 保留旧文件的最大天数
	MaxAge int `json:"max_age" yaml:"max_age"`
	// 是否压缩旧文件
	Compress bool `json:"compress" yaml:"compress"`
	// 是否在日志中显示调用位置
	ShowCaller bool `json:"show_caller" yaml:"show_caller"`
	// 是否输出到控制台
	EnableConsole bool `json:"enable_console" yaml:"enable_console"`
	// 是否使用 JSON 格式输出
	EnableJSON bool `json:"enable_json" yaml:"enable_json"`
	// 是否开启堆栈跟踪 建议error以上再开
	EnableStackTrace bool `json:"enable_stack_trace" yaml:"enable_stack_trace"`
	// 是否开启堆栈跟踪 建议error以上再开
	StackTraceLevel string `json:"stack_trace_level" yaml:"stack_trace_level"`
	// 时间格式
	TimeFormat string `json:"time_format" yaml:"time_format"`
	// 定义元信息，这些信息是一条日志文件中必须的字段
	LogConfigMeta LogConfigMeta `json:"log_config_meta" yaml:"log_config_meta"`
	// 定义元信息，这些信息是一条日志文件中业务自定义的字段
	CustomFields map[string]interface{} `json:"custom_fields" yaml:"custom_fields"`
}

// LogConfigMeta 日志元数据配置
type LogConfigMeta struct {
	TimeKey       string `json:"time_key" yaml:"time_key"`
	LevelKey      string `json:"level_key" yaml:"level_key"`
	NameKey       string `json:"name_key" yaml:"name_key"`
	CallerKey     string `json:"caller_key" yaml:"caller_key"`
	MessageKey    string `json:"message_key" yaml:"message_key"`
	StacktraceKey string `json:"stacktrace_key" yaml:"stacktrace_key"`
	AppName       string `json:"app_name" yaml:"app_name"`
	Env           string `json:"env" yaml:"env"`
}

// Initial 对日志中没over的参数做初始化
func (c *LoggerConfig) Initial() {
	if c.Level == "" {
		c.Level = "info"
	}
	if c.OutputDir == "" {
		c.OutputDir = "logs"
	}
	if c.MaxSize == 0 {
		c.MaxSize = 200
	}
	if c.MaxBackups == 0 {
		c.MaxBackups = 30
	}
	if c.MaxAge == 0 {
		c.MaxAge = 7
	}
	if c.TimeFormat == "" {
		c.TimeFormat = "2006-01-02 15:04:05"
	}
	if c.FileNameFormat == "" {
		c.FileNameFormat = "APP-%Y%m%d.log"
	}
	if c.StackTraceLevel == "" {
		c.StackTraceLevel = "error"
	}

	// 初始化日志元数据字段
	if c.LogConfigMeta.TimeKey == "" {
		c.LogConfigMeta.TimeKey = "time"
	}
	if c.LogConfigMeta.LevelKey == "" {
		c.LogConfigMeta.LevelKey = "level"
	}
	if c.LogConfigMeta.NameKey == "" {
		c.LogConfigMeta.NameKey = "logging"
	}
	if c.LogConfigMeta.CallerKey == "" {
		c.LogConfigMeta.CallerKey = "caller"
	}
	if c.LogConfigMeta.MessageKey == "" {
		c.LogConfigMeta.MessageKey = "msg"
	}
	if c.LogConfigMeta.StacktraceKey == "" {
		c.LogConfigMeta.StacktraceKey = "stack"
	}
}

func DefaultLogConfig() *LoggerConfig {
	loggerCfg := new(LoggerConfig)
	loggerCfg.Initial()
	return loggerCfg
}
