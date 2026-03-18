package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type AuditLog struct {
	Id        int       `json:"id"`
	Entity    string    `json:"entity"`
	EntityId  int       `json:"entity_id"`
	Action    string    `json:"action"`
	ActorId   *int      `json:"actor_id"`
	OldData   []byte    `json:"ols_data"`
	NewData   []byte    `json:"new_data"`
	ChangedAt time.Time `json:"changed_at"`
}

type AuditLogRepository interface {
	Log(ctx context.Context, data AuditLog) error
}

type AuditLogRepo struct {
	db *sql.DB
}

func NewAuditLogRepo(db *sql.DB) *AuditLogRepo {
	return &AuditLogRepo{db: db}
}

func (r *AuditLogRepo) Log(ctx context.Context, tx *sql.Tx, data AuditLog) error {
	query := `
		insert into audit_logs (entity, entity_id, action, actor_id, old_data, new_data)
		values($1, $2, $3, $4, $5, $6)
	`

	_, err := tx.ExecContext(ctx, query, data.Entity, data.EntityId, data.Action, data.ActorId, data.OldData, data.NewData)
	if err != nil {
		return fmt.Errorf("Log: %w", err)
	}

	return nil
}

func toJSON(data any) []byte {
	b, _ := json.Marshal(data)
	return b
}
