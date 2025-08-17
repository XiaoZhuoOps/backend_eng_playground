

# **A Junior Engineer's Guide to Message Queue Selection and Usage**

## **Part 1: The MQ Landscape at a Glance**

This initial section provides a high-level, comparative view of the message queue ecosystem. It is designed to serve as a quick-reference tool, enabling engineers to rapidly understand the key players, their core strengths, and the fundamental trade-offs involved. This part will help in forming a preliminary choice, which can be further validated and refined using the in-depth analyses in the subsequent sections.

### **Cheat Sheet Matrix: A Comparative Overview of Popular Message Queues**

The selection of a message queue (MQ) is a critical architectural decision with long-term implications for a system's performance, scalability, and reliability. The following matrix compares several leading MQ systems across essential attributes. It is designed not just to list features, but to provide the necessary context to understand the inherent trade-offs each system makes.1

One of the most significant columns in this table is the "Primary Messaging Model." This is not merely a feature but an architectural first principle. Systems built on a "Distributed Log" model are fundamentally designed as immutable, append-only logs. This structure is optimized for high-throughput sequential writes and reads, making message replay a natural and inherent capability because messages are not deleted upon consumption.2 This design is ideal for event streaming and data pipelines where historical data access is valuable.

Conversely, systems based on a "Traditional Broker" model treat the queue as a mutable data structure. They manage the state of individual messages, tracking acknowledgments and removing messages after they are successfully processed. This per-message management introduces overhead that can limit peak throughput compared to log-based systems. However, it enables more complex, broker-side logic, such as sophisticated message routing and individual message prioritization, making these systems exceptionally well-suited for task distribution and service decoupling in complex microservice architectures.3 Therefore, an engineer's first decision point should be to determine whether the application requires a replayable log for streaming data or a transient task queue for decoupling services. This single decision will substantially narrow the field of appropriate choices.

| Attribute | Apache Kafka | RabbitMQ | Apache RocketMQ | Apache Pulsar | AWS SQS |
| :---- | :---- | :---- | :---- | :---- | :---- |
| **Primary Messaging Model** | Distributed Log (Streaming) | Traditional Broker (Queuing) | Traditional Broker (Queuing) | Distributed Log (Streaming) | Traditional Broker (Queuing) |
| **Throughput** | Very High (600+ MB/s) 5 | Medium (30-50 MB/s) 5 | High (Benchmarks vary, but designed for high volume) 6 | Very High (300-600 MB/s) 5 | High (Standard queues have "nearly unlimited" throughput; FIFO up to 70,000 TPS) 9 |
| **Latency (p99)** | Low (\~5 ms at high throughput) 5 | Very Low (\~1 ms at low throughput) 5 | Low (Designed for financial-grade low latency) 11 | Low (\~10-25 ms at high throughput) 5 | Low to Medium (Typically single-digit to low double-digit ms) 14 |
| **Data Persistence** | File-based replicated log with configurable retention 16 | Durable queues (file-based); messages deleted on ack 16 | File-based storage with configurable retention 18 | Tiered Storage (Apache BookKeeper \+ S3/GCS) 16 | Managed, redundant storage (up to 14 days) 21 |
| **Delivery Guarantees** | At-least-once, Exactly-once 16 | At-most-once, At-least-once 16 | At-least-once, Exactly-once (for ordered messages) 23 | At-most-once, At-least-once, Exactly-once 16 | At-least-once (Standard), Exactly-once (FIFO) 25 |
| **Primary Use Case** | High-throughput event streaming, log aggregation, real-time analytics 26 | Complex task queues, microservice RPC, flexible message routing 1 | Financial-grade transactional messaging, e-commerce order systems 28 | Cloud-native, multi-tenant unified messaging & streaming 30 | Fully managed, serverless decoupling for AWS workloads 26 |
| **Key Differentiator** | Replayable, partitioned log with a vast ecosystem (Kafka Streams, Connect) 17 | Flexible routing via exchanges (direct, topic, fanout, headers) 1 | Financial-grade reliability with first-class support for ordered & transactional messages 23 | Built-in multi-tenancy & geo-replication with decoupled compute/storage architecture 16 | Fully managed, serverless, and deeply integrated into the AWS ecosystem 22 |

### **Decision Tree: Choosing Your Message Queue**

This flowchart provides a structured path to help select an appropriate message queue based on common project requirements. It begins with high-level architectural goals and progressively narrows the choices based on more specific technical needs.

Code snippet

graph TD  
    A \--\> B{Real-time data streaming, log aggregation, or event sourcing?};  
    A \--\> C{Decoupling services, running background tasks, or complex message routing?};  
    A \--\> D{Simple, maintenance-free queue for an application built on AWS?};

    B \-- No \--\> C;  
    B \-- Yes \--\> E{Is built-in multi-tenancy or geo-replication a critical requirement?};

    C \-- No \--\> F{Do you need financial-grade reliability and strict message ordering for transactions?};  
    C \-- Yes \--\> G;

    D \--\> H;

    E \-- Yes \--\> I;  
    E \-- No \--\> J;

    F \-- Yes \--\> K;  
    F \-- No \--\> L;

    subgraph Legend  
        direction LR  
        StartNode  
        DecisionNode{Decision}  
        ProcessNode  
    end

    style G fill:\#d4edda,stroke:\#c3e6cb,stroke-width:2px  
    style H fill:\#d4edda,stroke:\#c3e6cb,stroke-width:2px  
    style I fill:\#d4edda,stroke:\#c3e6cb,stroke-width:2px  
    style J fill:\#d4edda,stroke:\#c3e6cb,stroke-width:2px  
    style K fill:\#d4edda,stroke:\#c3e6cb,stroke-width:2px  
    style L fill:\#d4edda,stroke:\#c3e6cb,stroke-width:2px

## **Part 2: Foundational Knowledge**

Before diving into specific technologies, it is essential to build a solid understanding of the fundamental principles of message queuing. This section explains what message queues are, the critical problems they solve in modern software architecture, and the common terminology used across different MQ systems.

### **Core Concepts: What is a Message Queue and What Problems Does It Solve?**

At its core, a message queue is a software component that enables asynchronous communication between different parts of a system.35 It acts as an intermediary, providing a lightweight buffer that temporarily stores messages until the receiving application is ready to process them.35 This model is analogous to sending a text message instead of making a phone call; the sender can dispatch the message without needing the receiver to be immediately available, and the receiver can process it at their own convenience.37 This simple concept of asynchronous, brokered communication is the foundation for solving several complex problems in distributed systems.

The value of a message queue can be understood through three primary benefits it provides: service decoupling, asynchronous processing, and load smoothing.

* **Service Decoupling:** In a tightly coupled system, components communicate directly with each other. If one component fails or becomes slow, it can cause a cascading failure that brings down the entire system. Message queues solve this by decoupling the sender (producer) from the receiver (consumer).37 The producer and consumer only need to know about the message queue, not each other. They can be written in different languages, run on different servers, and operate independently.39 If a consumer service goes offline, the producer can continue to accept requests and place messages in the queue. The messages will be safely stored until the consumer comes back online and begins processing them, thus enhancing the system's overall resilience and fault tolerance.37  
* **Asynchronous Processing:** Many user-facing operations involve a quick initial action followed by several slower, background tasks. For instance, when a user places an order on an e-commerce site, the immediate response should be fast. However, the subsequent processes—charging the credit card, updating inventory, notifying the shipping department, and sending a confirmation email—can take several seconds.41 Performing these tasks synchronously would force the user to wait, creating a poor experience. With a message queue, the web server can instantly place an "OrderPlaced" message onto a queue and return a confirmation to the user. Separate, dedicated services can then consume this message asynchronously to perform the long-running tasks in the background, significantly improving system responsiveness and throughput.37  
* **Load Smoothing and Balancing:** Systems often face unpredictable traffic patterns, with sudden spikes that can overwhelm services. A message queue acts as a buffer, absorbing these bursts of activity.37 If an application receives thousands of requests in a short period, instead of trying to process them all at once and potentially crashing, it can enqueue them. Consumer services can then pull messages from the queue and process them at a steady, sustainable rate.37 This "load smoothing" prevents system overloads. Furthermore, this model naturally enables load balancing. By running multiple instances of a consumer service that all read from the same queue, the workload is automatically distributed among them. This is known as the "Competing Consumers" pattern, and it allows for easy horizontal scaling of processing capacity by simply adding more consumer instances.37

