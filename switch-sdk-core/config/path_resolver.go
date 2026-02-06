package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	APP_CONFIG_PATH        = "APP_CONFIG_PATH"
	APP_CONFIG_SEARCH_DIRS = "APP_CONFIG_SEARCH_DIRS"
)

// APP_CONFIG_SEARCH_DIRS 定义几个默认的配置文件名
var APP_CONFIG_SEARCH_DIRS_DEFAULT = []string{"etc", "config", "resource", "configs", "cfg"}

// resolvePath 解析给定的路径元素,如果拼接后的路径是绝对路径，则直接返回,如果是相对路径，则根据当前工作目录进行解析
func resolvePath(pathElements ...string) (string, error) {
	path := filepath.Join(pathElements...)

	// 如果路径已经是绝对路径，直接验证并返回
	if filepath.IsAbs(path) {
		return path, nil
	}

	// 如果是相对路径，则相对于当前工作目录来构建绝对路径
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, path), nil
}

// findConfigPath 实现了智能的配置路径查找策略,提供默认的和系统变量找配置文件路径，同时根据filePattern确定是否是配置文件
func findConfigPath(filePattern string) (string, error) {
	//文件或者文件夹命中则返回
	if configPath := os.Getenv(APP_CONFIG_PATH); configPath != "" {
		return configPath, nil
	}

	// 确定要搜索的目录名列表，如果环境变量未设置，则使用一组默认值(当不在from中或者环境变量APP_CONFIG_PATH中指定具体的路径，则将搜索配置目录)
	searchDirs := APP_CONFIG_SEARCH_DIRS_DEFAULT
	if searchDirsStr := os.Getenv(APP_CONFIG_SEARCH_DIRS); searchDirsStr != "" {
		searchDirs = strings.Split(searchDirsStr, ",")
	}
	//以工作空间为根进行查找
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	dir := wd
	for {
		for _, dirName := range searchDirs {
			potentialPath := filepath.Join(dir, dirName)
			// 检查目录是否存在
			if _, err := os.Stat(potentialPath); err == nil {
				var found bool
				_ = filepath.WalkDir(potentialPath, func(path string, d os.DirEntry, err error) error {
					if err != nil {
						return nil
					}
					if !d.IsDir() {
						if matched, _ := filepath.Match(filePattern, d.Name()); matched {
							found = true
							return filepath.SkipAll
						}
					}
					return nil
				})

				if found {
					return filepath.Abs(potentialPath)
				}
			}
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			break
		}
		dir = parentDir
	}

	return "", fmt.Errorf("no config directory (%s) found in any parent directory of %s", strings.Join(searchDirs, ", "), wd)
}
