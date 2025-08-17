package web

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

// OutboxPoller is a simplified background routine that processes "NEW" outbox records
// and sends them to MQ, then marks them as "SENT" on success.
type OutboxPoller struct {
	db       *sql.DB
	producer producer.Producer // or any MQ client
	interval time.Duration
}

// Start begins the periodic polling in a background goroutine.
// In real systems, you might want a graceful shutdown approach.
func (op *OutboxPoller) Start(ctx context.Context) {
	ticker := time.NewTicker(op.interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Println("OutboxPoller shutting down")
				return
			case <-ticker.C:
				op.processOutbox(ctx)
			}
		}
	}()
}

func (op *OutboxPoller) processOutbox(ctx context.Context) {
	rows, err := op.db.QueryContext(ctx, 
		`SELECT id, event_type, payload FROM outbox_events WHERE status='NEW' ORDER BY id ASC LIMIT 50`)
	if err != nil {
		log.Printf("processOutbox query error: %v\n", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id        int64
			eventType string
			payload   string
		)
		if err := rows.Scan(&id, &eventType, &payload); err != nil {
			log.Printf("row scan error: %v\n", err)
			continue
		}

		// Send to MQ
		if sendErr := op.sendToMQ(ctx, eventType, payload); sendErr != nil {
			log.Printf("sendToMQ error: %v\n", sendErr)
			// do not update outbox yet, so we can retry later
			continue
		}

		// Mark as SENT on success
		if _, err := op.db.ExecContext(ctx, 
			`UPDATE outbox_events SET status='SENT', updated_at=? WHERE id=?`,
			time.Now(), id,
		); err != nil {
			log.Printf("failed to update outbox event status, id=%d, err=%v\n", id, err)
		}
	}

	if err := rows.Err(); err != nil {
		log.Printf("rows iteration error: %v\n", err)
	}
}

// sendToMQ is a stub that constructs a message and sends it to RocketMQ or other MQ.
func (op *OutboxPoller) sendToMQ(ctx context.Context, eventType string, payload string) error {
	// parse the payload if needed, etc.
	msg := &primitive.Message{
		Topic: "web_deletion_topic",
		Body:  []byte(payload), // or you could store a structured JSON with "code" etc.
	}
	// In real code, handle the result, retries, etc.
	result, err := op.producer.SendSync(ctx, msg)
	if err != nil {
		return fmt.Errorf("producer.SendSync error: %w", err)
	}
	if result.Status != primitive.SendOK {
		return fmt.Errorf("send status not OK: %v", result.Status)
	}
	return nil
}