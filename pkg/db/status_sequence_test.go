package db_test

import (
	"sync"
	"testing"

	"github.com/m-mizutani/catbox/pkg/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type seqStore struct {
	client  interfaces.DBClient
	results []int64
}

func (x *seqStore) retrieve() error {
	seq, err := x.client.RetrieveStatusSequence()
	if err != nil {
		return err
	}

	x.results = append(x.results, seq)
	return nil
}

func TestScanSequence(t *testing.T) {
	t.Run("Retrieve unique sequence number", func(t *testing.T) {
		const threadNum int64 = 10
		const loopNum int64 = 20

		client := newTestTable(t)
		defer deleteTestTable(t, client)

		wg := &sync.WaitGroup{}

		var stores []*seqStore
		for i := int64(0); i < threadNum; i++ {
			wg.Add(1)
			store := &seqStore{client: client}
			stores = append(stores, store)

			go func() {
				defer wg.Done()
				for p := int64(0); p < loopNum; p++ {
					require.NoError(t, store.retrieve())
				}
			}()
		}

		wg.Wait()
		max := threadNum * loopNum
		results := make([]bool, max+1) // seq starts from 1
		for _, store := range stores {
			for _, r := range store.results {
				require.Greater(t, r, int64(0))
				require.LessOrEqual(t, r, max)
				assert.False(t, results[r])
				results[r] = true
			}
		}

		for r := int64(1); r <= max; r++ {
			assert.True(t, results[r])
		}
	})
}
