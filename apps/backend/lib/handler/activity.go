package handler

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/felipeserra/workagents/apps/backend/lib/db"
)

// logActivity registra uma ação no activity log
func logActivity(companyID, actorID, action, targetType, targetID string, metadata map[string]string) {
	id := uuid.New().String()
	metaJSON := "{}"
	if metadata != nil {
		b, err := json.Marshal(metadata)
		if err == nil {
			metaJSON = string(b)
		}
	}

	_, err := db.DB.Exec(
		`INSERT INTO activity_logs (id, company_id, actor_id, action, target_type, target_id, metadata)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, companyID, actorID, action, targetType, targetID, metaJSON,
	)
	if err != nil {
		log.Printf("[activity] failed to log: %v", err)
	}
}
