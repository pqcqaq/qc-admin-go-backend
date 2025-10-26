package schema

import (
	"go-backend/database/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Area 地区实体(存储全国行政区划数据)
type Area struct {
	ent.Schema
}

func (Area) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_areas"},
	}
}

func (Area) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (Area) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			MaxLen(32).
			NotEmpty().
			Comment("地区名称"),
		// 拼音首字母
		field.String("spell").
			MaxLen(8).
			NotEmpty().
			Comment("拼音首字母"),
		field.Enum("level").
			Values("country", "province", "city", "district", "street").
			Comment("层级类型"),
		field.Int("depth").
			Min(0).
			Max(4).
			Comment("深度: 0=国家、1=省、2=市、3=区、4=街道"),
		field.String("code").
			MaxLen(12).
			NotEmpty().
			Comment("地区编码(国家标准行政区划代码)"),
		field.Float("latitude").
			Comment("纬度"),
		field.Float("longitude").
			Comment("经度"),
		field.Uint64("parent_id").
			Optional().
			Comment("上级地区ID"),
		field.String("color").
			MaxLen(20).
			Optional().
			Comment("界面显示颜色"),
	}
}

func (Area) Edges() []ent.Edge {
	return []ent.Edge{
		// 自关联: 上级地区
		edge.To("parent", Area.Type).
			Field("parent_id").
			Unique().
			From("children"),
		// 反向边: 下级地区
		edge.From("addresses", Address.Type).
			Ref("area"),
		edge.From("stations", Station.Type).
			Ref("area"),
		edge.From("subways", Subway.Type).
			Ref("area"),
	}
}

func (Area) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code", "delete_time").Unique(),
		index.Fields("name"),
		index.Fields("level"),
		index.Fields("depth"),
		index.Fields("parent_id"),
	}
}

// Address 地址实体(用户收货地址等)
type Address struct {
	ent.Schema
}

func (Address) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_addresses"},
	}
}

func (Address) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (Address) Fields() []ent.Field {
	return []ent.Field{
		field.String("detail").
			MaxLen(256).
			NotEmpty().
			Comment("详细地址(街道门牌号等)"),
		field.Uint64("area_id").
			Comment("所在地区ID"),
		field.String("phone").
			MaxLen(20).
			NotEmpty().
			Comment("联系电话"),
		field.String("name").
			MaxLen(32).
			NotEmpty().
			Comment("收件人姓名"),
		field.Bool("is_default").
			Default(false).
			Comment("是否为默认地址"),
		field.Text("remark").
			Optional().
			Comment("备注信息"),
		field.String("entity").
			MaxLen(32).
			Optional().
			Comment("关联的实体类型(如user、store)"),
		field.String("entity_id").
			MaxLen(64).
			Optional().
			Comment("关联的实体ID"),
	}
}

func (Address) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("area", Area.Type).
			Field("area_id").
			Unique().
			Required().
			Comment("所在地区"),
	}
}

func (Address) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("entity", "entity_id"),
		index.Fields("is_default"),
		index.Fields("area_id"),
	}
}

// Station 站点实体(交通站点信息)
type Station struct {
	ent.Schema
}

func (Station) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_stations"},
	}
}

func (Station) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (Station) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			MaxLen(32).
			NotEmpty().
			Comment("站点名称"),
		field.Uint64("area_id").
			Comment("所在城市ID"),
	}
}

func (Station) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("area", Area.Type).
			Field("area_id").
			Unique().
			Required().
			Comment("所在城市"),
		edge.To("subway_stations", SubwayStation.Type).
			Comment("所属的地铁线路"),
	}
}

func (Station) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name", "area_id", "delete_time").Unique(),
		index.Fields("area_id"),
	}
}

// Subway 地铁线路实体
type Subway struct {
	ent.Schema
}

func (Subway) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_subways"},
	}
}

func (Subway) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (Subway) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			MaxLen(32).
			NotEmpty().
			Comment("线路名称(如1号线)"),
		field.Uint64("area_id").
			Comment("所在城市ID"),
		field.String("color").
			MaxLen(20).
			Optional().
			Comment("线路颜色标识"),
	}
}

func (Subway) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("area", Area.Type).
			Field("area_id").
			Unique().
			Required().
			Comment("所在城市"),
		edge.To("subway_stations", SubwayStation.Type).
			Comment("包含的站点"),
	}
}

func (Subway) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name", "area_id", "delete_time").Unique(),
		index.Fields("area_id"),
	}
}

// SubwayStation 地铁站点连接表(多对多关系中间表)
type SubwayStation struct {
	ent.Schema
}

func (SubwayStation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_subway_stations"},
	}
}

func (SubwayStation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (SubwayStation) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("station_id").
			Comment("站点ID"),
		field.Uint64("subway_id").
			Comment("地铁线路ID"),
		field.Int("sequence").
			Optional().
			Comment("在线路中的顺序(从起点到终点)"),
	}
}

func (SubwayStation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("station", Station.Type).
			Field("station_id").
			Unique().
			Required().
			Comment("站点"),
		edge.To("subway", Subway.Type).
			Field("subway_id").
			Unique().
			Required().
			Comment("地铁线路"),
	}
}

func (SubwayStation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("station_id", "subway_id", "delete_time").Unique(),
		index.Fields("subway_id", "sequence"),
	}
}
