# Storage Selection Guide V2.4

## 1. Usage Instructions

This guide delivers dense insights on storage systems for rapid, informed decisions.

- **Leaders**: Use `2. Cheat Sheet` for quick comparisons.
- **Engineers**: Follow `3. Decision Tree` for guided selection, then dive into sections.

## 2. Cheat Sheet Matrix

| Category | Capacity | QPS | Latency | Summary | Internal |
|----------|----------|-----|---------|---------|----------|
| **Relational (SQL)** | GB-10TB | 1K-100K | 1-100ms | Structured data with strong transactions. | `NDB` |
| **KV (In-Memory)** | GB-10TB | 100K-1M+ | <1ms | Ultra-low-latency in-memory dictionary. | `Redis` |
| **KV (On-Disk)** | TB-PB | 100K-10M+ | 1-10ms | Scalable persistent dictionary for throughput. | `FooKV`, `Abase` |
| **Document** | GB-TB | 1K-100K | 1-50ms | Flexible JSON storage for agile development. | - |
| **Wide-Column** | 10TB-PB+ | 10K-1M | 5-50ms | Sparse super-table for massive datasets. | `FooTable` |
| **Time-Series** | TB-PB | 10K-1M | 10ms-1s | Timestamp-optimized for metrics/monitoring. | - |
| **Search Engine** | TB-PB | 1K-100K | 50ms-1s | Full-text search and aggregation expert. | `ES` |
| **Object Storage** | PB-EB | N/A | 100ms-1s | Infinite blob storage for unstructured files. | - |
| **Graph** | GB-TB | 1K-50K | 10ms-1s | Relationship traversal for connected data. | `FooGraph` |

## 3. Decision Tree

Step-by-step guide to selection.

1. **Need strict multi-item ACID? (e.g., finance/orders)**
   - **Yes**: Relational (SQL). Internal: `NDB`. See §5.
   - **No**: Continue.

2. **Primary Pattern?**
   - **Key-Value Lookup**: KV Store (§6).
     - <1ms latency (cache/session): In-Memory (`Redis`).
     - Massive writes, eventual consistency: On-Disk (`Abase`).
     - Strong consistency: On-Disk (`FooKV`).
   - **Full-Text/Analytics**: Search Engine (§8). Internal: `ES`.
   - **Large Blobs (images/videos)**: Object Storage (§12).
   - **Relationships (social/fraud)**: Graph (§11). Internal: `FooGraph`.
   - **Timestamped Metrics (IoT/monitoring)**: Time-Series (§10).
   - **Sparse Massive Tables**: Wide-Column (§7). Internal: `FooTable`.
   - **JSON Objects**: Document (§9).

## 4. Core Concepts

### 4.1. Engines: B+Tree vs. LSM-Tree

- **B+Tree (Read-Optimized)**: Sorted pages for stable, fast reads; in-place updates cause splits/merges (write amplification). Used: `MySQL (InnoDB)`, `PostgreSQL`, `NDB`.
- **LSM-Tree (Write-Optimized)**: Buffer writes in MemTable, flush to immutable SSTables; background compaction. Fast writes, variable reads (read amplification). Used: `RocksDB`, `HBase`, `Cassandra`, `FooKV`, `Abase`.

### 4.2. Consistency: ACID vs. BASE

- **ACID**: Atomic, Consistent, Isolated, Durable; ensures correctness, trades performance/availability. Systems: Relational (`NDB`, MySQL, PostgreSQL).
- **BASE**: Basically Available, Soft State, Eventually Consistent; prioritizes availability/scalability over immediate consistency (CAP trade-off). Systems: `Abase`, `Cassandra`, `ES`.

### 4.3. Key Terminology

