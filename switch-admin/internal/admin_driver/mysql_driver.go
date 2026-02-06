package admin_driver

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gitee.com/fatzeng/switch-admin/internal/config"
	"gitee.com/fatzeng/switch-sdk-core/driver"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"gitee.com/fatzeng/switch-sdk-core/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const MySQLDriverType driver.DriverType = "mysql"

type MysqlDriver struct {
	db         *gorm.DB
	driverName string
}

func (m *MysqlDriver) Start(ctx context.Context) error {
	return nil
}

func (m *MysqlDriver) RecreateFromConfig() (driver.Driver, error) {
	return nil, nil
}

func (m *MysqlDriver) Validate(driver driver.Driver) error {
	return nil
}

func (m *MysqlDriver) Close() error {
	if m.db != nil {
		sqlDB, err := m.db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

func (m *MysqlDriver) GetDriverType() driver.DriverType {
	return MySQLDriverType
}

func (m *MysqlDriver) GetDriverName() string {
	return m.driverName
}

func (m *MysqlDriver) SetDriverMeta(name string) {
	m.driverName = name
}

func (m *MysqlDriver) SetFailureCallback(callback driver.DriverFailureCallback) {
}

func (m *MysqlDriver) Db() *gorm.DB {
	return m.db
}

func GetMysql(name string) (*MysqlDriver, error) {
	driverInstance, err := driver.GetManager().GetDriver(MySQLDriverType, name)
	if err != nil {
		return nil, err
	}
	return driverInstance.(*MysqlDriver), nil
}

func CreateMysql(c *config.MySQLConfig, name string) error {
	_, err := driver.GetManager().Register(MySQLDriverType, name, func() (driver.Driver, error) {
		db, err := createMysql(c)
		if err != nil {
			return nil, err
		}
		return &MysqlDriver{
			db: db,
		}, nil
	})
	return err
}

func createMysql(dbc *config.MySQLConfig) (*gorm.DB, error) {
	if dbc == nil || dbc.Host == "" {
		startErr := fmt.Sprint("Need to configure MySQL as the startup option")
		logger.Logger.Panicf(startErr)
		return nil, fmt.Errorf(startErr)
	}

	dsnWithoutDb := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local",
		dbc.Username,
		dbc.Password,
		dbc.Host,
		dbc.Port,
	)

	tempDB, err := sql.Open("mysql", dsnWithoutDb)
	if err != nil {
		logger.Logger.Panicf("Failed to open connection to MySQL server: %v", err)
		return nil, err
	}
	defer tempDB.Close()

	createDbSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbc.DBName)
	_, err = tempDB.Exec(createDbSQL)
	if err != nil {
		logger.Logger.Panicf("Failed to execute 'CREATE DATABASE' for db '%s': %v", dbc.DBName, err)
		return nil, err
	}
	//数据库需要创建，switch库
	logger.Logger.Infof("Database '%s' is ready", dbc.DBName)

	// 解析日志级别
	var logLevel gormlogger.LogLevel
	switch dbc.Logger.LogLevel {
	case "silent":
		logLevel = gormlogger.Silent
	case "error":
		logLevel = gormlogger.Error
	case "warn":
		logLevel = gormlogger.Warn
	case "info":
		logLevel = gormlogger.Info
	default:
		logLevel = gormlogger.Warn
	}

	// 解析慢SQL阈值
	slowThreshold, err := time.ParseDuration(dbc.Logger.SlowThreshold)
	if err != nil {
		slowThreshold = time.Second // 默认1秒
	}

	db := mysql.New(mysql.Config{
		DSN: dbc.ToDSN(),
	})
	my, err := gorm.Open(db, &gorm.Config{
		PrepareStmt: true,
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger: &gormLoggerAdapter{
			logLevel:                  logLevel,
			slowThreshold:             slowThreshold,
			ignoreRecordNotFoundError: dbc.Logger.IgnoreRecordNotFoundError,
			parameterizedQueries:      dbc.Logger.ParameterizedQueries,
			colorful:                  dbc.Logger.Colorful,
		},
	})

	if err != nil {
		return nil, err
	}

	sqlDB, err := my.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池配置
	if dbc.Pool.MinConnections > 0 {
		sqlDB.SetMaxIdleConns(dbc.Pool.MinConnections)
	}
	if dbc.Pool.MaxConnections > 0 {
		sqlDB.SetMaxOpenConns(dbc.Pool.MaxConnections)
	}
	if dbc.Pool.MaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(dbc.Pool.MaxIdleTime) * time.Second)
	}
	if dbc.Pool.ConnectTimeout > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(dbc.Pool.ConnectTimeout) * time.Second)
	}

	if err = my.AutoMigrate(
		&admin_model.NamespaceApprovalForm{},
		&admin_model.Approval{},
		&admin_model.SwitchApproval{},
		&model.Driver{},
		&admin_model.Environment{},
		&admin_model.Namespace{},
		&admin_model.Permission{},
		&admin_model.Role{},
		&admin_model.NamespaceUserRole{},
		&admin_model.RolePermission{},
		&admin_model.SwitchConfig{},
		&admin_model.SwitchFactor{},
		&admin_model.SwitchSnapshot{},
		&admin_model.User{},
		&admin_model.NamespaceMembers{},
		&model.SwitchModel{},
	); err != nil {
		return nil, fmt.Errorf("failed to auto migrate database: %w", err)
	}

	logger.Logger.Info("MySQL admin_driver initialized successfully.")

	return my, nil
}
