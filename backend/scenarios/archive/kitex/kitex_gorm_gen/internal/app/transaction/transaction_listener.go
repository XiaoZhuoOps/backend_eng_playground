package web

import (
	"context"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/cloudwego/kitex-examples/bizdemo/kitex_gorm_gen/internal/repository/model/query"
)

// DeleteWebTransactionListener executes and checks the local transaction 
// triggered by the half-message. This ensures atomic deletion in MySQL + Redis.
type DeleteWebTransactionListener struct {
	db
	redis
}

// ExecuteLocalTransaction is called after RocketMQ successfully receives 
// the half-transaction message. It should handle the core DB and Redis logic.
func (l *DeleteWebTransactionListener) ExecuteLocalTransaction(msg *primitive.Message) primitive.LocalTransactionState {
	// For simplicity, just parse 'code' from the message body
	// If you're sending JSON with adv, code, name, you'd unmarshal first.
	code := string(msg.Body)

	// Start DB transaction (mocked or real)
	tx, err := l.db.Begin()
	if err != nil {
		return primitive.RollbackMessageState
	}
	defer func() {
		// In real code, use a deferred error check for rollback if not committed.
		_ = tx.Rollback()
	}()

	// 1) Delete from MySQL (soft-delete or actual)
	err = DeleteWeb(context.Background(), code)
	if err != nil {
		return primitive.RollbackMessageState
	}

	// 2) Attempt redis deletion
	err = l.redisClient.Del(context.Background(), code).Err()
	if err != nil {
		return primitive.RollbackMessageState
	}

	// Commit transaction
	if commitErr := tx.Commit(); commitErr != nil {
		return primitive.RollbackMessageState
	}

	// If all steps succeed, commit the message
	return primitive.CommitMessageState
}

// CheckLocalTransaction is called when RocketMQ is uncertain about 
// the transaction status. Here you verify the DB + Redis state.
func (l *DeleteWebTransactionListener) CheckLocalTransaction(msg *primitive.Message) primitive.LocalTransactionState {
	code := string(msg.Body)

	// 1) Double-check if the DB record is actually marked deleted
	deleted, err := l.checkIfDeletedInDB(code)
	if err != nil {
		return primitive.RollbackMessageState
	}
	if !deleted {
		return primitive.RollbackMessageState
	}

	// 2) Ensure Redis is also removed (idempotent)
	_, _ = l.redisClient.Del(context.Background(), code).Result()

	return primitive.CommitMessageState
}