### **Key Terminology: The Language of Messaging**

The world of message queues has its own specific vocabulary. Understanding these terms is crucial for comprehending technical documentation and effectively communicating with other engineers.

* **Message:** The fundamental unit of data transmitted through the MQ system. It is typically a byte array containing the payload (the actual data, like a JSON object representing an order) and often includes a header with metadata, such as a timestamp or a routing key.44 It is the "letter" being sent through the system.  
* **Producer/Publisher:** The application component that creates messages and sends them to the message queue system.44 In an e-commerce system, the order service that creates an "OrderPlaced" message is a producer.  
* **Consumer/Subscriber:** The application component that connects to the queue, retrieves messages, and processes them.44 The shipping service that processes "OrderPlaced" messages is a consumer.  
* **Broker:** The central server or cluster of servers that forms the message queue system itself. The broker is responsible for receiving messages from producers, storing them, and delivering them to the appropriate consumers.2 It is the "post office" that manages the entire mail delivery process.  
* **Queue:** A data structure, typically operating on a First-In-First-Out (FIFO) basis, that stores messages until they are consumed.44 In a point-to-point messaging model, each message in a queue is delivered to and processed by exactly one consumer.45  
* **Topic:** A named logical channel to which producers publish messages. In a publish-subscribe model, a message sent to a topic is broadcast to all consumers that have subscribed to that topic.46 This allows for one-to-many communication.  
* **Publisher/Subscriber (Pub/Sub):** A messaging pattern where publishers are not programmed to send their messages to specific receivers. Instead, they publish messages to topics without knowledge of what subscribers there may be.50 Subscribers express interest in one or more topics and receive all messages published to those topics, without knowledge of the publishers.50 This creates a highly decoupled, many-to-many communication model.52  
* **Acknowledgment (Ack):** A signal sent from a consumer back to the broker to confirm that a message has been successfully received and processed.36 This is the core mechanism for ensuring reliable delivery. The broker will not consider a message fully delivered (and thus will not delete it from the queue) until it receives an acknowledgment. If a consumer fails before sending an ack, the broker will redeliver the message to another consumer.54  
* **Partition:** A fundamental concept in log-based systems like Kafka. A topic is divided into multiple partitions, and each partition is an ordered, immutable sequence of messages.55 Partitioning is the primary mechanism for achieving parallelism and scalability, as different partitions can be hosted on different brokers and consumed by different consumers simultaneously.57  
* **Consumer Group:** A concept central to Kafka and RocketMQ where multiple consumer instances are grouped together under a single logical identity (the group ID).55 The broker distributes the partitions of a topic among the members of a consumer group, ensuring that each partition is consumed by only one member at a time. This allows a group of consumers to work together to process all messages from a topic in a parallel, load-balanced fashion.60

It is important to note that the term "Topic" has different meanings in different systems, which can be a significant point of confusion. In Kafka, a topic is a physical storage abstraction—a partitioned log where data is durably stored.55 A single Kafka topic can support a queue-like pattern for one consumer group and a pub/sub pattern for multiple, independent consumer groups. In contrast, RabbitMQ does not have a "Topic" as a storage entity. Instead, it has a "Topic Exchange," which is a routing mechanism that uses pattern matching to deliver messages to one or more queues.62 In RabbitMQ, the "topic" is a concept within the routing logic, not the storage layer. Understanding this distinction is key to grasping the architectural differences between these systems.

## **Part 3: In-Depth Product Analysis**

This section provides a detailed examination of three major message queue systems: Apache Kafka, RabbitMQ, and Apache RocketMQ. For each system, the analysis covers its core architecture, typical use cases, essential configuration parameters, and common pitfalls that junior engineers may encounter. This deep dive is designed to equip engineers with the practical knowledge needed to make a final, informed decision and to implement their chosen solution effectively.

### **Apache Kafka: The Distributed Streaming Platform**

Apache Kafka is an open-source distributed event streaming platform, originally developed at LinkedIn and now one of the most active projects of the Apache Software Foundation.5 It is designed to handle high-throughput, real-time data feeds and is fundamentally architected as a distributed, replicated, and persistent commit log.5

#### **Overview**

* **Key Characteristics:** Kafka's architecture is built around an append-only, immutable log structure, which provides exceptional throughput for both writing and reading data streams.36 It employs a pull-based consumption model, where consumers are responsible for tracking their own progress (known as an "offset") through the log.55 This design gives consumers significant control over their consumption patterns. Fault tolerance and high durability are achieved through the replication of data across multiple broker nodes in the cluster.56  
* **Pros:** The primary advantage of Kafka is its extreme performance and scalability for streaming workloads, capable of handling millions of messages per second.5 Its log-based architecture enables powerful features like message replay, allowing multiple applications to consume the same data stream independently and at different times.26 The Kafka ecosystem is vast and mature, including powerful libraries like Kafka Streams for stream processing and Kafka Connect for integrating with hundreds of external data systems.17  
* **Cons:** Kafka's power comes at the cost of higher operational complexity. Managing a Kafka cluster, which historically required an external Zookeeper cluster for coordination (though newer versions are migrating to a self-managed KRaft protocol), is more involved than managing a single-node RabbitMQ instance.4 Its routing capabilities are simpler than those of traditional brokers; complex routing logic must be implemented in the consumer applications. While excellent for high-throughput scenarios, its latency at lower message rates can be higher than systems like RabbitMQ.5  
* **Typical Use Cases:** Kafka excels in scenarios that involve processing continuous streams of data. Common use cases include real-time analytics dashboards, event sourcing architectures where every state change is captured as an event, aggregating logs from thousands of microservices, and implementing Change Data Capture (CDC) pipelines that stream changes from databases to other systems.26

#### **Underlying Architecture**

The core of Kafka's design is the distributed commit log. A Kafka topic is not a traditional queue; it is a log of records that is partitioned and replicated across multiple servers, or "brokers".3

* **The Log is Everything:** When a message (or "event") is published to a topic, it is appended to the end of one of its partitions. Each message within a partition is assigned a unique, sequential ID called an offset.55 Unlike a traditional queue, messages are not deleted when a consumer reads them. Instead, they are retained based on a configurable policy (e.g., for 7 days or until the topic reaches a certain size).56 This retention is what makes message replay possible; a consumer can "rewind" and re-process messages from any offset it chooses.  
* **Partitions and Consumer Groups:** These two concepts are the key to Kafka's scalability and are often the most misunderstood by newcomers.  
  * A topic is divided into one or more **partitions**. The partition is the fundamental unit of parallelism in Kafka. Each partition is an independent, ordered log that can be placed on a different broker in the cluster.55  
  * A **consumer group** is a set of consumer application instances that share a common group.id and work together to consume a topic.55  
  * The Kafka broker assigns each partition of a topic to **exactly one** consumer instance within a given consumer group at any point in time.67 This strict one-to-one mapping between a partition and a consumer (within a group) is how Kafka ensures that messages within a partition are processed in order while allowing the overall workload to be parallelized across the group.  
  * The relationship between the number of partitions and the number of consumers in a group is critical. If a topic has N partitions, a consumer group can have at most N active consumers. If more than N consumers are added to the group, the excess consumers will remain idle, as there are no partitions left to assign to them.67 Conversely, if there are fewer consumers than partitions, some consumers will be assigned multiple partitions.68

