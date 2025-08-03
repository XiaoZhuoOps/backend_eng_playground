Solution 1: Binlog Listener + Kafka Pub/Sub

Tech: MySQL binlog (Canal/ Maxwell), Kafka for cross-DC replication, Golang consumer in ToC.
Logic Flow: ToB updates MySQL -> Binlog captured -> Publish to Kafka topic -> Consume in DC B -> Invalidate Redis/local caches.
Data Flow: Update event (key, op type) from A to B via Kafka; eviction uses Redis DEL, local cache invalidate.
Corner Cases: High update volume overwhelming queue (handle with partitioning); network partition (use Kafka retries, fallback TTL); inconsistent binlog parse (validate event schema).
Pros: Reliable delivery, scalable, low latency (~200ms + queue time); uses existing infra.
Cons: Kafka setup complexity for cross-DC; potential added load if not optimized.