- **Write Amplification**: Ratio of data written to storage vs. user data (e.g., B+Tree splits or LSM compaction rewrite multiples of original data, increasing I/O and wear on SSDs).
- **Read Amplification**: Extra data scanned during reads (e.g., LSM checking multiple SSTables or B+Tree traversing branches).
- **Copy-on-Write (CoW)**: Technique to avoid in-place updates by copying data before modification; ensures crash consistency, used in filesystems like Btrfs or databases for snapshots (trades space for safety).
- **Compaction**: Background process merging/reorganizing data files to remove duplicates/deletions (e.g., LSM-Tree merges SSTables to reduce read amplification).
- **MemTable**: In-memory buffer for recent writes (e.g., in LSM-Tree, a sorted map flushed to disk when full).
- **SSTable (Sorted String Table)**: Immutable on-disk file in LSM-Tree, containing sorted key-value pairs for efficient lookups.
- **Inverted Index**: Mapping from terms to document locations (core to search engines for fast full-text queries).
- **Sharding**: Horizontal partitioning of data across nodes by key ranges to scale capacity/load (e.g., in distributed KV or relational databases).
- **Replication**: Duplicating data across nodes for availability/durability (e.g., master-slave or quorum-based).
- **Consensus**: Algorithm ensuring agreement among distributed nodes (e.g., Raft in `FooKV` or Paxos in Spanner for strong consistency).
- **CAP Theorem**: In distributed systems, you can have at most two of Consistency, Availability, Partition Tolerance; most scale-out systems prioritize AP (BASE) over CP (ACID).
- **WAL (Write-Ahead Logging)**: Logs operations before applying to main storage for durability and crash recovery (e.g., in databases like MySQL or Redis AOF).
- **CAS (Compare-And-Swap)**: Atomic operation that compares a value to an expected one and swaps if matched; used for optimistic concurrency control.
- **OLAP (Online Analytical Processing)**: Systems optimized for complex, read-heavy analytics queries on large datasets (e.g., data warehouses like Snowflake).
- **OLTP (Online Transaction Processing)**: Systems handling high-volume, low-latency transactions (e.g., relational databases for e-commerce).
- **MVCC (Multi-Version Concurrency Control)**: Manages concurrent access by keeping multiple versions of data; enables non-blocking reads (e.g., in PostgreSQL or MongoDB).
- **Bloom Filter**: Probabilistic structure to test set membership quickly, reducing unnecessary disk I/O (e.g., in LSM-Tree reads or Cassandra).
- **Erasure Coding**: Fault-tolerant encoding that reconstructs data from subsets (more space-efficient than replication; used in object storage like S3).
- **Snapshot**: Point-in-time, read-only copy of data (often via CoW; enables backups or versioning without blocking writes).
- **Hotspot**: Uneven data access or distribution causing performance bottlenecks (e.g., poor key design in sharded systems overloading nodes).
- **Quorum**: Minimum number of nodes required for a distributed operation to succeed (e.g., in replication for reads/writes to ensure consistency).
- **Journaling**: File system technique logging metadata or data changes for fast crash recovery (similar to WAL; used in ext4 or NTFS).
- **Partition Tolerance**: Ability to handle network partitions (part of CAP; most distributed systems are designed to be partition-tolerant).
- **Durability**: ACID property ensuring committed data survives crashes (achieved via WAL or replication).
- **Isolation Levels**: ACID controls for transaction visibility (e.g., Serializable prevents anomalies, Read Committed allows non-repeatable reads; tunable in SQL databases).

## 5. Relational (SQL)

- **Overview**: Structured tables with schemas, JOINs, ACID for integrity.
- **Uses**: Finance, e-commerce, ERPs.
- **Internal**: `NDB`.
- **Principles**: Declarative SQL; relational model (keys); ACID transactions.
- **Trade-offs**: Pros: Strong guarantees, mature ecosystem, queries. Cons: Hard scaling (sharding), rigid schema, poor for massive writes.
- **Benchmarks & Practices**:
  - **MySQL**: [Perf] 1K-100K QPS, GB-TB; [Impl] B+Tree (InnoDB), master-slave replication; [Note] Popular, easy; use pooling, replicas for reads; shard last-resort.
  - **PostgreSQL**: [Perf] Similar to MySQL, advanced features; [Impl] B+Tree, rich extensions (JSON/GIS); [Note] SQL-compliant; best for complex queries; Netflix uses for secondary indexes.
  - **Google Spanner**: [Perf] Global, 100K+ QPS, TB+; [Impl] Distributed, TrueTime clocks for ACID; [Note] High cost/complexity; Google uses for AdWords.
  - **Internal NDB**: [Perf] GB-10TB, 1K-100K QPS; [Impl] B+Tree, ACID; [Note] Reliable for critical data; scale via replicas.

## 6. Key-Value (KV)

