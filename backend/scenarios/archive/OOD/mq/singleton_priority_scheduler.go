package flow

import (
	"container/heap"
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Define priority levels
type priority int

const (
	p0 priority = iota // Highest priority
	p1
	p2
	p3 // Lowest priority
)

// tag2Priority maps message tags to priorities
var tag2Priority = map[string]priority{
	"tag1": p0,
	"tag2": p1,
	"tag3": p2,
	"tag4": p3,
}

// Item represents a single message item in the priority queue.
type Item struct {
	SubmitChan chan struct{} // Consumer closes this to inform scheduler it's safe to enqueue
	ReadyChan  chan struct{} // Scheduler closes this to inform consumer item is ready to process
	Priority   priority
	Index      int
}

// PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool {
	// Lower numeric value => higher priority
	return pq[i].Priority < pq[j].Priority
}
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}
func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*Item)
	*pq = append(*pq, item)
	item.Index = len(*pq) - 1
}
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	if n == 0 {
		return nil
	}
	item := old[n-1]
	old[n-1] = nil // Avoid memory leak
	item.Index = -1
	*pq = old[:n-1]
	return item
}

// Scheduler coordinates message distribution based on priority and rate limits.
type Scheduler struct {
	enableLimit bool
	rateLimit   int64

	queue        PriorityQueue
	queueMutex   sync.Mutex
	queueRecv    chan *Item
	shutdownChan chan struct{}

	limiter *rate.Limiter
}

// NewScheduler creates a new Scheduler instance.
func NewScheduler(enableLimit bool, rateLimit int64) *Scheduler {
	var limiter *rate.Limiter
	if enableLimit && rateLimit > 0 {
		// rateLimit messages per second, with burst size equal to rateLimit
		limiter = rate.NewLimiter(rate.Limit(rateLimit), int(rateLimit))
	}
	s := &Scheduler{
		enableLimit:  enableLimit,
		rateLimit:    rateLimit,
		queue:        make(PriorityQueue, 0),
		queueRecv:    make(chan *Item, 1000),
		shutdownChan: make(chan struct{}),
		limiter:      limiter,
	}
	heap.Init(&s.queue) // Initialize the heap
	go s.run()
	return s
}

// run starts the scheduler's main loop.
func (s *Scheduler) run() {
	for {
		select {
		case <-s.shutdownChan:
			return
		case item := <-s.queueRecv:
			// Wait for the consumer to close SubmitChan to ensure item is ready to enqueue
			<-item.SubmitChan

			s.queueMutex.Lock()
			heap.Push(&s.queue, item)
			s.queueMutex.Unlock()

		default:
			s.queueMutex.Lock()
			if s.queue.Len() > 0 {
				item := heap.Pop(&s.queue).(*Item)
				s.queueMutex.Unlock()

				// Apply rate limiting
				if s.enableLimit && s.limiter != nil {
					if err := s.limiter.Wait(context.Background()); err != nil {
						log.Printf("Rate limiter error: %v", err)
						continue
					}
				}

				// Signal the consumer that the item is ready to process
				close(item.ReadyChan)

			} else {
				s.queueMutex.Unlock()
				// If queue is empty, sleep briefly to avoid busy waiting
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}

// wait enqueues an item and returns a channel that signals when it's ready to process.
// The consumer will close SubmitChan to notify the scheduler that the item can be added to the queue,
// and eventually the scheduler will close ReadyChan to let consumer proceed.
func (s *Scheduler) wait(p priority) chan struct{} {
	item := &Item{
		SubmitChan: make(chan struct{}),
		ReadyChan:  make(chan struct{}),
		Priority:   p,
	}
	// Send item to scheduler
	s.queueRecv <- item
	// The consumer now closes SubmitChan to tell the scheduler the item is fully submitted
	close(item.SubmitChan)
	// The consumer waits on ReadyChan, which will be closed by the scheduler when ready.
	return item.ReadyChan
}

// Shutdown gracefully stops the scheduler.
func (s *Scheduler) Shutdown() {
	close(s.shutdownChan)
}

// schedulerConsumer associates a priority with the global scheduler.
type schedulerConsumer struct {
	p priority
	s *Scheduler
}

// handleMsg is called when a message arrives. It enqueues the message and then waits for the scheduler's signal.
func (sc *schedulerConsumer) handleMsg(msg string) {
	readyChan := sc.s.wait(sc.p)
	// Wait until the scheduler signals that the message is ready
	<-readyChan
	// Now process the message
	log.Printf("[Consumer priority=%d] Processing: %s", sc.p, msg)
}

// Config represents consumer configuration.
type Config struct {
	topic         string
	consumerGroup string
}

func (c Config) getTopic() string { return c.topic }
func (c Config) getConsumerGroup() string {
	return c.consumerGroup
}
func (c Config) SetConsumerGroup(cg string) Config {
	c.consumerGroup = cg
	return c
}

func copyConfig(c Config) Config {
	return Config{
		topic:         c.topic,
		consumerGroup: c.consumerGroup,
	}
}

// InitRmqConsumerWithTag initializes a RocketMQ consumer for a specific tag
// This is mocked for demonstration purposes.
func InitRmqConsumerWithTag(cfg Config, handleFunc func(msg string)) {
	// In a real scenario, set up RocketMQ consumer here.
	// For demonstration, simulate message arrival:
	go func() {
		for i := 0; i < 10; i++ {
			msg := fmt.Sprintf("Message_%s_%d", cfg.consumerGroup, i)
			handleFunc(msg)
			time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
		}
	}()
}

// InitMqConsumerWithPriority sets up consumers for each tag with different priorities.
func InitMqConsumerWithPriority(cfg Config, s *Scheduler) {
	topic := cfg.getTopic()
	cg := cfg.getConsumerGroup()

	for tag, p := range tag2Priority {
		subCg := fmt.Sprintf("%s-%s", cg, tag)
		consumer := &schedulerConsumer{
			p: p,
			s: s,
		}
		_ = topic // Use topic in real consumer setup
		InitRmqConsumerWithTag(copyConfig(cfg).SetConsumerGroup(subCg), consumer.handleMsg)
	}
}

// Main function demonstrating the scheduler and consumers.
func main() {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Create a global scheduler with rate limit enabled (e.g., 5 messages/sec)
	scheduler := NewScheduler(true, 5)

	// Ensure the scheduler is shutdown gracefully on exit
	defer scheduler.Shutdown()

	// Initialize consumers with different priorities
	cfg := Config{
		topic:         "TestTopic",
		consumerGroup: "BaseConsumerGroup",
	}

	InitMqConsumerWithPriority(cfg, scheduler)

	// Run for a specified duration to observe processing
	runDuration := 10 * time.Second
	log.Printf("Scheduler running for %v...", runDuration)
	time.Sleep(runDuration)
	log.Println("Shutting down scheduler...")
}
