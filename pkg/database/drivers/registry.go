package drivers

import (
	"fmt"
	"sort"
	"sync"
)

// DriverInfo 数据库驱动信息
type DriverInfo struct {
	Name        string // 驱动名称 (如 "mysql", "postgres")
	DisplayName string // 显示名称 (如 "MySQL", "PostgreSQL")
	IsLoaded    bool   // 是否已加载
}

var (
	// 已注册的驱动
	registeredDrivers = make(map[string]*DriverInfo)
	driverMutex       sync.RWMutex
)

// RegisterDriver 注册数据库驱动
func RegisterDriver(name, displayName string) {
	driverMutex.Lock()
	defer driverMutex.Unlock()

	registeredDrivers[name] = &DriverInfo{
		Name:        name,
		DisplayName: displayName,
		IsLoaded:    true,
	}
}

// GetRegisteredDrivers 获取所有已注册的驱动
func GetRegisteredDrivers() map[string]*DriverInfo {
	driverMutex.RLock()
	defer driverMutex.RUnlock()

	// 创建副本以避免并发问题
	drivers := make(map[string]*DriverInfo)
	for k, v := range registeredDrivers {
		drivers[k] = &DriverInfo{
			Name:        v.Name,
			DisplayName: v.DisplayName,
			IsLoaded:    v.IsLoaded,
		}
	}
	return drivers
}

// GetSupportedDrivers 获取支持的驱动列表（按字母序排序）
func GetSupportedDrivers() []string {
	driverMutex.RLock()
	defer driverMutex.RUnlock()

	var drivers []string
	for name := range registeredDrivers {
		drivers = append(drivers, name)
	}
	sort.Strings(drivers)
	return drivers
}

// IsDriverSupported 检查驱动是否被支持
func IsDriverSupported(driverName string) bool {
	driverMutex.RLock()
	defer driverMutex.RUnlock()

	_, exists := registeredDrivers[driverName]
	return exists
}

// GetDriverInfo 获取特定驱动的信息
func GetDriverInfo(driverName string) (*DriverInfo, error) {
	driverMutex.RLock()
	defer driverMutex.RUnlock()

	if info, exists := registeredDrivers[driverName]; exists {
		return &DriverInfo{
			Name:        info.Name,
			DisplayName: info.DisplayName,
			IsLoaded:    info.IsLoaded,
		}, nil
	}

	return nil, fmt.Errorf("driver '%s' is not registered", driverName)
}

// ListDrivers 列出所有已注册的驱动信息
func ListDrivers() string {
	drivers := GetRegisteredDrivers()
	if len(drivers) == 0 {
		return "No database drivers loaded"
	}

	result := fmt.Sprintf("Loaded database drivers (%d):\n", len(drivers))
	for _, name := range GetSupportedDrivers() {
		if info, exists := drivers[name]; exists {
			result += fmt.Sprintf("  - %s (%s)\n", info.DisplayName, info.Name)
		}
	}
	return result
}
