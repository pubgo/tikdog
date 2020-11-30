package schema

import (
	"github.com/facebook/ent"
	"github.com/facebook/ent/schema/field"
)

// WatcherFile holds the schema definition for the WatcherFile entity.
type WatcherFile struct {
	ent.Schema
}

// Fields of the WatcherFile.
func (WatcherFile) Fields() []ent.Field {
	return []ent.Field{
		field.Int("age").
			Positive(),
		field.String("name").
			Default("unknown"),
	}
}

// Edges of the WatcherFile.
func (WatcherFile) Edges() []ent.Edge {
	return nil
}