- **Overview**: Simple Map<Key,Value>; fast Get/Put/Delete.
- **Uses**: Cache, sessions, metadata.
- **Internal**: `FooKV` (consistent), `Abase` (throughput), `Redis` (in-memory).
- **Principles**: Flexible values; scalability via hashing.
- **Trade-offs**: Pros: High perf/scale, flexible. Cons: Weak queries, key design critical.
- **Benchmarks & Practices**:
  - **Redis**: [Perf] 100K-1M+ QPS, GB-10TB; [Impl] In-memory, single-threaded, rich structures (lists/sets); [Note] Data loss on crash, eviction on full memory; Twitter/Netflix use for caching; enable AOF/RDB for persistence.
  - **Abase**: [Perf] 100K-10M+ QPS, TB-PB; [Impl] LSM-Tree (RocksDB-like), eventual consistency; [Note] Write-optimized, tolerate inconsistencies; for logs/IoT; design keys to avoid hotspots.
  - **FooKV**: [Perf] 100K-1M QPS, TB-PB; [Impl] LSM-Tree with Raft consensus; [Note] Strong consistency; for metadata/transactions; high reliability.
  - **Cassandra**: [Perf] 100K-1M+ QPS, TB-PB; [Impl] LSM-Tree, masterless ring; [Note] High availability; Facebook uses for inbox; tune quorum for consistency.
  - **DynamoDB**: [Perf] Auto-scales to 10M+ QPS, PB; [Impl] AWS-managed, eventual/strong modes; [Note] Pay-per-use; Amazon uses for shopping cart.
  - **TiKV**: [Perf] 100K-1M QPS, TB-PB; [Impl] RocksDB + Raft; [Note] Strong consistency; basis for TiDB.
  - **etcd**: [Perf] 10K QPS, GB-TB; [Impl] KV with Raft; [Note] For config/metadata; Kubernetes uses as control plane store.

## 7. Wide-Column

- **Overview**: Sparse table: Map<RowKey, Columns>; range scans.
- **Uses**: Logs, IoT, metrics.
- **Internal**: `FooTable`.
- **Principles**: LSM-Tree; dynamic columns in families.
- **Trade-offs**: Pros: PB-scale, high writes, flexible. Cons: No ACID/JOINs, RowKey critical.
- **Benchmarks & Practices**:
  - **FooTable**: [Perf] 10K-1M QPS, 10TB-PB+; [Impl] LSM-Tree; [Note] For massive logs; internal for event tracking.
  - **Google Bigtable**: [Perf] 10K-1M QPS, PB+; [Impl] LSM-Tree on Colossus; [Note] Google uses for Search/Maps; high durability.
  - **Apache HBase**: [Perf] 10K-1M QPS, TB-PB; [Impl] LSM-Tree on HDFS; [Note] Hadoop-integrated; compaction overhead; tune for write-heavy.
  - **Apache Cassandra**: [Perf] 100K-1M+ QPS, TB-PB; [Impl] LSM-Tree, decentralized; [Note] No single failure point; Netflix uses for user data; reverse timestamps in RowKey.
  - **ScyllaDB**: [Perf] 1M+ QPS, TB-PB; [Impl] C++ Cassandra rewrite; [Note] Lower latency; for high-perf needs.

## 8. Search Engines

- **Overview**: Inverted index for search/aggregation.
- **Uses**: Logs, BI, search.
- **Internal**: `ES`.
- **Principles**: Term -> DocID; real-time indexing.
- **Trade-offs**: Pros: Powerful search/aggs, scalable. Cons: Not primary store, high write cost, complex ops.
- **Benchmarks & Practices**:
  - **ES (Elasticsearch)**: [Perf] 1K-100K QPS, TB-PB; [Impl] Lucene-based, distributed shards; [Note] Eventual consistency; Uber uses for logs; use ILM for hot/warm tiers.
  - **OpenSearch**: [Perf] Similar to ES; [Impl] ES fork; [Note] AWS-managed option; open-source alternative.

## 9. Document

- **Overview**: Schema-free JSON/BSON docs.
- **Uses**: CMS, profiles, agile apps.
- **Principles**: Nested structures; object mapping.
- **Trade-offs**: Pros: Flexible, dev-friendly, queries. Cons: Weak consistency, duplication, poor JOINs.
- **Benchmarks & Practices**:
  - **MongoDB**: [Perf] 1K-100K QPS, GB-TB; [Impl] WiredTiger (B+Tree/LSM), sharding; [Note] Multi-doc txns in v4+; use validation; embed for perf.
  - **Azure Cosmos DB**: [Perf] Auto-scales, global; [Impl] Multi-model, RU-based; [Note] Mongo-compatible; Microsoft uses for Xbox.

## 10. Time-Series (TSDB)

- **Overview**: (Metric, Time, Tags, Value); compressed storage.
- **Uses**: Monitoring, IoT.
- **Principles**: Columnar, compression (Gorilla), aggregations.
- **Trade-offs**: Pros: High ingest/compression, fast queries. Cons: Specialized model.
- **Benchmarks & Practices**:
  - **InfluxDB**: [Perf] 10K-1M QPS, TB-PB; [Impl] TSM engine, push model; [Note] TICK stack; downsample for retention.
  - **Prometheus**: [Perf] 10K-1M samples/sec, TB; [Impl] Pull model, local TSDB; [Note] K8s standard; use federation for scale.

