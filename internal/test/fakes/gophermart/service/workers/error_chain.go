//nolint:revive,err113
package fakes

import (
	"fmt"
)

type TestError struct {
	chain *[]bool
}

func NewTestError(counts ...uint) TestError {
	chain := make([]bool, 0)

	for _, count := range counts {
		if count == 0 {
			chain = append(chain, false)

			continue
		}

		for range count {
			chain = append(chain, true)
		}
	}

	return TestError{
		chain: &chain,
	}
}

func (te TestError) Next() error {
	if te.chain == nil {
		return nil
	}

	remaining := len(*te.chain)
	if remaining == 0 {
		return nil
	}

	wantErr := (*te.chain)[0]
	*te.chain = (*te.chain)[1:]

	if wantErr {
		return fmt.Errorf("test error %d", remaining)
	}

	return nil
}
