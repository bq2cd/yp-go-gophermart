//nolint:revive,err113
package fakes

import (
	"errors"
	"fmt"
	"sync"
)

type testErrorCount struct {
	err   error
	count uint
}

type testErrorChain struct {
	mu    sync.RWMutex
	chain []*testErrorCount
}

func (tc *testErrorChain) calculateRemaining() uint {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	var remaining uint
	for _, errCount := range tc.chain {
		if errCount == nil {
			remaining++
		} else {
			remaining += errCount.count
		}
	}

	return remaining
}

func (tc *testErrorChain) getNextErrorCount() *testErrorCount {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	for len(tc.chain) > 0 {
		errCount := tc.chain[0]

		if errCount == nil || errCount.count == 0 {
			tc.chain = tc.chain[1:]
		}

		if errCount == nil {
			return nil
		}

		if errCount.count == 0 {
			continue
		}

		errCount.count--

		return errCount
	}

	return nil
}

type TestError struct {
	*testErrorChain
}

func NewTestError(counts ...uint) TestError {
	return NewTestErrorCustom(nil, counts...)
}

func NewTestErrorCustom(custom error, counts ...uint) TestError {
	chain := make([]*testErrorCount, 0)

	err := custom
	if err == nil {
		err = errors.New("test error")
	}

	for _, count := range counts {
		var errCount *testErrorCount
		if count > 0 {
			errCount = &testErrorCount{err: err, count: count}
		}

		chain = append(chain, errCount)
	}

	return TestError{
		testErrorChain: &testErrorChain{
			mu:    sync.RWMutex{},
			chain: chain,
		},
	}
}

func (te TestError) Next() error {
	if te.testErrorChain == nil {
		return nil
	}

	remaining := te.calculateRemaining()

	errCount := te.getNextErrorCount()
	if errCount == nil {
		return nil
	}

	return fmt.Errorf("%w %d", errCount.err, remaining)
}
