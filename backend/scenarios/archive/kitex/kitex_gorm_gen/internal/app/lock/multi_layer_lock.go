package web

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudwego/kitex-examples/bizdemo/kitex_gorm_gen/internal/repository/model/query"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	goredislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Redsync instance for distributed locking. This should typically be a singleton or globally shared.
var (
	rs             *redsync.Redsync
	ErrNoRedsync   = errors.New("redsync is not initialized")
	ErrLockFailure = errors.New("failed to acquire distributed lock")
)

// InitializeRedisSync initializes a Redsync instance with a Redis pool.
//
// Production Considerations / Best Practices:
// 1. Use environment variables or configuration files to manage Redis connection settings.
// 2. Implement robust error handling/logging if Redis is unreachable.
// 3. Wrap this in your main application startup so that it’s called only once.
func InitializeRedisSync(redisAddr string, redisPassword string, db int) error {
	client := goredislib.NewClient(&goredislib.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       db,
	})

	// Ping to ensure connectivity
	if err := client.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("failed to ping redis: %w", err)
	}

	pool := goredis.NewPool(client)
	rs = redsync.New(pool)
	return nil
}

// UpdateWebWithOptimisticLock updates the AnalyticsWeb record using optimistic locking.
//
// Steps:
// 1. Checks whether a record with (web_code = code AND version = version) exists.
// 2. If found, updates the 'extra' field and increments 'version' by 1.
// 3. If the RowsAffected is zero, it indicates the record was modified by another transaction.
//
// Potential Corner Cases:
// 1. Record not found with the specified version -> RowsAffected = 0 -> returns an error.
// 2. Concurrency conflict -> also results in RowsAffected = 0.
// 3. Large amounts of concurrent updates -> consider retries or backoff if needed.
func UpdateWebWithOptimisticLock(ctx context.Context, code string, extra string, version int32) error {
	web := query.AnalyticsWeb

	result, err := web.WithContext(ctx).
		Where(
			web.WebCode.Eq(code),
			web.Version.Eq(version),
		).
		Updates(map[string]interface{}{
			"extra":   extra,
			"version": gorm.Expr("version + ?", 1),
		})
	if err != nil {
		return fmt.Errorf("failed to update AnalyticsWeb with optimistic lock: %w", err)
	}
	if result.RowsAffected == 0 {
		return errors.New("optimistic lock failed: record has been modified or does not match the expected version")
	}

	return nil
}

// UpdateWebWithPessimisticLock updates the AnalyticsWeb record using a "SELECT ... FOR UPDATE" style lock (row-level).
//
// Steps:
// 1. Begins a transaction.
// 2. Queries the row with a "UPDATE" lock (pessimistic lock).
// 3. Updates the row within the same transaction.
//
// Potential Corner Cases:
// 1. Lock wait timeouts in high-contention environments (MySQL can throw an error if the lock can’t be acquired).
// 2. Transaction deadlocks -> recommend adding retry logic in production or gracefully handling deadlock errors.
func UpdateWebWithPessimisticLock(ctx context.Context, code string, extra string) error {
	return query.Q.Transaction(func(tx *query.Query) error {
		// Acquire a row lock on the record we want to update
		// Note!!!: Subsequent Transactions will **block and wait** until the lock is released by the first transaction.
		existingWeb, err := tx.AnalyticsWeb.WithContext(ctx).
			Where(tx.AnalyticsWeb.WebCode.Eq(code)).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First()
		if err != nil {
			return fmt.Errorf("failed to acquire pessimistic lock: %w", err)
		}
		if existingWeb == nil {
			return fmt.Errorf("no record found with WebCode=%s", code)
		}

		// Update the record
		res, err := tx.AnalyticsWeb.WithContext(ctx).
			Where(tx.AnalyticsWeb.WebCode.Eq(code)).
			Updates(map[string]interface{}{
				"extra":   extra,
				"version": gorm.Expr("version + ?", 1),
			})
		if err != nil {
			return fmt.Errorf("failed to update AnalyticsWeb with pessimistic lock: %w", err)
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("pessimistic lock update failed for WebCode=%s", code)
		}

		return nil
	})
}

// UpdateWebWithDistributedLock updates the AnalyticsWeb record using a Redis-based distributed lock via Redsync.
//
// Steps:
// 1. Acquires a distributed lock on the key "lock:web:{code}".
// 2. Updates the row in MySQL.
// 3. Releases the distributed lock.
//
// Potential Corner Cases:
// 1. The lock might expire before the update completes if Expiry is too short (you may want a watch-dog mechanism).
// 2. Redis connectivity issues -> the lock cannot be acquired or released -> handle errors robustly.
// 3. If unlocking fails, there's a risk of holding the lock longer than intended -> consider alerting/monitoring.
func UpdateWebWithDistributedLock(ctx context.Context, code string, extra string) error {
	if rs == nil {
		return ErrNoRedsync
	}

	// Mutex key should be unique for each resource, here it’s "lock:web:{code}"
	mutex := rs.NewMutex(
		"lock:web:"+code,
		redsync.WithExpiry(10*time.Second),
		redsync.WithRetryDelay(200*time.Millisecond),
		redsync.WithTries(3), // example: try 3 times
	)

	// Acquire the lock
	if err := mutex.Lock(); err != nil {
		return fmt.Errorf("%w for code=%s: %v", ErrLockFailure, code, err)
	}
	defer func() {
		// Attempt to unlock
		if ok, err := mutex.Unlock(); !ok || err != nil {
			// If failed, log or handle carefully
			fmt.Printf("failed to unlock (web_code=%s): %v\n", code, err)
		}
	}()

	// Perform the MySQL update
	web := query.AnalyticsWeb
	result, err := web.WithContext(ctx).
		Where(web.WebCode.Eq(code)).
		Updates(map[string]interface{}{
			"extra": extra,
		})
	if err != nil {
		return fmt.Errorf("failed to update AnalyticsWeb with distributed lock: %w", err)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no records found with WebCode=%s", code)
	}

	return nil
}
