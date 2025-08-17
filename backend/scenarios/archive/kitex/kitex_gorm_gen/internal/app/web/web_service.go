package web

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/kitex-examples/bizdemo/kitex_gorm_gen/internal/mq/rmq"
	"github.com/cloudwego/kitex-examples/bizdemo/kitex_gorm_gen/internal/repository/model/query"
)

/*
# Task
Implement the function as below requirements:
1. logic
 1. delete the record with code in mysql
 2. send a mq to all the downstreams notifying the deletion (rocketmq broadcast mode)
 3. delte the redis cache (key: code, value: model.analyticsWeb)

2. you must adhere to big company's best practices, as standard as possible
3. you must consier corner cases as more as possible
4. must implement the strong distributed transaction, i.e, if mq send or redis deletion fails, rollback the previous operastions
5. provide all solutions, pros and cons, and finally select the best one to code below

# Answer
Below are some common patterns/approaches used by large enterprises:

1. **Two-Phase Commit (2PC)**  
   - **How it works**: A coordinating service (transaction manager) performs 
   a "prepare" phase on all participants, 
   and if all prepare requests succeed, it issues a "commit" phase.  
   - **Pros**: Strong consistency; if all participants can handle 2PC properly, it ensures an atomic commit/rollback.  
   - **Cons**:  
     - Complex to implement and maintain.  
     - Performance overhead due to locking of resources during prepare phase.  
     - Can lead to blocking issues if the coordinator goes down.  

2. **TCC (Try-Confirm-Cancel)**  
   - **How it works**: Each service implements a "try" (reserve resources),
			"confirm" (final commit),
			"cancel" (release resources).  
   - **Pros**: Fine-grained control over reserving, confirming, or canceling resources.  
   - **Cons**:  
     - More code overhead (must implement try/confirm/cancel steps in each service).  
     - You need to define these steps carefully for each participant.  

3. **Outbox & Polling** (a type of **Saga** pattern)  
   - **How it works**:  
     1. Wrap the DB operation and a record of the "event to publish" in the same local DB transaction.  
     2. After the local transaction commits, a separate "outbox" processor reads the pending events and 
     publishes them to the MQ.  
     3. If sending to MQ fails, the outbox row remains, and it can retry.  
   - **Pros**:  
     - Well-known approach for achieving *eventual consistency*.  
     - Scalable and decoupled.  
   - **Cons**:  
     - *Not strictly synchronous strong consistency.*  
     - Requires an additional outbox table and a reliable process to poll and publish.  

4. **RocketMQ Transaction Messages**  
   - **How it works**:  
     1. Send a *half* message to RocketMQ.  
     2. Execute the local transaction in MySQL 
     (and possibly Redis if you want it done in the same local transaction).  
     3. Based on success/failure, commit or rollback the half message.  
     4. RocketMQ may invoke a check callback to confirm the final state if uncertain.  
   - **Pros**:  
     - Provides a near 2PC-like semantic (the message is only visible to consumers after the local transaction commits).  
     - Less overhead than a fully external 2PC coordinator.  
   - **Cons**:  
     - Involves extra complexity in code: you have to implement the transaction message callback logic.  
     - If your business logic is large, you need to carefully manage how to detect and re-check local transaction state. 
*/
func DeleteWebStrongConsistency(ctx context.Context, adv int64, code string, name string) error {
	// Build the payload to send to MQ
	payload := struct {
		Adv  int64  `json:"adv"`
		Code string `json:"code"`
		Name string `json:"name"`
	}{
		Adv:  adv,
		Code: code,
		Name: name,
	}

	msgBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	// Create the transaction listener (in real practice, you'd ensure
	// proper initialization of db, redisClient, etc.)
	bizListener := &DeleteWebTransactionListener{
		Query: query.Query{
			// db and redisClient would be set properly here, e.g.:
			// db:          globalDB,
			// redisClient: globalRedis,
		},
	}

	// Invoke the distributed transaction via RocketMQ
	if err := rmq.SendMsgWithLocalTransaction(ctx, "web_deletion_topic", msgBytes, bizListener); err != nil {
		return fmt.Errorf("SendMsgWithLocalTransaction error: %v", err)
	}

	return nil
}

// DeleteWebStrongConsistencyOutbox illustrates a function that uses the
// Outbox & Polling approach. The DB transaction atomically (a) deletes
// the record in "web_table" and (b) inserts an event into "outbox_events".
func DeleteWebStrongConsistencyOutbox(ctx context.Context, db *sql.DB, code string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin tx: %v", err)
	}
	defer func() {
		_ = tx.Rollback() // safe to call even if commit already happened
	}()

	// 1) Soft-delete (or actual delete) from web_table
	if _, err = tx.ExecContext(ctx, "UPDATE web_table SET deleted=1 WHERE code=?", code); err != nil {
		return fmt.Errorf("delete from web_table failed: %v", err)
	}

	// 2) Insert outbox event to notify downstream that `code` has been deleted
	eventPayload := struct {
		Code  string `json:"code"`
		Topic string `json:"topic"`
	}{
		Code:  code,
		Topic: "web_deletion_topic",
	}

	payloadBytes, err := json.Marshal(eventPayload)
	if err != nil {
		return fmt.Errorf("json marshal failed: %v", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO outbox_events (event_type, payload, status, created_at) VALUES (?, ?, ?, ?)`,
		"WEB_DELETE",
		string(payloadBytes),
		"NEW",
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("insert outbox_events failed: %v", err)
	}

	// 3) Commit if everything is okay
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %v", err)
	}

	return nil
}