package web

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudwego/kitex-examples/bizdemo/kitex_gorm_gen/internal/repository/model/model"
	"github.com/cloudwego/kitex-examples/bizdemo/kitex_gorm_gen/internal/repository/model/query"
	"sync"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreateWeb(ctx context.Context, code, name string, advID int64) error {
	/*
		INSERT INTO `analytics_web` (`id`, `web_code`, `name`, `created_at`, `updated_at`, `deleted_at`)
		VALUES (UUID_TO_BIN(UUID()), 'test_code', 'test_name', NOW(), NOW(), NULL)
	*/
	// Start a database transaction
	err := query.Q.Transaction(func(tx *query.Query) error {
		webRepo := tx.AnalyticsWeb
		relationRepo := tx.AnalyticsWebAdvertiserRelation

		// Generate UUIDs for the new records
		webUUID, err := uuid.NewUUID()
		if err != nil {
			return fmt.Errorf("failed to generate web UUID: %w", err)
		}

		relationUUID, err := uuid.NewUUID()
		if err != nil {
			return fmt.Errorf("failed to generate relation UUID: %w", err)
		}

		// Create a new AnalyticsWeb record
		newWeb := &model.AnalyticsWeb{
			ID:      int64(webUUID.ID()),
			WebCode: code,
			Name:    name,
		}

		if err := webRepo.WithContext(ctx).Create(newWeb); err != nil {
			return fmt.Errorf("failed to create AnalyticsWeb: %w", err)
		}

		// Create a new AnalyticsWebAdvertiserRelation record
		newRelation := &model.AnalyticsWebAdvertiserRelation{
			ID:    int64(relationUUID.ID()),
			AdvID: advID,
			WebID: int64(webUUID.ID()),
		}

		if err := relationRepo.WithContext(ctx).Create(newRelation); err != nil {
			return fmt.Errorf("failed to create AnalyticsWebAdvertiserRelation: %w", err)
		}

		return nil
	})

	if err != nil {
		// Consider logging the error using a logging framework
		return err
	}

	return nil
}

func ValidateWebParams(ctx context.Context, adv int64, code string, name string) (*model.AnalyticsWeb, error) {
	/*
		select
			*
		from
			analytics_web_adv_relation a
		inner join
			analytics_web b
		on
			a.web_id = b.id
		where
			a.adv_id = $adv_id
			and b.delete_at is null
		limit
			1
	*/
	/*
		 	SELECT `analytics_web`.`web_code` FROM `analytics_web` INNER JOIN `analytics_web_advertiser_relation`
			ON `analytics_web_advertiser_relation`.`web_id` = `analytics_web`.`id` WHERE `analytics_web_advertiser_relation`.`adv_id` = 9277
		 	AND `analytics_web`.`delete_at` IS NULL ORDER BY `analytics_web`.`id` LIMIT 1
	*/
	web := query.AnalyticsWeb
	relation := query.AnalyticsWebAdvertiserRelation
	res, err := web.WithContext(ctx).Select(web.WebCode).
		Join(relation, relation.WebID.EqCol(web.ID)).
		//Where(relation.AdvID.Eq(int64(adv)), web.DeletedAt.IsNull()).
		Where(relation.AdvID.Eq(int64(adv))).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no records found with AdvID: %d", adv)
		}
		// 如果查询失败(包括找不到记录)，返回错误
		return nil, err
	}
	return res, nil
}

func QueryWeb(ctx context.Context, code string) ([]*model.AnalyticsWeb, error) {
	web := query.AnalyticsWeb
	return web.WithContext(ctx).Where(web.WebCode.Eq(code)).Find()
}

func UpdateWebAAMConfig(ctx context.Context, code string, extra string) error {
	/*
		UPDATE `analytics_web`
		SET `extra`='test_extra',`updated_at`='2024-12-14 22:18:22.701'
		WHERE `analytics_web`.`web_code` = 'test_code' AND `analytics_web`.`deleted_at` IS NULL
	*/
	web := query.AnalyticsWeb
	result, err := web.WithContext(ctx).Where(web.WebCode.Eq(code)).
		Updates(map[string]interface{}{
			"extra": extra,
		})
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no records found with WebCode: %s", code)
	}
	return nil
}

func DeleteWeb(ctx context.Context, code string) error {
	/*
		delete from analytics_web where web_code = $code
	*/
	// 软删除
	/*
		UPDATE `analytics_web` SET `deleted_at`='2024-12-14 22:24:28.101'
		WHERE `analytics_web`.`web_code` = 'non_exist_code' AND `analytics_web`.`deleted_at` IS NULL
	*/
	web := query.AnalyticsWeb

	_, err := web.WithContext(ctx).Where(web.WebCode.Eq(code)).Delete()
	if err != nil {
		return err
	}
	return nil
}

func MGetWebByCode(ctx context.Context, codes []string) ([]*model.AnalyticsWeb, error) {
	remember.MGet(ctx, redisCli, codes, durationWithRand, r.mSlowGetWebsByCode, keyFmt, options)
}

func Mget(ctx context.Context, codes []string, f func((ctx context.Context, codes []string) (map[string]interface{}, error) {
	buckets := utils.Makebuckets(codes)
	wg := sync.WaitGroup{}
	wg.Add(len(buckets))
	outputs := make(chan interface{}, len(codes))

	for idx := range buckets {
		go func(idx int, output chan <- interface{}) {
			hit, miss, err := redis.MGet(codes)
			for idx := range hit {

			}
			for idx := range miss {

			}
		}(idx, outputs)
	}
}

type CacheableRepo struct {
	redisCli go.Redis

}

func (r *CacheableRepo) mSlowGetWebsByCode(ctx context.Context, codes []string) (map[string]interface{}, error) {
	caller.Table("analytics_web").
	Select("analytics_web.*, adv_id, analytics_web_events.*).
	Joins("left join analytics_web_adv_relation on analytics_web.id = analytics_web_adv_relation.pixel_id").
	Joins("left join analytics_web_events on analytics_web.id = analytics_web_events.pixel_id and analytics_web_events.delete_at is null").
	Joins("left join analytics_asset on analytics_web.asset_id = analytics_asset.asset_id")
	Where("analytics_web.delete_at is NULL").
	Where("analytics_web.codes in (?)", codes).
	Order("analytics_web_event.update_at desc").
	Find(&result)
}

/*
analytics_web_events

 */


 results, err := product.WithContext(ctx).
		Distinct(product.Code).
		Join(productUser, productUser.ProductID.EqCol(product.ID)).
		Join(user, user.ID.EqCol(productUser.UserID)).
		Where(user.Region.Eq(region), user.Status.Eq(1)).
		Where(user.DeletedAt.IsNull(), product.DeletedAt.IsNull()).
		Find()

	if err != nil {
		return nil, err
	}

	var productCodes []string
	for _, result := range results {
		productCodes = append(productCodes, result.Code)
	}

	return productCodes, nil