## 11. Graph

- **Overview**: Nodes/Edges with properties.
- **Uses**: Networks, recommendations.
- **Internal**: `FooGraph`.
- **Principles**: Fast traversals (e.g., Cypher).
- **Trade-offs**: Pros: Efficient relations. Cons: Niche, not for bulk.
- **Benchmarks & Practices**:
  - **FooGraph**: [Perf] 1K-50K QPS, GB-TB; [Impl] Graph-native; [Note] Internal for relations.
  - **Neo4j**: [Perf] 1K-50K QPS, GB-TB; [Impl] Native storage, Cypher; [Note] PayPal uses for fraud; limit depth for perf.

## 12. Object Storage

- **Overview**: Key/Data/Metadata; HTTP API.
- **Uses**: Blobs, data lakes.
- **Principles**: Erasure coding for durability.
- **Trade-offs**: Pros: Infinite scale, low cost. Cons: High latency, immutable objects.
- **Benchmarks & Practices**:
  - **Amazon S3**: [Perf] PB-EB, 100ms-1s; [Impl] Distributed, 11 9s durability; [Note] Eventual for overwrites; data lake base.
  - **MinIO**: [Perf] Similar, on-prem; [Impl] S3-compatible; [Note] Private deployments; integrate with Spark.

## 13. Common Pitfalls

Avoid these traps in storage design/usage:

- **General**: Ignoring monitoring for space/IO/QPS; leads to outages. Always set alerts for 80% utilization. Overlooking data modeling (e.g., poor key design) causes hotspots and uneven scaling.
- **Relational (SQL, e.g., MySQL/PostgreSQL/NDB)**: Missing/inappropriate indexes cause full table scans and high latency; always EXPLAIN queries. Overusing JOINs on large tables without indexes leads to exponential slowdowns. Neglecting transaction isolation levels can cause dirty reads or deadlocks.
- **KV (In-Memory, e.g., Redis)**: Large keys (e.g., huge lists/sets) slow operations and fragment memory; split them. Hot keys overload single threads/instances; use consistent hashing or client-side sharding. Forgetting persistence (AOF/RDB) risks data loss on crashes.
- **KV (On-Disk, e.g., Abase/FooKV/Cassandra)**: Poor key design creates hotspots (e.g., sequential keys all hit one shard); randomize or composite keys. Underestimating compaction overhead spikes CPU/IO during merges; tune compaction strategies. Relying on eventual consistency without app-level handling leads to stale reads.
- **Wide-Column (e.g., FooTable/HBase/Cassandra)**: Bad RowKey design (e.g., no salting) causes hotspots and imbalanced clusters. Writing too many columns per row bloats storage; limit to thousands. Ignoring tombstones (deleted data markers) accumulates during frequent deletes, inflating read amplification.
- **Search Engine (e.g., ES/OpenSearch)**: Over-indexing fields consumes excessive memory/CPU; index only what's queried. High-cardinality fields (e.g., unique IDs) explode index size; use doc_values sparingly. Not managing shard/replica counts leads to OOM or slow queries in large clusters.
- **Document (e.g., MongoDB)**: Allowing unbounded document growth (e.g., large arrays) hurts performance; cap sizes or use bucketing. Over-denormalizing data causes inconsistencies during updates; balance embedding vs. referencing. Ignoring schema validation in schema-free mode leads to data corruption over time.
- **Time-Series (e.g., InfluxDB/Prometheus)**: High-cardinality tags (e.g., unique user IDs) explode series count and memory; limit to low-cardinality dimensions. Not downsampling old data wastes storage; configure retention policies. Pull models (Prometheus) can overload scrape targets if intervals are too aggressive.
- **Graph (e.g., FooGraph/Neo4j)**: Deep traversals without limits cause query explosions (e.g., full graph scans); add depth caps. Dense nodes (millions of edges) slow traversals; partition or use supernodes. Treating it as a general DB wastes its strengths; use only for relation-heavy workloads.
- **Object Storage (e.g., S3/MinIO)**: Frequent small writes incur high costs (per-request fees); batch them. Assuming strong consistency on overwrites leads to surprises (eventual in S3); use versioning. Not setting lifecycle policies accumulates stale data, inflating bills.

## 14. official doc
