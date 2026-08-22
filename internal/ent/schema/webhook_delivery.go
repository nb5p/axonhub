package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/scopes"
)

// WebhookDelivery stores a sanitized audit record for one outbound notification.
// Secrets are removed before this entity is written; it must never hold a raw
// bearer token, Bark device key, proxy credential, or other request secret.
type WebhookDelivery struct {
	ent.Schema
}

func (WebhookDelivery) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

func (WebhookDelivery) Fields() []ent.Field {
	return []ent.Field{
		field.String("event"),
		field.String("target_name"),
		field.String("target_type"),
		field.String("url"),
		field.String("method"),
		field.JSON("request_headers", []objects.HeaderEntry{}).
			Default([]objects.HeaderEntry{}),
		field.String("request_body").
			SchemaType(map[string]string{dialect.MySQL: "mediumtext"}),
		field.Enum("status").Values("success", "failed"),
		field.Int("response_status").Default(0),
		field.String("error_message").
			Optional().
			SchemaType(map[string]string{dialect.MySQL: "mediumtext"}),
	}
}

func (WebhookDelivery) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.Skip(entgql.SkipAll),
	}
}

func (WebhookDelivery) Policy() ent.Policy {
	return scopes.Policy{
		Query: scopes.QueryPolicy{
			scopes.OwnerRule(),
			scopes.UserReadScopeRule(scopes.ScopeReadSettings),
		},
		Mutation: scopes.MutationPolicy{
			scopes.OwnerRule(),
			scopes.UserWriteScopeRule(scopes.ScopeWriteSettings),
		},
	}
}
