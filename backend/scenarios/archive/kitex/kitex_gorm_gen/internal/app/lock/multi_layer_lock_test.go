package web

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUpdateWebWithOptimisticLock(t *testing.T) {
	ctx := context.Background()
	code := "code:" + uuid.NewString()
	name := "name:" + uuid.NewString()
	adv := uuid.New().ID()
	version := int32(0)

	extra := "updated_extra"

	// Initialize the record
	err := CreateWeb(context.Background(), code, name, int64(adv))
	assert.Nil(t, err)
	webs, err := QueryWeb(ctx, code)
	assert.Nil(t, err)
	assert.Equal(t, len(webs), 1)
	assert.Equal(t, webs[0].Version, version)

	// Initial version should be 0
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := UpdateWebWithOptimisticLock(ctx, code, extra, version)
			t.Logf("error: %+v", err)
		}()
	}

	wg.Wait()

	// Select the record again
	webs, err = QueryWeb(ctx, code)
	assert.Nil(t, err)
	assert.Equal(t, len(webs), 1)
	assert.Equal(t, webs[0].Version, version+1)
}

func TestUpdateWebWithPessimisticLock(t *testing.T) {
	ctx := context.Background()
	code := "code:" + uuid.NewString()
	name := "name:" + uuid.NewString()
	adv := uuid.New().ID()
	version := int32(0)
	times := 10

	extra := "updated_extra"

	// Initialize the record
	err := CreateWeb(context.Background(), code, name, int64(adv))
	assert.Nil(t, err)
	webs, err := QueryWeb(ctx, code)
	assert.Nil(t, err)
	assert.Equal(t, len(webs), 1)
	assert.Equal(t, webs[0].Version, version)

	// Initial version should be 0
	var wg sync.WaitGroup
	for i := 0; i < times; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := UpdateWebWithPessimisticLock(ctx, code, extra)
			t.Logf("error: %+v", err)
		}()
	}

	wg.Wait()

	// Select the record again
	webs, err = QueryWeb(ctx, code)
	assert.Nil(t, err)
	assert.Equal(t, len(webs), 1)
	assert.Equal(t, webs[0].Version, times)
}
