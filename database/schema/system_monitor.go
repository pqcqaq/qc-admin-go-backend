package schema

import (
	"go-backend/database/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SystemMonitor holds the schema definition for the SystemMonitor entity.
type SystemMonitor struct {
	ent.Schema
}

func (SystemMonitor) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_system_monitor"},
	}
}

// Mixin returns SystemMonitor mixed-in fields.
func (SystemMonitor) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
	}
}

// Fields of the SystemMonitor.
func (SystemMonitor) Fields() []ent.Field {
	return []ent.Field{
		// CPU 使用率 (百分比)
		field.Float("cpu_usage_percent").
			Comment("CPU使用率(%)").
			Min(0).
			Max(100),

		// CPU 核心数
		field.Int("cpu_cores").
			Comment("CPU核心数").
			Positive(),

		// 内存信息
		field.Uint64("memory_total").
			Comment("总内存(字节)"),

		field.Uint64("memory_used").
			Comment("已使用内存(字节)"),

		field.Uint64("memory_free").
			Comment("空闲内存(字节)"),

		field.Float("memory_usage_percent").
			Comment("内存使用率(%)").
			Min(0).
			Max(100),

		// 磁盘信息
		field.Uint64("disk_total").
			Comment("总磁盘空间(字节)"),

		field.Uint64("disk_used").
			Comment("已使用磁盘空间(字节)"),

		field.Uint64("disk_free").
			Comment("空闲磁盘空间(字节)"),

		field.Float("disk_usage_percent").
			Comment("磁盘使用率(%)").
			Min(0).
			Max(100),

		// 网络信息
		field.Uint64("network_bytes_sent").
			Comment("网络发送字节数").
			Default(0),

		field.Uint64("network_bytes_recv").
			Comment("网络接收字节数").
			Default(0),

		// 系统信息
		field.String("os").
			Comment("操作系统").
			MaxLen(50),

		field.String("platform").
			Comment("平台").
			MaxLen(50),

		field.String("platform_version").
			Comment("平台版本").
			MaxLen(100),

		field.String("hostname").
			Comment("主机名").
			MaxLen(255),

		// Go运行时信息
		field.Int("goroutines_count").
			Comment("Goroutine数量").
			NonNegative(),

		field.Uint64("heap_alloc").
			Comment("堆内存分配(字节)"),

		field.Uint64("heap_sys").
			Comment("堆系统内存(字节)"),

		field.Uint32("gc_count").
			Comment("GC次数"),

		// 系统负载 (仅Unix系统)
		field.Float("load_avg_1").
			Comment("1分钟平均负载").
			Optional(),

		field.Float("load_avg_5").
			Comment("5分钟平均负载").
			Optional(),

		field.Float("load_avg_15").
			Comment("15分钟平均负载").
			Optional(),

		// 系统运行时间(秒)
		field.Uint64("uptime").
			Comment("系统运行时间(秒)"),

		// 记录时间戳
		field.Time("recorded_at").
			Comment("记录时间"),
	}
}

// Indexes of the SystemMonitor.
func (SystemMonitor) Indexes() []ent.Index {
	return []ent.Index{
		// 按记录时间索引
		index.Fields("recorded_at").
			Annotations(entsql.Desc()),
		// 按创建时间索引
		index.Fields("create_time").
			Annotations(entsql.Desc()),
	}
}
