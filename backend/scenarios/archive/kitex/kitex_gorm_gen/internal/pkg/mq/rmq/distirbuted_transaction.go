package rmq

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/cloudwego/kitex-examples/bizdemo/kitex_gorm_gen/internal/app/web"
)

func SendMsgWithLocalTransaction(ctx context.Context, topic string, msg []byte, bizListener primitive.TransactionListener) error {
	/*
	  We will send a half-transaction message to RocketMQ.
	  The message will trigger ExecuteLocalTransaction.
	*/
	// 1) Create RocketMQ producer with the given transaction listener
	producer, err := rocketmq.NewTransactionProducer(bizListener)
	if err != nil {
		return fmt.Errorf("failed to create transaction producer: %v", err)
	}
	defer producer.Shutdown()

	if err := producer.Start(); err != nil {
		return fmt.Errorf("failed to start producer: %v", err)
	}

	// 2) Construct the half-transaction message
	message := &primitive.Message{
		Topic: topic,
		Body:  msg,
		// Additional settings (tags, keys) go here if needed
	}

	// 3) Send half message
	res, err := producer.SendMessageInTransaction(ctx, message)
	if err != nil {
		return fmt.Errorf("send half message failed: %v", err)
	}

	// 4) Validate result
	if res.Status != primitive.SendOK {
		return fmt.Errorf("send message status not OK: %v", res.Status)
	}

	// Once local transaction is committed in ExecuteLocalTransaction, 
	// the message gets broadcasted to downstream consumers.
	return nil
}
