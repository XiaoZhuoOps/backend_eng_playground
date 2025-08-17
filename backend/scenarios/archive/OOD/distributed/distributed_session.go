package distributed

/*
Context
你正在开发一个基于 Golang 的分布式广告 pixel 日志流处理系统。每条日志都有一个唯一的 session_id，
同一个 session 内的所有日志都应归属于同一个广告平台（source_ad_platform）。
由于页面跳转或参数丢失，一个会话中只有第一条或前几条日志能准确计算出平台信息。

Objective
你需要设计一个基于 Redis 的分布式会話方案，用会话内第一条有效的平台信息来填充（enrich）
后续所有消息的平台信息，以确保数据一致性。

Initial Problem
旧方案中，为 session 设置一个固定的长 TTL（如 70 分钟），导致 Redis 内存压力较大。

Data Insights
- 50% 的 session 在 10 分钟内结束。
- 99% 的 session 在 70 分钟内结束。
- 一个 session 如果超过 30 分钟不活动（没有新日志），就会被视为结束。

New Requirements
1.  初始 TTL 设置为 40 分钟。
2.  当 session 的剩余 TTL 不足 30 分钟时，为其续期一次。
3.  续期时长为 30 分钟。
4.  每个 session 只允许续期一次，确保其最大总生命周期不超过 70 分钟（40分钟初始 + 30分钟续期）。
5.  方案必须只使用基本的 Redis 命令（如 GET, SET, SETNX, TTL），减少操作次数和内存占用。
*/

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

// --- Constants ---

const (
	// othersPlatform is a constant for non-informative platform values.
	othersPlatform = "others"
	// renewedMarker is the suffix used to indicate a session has been renewed.
	renewedMarker = ":1"
	// notRenewedMarker is the suffix for a session that has not been renewed.
	notRenewedMarker = ":0"
	// sessionKeyPrefix is the prefix for all session-related keys in Redis.
	sessionKeyPrefix = "session:"
)

// --- Configuration ---

// SessionManagerConfig holds the TTL configuration for session management.
type SessionManagerConfig struct {
	InitialTTL       time.Duration
	RenewalTTL       time.Duration
	RenewalThreshold time.Duration
}

// NewProductionConfig provides the default production configuration based on the requirements.
// Initial TTL: 40min. Renew by 30min when <=30min remain. Max total TTL: 70min.
func NewProductionConfig() SessionManagerConfig {
	return SessionManagerConfig{
		InitialTTL:       40 * time.Minute,
		RenewalTTL:       30 * time.Minute,
		RenewalThreshold: 30 * time.Minute,
	}
}

// --- Session Manager ---

// SessionManager handles the logic for enriching session data using Redis.
type SessionManager struct {
	rdb    *redis.Client
	config SessionManagerConfig
}

// NewSessionManager creates a new SessionManager with the given Redis client and configuration.
func NewSessionManager(rdb *redis.Client, config SessionManagerConfig) *SessionManager {
	if rdb == nil {
		panic("redis client cannot be nil")
	}
	return &SessionManager{
		rdb:    rdb,
		config: config,
	}
}

// EnrichSourceAdPlatform is the core function to get, set, or renew a session's ad platform.
// It uses a Redis Pipeline to fetch session value and TTL in a single network request.
func (s *SessionManager) EnrichSourceAdPlatform(ctx context.Context, sessionID, sourceAdPlatform string) (string, error) {
	key := sessionKeyPrefix + sessionID

	// --- 1. Atomically fetch GET and TTL using a Pipeline ---
	var storedValue string
	var remainingTTL time.Duration

	// Create a pipeline to send multiple commands in one go.
	pipe := s.rdb.Pipeline()
	getCmd := pipe.Get(ctx, key)
	ttlCmd := pipe.TTL(ctx, key)
	_, pipeErr := pipe.Exec(ctx) // Errors from individual commands are checked below.

	// --- 2. Check GET command result first ---
	storedValue, getErr := getCmd.Result()

	switch {
	case getErr == nil:
		// Session exists. The TTL command also executed.
		remainingTTL = ttlCmd.Val() // .Val() is safe as it returns zero value on error.
		return s.handleExistingSession(ctx, key, storedValue, remainingTTL)

	case errors.Is(getErr, redis.Nil):
		// Session does not exist. This is the first message for this session.
		return s.handleFirstMessage(ctx, key, sourceAdPlatform)

	default:
		// An unexpected Redis error occurred during GET or pipeline execution.
		// `pipeErr` would likely be non-nil here.
		if pipeErr != nil {
			return sourceAdPlatform, fmt.Errorf("redis pipeline error: %w", pipeErr)
		}
		return sourceAdPlatform, fmt.Errorf("redis GET error: %w", getErr)
	}
}

// handleExistingSession deals with a session that already has a value in Redis.
// It takes `remainingTTL` as an argument, avoiding another network call.
func (s *SessionManager) handleExistingSession(ctx context.Context, key, storedValue string, remainingTTL time.Duration) (string, error) {
	platform, renewed := parseStoredValue(storedValue)

	// If already renewed OR if TTL is not in the renewal window, just return.
	if renewed || remainingTTL <= 0 || remainingTTL > s.config.RenewalThreshold {
		return platform, nil
	}

	// --- Apply Renewal Policy ---
	newTTL := remainingTTL + s.config.RenewalTTL
	newValue := platform + renewedMarker

	// Atomically set the new value and new expiration in a single SET command.
	if err := s.rdb.Set(ctx, key, newValue, newTTL).Err(); err != nil {
		// If renewal fails, we still successfully served the request with the old data.
		// Log the error for monitoring.
		// log.Printf("failed to renew session for key %s: %v", key, err)
	}

	return platform, nil
}

// handleFirstMessage deals with the first message of a new session.
// It sets the initial value and TTL in Redis, handling potential race conditions.
func (s *SessionManager) handleFirstMessage(ctx context.Context, key, sourceAdPlatform string) (string, error) {
	// If the first message has a non-informative platform, do not store it.
	if sourceAdPlatform == othersPlatform {
		return sourceAdPlatform, nil
	}

	// Prepare the value with the "not renewed" marker.
	valueToStore := sourceAdPlatform + notRenewedMarker

	// --- Atomic SETNX ---
	// Use SETNX to ensure only the first request can set the initial value.
	wasSet, err := s.rdb.SetNX(ctx, key, valueToStore, s.config.InitialTTL).Result()
	if err != nil {
		return sourceAdPlatform, fmt.Errorf("redis SETNX error: %w", err)
	}

	// If SETNX succeeded, this request won the race. Return the current platform.
	if wasSet {
		return sourceAdPlatform, nil
	}

	// --- Handle Race Condition ---
	// If wasSet is false, another request set the value between our initial GET and this SETNX.
	// We simply need to get the value that the other request successfully set.
	finalValue, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		// If this final GET fails, something is wrong. Return the original platform as a fallback.
		return sourceAdPlatform, fmt.Errorf("redis GET after SETNX race failed: %w", err)
	}
	
	platform, _ := parseStoredValue(finalValue)
	return platform, nil
}

// parseStoredValue splits the value from Redis (e.g., "tiktok:0") into the platform and a boolean indicating renewal status.
func parseStoredValue(value string) (platform string, renewed bool) {
	if strings.HasSuffix(value, renewedMarker) {
		return strings.TrimSuffix(value, renewedMarker), true
	}
	if strings.HasSuffix(value, notRenewedMarker) {
		return strings.TrimSuffix(value, notRenewedMarker), false
	}
	// Fallback for values that don't match the format (e.g., from an older system).
	return value, false
}