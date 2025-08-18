package events

import (
	"context"
	"reflect"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/mixin"
)

// QueryEventMixin 查询事件驱动的mixin，用于拦截查询操作
type QueryEventMixin struct {
	mixin.Schema
}

// Interceptors 返回查询拦截器
func (QueryEventMixin) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		// 查询拦截器 - 拦截所有查询操作
		ent.InterceptFunc(func(next ent.Querier) ent.Querier {
			return ent.QuerierFunc(func(ctx context.Context, query ent.Query) (ent.Value, error) {
				// 创建并发布 pre-select 事件
				preEvent := NewQueryEvent(EventTypePreSelect, ctx, query)
				if err := Publish(ctx, preEvent); err != nil {
					return nil, err
				}

				// 执行实际的查询
				value, err := next.Query(ctx, query)

				// 创建并发布 post-select 事件
				postEvent := NewQueryEventWithValue(EventTypePostSelect, ctx, query, value, err)
				Publish(ctx, postEvent)

				return value, err
			})
		}),
	}
}

// NewQueryEvent 创建查询事件
func NewQueryEvent(eventType EventType, ctx context.Context, query ent.Query) *Event {
	return &Event{
		Type:       eventType,
		EntityType: getQueryEntityType(query),
		Operation:  getQueryOperation(query),
		Context:    ctx,
		Mutation:   nil, // 查询操作没有 mutation
		Fields:     extractQueryFields(query),
		Timestamp:  time.Now(),
	}
}

// NewQueryEventWithValue 创建带有查询结果的查询事件
func NewQueryEventWithValue(eventType EventType, ctx context.Context, query ent.Query, value ent.Value, err error) *Event {
	return &Event{
		Type:       eventType,
		EntityType: getQueryEntityType(query),
		Operation:  getQueryOperation(query),
		Context:    ctx,
		Mutation:   nil,
		NewValue:   value,
		Error:      err,
		Fields:     extractQueryFields(query),
		Timestamp:  time.Now(),
	}
}

// getQueryEntityType 从查询中提取实体类型
func getQueryEntityType(query ent.Query) string {
	// 使用类型断言来获取具体的查询类型
	switch query := query.(type) {
	case interface{ Table() string }:
		return query.Table()
	}

	// 备用方案：从类型名称推断
	queryType := getQueryTypeName(query)
	if len(queryType) > 5 && queryType[len(queryType)-5:] == "Query" {
		return queryType[:len(queryType)-5]
	}
	return "Unknown"
}

// getQueryTypeName 获取查询对象的类型名称
func getQueryTypeName(query ent.Query) string {
	t := reflect.TypeOf(query)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

// getQueryOperation 从查询中推断操作类型
func getQueryOperation(query ent.Query) ent.Op {
	// 所有查询操作统一使用 OpSelect 操作类型
	return OpSelect
}

// extractQueryFields 从查询中提取相关字段信息
func extractQueryFields(query ent.Query) map[string]any {
	fields := make(map[string]any)

	// 尝试提取查询的一些基本信息
	if q, ok := query.(interface{ String() string }); ok {
		fields["query_string"] = q.String()
	}

	// 可以根据需要添加更多字段提取逻辑
	return fields
}
