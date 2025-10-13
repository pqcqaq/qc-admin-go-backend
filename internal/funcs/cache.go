package funcs

import (
	"context"
	"go-backend/database/ent"
	"go-backend/database/ent/apiauth"
	"go-backend/pkg/database"
	"go-backend/pkg/logging"
	"go-backend/pkg/utils"
	"go-backend/shared/models"
	"sync"
)

type WsCache struct {
	id          uint64
	action      string
	active      bool
	topic       string
	permissions []*models.PermissionResponse
	isPublic    bool
}

var wsCache map[uint64]*WsCache
var wsCacheLock sync.RWMutex

func queryWsList() []*ent.APIAuth {
	records, err := database.Client.APIAuth.Query().
		Where(apiauth.TypeEQ(apiauth.TypeWebsocket)).
		WithPermissions().
		All(context.Background())

	if err != nil {
		if ent.IsNotFound(err) {
			return make([]*ent.APIAuth, 0)
		}
		return nil
	}
	return records
}

func makeCache(records []*ent.APIAuth) map[uint64]*WsCache {
	cacheMap := make(map[uint64]*WsCache)
	for _, record := range records {
		var permissionsList []*models.PermissionResponse
		for _, perm := range record.Edges.Permissions {
			permissionsList = append(permissionsList, &models.PermissionResponse{
				ID:          utils.Uint64ToString(perm.ID),
				Action:      perm.Action,
				Description: perm.Description,
			})
		}
		cacheMap[record.ID] = &WsCache{
			id:          record.ID,
			action:      record.Method,
			topic:       record.Path,
			active:      record.IsActive,
			permissions: permissionsList,
			isPublic:    record.IsPublic,
		}
	}
	return cacheMap
}

func updateCache(item *ent.APIAuth) {
	wsCacheLock.Lock()
	oldOne := wsCache[item.ID]
	// 如果老的存在并且新的类型是http，则删除
	if oldOne != nil && item.Type == apiauth.TypeHTTP {
		delete(wsCache, item.ID)
		logging.Info("WebSocket cache deleted for ID %d due to type change to HTTP", item.ID)
		wsCacheLock.Unlock()
		return
	}
	wsCacheLock.Unlock()
	var permissionsList []*models.PermissionResponse
	for _, perm := range item.Edges.Permissions {
		permissionsList = append(permissionsList, &models.PermissionResponse{
			ID:          utils.Uint64ToString(perm.ID),
			Action:      perm.Action,
			Description: perm.Description,
		})
	}
	newOne := &WsCache{
		id:          item.ID,
		action:      item.Method,
		topic:       item.Path,
		active:      item.IsActive,
		permissions: permissionsList,
		isPublic:    item.IsPublic,
	}
	wsCacheLock.Lock()
	wsCache[item.ID] = newOne // 无论存在与否，直接设置
	wsCacheLock.Unlock()

	logging.Info("WebSocket cache updated for ID %d", item.ID)
}

func deleteByKey(id uint64) {
	wsCacheLock.Lock()

	delete(wsCache, id)
	wsCacheLock.Unlock()

	logging.Info("WebSocket cache deleted for ID %d", id)
}

func IsTopicAllowed(topic string, userId uint64, action string) (bool, []*models.PermissionResponse) {
	wsCacheLock.RLock()

	ctx := context.Background()
	var matchedCache []*WsCache = make([]*WsCache, 0)
	for _, cache := range wsCache {
		if cache.action != action || !cache.active {
			continue
		}
		matched := utils.MatchTopic(cache.topic, topic)
		if matched {
			if cache.isPublic {
				wsCacheLock.RUnlock() // 释放读锁
				logging.Info("WebSocket topic %s is public, allowed, action: %s", topic, action)
				return true, cache.permissions
			}
			matchedCache = append(matchedCache, cache)
		}
	}

	// 若没找到则不允许
	if len(matchedCache) == 0 {
		wsCacheLock.RUnlock()
		logging.Warn("WebSocket topic %s not found in cache, denied, action: %s", topic, action)
		return false, nil
	}

	var reqPerms []string = make([]string, 0)
	for _, cache := range matchedCache {
		for _, perm := range cache.permissions {
			reqPerms = append(reqPerms, perm.Action)
		}
	}

	wsCacheLock.RUnlock()
	has, err := HasAnyPermissionsOptimized(ctx, userId, reqPerms)

	if err != nil {
		logging.Error("Error checking permissions for user %d on topic %s: %v", userId, topic, err)
		return false, nil
	}
	return has, nil
}
