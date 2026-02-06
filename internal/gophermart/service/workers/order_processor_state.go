package workers

import (
	"context"
	"log/slog"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
)

type orderProcessorState struct {
	mu             sync.RWMutex
	queue          OrderQueue
	ordersInFlight map[domain.OrderID]struct{}
	notifyCh       chan struct{}
	startedCh      chan struct{}
	nextWakeupTime time.Time
	pendingOrders  map[domain.OrderID]struct{}
}

func newOrderProcessorState(queue OrderQueue) *orderProcessorState {
	return &orderProcessorState{
		mu:             sync.RWMutex{},
		queue:          queue,
		ordersInFlight: make(map[domain.OrderID]struct{}),
		notifyCh:       make(chan struct{}, 1),
		startedCh:      nil,
		nextWakeupTime: time.Time{},
		pendingOrders:  make(map[domain.OrderID]struct{}),
	}
}

// EnqueueEvent attempts to put event back into the [OrderQueue].
// It is assumed that the event already has correct [processAfter] property.
// Errors from [OrderQueue] are ignored because the same order will be picked up
// again by the [NextEvent] call if this method fails.
func (s *orderProcessorState) EnqueueEvent(ctx context.Context, event OrderEvent) {
	slog.Debug("processor: enqueue",
		slog.Any("event", event),
	)

	err := s.queue.PostponeOrderProcessing(ctx, event.ProcessableOrder, event.processAfter)
	if err != nil {
		return
	}

	s.addPendingOrder(event.OrderID, event.processAfter)
}

// NextEvent pops first event from the front of the queue and
// increments inflight counter.
// It will return [false] as the second argument if the queue is empty.
func (s *orderProcessorState) NextEvent(ctx context.Context, delayConfig OrderEventDelayConfig) (OrderEvent, bool) {
	var (
		event    OrderEvent
		zeroTime time.Time
	)

	slog.DebugContext(ctx, "processor: fetching next event from queue",
		slog.Any("ctx", ctx),
	)

	order, retries, err := s.queue.GetNextProcessableOrder(ctx, s.getOrdersInFlight())
	if err != nil {
		return event, false
	}

	s.incrementEventsInFlight(order.OrderID)
	s.removePendingOrder(order.OrderID)

	event = OrderEvent{
		ProcessableOrder: order,
		processAfter:     zeroTime,
		retries:          retries,
		delayConfig:      delayConfig,
	}

	return event, true
}

// DecrementEventsInFlight reduces internal inflight counter by one.
// It is being called from multiple places when an event processing
// is finished (either event is discarded or put back into queue).
func (s *orderProcessorState) DecrementEventsInFlight(orderID domain.OrderID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	inflight := len(s.ordersInFlight)

	delete(s.ordersInFlight, orderID)

	slog.Debug("processor: inflight--",
		slog.Int("before", inflight),
		slog.Int("after", len(s.ordersInFlight)),
		slog.Uint64("order_id", uint64(orderID)),
		slog.Any("orders", s.ordersInFlight),
	)
}

// NotifyC exposes internal notification channel for updates consumption
// by the main loop in [OrderProcessor].
func (s *orderProcessorState) NotifyC() <-chan struct{} {
	return s.notifyCh
}

// TriggerQueueProcessing will notify main loop of the [OrderProcessor]
// about new events in the queue.
func (s *orderProcessorState) TriggerQueueProcessing() {
	// Ensure we never block on sending to notify channel.
	select {
	case <-s.notifyCh:
	default:
	}

	s.notifyCh <- struct{}{}

	slog.Debug("processor: triggered queue processing")
}

// WakeupC creates and returns a channel based on the next wakeup time.
// The next wakeup time is set when an event is put back into the queue
// via [EnqueueEvent] method.
// The next wakeup time means that there have been some postponed events
// and we need to wait for them to become ready for processing.
func (s *orderProcessorState) WakeupC() <-chan time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return time.After(time.Until(s.nextWakeupTime))
}

// Start creates and closes a special channel to indicate that
// queue processing has started.
func (s *orderProcessorState) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.startedCh = make(chan struct{})

	close(s.startedCh)

	slog.Debug("processor: activating internal state")
}

// HasStarted returns [true] when queue processing is started
// by [OrderProcessor] via [Start] method.
func (s *orderProcessorState) HasStarted() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	select {
	case <-s.startedCh:
		return true
	default:
		return false
	}
}

// Stop resets all internal counters to ensure [HasFinished] returns [true] in the end.
// This method is called by [OrderProcessor] when running context has been canceled.
func (s *orderProcessorState) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	slog.Debug("processor: deactivating internal state",
		slog.Any("pending_orders", s.pendingOrders),
		slog.Any("inflight_orders", s.ordersInFlight),
	)

	clear(s.pendingOrders)
}

// HasFinished returns [true] when [OrderProcessor] finishes
// processing of all enqueued events after it has been shutdown.
// It is primarily used in tests.
func (s *orderProcessorState) HasFinished() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hasProcessingFinished := len(s.ordersInFlight) == 0
	hasNoPendingEvents := len(s.pendingOrders) == 0

	return hasProcessingFinished && hasNoPendingEvents
}

func (s *orderProcessorState) addPendingOrder(orderID domain.OrderID, processAfter time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pendingOrders[orderID] = struct{}{}

	if s.nextWakeupTime.After(processAfter) {
		return
	}

	s.nextWakeupTime = processAfter

	slog.Debug("processor: updated next wakeup time",
		slog.Time("next_wakeup", s.nextWakeupTime),
		slog.Uint64("order_id", uint64(orderID)),
		slog.Any("pending_orders", s.pendingOrders),
	)
}

func (s *orderProcessorState) removePendingOrder(orderID domain.OrderID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.pendingOrders, orderID)

	slog.Debug("processor: removed pending order",
		slog.Uint64("order_id", uint64(orderID)),
		slog.Any("remaining_orders", s.pendingOrders),
	)
}

func (s *orderProcessorState) getOrdersInFlight() []domain.OrderID {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return slices.Collect(maps.Keys(s.ordersInFlight))
}

func (s *orderProcessorState) incrementEventsInFlight(orderID domain.OrderID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	inflight := len(s.ordersInFlight)

	s.ordersInFlight[orderID] = struct{}{}

	slog.Debug("processor: inflight++",
		slog.Int("before", inflight),
		slog.Int("after", len(s.ordersInFlight)),
		slog.Uint64("order_id", uint64(orderID)),
	)
}
