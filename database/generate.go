//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate --target "./ent" --feature schema/snapshot --feature privacy --feature entql --feature intercept --feature modifier --feature versioned-migration --feature sql/versioned-migration --feature namedges --feature sql/lock --feature sql/execquery ./schema

package database
