package funcs

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"go-backend/database/ent"
	"go-backend/database/ent/loginrecord"
	"go-backend/pkg/database"
	"go-backend/pkg/logging"

	"github.com/gin-gonic/gin"
)

// LoginRecordStatus 登录记录状态常量
const (
	LoginStatusSuccess = "success"
	LoginStatusFailed  = "failed"
	LoginStatusLocked  = "locked"
)

// LoginRecordParams 登录记录参数
type LoginRecordParams struct {
	UserID         uint64
	Identifier     string
	CredentialType string
	IPAddress      string
	UserAgent      string
	DeviceInfo     string
	Location       string
	Status         string
	FailureReason  string
	SessionID      string
	Metadata       map[string]interface{}
	ClientId       uint64
}

// CreateLoginRecord 创建登录记录
func CreateLoginRecord(ctx context.Context, params LoginRecordParams) (*ent.LoginRecord, error) {
	// 解析设备信息
	if params.DeviceInfo == "" {
		params.DeviceInfo = parseDeviceInfo(params.UserAgent)
	}

	// 获取地理位置信息（如果未提供）
	if params.Location == "" {
		params.Location = getLocationFromIP(params.IPAddress)
	}

	// 创建登录记录
	builder := database.Client.LoginRecord.Create().
		SetUserID(params.UserID).
		SetIdentifier(params.Identifier).
		SetCredentialType(loginrecord.CredentialType(params.CredentialType)).
		SetIPAddress(params.IPAddress).
		SetStatus(loginrecord.Status(params.Status)).
		SetClientID(params.ClientId)

	// 设置可选字段
	if params.UserAgent != "" {
		builder = builder.SetUserAgent(params.UserAgent)
	}
	if params.DeviceInfo != "" {
		builder = builder.SetDeviceInfo(params.DeviceInfo)
	}
	if params.Location != "" {
		builder = builder.SetLocation(params.Location)
	}
	if params.FailureReason != "" {
		builder = builder.SetFailureReason(params.FailureReason)
	}
	if params.SessionID != "" {
		builder = builder.SetSessionID(params.SessionID)
	}
	if params.Metadata != nil {
		builder = builder.SetMetadata(params.Metadata)
	}

	record, err := builder.Save(ctx)
	if err != nil {
		logging.Error("创建登录记录失败: %v", err)
		return nil, fmt.Errorf("创建登录记录失败: %w", err)
	}

	return record, nil
}

// UpdateLoginRecordLogout 更新登录记录的退出信息
func UpdateLoginRecordLogout(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("会话ID不能为空")
	}

	// 查找对应的登录记录
	loginRecord, err := database.Client.LoginRecord.Query().
		Where(
			loginrecord.SessionIDEQ(sessionID),
			loginrecord.StatusEQ(loginrecord.StatusSuccess),
			loginrecord.LogoutTimeIsNil(),
		).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("未找到对应的登录记录")
		}
		return fmt.Errorf("查询登录记录失败: %w", err)
	}

	// 计算会话持续时间
	duration := int(time.Since(loginRecord.CreateTime).Seconds())

	// 更新退出时间和持续时间
	_, err = loginRecord.Update().
		SetLogoutTime(time.Now()).
		SetDuration(duration).
		Save(ctx)

	if err != nil {
		logging.Error("更新登录记录退出信息失败: %v", err)
		return fmt.Errorf("更新登录记录失败: %w", err)
	}

	return nil
}

// CreateLoginRecordFromGinContext 从Gin上下文创建登录记录
func CreateLoginRecordFromGinContext(ctx context.Context, c *gin.Context, userID uint64, identifier, credentialType, status, failureReason, sessionID string, clientId uint64) (*ent.LoginRecord, error) {
	params := LoginRecordParams{
		UserID:         userID,
		Identifier:     identifier,
		CredentialType: credentialType,
		IPAddress:      getClientIP(c),
		UserAgent:      c.GetHeader("User-Agent"),
		Status:         status,
		FailureReason:  failureReason,
		SessionID:      sessionID,
		Metadata: map[string]interface{}{
			"referer":      c.GetHeader("Referer"),
			"accept":       c.GetHeader("Accept"),
			"content_type": c.GetHeader("Content-Type"),
			"request_time": time.Now().Unix(),
		},
		ClientId: clientId,
	}

	return CreateLoginRecord(ctx, params)
}

