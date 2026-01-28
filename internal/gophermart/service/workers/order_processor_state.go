package workers

import (
	"log/slog"
	"sync"
)

type orderProcessorState struct {
	mu             sync.Mutex
	queue          OrderQueue
	isClosed       bool
	eventsInFlight uint64
	notifyCh       chan struct{}
}

func newOrderProcessorState(queue OrderQueue) *orderProcessorState {
	return &orderProcessorState{
		mu:             sync.Mutex{},
		queue:          queue,
		isClosed:       false,
		eventsInFlight: 0,
		notifyCh:       make(chan struct{}, 1),
	}
}

// IsClosed returns [true] when [OrderProcessor] has started
// a shutdown process and has stopped accepting new orders.
// It is primarily used in tests.
func (s *orderProcessorState) IsClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.isClosed
}

// HasFinished returns [true] when [OrderProcessor] finishes
// processing of all enqueued events after it has been shutdown.
// It is primarily used in tests.
func (s *orderProcessorState) HasFinished() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	isEmptyQueue := s.queue.Len() == 0
	hasProcessingFinished := s.eventsInFlight == 0

	return isEmptyQueue && hasProcessingFinished
}

// EnqueueEvent take an event and puts into the back of the in-memory queue.
func (s *orderProcessorState) EnqueueEvent(event OrderEvent) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isClosed {
		return false
	}

	slog.Debug("processor: enqueue",
		slog.Any("event", event),
		slog.Int("queue_size", s.queue.Len()),
	)

	s.queue.PushBack(event)
	s.notifyMainLoop()

	return true
}

// DecrementEventsInFlight reduces internal inflight counter by one.
// It is being called from multiple places when an event processing
// is finished (either event is discarded or put back into queue).
func (s *orderProcessorState) DecrementEventsInFlight() {
	s.mu.Lock()
	defer s.mu.Unlock()

	inflight := s.eventsInFlight

	if s.eventsInFlight > 0 {
		s.eventsInFlight--
	}

	slog.Debug("processor: inflight--",
		slog.Uint64("before", inflight),
		slog.Uint64("after", s.eventsInFlight),
		slog.Int("queue_size", s.queue.Len()),
	)
}

// SetIsClosed updates internal [isClosed] property to indicate whether the queue can accept new events or not.
func (s *orderProcessorState) SetIsClosed(value bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.isClosed = value
}

// NextEvent pops first event from the front of the queue and
// increments inflight counter.
// It will return [false] as the second argument if the queue is empty.
func (s *orderProcessorState) NextEvent() (OrderEvent, bool) {
	var event OrderEvent

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.queue.Len() == 0 {
		return event, false
	}

	event = s.queue.PopFront()

	inflight := s.eventsInFlight
	s.eventsInFlight++

	slog.Debug("processor: inflight++",
		slog.Uint64("before", inflight),
		slog.Uint64("after", s.eventsInFlight),
		slog.Int("queue_size", s.queue.Len()),
	)

	return event, true
}

// QueueSize returns current size of the queue.
func (s *orderProcessorState) QueueSize() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.queue.Len()
}

// NotifyC exposes internal notification channel for updates consumption
// by the main loop in [OrderProcessor].
func (s *orderProcessorState) NotifyC() <-chan struct{} {
	return s.notifyCh
}

func (s *orderProcessorState) notifyMainLoop() {
	// Ensure we never block on sending to notify channel.
	select {
	case <-s.notifyCh:
	default:
	}

	s.notifyCh <- struct{}{}
}
