package xdccachesync

import (
	"backend_eng_playground/pkg/repo/model/model"
	"backend_eng_playground/pkg/repo/model/query"
	"context"

	"github.com/google/uuid"
)

func CreateTestWebTracker(ctx context.Context, name string, uid int64) error {
	return query.Q.Transaction(func(tx *query.Query) error {
		var (
			webTrackerRepo             = tx.WebTracker
			webTrackerUserRelationRepo = tx.WebTrackerUserRelation
		)

		trackerUUID, err := uuid.NewUUID()
		if err != nil {
			return err
		}

		relationUUID, err := uuid.NewUUID()
		if err != nil {
			return err
		}

		webTracker := &model.WebTracker{
			ID:      int64(trackerUUID.ID()),
			Code:    "TEST_" + trackerUUID.String()[:8],
			Name:    name,
			Mode:    1,
			Extra:   "{}",
			Version: 1,
		}

		if err := webTrackerRepo.WithContext(ctx).Create(webTracker); err != nil {
			return err
		}

		webTrackerUserRelation := &model.WebTrackerUserRelation{
			ID:        int64(relationUUID.ID()),
			TrackerID: webTracker.ID,
			UserID:    uid,
		}

		if err := webTrackerUserRelationRepo.WithContext(ctx).Create(webTrackerUserRelation); err != nil {
			return err
		}

		return nil
	})
}
