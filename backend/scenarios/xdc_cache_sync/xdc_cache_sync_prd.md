# PRD: Cross-DC Cache Sync for ToC Service

## 1. Overview

### 1.1 Background

High-traffic ToC web service in Golang (Gin, GORM), peak QPS millions. Flow: Request -> Local Cache -> Redis -> MySQL.

Table in /backend/pkg/repo/sql/tracker_web.sql, Extra JSON for configs.

ToC reads only, in DC B (US).

ToB writes/updates MySQL, low QPS, in DC A (SG).

Updates sync via binlog A to B, but no cache invalidation, causing 30min inconsistency (Redis TTL).

### 1.2 Problem

30min stale data affects UX.

### 1.3 Objectives

- Inconsistency <5min.
- Maintain high QPS.
- Reliable cross-DC, no SPOF.

## 2. Requirements

### 2.1 Functional

- Active cache invalidation (local, Redis) in B on A updates.
- Cross-DC sync via binlog or pub/sub.
- Fallback to TTL.
- Metrics: sync latency, failures, cache hits.

### 2.2 Non-Functional

- Latency +<100ms, handle peaks.
- Scalable.
- 99.99% uptime.
- Secure transmission.
- Low cost, use existing infra (RocketMQ, Kafka, etc.).

## 3. Constraints and Context

- Local cache TTL 15min.
- Redis TTL 30min, clustered.
- Binlog sync ~200ms+ A to B.
- Minimize load.
- Real-time single updates via ToB.
- Available: RocketMQ, Kafka, Redis, FaaS, Spark, Flink, etc.
- Services distributed: ToB (A) dozens instances in DC A (SG), ToC (B) hundreds in DC B (US).
- Simulate distributed scenarios on single machine.

## 4. Proposed Solution

### 4.1 High-Level Design

- Capture updates via binlog in A.
- Publish to cross-DC queue (Kafka/RocketMQ).
- Consume in B, invalidate caches.
- Fallback TTL.
- Target <5min delay.

### 4.2 Components

- Producer: Binlog consumer in A (FaaS).
- Queue: Kafka/RocketMQ for events.
- Consumer: In ToC, evict Redis/local.
- Cache: Eviction logic for single updates.

### 4.3 Data Flow

1. ToB updates MySQL in A.
2. Binlog captured (~200ms), publish to queue.
3. Queue replicates to B.
4. Consumer evicts Redis/local keys.
5. Requests fetch fresh from MySQL on miss.

## 5. Implementation Plan

1. Design, select infra (1-2d).
2. Setup queue (2d).
3. Implement producer (3d).
4. Implement consumer (3d).
5. Testing (4d).
6. Deploy, monitor (2d).

## 6. Risks and Mitigations

- Queue delay >5min: Use fast queue, latency alerts.
- Redis load increase: Optimize evictions, monitor.
- Data loss: Reliable delivery with retries.

## 7. Success Metrics

- Delay <5min.
- Load increase <5%.
- Hit rate >90%.
- Handle real-time updates without degradation.
