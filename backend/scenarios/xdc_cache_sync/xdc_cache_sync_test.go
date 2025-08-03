package xdccachesync

import (
	"backend_eng_playground/pkg/repo"
	"backend_eng_playground/pkg/repo/model/model"
	"backend_eng_playground/pkg/repo/model/query"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXdcCacheSync(t *testing.T) {
	repo.Init()

	ctx := context.Background()
	name := "Test Tracker"
	uid := int64(12345)

	err := CreateTestWebTracker(ctx, name, uid)
	assert.Nil(t, err, "Failed to create test web tracker")

	var tracker *model.WebTracker
	tracker, err = query.Q.WebTracker.WithContext(ctx).Where(query.Q.WebTracker.Name.Eq(name)).First()
	assert.Nil(t, err, "Failed to query the created tracker")

	t.Logf("%v", tracker)
	assert.Equal(t, name, tracker.Name)
}
