package flow

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"
	// Pseudo import for RocketMQ
	// "github.com/apache/rocketmq-client-go/v2"
	// "github.com/apache/rocketmq-client-go/v2/consumer"
	// "github.com/apache/rocketmq-client-go/v2/primitive"
)

// Scheduler defines the common operations of a distributed priority scheduler.
type Scheduler interface {
	Start(ctx context.Context) error
	Stop()
	RegisterConsumerGroup(consumerGroup *ConsumerGroup)
	EnqueueMessage(topic string, msg *Message)
}

// Message is a simplified representation of a message to be consumed.
type Message struct {
	Topic    string
	Body     []byte
	Priority int // Lower number => Higher priority
}

// ConsumerGroup holds info about a single consumer group.
type ConsumerGroup struct {
	Name               string
	AllowedTopics      []string
	RateLimitPerSecond int
}

// tokenBucket is a simple token-bucket rate limiter.
type tokenBucket struct {
	capacity   int
	tokens     int
	refillRate int // tokens per second
	mu         sync.Mutex
}

func newTokenBucket(capacity, refillRate int) *tokenBucket {
	return &tokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
	}
}

func (tb *tokenBucket) tryConsume() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

func (tb *tokenBucket) refill() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.tokens += tb.refillRate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
}

// priorityQueue maintains a slice of messages sorted by ascending priority.
type priorityQueue []*Message

func (pq priorityQueue) Len() int {
	return len(pq)
}

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].Priority < pq[j].Priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

// DistributedPriorityScheduler is the core implementation of our scheduler.
type DistributedPriorityScheduler struct {
	mu              sync.Mutex
	topics          map[string]priorityQueue // each topic has a priority queue
	consumerGroups  []*ConsumerGroup
	rateLimiters    map[string]*tokenBucket // each consumerGroup has a rate limiter
	rocketMQClients map[string]RocketMQConsumerClient

	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	isRunning bool
}

// RocketMQConsumerClient is a pseudo-interface to represent a distributed consumer client.
type RocketMQConsumerClient interface {
	Subscribe(topic string) error
	Start() error
	Shutdown() error
}

// newRocketMQConsumerClient is our factory method to create a new RocketMQ consumer client.
func newRocketMQConsumerClient(consumerGroupName string) RocketMQConsumerClient {
	// Pseudo-implementation: In real usage, you'd create and configure the rocketmq client here
	return &mockRocketMQClient{
		name: consumerGroupName,
	}
}

// mockRocketMQClient is a mock struct to simulate the behavior of a real RocketMQ client.
type mockRocketMQClient struct {
	name string
}

func (m *mockRocketMQClient) Subscribe(topic string) error {
	// Pseudo: In real usage, set up subscription logic
	fmt.Printf("[mockRocketMQClient] Group:%s subscribed to topic:%s\n", m.name, topic)
	return nil
}

func (m *mockRocketMQClient) Start() error {
	// Pseudo: Start listening for messages from the broker
	fmt.Printf("[mockRocketMQClient] Group:%s started.\n", m.name)
	return nil
}

func (m *mockRocketMQClient) Shutdown() error {
	// Pseudo: Gracefully stop
	fmt.Printf("[mockRocketMQClient] Group:%s shutdown.\n", m.name)
	return nil
}

// NewDistributedPriorityScheduler constructs a new DistributedPriorityScheduler.
func NewDistributedPriorityScheduler() *DistributedPriorityScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &DistributedPriorityScheduler{
		topics:          make(map[string]priorityQueue),
		consumerGroups:  make([]*ConsumerGroup, 0),
		rateLimiters:    make(map[string]*tokenBucket),
		rocketMQClients: make(map[string]RocketMQConsumerClient),
		ctx:             ctx,
		cancel:          cancel,
	}
}