#### **Configuration and Tuning**

Proper configuration is vital for a stable and performant Kafka deployment. Below are some of the most critical parameters for producers, topics, and consumers.

* **Producer Configuration:**  
  * bootstrap.servers: A comma-separated list of broker host:port pairs that the producer will use for its initial connection to the cluster.70  
  * acks: This setting controls the durability of produced messages.  
    * acks=0: The producer does not wait for any acknowledgment from the broker (fire-and-forget). This offers the lowest latency but risks message loss.72  
    * acks=1: The producer waits for an acknowledgment from the leader broker of the partition. This is the default and provides a good balance of durability and performance.73  
    * acks=all: The producer waits for an acknowledgment from the leader and all in-sync replicas. This provides the strongest durability guarantee but at the cost of higher latency.73  
  * retries and enable.idempotence=true: Setting idempotence to true ensures that the producer will not introduce duplicate messages in the event of retries due to network errors. This is a cornerstone of achieving exactly-once processing semantics.60  
* **Topic Configuration:**  
  * partitions: The number of partitions for the topic. This should be chosen based on the desired level of consumer parallelism. It's often better to over-partition slightly than to under-partition, as adding partitions later can be complex.66  
  * replication.factor: The number of copies of each partition to maintain. A replication factor of 3 is standard for production environments to ensure high availability and fault tolerance.56  
* **Consumer Configuration:**  
  * group.id: The unique string that identifies the consumer group this consumer belongs to.71  
  * auto.offset.reset: This defines what the consumer should do when there is no initial offset stored for its group (e.g., when the group is new).  
    * earliest: The consumer will start reading from the very beginning of the partition.73  
    * latest: The consumer will start reading only new messages that are produced after it subscribes. This is the default.73 This setting is a frequent source of confusion and unexpected behavior for junior engineers.  
  * enable.auto.commit: If true, the consumer's offset is automatically committed in the background at a regular interval. For more reliable processing, it is often recommended to set this to false and commit offsets manually after a message has been successfully processed.72

#### **Common Pitfalls**

* **Partition Count vs. Consumer Count Mismatch:** A common scaling mistake is to add more consumer instances to a group than there are partitions in the topic. As explained in the architecture section, these extra consumers will sit idle.55 To increase parallelism beyond the number of partitions, the topic itself must be re-partitioned, which can be a complex and disruptive operation that affects message ordering guarantees if keys are used.  
* **Improper Offset Management and Message Loss:** The most subtle and dangerous pitfall for new users is related to offset committing. If enable.auto.commit is true, the client library commits the last received offset periodically. It is possible for a consumer to fetch a batch of messages, the client to commit the offset for that batch, and then the application to crash before it has finished processing all the messages in the batch. Upon restart, the consumer will begin from the *committed* offset, and the unprocessed messages from the previous batch will be lost forever. Manually committing the offset after processing is complete provides much stronger at-least-once delivery guarantees.72  
* **Non-Idempotent Consumers:** Kafka provides at-least-once delivery semantics by default. This means that under certain failure scenarios (like a consumer crashing after processing a message but before committing its offset, or during a consumer group rebalance), a message may be delivered more than once. The consumer's processing logic *must* be designed to be idempotent—meaning that processing the same message multiple times has the same effect as processing it once. For example, instead of using a simple INSERT statement into a database, one might use an UPSERT (update or insert) operation based on a unique message ID.74  
* **Ignoring Key-based Ordering:** Junior engineers often assume Kafka provides global message ordering, which it does not. Kafka only guarantees strict FIFO ordering *within a single partition*.26 If related events must be processed in order (e.g., all updates for a specific customer account), they must all be produced with the same message key (e.g., the  
  customer\_id). The producer's partitioner will use a hash of this key to ensure that all messages for that key are always sent to the same partition, thus preserving their order.60

#### **Resources**

For further study, the official Apache Kafka documentation is the most comprehensive and authoritative resource.

