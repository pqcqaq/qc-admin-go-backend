//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate --target "./ent" --feature privacy --feature entql --feature intercept --feature modifier --feature versioned-migration --feature sql/versioned-migration  ./schema

package database