// GetUserLoginRecords 获取用户的登录记录
func GetUserLoginRecords(ctx context.Context, userID uint64, limit, offset int) ([]*ent.LoginRecord, error) {
	query := database.Client.LoginRecord.Query().
		Where(loginrecord.UserIDEQ(userID)).
		Order(ent.Desc(loginrecord.FieldCreateTime))

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	records, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询用户登录记录失败: %w", err)
	}

	return records, nil
}

// GetLoginRecordsByStatus 根据状态获取登录记录
func GetLoginRecordsByStatus(ctx context.Context, status string, limit, offset int) ([]*ent.LoginRecord, error) {
	query := database.Client.LoginRecord.Query().
		Where(loginrecord.StatusEQ(loginrecord.Status(status))).
		Order(ent.Desc(loginrecord.FieldCreateTime))

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	records, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询登录记录失败: %w", err)
	}

	return records, nil
}

// GetRecentFailedLoginAttempts 获取最近的失败登录尝试
func GetRecentFailedLoginAttempts(ctx context.Context, identifier string, minutes int) (int, error) {
	since := time.Now().Add(-time.Duration(minutes) * time.Minute)

	count, err := database.Client.LoginRecord.Query().
		Where(
			loginrecord.IdentifierEQ(identifier),
			loginrecord.StatusEQ(loginrecord.StatusFailed),
			loginrecord.CreateTimeGTE(since),
		).
		Count(ctx)

	if err != nil {
		return 0, fmt.Errorf("查询失败登录记录失败: %w", err)
	}

	return count, nil
}

// parseDeviceInfo 解析设备信息
func parseDeviceInfo(userAgent string) string {
	if userAgent == "" {
		return "Unknown"
	}

	ua := strings.ToLower(userAgent)

	// 检测操作系统
	var os string
	switch {
	case strings.Contains(ua, "windows"):
		os = "Windows"
	case strings.Contains(ua, "macintosh") || strings.Contains(ua, "mac os"):
		os = "macOS"
	case strings.Contains(ua, "linux"):
		os = "Linux"
	case strings.Contains(ua, "android"):
		os = "Android"
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad"):
		os = "iOS"
	default:
		os = "Unknown"
	}

	// 检测浏览器
	var browser string
	switch {
	case strings.Contains(ua, "chrome") && !strings.Contains(ua, "edg"):
		browser = "Chrome"
	case strings.Contains(ua, "firefox"):
		browser = "Firefox"
	case strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome"):
		browser = "Safari"
	case strings.Contains(ua, "edg"):
		browser = "Edge"
	case strings.Contains(ua, "opera"):
		browser = "Opera"
	default:
		browser = "Unknown"
	}

	return fmt.Sprintf("%s/%s", os, browser)
}

// getClientIP 获取客户端真实IP
func getClientIP(c *gin.Context) string {
	// 尝试从各种头部获取真实IP
	headers := []string{"X-Forwarded-For", "X-Real-IP", "X-Client-IP"}

	for _, header := range headers {
		ip := c.GetHeader(header)
		if ip != "" {
			// X-Forwarded-For 可能包含多个IP，取第一个
			if header == "X-Forwarded-For" {
				ips := strings.Split(ip, ",")
				ip = strings.TrimSpace(ips[0])
			}

			// 验证IP格式
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// 如果没有找到，使用RemoteAddr
	ip := c.ClientIP()
	if ip == "" {
		ip = "unknown"
	}

	return ip
}

// getLocationFromIP 从IP获取地理位置（简化版本）
func getLocationFromIP(ip string) string {
	// 这里可以集成第三方IP地理位置服务
	// 目前返回简化信息
	if ip == "" || ip == "unknown" {
		return "Unknown"
	}

	// 检查是否为本地IP
	if isLocalIP(ip) {
		return "Local"
	}

	// 这里可以调用IP地理位置API
	// 例如：ipapi.co, ip-api.com 等
	return "Unknown"
}

// isLocalIP 检查是否为本地IP
func isLocalIP(ip string) bool {
	if ip == "127.0.0.1" || ip == "::1" || ip == "localhost" {
		return true
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// 检查私有IP段
	private := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}

	for _, cidr := range private {
		_, subnet, _ := net.ParseCIDR(cidr)
		if subnet.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// CleanupOldLoginRecords 清理旧的登录记录
func CleanupOldLoginRecords(ctx context.Context, days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)

	count, err := database.Client.LoginRecord.Delete().
		Where(loginrecord.CreateTimeLT(cutoff)).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("清理旧登录记录失败: %w", err)
	}

	logging.Info("清理了 %d 条超过 %d 天的登录记录", count, days)
	return nil
}