// Start begins all necessary background routines for the scheduler.
func (s *DistributedPriorityScheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return fmt.Errorf("scheduler already running")
	}
	s.isRunning = true
	s.mu.Unlock()

	// Merge parent context with internal context
	s.ctx, s.cancel = context.WithCancel(ctx)

	// Start RocketMQ clients for each consumer group
	for _, group := range s.consumerGroups {
		client := newRocketMQConsumerClient(group.Name)
		s.rocketMQClients[group.Name] = client
		for _, topic := range group.AllowedTopics {
			err := client.Subscribe(topic)
			if err != nil {
				return err
			}
		}
		if err := client.Start(); err != nil {
			return err
		}
	}

	// Start a goroutine to handle message consumption
	s.wg.Add(1)
	go s.dispatchLoop()

	// Start a goroutine to periodically refill rate limiters
	s.wg.Add(1)
	go s.refillLoop()

	log.Println("[Scheduler] Started successfully.")
	return nil
}

// Stop gracefully stops the scheduler and all RocketMQ consumers.
func (s *DistributedPriorityScheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return
	}

	// Signal all goroutines to stop
	s.cancel()
	s.isRunning = false

	// Wait for goroutines to finish
	s.wg.Wait()

	// Shutdown RocketMQ clients
	for groupName, client := range s.rocketMQClients {
		if err := client.Shutdown(); err != nil {
			log.Printf("[Scheduler] Error shutting down client for %s: %v\n", groupName, err)
		}
	}

	log.Println("[Scheduler] Stopped gracefully.")
}

// RegisterConsumerGroup attaches a consumer group to the scheduler.
// It also creates a rate limiter for the group if specified.
func (s *DistributedPriorityScheduler) RegisterConsumerGroup(consumerGroup *ConsumerGroup) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.consumerGroups = append(s.consumerGroups, consumerGroup)

	// Create a token bucket for rate limiting if RateLimitPerSecond > 0
	if consumerGroup.RateLimitPerSecond > 0 {
		s.rateLimiters[consumerGroup.Name] = newTokenBucket(
			consumerGroup.RateLimitPerSecond,
			consumerGroup.RateLimitPerSecond,
		)
	}
}

// EnqueueMessage places a message into the respective topic's priority queue.
func (s *DistributedPriorityScheduler) EnqueueMessage(topic string, msg *Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pq, exists := s.topics[topic]
	if !exists {
		pq = make([]*Message, 0)
	}
	pq = append(pq, msg)
	s.topics[topic] = pq
}

// dispatchLoop continuously dispatches messages from the priority queues
// to any consumer group that is allowed to handle the topic.
func (s *DistributedPriorityScheduler) dispatchLoop() {
	defer s.wg.Done()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.mu.Lock()
			for topic, pq := range s.topics {
				if len(pq) == 0 {
					continue
				}

				// Sort the slice by ascending priority
				sort.Sort(pq)
				msg := pq[0]
				pq = pq[1:]
				s.topics[topic] = pq

				// Find consumer groups that can process this topic
				for _, group := range s.consumerGroups {
					if !containsTopic(group.AllowedTopics, topic) {
						continue
					}

					// Rate-limit check
					tb, found := s.rateLimiters[group.Name]
					if found && !tb.tryConsume() {
						// Rate limit exceeded for this group, requeue for next iteration
						s.topics[topic] = append([]*Message{msg}, s.topics[topic]...)
						break
					}

					// Simulate "sending" the message to the consumer group
					log.Printf("[Scheduler] Dispatching message to group: %s (topic=%s, priority=%d)\n",
						group.Name, msg.Topic, msg.Priority)
					break
				}
			}
			s.mu.Unlock()
		}
	}
}

// refillLoop periodically refills the token buckets for each consumer group.
func (s *DistributedPriorityScheduler) refillLoop() {
	defer s.wg.Done()
	refillTicker := time.NewTicker(time.Second)
	defer refillTicker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-refillTicker.C:
			s.mu.Lock()
			for groupName, tb := range s.rateLimiters {
				tb.refill()
				log.Printf("[Scheduler] Refilled token bucket for group: %s\n", groupName)
			}
			s.mu.Unlock()
		}
	}
}

// containsTopic checks if a topic is in a given list of topics.
func containsTopic(topics []string, t string) bool {
	for _, topic := range topics {
		if topic == t {
			return true
		}
	}
	return false
}