* **Official Documentation:** [https://kafka.apache.org/documentation/](https://kafka.apache.org/documentation/) 56

### **RabbitMQ: The Versatile Message Broker**

RabbitMQ is one of the most popular and mature open-source message brokers. It is an implementation of the Advanced Message Queuing Protocol (AMQP) 0-9-1 and is known for its reliability, flexibility, and ease of use.62 It follows a "smart broker, dumb consumer" philosophy, where the broker contains rich routing logic, allowing consumers to be simple and focused on business logic.1

#### **Overview**

* **Key Characteristics:** RabbitMQ's core strength lies in its highly flexible and powerful message routing capabilities, which are managed by a component called an "exchange".1 It operates on a push-based model, where the broker actively pushes messages to registered consumers.17 It supports a wide range of messaging patterns, including point-to-point, publish/subscribe, and request/reply (RPC).1  
* **Pros:** RabbitMQ is mature, highly reliable, and relatively easy to set up for simple use cases. Its powerful routing features make it ideal for complex microservice architectures where messages need to be delivered to different queues based on their content or metadata.63 It also provides a comprehensive management UI that is excellent for monitoring and administration.62  
* **Cons:** Compared to log-based systems like Kafka, RabbitMQ has lower peak throughput.5 It is not designed for message replay, as messages are typically deleted from queues once they are acknowledged by a consumer.16 If the routing logic becomes overly complex or queues become very long, the broker itself can become a performance bottleneck.  
* **Typical Use Cases:** RabbitMQ is exceptionally well-suited for background job processing, where a web application offloads long-running tasks (like generating a report or sending an email) to a pool of worker processes.26 It is also widely used for inter-service communication in microservice architectures and for implementing remote procedure call (RPC) patterns where a service needs to request an action from another and receive a response asynchronously.77

#### **Underlying Architecture**

RabbitMQ's architecture is defined by the AMQP 0-9-1 model, which describes a specific flow of messages from producer to consumer via the broker.

* **The AMQP Model Flow:** The message lifecycle in RabbitMQ is distinct from Kafka's. A producer does not send a message directly to a queue. Instead, the flow is as follows: **Producer \-\> Exchange \-\> Binding \-\> Queue \-\> Consumer**.62 The producer publishes a message to an exchange. The exchange then uses rules defined by bindings to route the message to one or more queues. Finally, consumers subscribed to those queues receive the message.  
* **Exchanges are the Routers:** The exchange is the heart of RabbitMQ's routing logic. It is responsible for receiving messages from producers and deciding which queues they should be delivered to. There are four primary exchange types, and understanding them is key to leveraging RabbitMQ's power 63:  
  * **Direct Exchange:** Routes a message to queues whose binding key exactly matches the message's routing key. This is useful for unicast, point-to-point messaging.  
  * **Fanout Exchange:** Ignores the routing key and broadcasts an incoming message to *all* queues that are bound to it. This is the simplest way to implement a publish/subscribe pattern.  
  * **Topic Exchange:** Routes messages to queues based on a wildcard match between the message's routing key and the pattern used in the queue binding. For example, a routing key of logs.error.auth could match binding patterns like logs.error.\* or logs.\#. This allows for flexible and powerful multicast routing.  
  * **Headers Exchange:** Ignores the routing key entirely and uses the attributes in the message header for routing. This is the most flexible but also the most complex routing mechanism.

#### **Configuration and Tuning**

Configuring RabbitMQ involves defining the properties of queues and tuning the behavior of consumers to match the application's workload.

* **Queue Properties:**  
  * durable: If true, the queue's metadata will survive a broker restart. For messages to also survive a restart, they must be published as "persistent" and sent to a durable queue.63  
  * auto-delete: If true, the queue is automatically deleted when its last consumer unsubscribes. This is useful for temporary queues.63  
  * max-length: Sets a limit on the number of messages a queue can hold. When the limit is reached, older messages are dropped from the head of the queue. This is a crucial setting to prevent unbounded queue growth.80  
* **Consumer Tuning:**  
  * prefetch\_count: This is a critical performance and load-balancing parameter. It defines the maximum number of unacknowledged messages that the broker will send to a consumer at one time. A low prefetch count (e.g., 1\) ensures that messages are distributed fairly among multiple consumers, preventing one fast consumer from hoarding all the messages while others sit idle. A higher prefetch count can improve throughput for a single consumer but may lead to unfair load distribution.81  
* **Best Practices:** For optimal performance, applications should use long-lived connections and create new channels for tasks as needed, rather than opening and closing connections frequently.81 Queues should be kept as short as possible. If long queues are an expected part of the workload, "lazy queues" can be used, which move messages to disk more aggressively to reduce RAM usage, albeit at the cost of throughput.80

#### **Common Pitfalls**

* **Improper Acknowledgment Handling:** This is a very common mistake for junior developers. RabbitMQ relies on acknowledgments to guarantee delivery. If a consumer uses "auto-ack" mode or simply forgets to send a manual ack after processing, a message can be lost if the consumer crashes before processing is complete. Conversely, if a consumer crashes *after* processing but *before* sending the ack, the broker will redeliver the message to another consumer upon reconnection, leading to duplicate processing. This necessitates that consumer logic be idempotent.83  
* **Unbounded Queues:** A slow or dead consumer is a silent killer in RabbitMQ. If producers are sending messages to a queue faster than consumers can process them, the queue will grow indefinitely. This can exhaust the broker's memory and disk space, eventually causing it to crash and become unresponsive. It is vital to always set a max-length policy on queues or have robust monitoring and alerting in place to detect growing queues.80  
* **Inefficient Connection and Channel Usage:** The TCP and AMQP handshake process for establishing a new connection is resource-intensive.81 A naive implementation might open a new connection for every message it sends, which will lead to terrible performance. The best practice is to establish a long-lived connection when the application starts and then open and close lightweight channels on that connection as needed. A common pattern is one connection per process, and one channel per thread.81  
* **Sharing Channels Between Threads:** Most RabbitMQ client libraries are not thread-safe at the channel level. Sharing a single channel object across multiple threads in an application will result in interleaved AMQP frames being sent to the broker, which will cause protocol errors and unpredictable connection closures. Each thread that needs to interact with RabbitMQ should have its own channel.81

#### **Resources**

The official RabbitMQ documentation and tutorials are excellent resources for getting started and diving deeper into its features.

* **Official Website:** [https://www.rabbitmq.com/](https://www.rabbitmq.com/) 85  
* **Client Libraries:** [https://www.rabbitmq.com/client-libraries](https://www.rabbitmq.com/client-libraries) 86

### **Apache RocketMQ: The Financial-Grade Messaging Platform**

Apache RocketMQ is a distributed messaging and streaming platform developed by Alibaba Group and later donated to the Apache Software Foundation.59 It was engineered to handle the massive scale and stringent reliability requirements of Alibaba's e-commerce and financial transaction systems. As such, its design prioritizes high performance, reliability, and first-class support for ordered and transactional messaging.28

#### **Overview**

* **Key Characteristics:** RocketMQ is known for its financial-grade reliability. It provides robust support for several advanced message types, including strictly ordered messages (FIFO), transactional messages that ensure consistency between message sending and local database operations, and scheduled/delayed messages.18  
* **Pros:** The platform's standout features are its excellent implementation of ordered and transactional messaging, which are often complex to achieve in other systems.59 It delivers high performance and is designed for massive scalability without external dependencies like Zookeeper.29  
* **Cons:** Compared to Kafka and RabbitMQ, RocketMQ has a smaller international community and a less extensive ecosystem of third-party tools and integrations. Its documentation, while improving, has historically been more accessible to Chinese-speaking developers, which can present a learning curve for others.  
* **Typical Use Cases:** RocketMQ is an ideal choice for systems where message order and transactional integrity are paramount. This includes e-commerce platforms for processing order pipelines (create, pay, ship, complete), financial systems for handling transactions, and any business workflow that requires a guaranteed sequence of events for a specific entity (e.g., all operations for a single user account).28

#### **Underlying Architecture**

RocketMQ features a decoupled architecture that is designed for high availability and easy scalability.

* **NameServer and Broker:** The architecture consists of four main components: Producers, Consumers, Brokers, and the NameServer.59  
  * **Broker:** The core component responsible for storing messages and handling requests from producers and consumers.  
  * **NameServer:** This is a lightweight service discovery component. Brokers register their metadata (address, topic information) with the NameServer. Producers and consumers then query the NameServer to discover which brokers host the topics they need to interact with.87 This design decouples the clients from the specific broker addresses and eliminates the need for a heavy coordination service like Zookeeper, simplifying operations.  
* **Consumption Models:** RocketMQ supports two primary consumption models for a consumer group 89:  
  * **Clustering:** This is the default and most common model. All consumer instances within a consumer group work together to consume the messages from the subscribed topic's queues. The load is distributed among the consumers, similar to a Kafka consumer group. Each message is processed by only one consumer in the group.  
  * **Broadcasting:** In this model, every message is delivered to *every* consumer instance within the consumer group. This is analogous to a classic publish/subscribe fanout pattern and is used when all consumers need to receive a copy of each message.

#### **Configuration and Tuning**

Tuning RocketMQ involves configuring the broker's storage behavior, the consumer's parallelism, and the underlying JVM and OS, which can have a significant impact on performance.

* **Broker Configuration:**  
  * flushDiskType: This critical parameter determines the message persistence strategy. ASYNC\_FLUSH provides higher throughput by flushing messages to disk in the background, while SYNC\_FLUSH provides maximum durability by ensuring each message is written to disk before an acknowledgment is sent to the producer. The latter is essential for financial-grade applications.90  
  * fileReservedTime: Sets the message retention period in hours, after which message files will be deleted.19  
* **Consumer Configuration:**  
  * consumeThreadMin / consumeThreadMax: These parameters control the size of the thread pool for a push consumer, allowing for fine-grained tuning of its concurrent processing capacity.90  
  * AllocateMessageQueueStrategy: Defines the algorithm for distributing message queues among the consumers in a group. The default is an average allocation strategy.89  
* **JVM and OS Tuning:** RocketMQ's performance is highly dependent on the underlying system configuration. The official documentation recommends specific JVM settings, such as using the G1 garbage collector and pre-allocating the heap. At the OS level, it is crucial to increase limits for open file descriptors and vm.max\_map\_count, as RocketMQ uses memory-mapped files extensively for message storage.92

#### **Common Pitfalls**

* **Inconsistent Subscription Relationships:** This is arguably the most critical and common pitfall in RocketMQ. All consumer instances belonging to the same consumer group **must** subscribe to the exact same set of topics and tags. If one consumer in a group subscribes to TopicA with Tag1, while another in the same group subscribes to TopicA with Tag2, the broker's load balancing and message delivery will become unpredictable, often resulting in messages not being delivered to the intended consumers or being lost entirely. This consistency must be strictly enforced at the code level across all consumer deployments.93  
* **Neglecting Consumer Idempotency:** Like other systems that provide at-least-once delivery guarantees, RocketMQ may redeliver a message in the event of a consumer failure or network issue. The business logic within the consumer must be designed to be idempotent to handle duplicate messages gracefully and prevent data corruption. The official documentation explicitly highlights this requirement.91  
* **Misunderstanding Ordered Messages:** RocketMQ's ordered message feature does not provide global ordering for all messages in a topic. It guarantees strict FIFO order only for messages that share the same "sharding key" and are sent to the same message queue.23 This is conceptually similar to using a partition key in Kafka. To process all events for a specific order in sequence, for example, all messages must be produced with the same order ID as the sharding key.  
* **Improper Retry and Dead-Letter Queue (DLQ) Handling:** When a consumer fails to process a message, RocketMQ will automatically retry delivery. If the failure is due to a transient issue (e.g., a temporary database deadlock), this is helpful. However, if the failure is caused by a persistent problem (e.g., a "poison pill" message with malformed data), the message will be retried repeatedly up to a configured maximum, after which it will be sent to a Dead-Letter Queue (DLQ). A poorly designed retry strategy can block the consumption of subsequent messages in an ordered queue or waste significant system resources. It is essential to have a robust strategy for monitoring and handling messages that land in the DLQ.95

#### **Resources**

The official Apache RocketMQ documentation provides concepts, quick-start guides, and best practices.

* **Official Website:** [https://rocketmq.apache.org/](https://rocketmq.apache.org/) 28  
* **Official Documentation:** [https://rocketmq.apache.org/docs/introduction/02concepts/](https://rocketmq.apache.org/docs/introduction/02concepts/) 59

## **Conclusions**

The selection and effective use of a message queue are pivotal decisions in the architecture of modern distributed systems. The analysis reveals that there is no single "best" message queue; instead, the optimal choice is deeply contingent on the specific requirements of the application, including its scalability needs, latency tolerance, data durability guarantees, and operational model.

The primary architectural distinction lies between **log-based streaming platforms** and **traditional message brokers**.

* **Apache Kafka** and **Apache Pulsar** represent the log-based model. They are engineered for extremely high throughput and are the superior choice for use cases centered around event streaming, real-time data pipelines, and log aggregation. Their ability to retain and replay messages makes them powerful tools for event sourcing and for serving multiple, independent downstream applications from a single data stream. The trade-off is generally higher operational complexity and potentially higher latency at low message volumes.  
* **RabbitMQ**, **Apache RocketMQ**, and **AWS SQS** embody the traditional broker model. They excel in scenarios requiring service decoupling, asynchronous task processing, and complex message routing. **RabbitMQ** stands out for its flexible routing capabilities, making it ideal for intricate microservice communication patterns. **RocketMQ** provides exceptional reliability for transactional and strictly ordered messaging, making it a strong contender for financial and e-commerce systems. **AWS SQS** offers a compelling serverless, fully-managed solution for applications built within the AWS ecosystem, eliminating operational overhead at the cost of vendor lock-in.

For a junior engineer, the path to successful implementation involves not only choosing the right tool but also understanding its underlying architecture and common failure modes. Key takeaways include:

1. **Understand the Model:** First, determine if the problem requires a replayable log or a transient task queue. This decision will guide the initial selection more than any other factor.  
2. **Embrace Idempotency:** All systems that guarantee at-least-once delivery can and will send duplicate messages under failure conditions. Consumer logic must be idempotent to prevent data corruption.  
3. **Respect the Scaling Unit:** In Kafka, the partition is the unit of parallelism. In RabbitMQ and RocketMQ, parallelism is often achieved by adding more consumers to a queue. Understanding and configuring this relationship correctly is fundamental to scaling the application.  
4. **Manage Resources:** Unbounded queues and improper connection management are common pitfalls that can bring down a broker. Proactive configuration of queue limits, consumer prefetch, and connection pooling is essential for stability.

By internalizing these foundational concepts and the specific architectural nuances of each system, a junior engineer can move beyond simply using a message queue as a black box and begin to leverage it as a powerful tool for building robust, scalable, and resilient distributed applications.

#### **Works cited**

1. Apache Kafka vs RabbitMQ vs AWS SNS/SQS: Which message broker is best?, accessed July 21, 2025, [https://ably.com/topic/apache-kafka-vs-rabbitmq-vs-aws-sns-sqs](https://ably.com/topic/apache-kafka-vs-rabbitmq-vs-aws-sns-sqs)  
2. What is a Message Broker? | VMware, accessed July 21, 2025, [https://www.vmware.com/topics/message-brokers](https://www.vmware.com/topics/message-brokers)  
3. Message Brokers: Queue-based vs Log-based \- DEV Community, accessed July 21, 2025, [https://dev.to/oleg\_potapov/message-brokers-queue-based-vs-log-based-2f21](https://dev.to/oleg_potapov/message-brokers-queue-based-vs-log-based-2f21)  
4. RabbitMQ vs Kafka \- Difference Between Message Queue Systems \- AWS, accessed July 21, 2025, [https://aws.amazon.com/compare/the-difference-between-rabbitmq-and-kafka/](https://aws.amazon.com/compare/the-difference-between-rabbitmq-and-kafka/)  
5. Benchmarking RabbitMQ vs Kafka vs Pulsar Performance \- Confluent, accessed July 21, 2025, [https://www.confluent.io/blog/kafka-fastest-messaging-system/](https://www.confluent.io/blog/kafka-fastest-messaging-system/)  
6. Testing Performance of Basic Edition RocketMQ 5.x Instances, accessed July 21, 2025, [https://support.huaweicloud.com/intl/en-us/usermanual-hrm/hrm\_ug\_065.html](https://support.huaweicloud.com/intl/en-us/usermanual-hrm/hrm_ug_065.html)  
7. Apache RocketMQ benchmarks \- OpenMessaging, accessed July 21, 2025, [https://openmessaging.cloud/docs/benchmarks/rocketmq/](https://openmessaging.cloud/docs/benchmarks/rocketmq/)  
8. Comparing Apache Pulsar vs. Apache Kafka | 2022 Benchmark Report \- StreamNative, accessed July 21, 2025, [https://streamnative.io/blog/apache-pulsar-vs-apache-kafka-2022-benchmark](https://streamnative.io/blog/apache-pulsar-vs-apache-kafka-2022-benchmark)  
9. Amazon SQS Features | Message Queuing Service \- AWS, accessed July 21, 2025, [https://aws.amazon.com/sqs/features/](https://aws.amazon.com/sqs/features/)  
10. Benchmarking Message Queue Latency \- Brave New Geek, accessed July 21, 2025, [https://bravenewgeek.com/benchmarking-message-queue-latency/](https://bravenewgeek.com/benchmarking-message-queue-latency/)  
11. Benchmarking Message Queues \- NSF Public Access Repository, accessed July 21, 2025, [https://par.nsf.gov/servlets/purl/10438973](https://par.nsf.gov/servlets/purl/10438973)  
12. The network delay between master and slave has a great performance impact · Issue \#3667 · apache/rocketmq \- GitHub, accessed July 21, 2025, [https://github.com/apache/rocketmq/issues/3667](https://github.com/apache/rocketmq/issues/3667)  
13. Benchmarking Pulsar and Kafka \- The Full Benchmark Report \- 2020 \- StreamNative, accessed July 21, 2025, [https://streamnative.io/blog/benchmarking-pulsar-and-kafka-report-2020](https://streamnative.io/blog/benchmarking-pulsar-and-kafka-report-2020)  
14. Performance of message brokers in event-driven architecture: Amazon SNS/SQS vs Apache Kafka Prestanda av meddelandeköer i händ \- DiVA, accessed July 21, 2025, [https://kth.diva-portal.org/smash/get/diva2:1801007/FULLTEXT01.pdf](https://kth.diva-portal.org/smash/get/diva2:1801007/FULLTEXT01.pdf)  
15. Optimizing Amazon Simple Queue Service (SQS) for speed and scale | AWS News Blog, accessed July 21, 2025, [https://aws.amazon.com/blogs/aws/optimizing-amazon-simple-queue-service-sqs-for-speed-and-scale/](https://aws.amazon.com/blogs/aws/optimizing-amazon-simple-queue-service-sqs-for-speed-and-scale/)  
16. Compare NATS | NATS Docs, accessed July 21, 2025, [https://docs.nats.io/nats-concepts/overview/compare-nats](https://docs.nats.io/nats-concepts/overview/compare-nats)  
17. A Comprehensive Comparison of ActiveMQ, Kafka, RabbitMQ, and Pulsar \- GTCSYS, accessed July 21, 2025, [https://gtcsys.com/a-comprehensive-comparison-of-activemq-kafka-rabbitmq-and-pulsar/](https://gtcsys.com/a-comprehensive-comparison-of-activemq-kafka-rabbitmq-and-pulsar/)  
18. Message \- Apache RocketMQ, accessed July 21, 2025, [https://rocketmq.apache.org/docs/domainModel/04message/](https://rocketmq.apache.org/docs/domainModel/04message/)  
19. Message Storage and Cleanup \- Apache RocketMQ, accessed July 21, 2025, [https://rocketmq.apache.org/docs/featureBehavior/11messagestorepolicy/](https://rocketmq.apache.org/docs/featureBehavior/11messagestorepolicy/)  
20. What is Apache Pulsar? | Datastax, accessed July 21, 2025, [https://www.datastax.com/blog/what-apache-pulsar](https://www.datastax.com/blog/what-apache-pulsar)  
21. Complete Guide to AWS SQS: Pros/Cons, Getting Started and Pro Tips \- Lumigo, accessed July 21, 2025, [https://lumigo.io/aws-sqs/](https://lumigo.io/aws-sqs/)  
22. Fully Managed Message Queuing – Amazon Simple Queue Service ..., accessed July 21, 2025, [https://aws.amazon.com/sqs/](https://aws.amazon.com/sqs/)  
23. Message Queue for Apache RocketMQ: Ordered Messages \- Alibaba Cloud Community, accessed July 21, 2025, [https://www.alibabacloud.com/blog/message-queue-for-apache-rocketmq-ordered-messages\_599727](https://www.alibabacloud.com/blog/message-queue-for-apache-rocketmq-ordered-messages_599727)  
24. Why transactions? \- Apache Pulsar, accessed July 21, 2025, [https://pulsar.apache.org/docs/next/txn-why/](https://pulsar.apache.org/docs/next/txn-why/)  
25. Amazon SQS FAQs | Message Queuing Service, accessed July 21, 2025, [https://www.amazonaws.cn/en/sqs/faqs/](https://www.amazonaws.cn/en/sqs/faqs/)  
26. Kafka vs. RabbitMQ vs. SQS – Key Differences & Best Use Cases \- Get SDE Ready, accessed July 21, 2025, [https://getsdeready.com/kafka-vs-rabbitmq-vs-sqs-which-message-queue-is-best-for-you/](https://getsdeready.com/kafka-vs-rabbitmq-vs-sqs-which-message-queue-is-best-for-you/)  
27. Kafka Tutorial for Beginners | Everything you need to get started \- YouTube, accessed July 21, 2025, [https://www.youtube.com/watch?v=QkdkLdMBuL0](https://www.youtube.com/watch?v=QkdkLdMBuL0)  
28. RocketMQ · 官方网站 | RocketMQ \- The Apache Software Foundation, accessed July 21, 2025, [https://rocketmq.apache.org/](https://rocketmq.apache.org/)  
29. Comparative Evaluation of Java Virtual Machine-Based Message Queue Services: A Study on Kafka, Artemis, Pulsar, and RocketMQ \- MDPI, accessed July 21, 2025, [https://www.mdpi.com/2079-9292/12/23/4792](https://www.mdpi.com/2079-9292/12/23/4792)  
30. Exploring Apache Pulsar Use Cases | Datastax, accessed July 21, 2025, [https://www.datastax.com/blog/apache-pulsar-use-cases](https://www.datastax.com/blog/apache-pulsar-use-cases)  
31. Pulsar Use Cases, accessed July 21, 2025, [https://pulsar.apache.org/use-cases/](https://pulsar.apache.org/use-cases/)  
32. Choosing a Message Broker: Kafka vs RabbitMQ vs AWS SQS/SNS \- DEV Community, accessed July 21, 2025, [https://dev.to/aspecto/choosing-a-message-broker-kafka-vs-rabbitmq-vs-aws-sqs-sns-20na](https://dev.to/aspecto/choosing-a-message-broker-kafka-vs-rabbitmq-vs-aws-sqs-sns-20na)  
33. Kafka | Confluent Documentation, accessed July 21, 2025, [https://docs.confluent.io/kafka/overview.html](https://docs.confluent.io/kafka/overview.html)  
34. Kafka vs Pulsar: Key Differences \- Optiblack, accessed July 21, 2025, [https://optiblack.com/insights/kafka-vs-pulsar-key-differences](https://optiblack.com/insights/kafka-vs-pulsar-key-differences)  
35. aws.amazon.com, accessed July 21, 2025, [https://aws.amazon.com/message-queue/\#:\~:text=Message%20Queue%20Basics\&text=Message%20queues%20allow%20different%20parts,to%20send%20and%20receive%20messages.](https://aws.amazon.com/message-queue/#:~:text=Message%20Queue%20Basics&text=Message%20queues%20allow%20different%20parts,to%20send%20and%20receive%20messages.)  
36. Common Problems in Message Queues \[With Solutions\] | by Abhinav Vinci \- Medium, accessed July 21, 2025, [https://medium.com/@vinciabhinav7/common-problems-in-message-queues-with-solutions-f0703c0bd5af](https://medium.com/@vinciabhinav7/common-problems-in-message-queues-with-solutions-f0703c0bd5af)  
37. What is a message queue? \- Dynatrace, accessed July 21, 2025, [https://www.dynatrace.com/news/blog/what-is-a-message-queue/](https://www.dynatrace.com/news/blog/what-is-a-message-queue/)  
38. What Is a Message Queue? | IBM, accessed July 21, 2025, [https://www.ibm.com/think/topics/message-queues](https://www.ibm.com/think/topics/message-queues)  
39. The Big Little Guide to Message Queues \- Sudhir Jonathan, accessed July 21, 2025, [https://sudhir.io/the-big-little-guide-to-message-queues](https://sudhir.io/the-big-little-guide-to-message-queues)  
40. Message Queue and Message Broker explained in easy terms \+ Terminology, accessed July 21, 2025, [https://teymorian.medium.com/message-queue-and-message-broker-explained-in-easy-terms-terminology-ca99356a96dc](https://teymorian.medium.com/message-queue-and-message-broker-explained-in-easy-terms-terminology-ca99356a96dc)  
41. What is Message Queue? And what's it's use case? : r/csharp \- Reddit, accessed July 21, 2025, [https://www.reddit.com/r/csharp/comments/166zz25/what\_is\_message\_queue\_and\_whats\_its\_use\_case/](https://www.reddit.com/r/csharp/comments/166zz25/what_is_message_queue_and_whats_its_use_case/)  
42. Message Queuing for Beginners \- The Non-developer Perspective \- CloudAMQP, accessed July 21, 2025, [https://www.cloudamqp.com/blog/message-queuing-for-beginners-the-non-developer-perspective.html](https://www.cloudamqp.com/blog/message-queuing-for-beginners-the-non-developer-perspective.html)  
43. Learning Notes \#24 – Competing Consumer | Messaging Queue Patterns \- Parottasalna, accessed July 21, 2025, [https://parottasalna.com/2025/01/01/learning-notes-24-competing-consumer-messaging-queue-patterns/](https://parottasalna.com/2025/01/01/learning-notes-24-competing-consumer-messaging-queue-patterns/)  
44. The Simple Guide to Message Queues \- DEV Community, accessed July 21, 2025, [https://dev.to/tamerlan\_dev/the-simple-guide-to-message-queues-82a](https://dev.to/tamerlan_dev/the-simple-guide-to-message-queues-82a)  
45. Beginners Guide to Message Queues: Benefits, 2 Types & Use Cases \- Learn \- Hevo Data, accessed July 21, 2025, [https://hevodata.com/learn/message-queues/](https://hevodata.com/learn/message-queues/)  
46. Introduction to Message Queues | Baeldung on Computer Science, accessed July 21, 2025, [https://www.baeldung.com/cs/message-queues](https://www.baeldung.com/cs/message-queues)  
47. Introduction to Message Brokers: A Beginner's Guide : Mastering \- DEV Community, accessed July 21, 2025, [https://dev.to/thedsdev/introduction-to-message-brokers-a-beginners-guide-mastering-15ih](https://dev.to/thedsdev/introduction-to-message-brokers-a-beginners-guide-mastering-15ih)  
48. Message Queues \- System Design \- GeeksforGeeks, accessed July 21, 2025, [https://www.geeksforgeeks.org/system-design/message-queues-system-design/](https://www.geeksforgeeks.org/system-design/message-queues-system-design/)  
49. Message broker – complete know-how, use cases and a step-by-step guide | TSH.io, accessed July 21, 2025, [https://tsh.io/blog/message-broker/](https://tsh.io/blog/message-broker/)  
50. Publish–subscribe pattern \- Wikipedia, accessed July 21, 2025, [https://en.wikipedia.org/wiki/Publish%E2%80%93subscribe\_pattern](https://en.wikipedia.org/wiki/Publish%E2%80%93subscribe_pattern)  
51. System Design Architecture : Publisher Subscriber Model | by Akshat Sharma | Medium, accessed July 21, 2025, [https://medium.com/@akshatsharma0610/system-design-architecture-publisher-subscriber-model-7430740085d5](https://medium.com/@akshatsharma0610/system-design-architecture-publisher-subscriber-model-7430740085d5)  
52. The publish-subscribe pattern: Everything you need to know about this messaging pattern | Contentful, accessed July 21, 2025, [https://www.contentful.com/blog/publish-subscribe-pattern/](https://www.contentful.com/blog/publish-subscribe-pattern/)  
53. per message acknowledgement in Kafka / RabbitMQ \- Stack Overflow, accessed July 21, 2025, [https://stackoverflow.com/questions/54963306/per-message-acknowledgement-in-kafka-rabbitmq](https://stackoverflow.com/questions/54963306/per-message-acknowledgement-in-kafka-rabbitmq)  
54. Guaranteed Messaging Acknowledgments \- Solace Docs, accessed July 21, 2025, [https://docs.solace.com/Messaging/Guaranteed-Msg/guaranteed-messaging-acknowledgments.htm](https://docs.solace.com/Messaging/Guaranteed-Msg/guaranteed-messaging-acknowledgments.htm)  
55. Apache Kafka — Understanding how to produce and consume messages? \- Medium, accessed July 21, 2025, [https://medium.com/@sirajul.anik/apache-kafka-understanding-how-to-produce-and-consume-messages-9744c612f40f](https://medium.com/@sirajul.anik/apache-kafka-understanding-how-to-produce-and-consume-messages-9744c612f40f)  
56. Apache Kafka documentation, accessed July 21, 2025, [https://kafka.apache.org/documentation/](https://kafka.apache.org/documentation/)  
57. Azure Service Bus \- Partitioned queues and topics \- Learn Microsoft, accessed July 21, 2025, [https://learn.microsoft.com/en-us/azure/service-bus-messaging/service-bus-partitioning](https://learn.microsoft.com/en-us/azure/service-bus-messaging/service-bus-partitioning)  
58. How to Design a Message Queue Architecture for System Design Interviews \- Design Gurus, accessed July 21, 2025, [https://www.designgurus.io/answers/detail/how-to-design-a-message-queue-for-system-design-interviews](https://www.designgurus.io/answers/detail/how-to-design-a-message-queue-for-system-design-interviews)  
59. Concepts | RocketMQ, accessed July 21, 2025, [https://rocketmq.apache.org/docs/introduction/02concepts/](https://rocketmq.apache.org/docs/introduction/02concepts/)  
60. Kafka Message Queue | Svix Resources, accessed July 21, 2025, [https://www.svix.com/resources/guides/kafka-message-queue/](https://www.svix.com/resources/guides/kafka-message-queue/)  
61. Kafka Consumer Groups Explained | Apache Kafka On The Go \- Confluent Developer, accessed July 21, 2025, [https://developer.confluent.io/learn-more/kafka-on-the-go/consumer-groups/](https://developer.confluent.io/learn-more/kafka-on-the-go/consumer-groups/)  
62. Part 1: RabbitMQ for beginners \- What is RabbitMQ? \- CloudAMQP, accessed July 21, 2025, [https://www.cloudamqp.com/blog/part1-rabbitmq-for-beginners-what-is-rabbitmq.html](https://www.cloudamqp.com/blog/part1-rabbitmq-for-beginners-what-is-rabbitmq.html)  
63. Understanding RabbitMQ: A Beginner's Guide \- DEV Community, accessed July 21, 2025, [https://dev.to/abhivyaktii/understanding-rabbitmq-a-beginners-guide-297c](https://dev.to/abhivyaktii/understanding-rabbitmq-a-beginners-guide-297c)  
64. Apache Kafka \- Wikipedia, accessed July 21, 2025, [https://en.wikipedia.org/wiki/Apache\_Kafka](https://en.wikipedia.org/wiki/Apache_Kafka)  
65. Apache Kafka for Beginners: A Comprehensive Guide \- DataCamp, accessed July 21, 2025, [https://www.datacamp.com/tutorial/apache-kafka-for-beginners-a-comprehensive-guide](https://www.datacamp.com/tutorial/apache-kafka-for-beginners-a-comprehensive-guide)  
66. Kafka topic partitioning strategies and best practices \- New Relic, accessed July 21, 2025, [https://newrelic.com/blog/best-practices/effective-strategies-kafka-topic-partitioning](https://newrelic.com/blog/best-practices/effective-strategies-kafka-topic-partitioning)  
67. A Beginner's Guide to Kafka® Consumers \- Instaclustr, accessed July 21, 2025, [https://www.instaclustr.com/blog/a-beginners-guide-to-kafka-consumers/](https://www.instaclustr.com/blog/a-beginners-guide-to-kafka-consumers/)  
68. Kafka Partitions and Consumer Groups in 6 mins | by Ahmed Gulab Khan \- Medium, accessed July 21, 2025, [https://medium.com/javarevisited/kafka-partitions-and-consumer-groups-in-6-mins-9e0e336c6c00](https://medium.com/javarevisited/kafka-partitions-and-consumer-groups-in-6-mins-9e0e336c6c00)  
69. Kafka Partition Strategies: Optimize Your Data Streaming \- Redpanda, accessed July 21, 2025, [https://www.redpanda.com/guides/kafka-tutorial-kafka-partition-strategy](https://www.redpanda.com/guides/kafka-tutorial-kafka-partition-strategy)  
70. Apache Kafka Simple Producer Example \- Tutorials Point, accessed July 21, 2025, [https://www.tutorialspoint.com/apache\_kafka/apache\_kafka\_simple\_producer\_example.htm](https://www.tutorialspoint.com/apache_kafka/apache_kafka_simple_producer_example.htm)  
71. Kafka Consumer Config—Configurations & use cases \- Redpanda, accessed July 21, 2025, [https://www.redpanda.com/guides/kafka-tutorial-kafka-consumer-config](https://www.redpanda.com/guides/kafka-tutorial-kafka-consumer-config)  
72. Kafka Anti-Patterns: Common Pitfalls and How to Avoid Them | by Shailendra \- Medium, accessed July 21, 2025, [https://medium.com/@shailendrasinghpatil/kafka-anti-patterns-common-pitfalls-and-how-to-avoid-them-833cdcf2df89](https://medium.com/@shailendrasinghpatil/kafka-anti-patterns-common-pitfalls-and-how-to-avoid-them-833cdcf2df89)  
73. Study Notes 6.5-6: Kafka Producer, Consumer & Configuration \- DEV Community, accessed July 21, 2025, [https://dev.to/pizofreude/study-notes-65-6-kafka-producer-consumer-configuration-2a0l](https://dev.to/pizofreude/study-notes-65-6-kafka-producer-consumer-configuration-2a0l)  
74. Beyond Hello World: Real-World Kafka Patterns (and Pitfalls) from Production | by Enu Mittal, accessed July 21, 2025, [https://levelup.gitconnected.com/beyond-hello-world-real-world-kafka-patterns-and-pitfalls-from-production-283cd408d874](https://levelup.gitconnected.com/beyond-hello-world-real-world-kafka-patterns-and-pitfalls-from-production-283cd408d874)  
75. Kafka consumer doesn't consume messages from existing topic \- Codemia, accessed July 21, 2025, [https://codemia.io/knowledge-hub/path/kafka\_consumer\_doesnt\_consume\_messages\_from\_existing\_topic](https://codemia.io/knowledge-hub/path/kafka_consumer_doesnt_consume_messages_from_existing_topic)  
76. RabbitMQ tutorial \- "Hello world\!", accessed July 21, 2025, [https://www.rabbitmq.com/tutorials/tutorial-one-python](https://www.rabbitmq.com/tutorials/tutorial-one-python)  
77. RabbitMQ Tutorials, accessed July 21, 2025, [https://www.rabbitmq.com/tutorials](https://www.rabbitmq.com/tutorials)  
78. Kafka vs. RabbitMQ vs. Pulsar: The Ultimate Messaging Showdown \- Medium, accessed July 21, 2025, [https://medium.com/@abdipor/kafka-vs-rabbitmq-vs-pulsar-the-ultimate-messaging-showdown-4a57deaa0c27](https://medium.com/@abdipor/kafka-vs-rabbitmq-vs-pulsar-the-ultimate-messaging-showdown-4a57deaa0c27)  
79. RabbitMQ: A Complete Guide to Message Broker, Performance, and Reliability \- Medium, accessed July 21, 2025, [https://medium.com/@erickzanetti/rabbitmq-a-complete-guide-to-message-broker-performance-and-reliability-3999ee776d85](https://medium.com/@erickzanetti/rabbitmq-a-complete-guide-to-message-broker-performance-and-reliability-3999ee776d85)  
80. RabbitMQ Performance Optimization | by Jawad Zaarour \- Medium, accessed July 21, 2025, [https://medium.com/@zjawad333/rabbitmq-performance-optimization-ef2ab64a0490](https://medium.com/@zjawad333/rabbitmq-performance-optimization-ef2ab64a0490)  
81. 13 Common RabbitMQ Mistakes and How to Avoid Them \- CloudAMQP, accessed July 21, 2025, [https://www.cloudamqp.com/blog/part4-rabbitmq-13-common-errors.html](https://www.cloudamqp.com/blog/part4-rabbitmq-13-common-errors.html)  
82. Part 2: RabbitMQ Best Practice for High Performance (High Throughput) \- CloudAMQP, accessed July 21, 2025, [https://www.cloudamqp.com/blog/part2-rabbitmq-best-practice-for-high-performance.html](https://www.cloudamqp.com/blog/part2-rabbitmq-best-practice-for-high-performance.html)  
83. RabbitMQ: Does not auto acknowledging mean another consumer can pick up the message? : r/csharp \- Reddit, accessed July 21, 2025, [https://www.reddit.com/r/csharp/comments/1136odu/rabbitmq\_does\_not\_auto\_acknowledging\_mean\_another/](https://www.reddit.com/r/csharp/comments/1136odu/rabbitmq_does_not_auto_acknowledging_mean_another/)  
84. Most Common 10 RabbitMQ Mistakes and How to Avoid Them | by Can Sener \- Dev Genius, accessed July 21, 2025, [https://blog.devgenius.io/most-common-10-rabbitmq-mistakes-and-how-to-avoid-them-f4b74af5885d](https://blog.devgenius.io/most-common-10-rabbitmq-mistakes-and-how-to-avoid-them-f4b74af5885d)  
85. RabbitMQ: One broker to queue them all | RabbitMQ, accessed July 21, 2025, [https://www.rabbitmq.com/](https://www.rabbitmq.com/)  
86. Client Documentation \- RabbitMQ, accessed July 21, 2025, [https://www.rabbitmq.com/client-libraries](https://www.rabbitmq.com/client-libraries)  
87. Domain Model \- Apache RocketMQ, accessed July 21, 2025, [https://rocketmq.apache.org/docs/domainModel/01main/](https://rocketmq.apache.org/docs/domainModel/01main/)  
88. RocketMQ \- Websoft9, accessed July 21, 2025, [https://support.websoft9.com/en/docs/next/rocketmq/](https://support.websoft9.com/en/docs/next/rocketmq/)  
89. Core Concept \- Apache RocketMQ, accessed July 21, 2025, [https://rocketmq.apache.org/docs/4.x/consumer/01concept2/](https://rocketmq.apache.org/docs/4.x/consumer/01concept2/)  
90. Basic Best Practices \- Apache RocketMQ, accessed July 21, 2025, [https://rocketmq.apache.org/docs/bestPractice/01bestpractice/](https://rocketmq.apache.org/docs/bestPractice/01bestpractice/)  
91. Basic Best Practices \- Apache RocketMQ, accessed July 21, 2025, [https://rocketmq.apache.org/docs/best-practice-consumer/](https://rocketmq.apache.org/docs/best-practice-consumer/)  
92. JVM/OS Configuration \- Apache RocketMQ, accessed July 21, 2025, [https://rocketmq.apache.org/docs/4.x/bestPractice/05JVMOS/](https://rocketmq.apache.org/docs/4.x/bestPractice/05JVMOS/)  
93. Consistent Subscription Relationship \- Apache RocketMQ, accessed July 21, 2025, [https://rocketmq.apache.org/docs/bestPractice/05subscribe/](https://rocketmq.apache.org/docs/bestPractice/05subscribe/)  
94. ApsaraMQ for RocketMQ:Load balancing policies \- Alibaba Cloud, accessed July 21, 2025, [https://www.alibabacloud.com/help/en/apsaramq-for-rocketmq/cloud-message-queue-rocketmq-4-x-series/developer-reference/load-balancing-policies](https://www.alibabacloud.com/help/en/apsaramq-for-rocketmq/cloud-message-queue-rocketmq-4-x-series/developer-reference/load-balancing-policies)  
95. Consumption Retry \- Apache RocketMQ, accessed July 21, 2025, [https://rocketmq.apache.org/docs/featureBehavior/10consumerretrypolicy/](https://rocketmq.apache.org/docs/featureBehavior/10consumerretrypolicy/)