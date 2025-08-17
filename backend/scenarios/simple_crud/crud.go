package simplecrud

import (
	"backend_eng_playground/pkg/repo/model/model"
	"backend_eng_playground/pkg/repo/model/query"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CreateWebProduct(ctx context.Context, name string, uid int64) (string, error) {
	var productCode string
	err := query.Q.Transaction(func(tx *query.Query) error {
		var (
			webProductRepo             = tx.WebProduct
			webProductUserRelationRepo = tx.WebProductUserRelation
		)

		productUUID, err := uuid.NewUUID()
		if err != nil {
			return err
		}

		relationUUID, err := uuid.NewUUID()
		if err != nil {
			return err
		}

		webProduct := &model.WebProduct{
			ID:      int64(productUUID.ID()),
			Code:    productUUID.String()[:8],
			Name:    name,
			Mode:    1,
			Extra:   "{}",
			Version: 1,
		}

		if err := webProductRepo.WithContext(ctx).Create(webProduct); err != nil {
			return err
		}

		webProductUserRelation := &model.WebProductUserRelation{
			ID:        int64(relationUUID.ID()),
			ProductID: webProduct.ID,
			UserID:    uid,
		}

		if err := webProductUserRelationRepo.WithContext(ctx).Create(webProductUserRelation); err != nil {
			return err
		}

		productCode = webProduct.Code
		return nil
	})

	if err != nil {
		return "", err
	}
	return productCode, nil
}

func MGetWebProductCodes(ctx context.Context, uid int64) ([]string, error) {
	var (
		product     = query.WebProduct
		productUser = query.WebProductUserRelation
	)

	results, err := product.WithContext(ctx).Select(product.Code).
		Join(productUser, productUser.ProductID.EqCol(product.ID)).
		Where(productUser.UserID.Eq(uid), product.DeletedAt.IsNull()).
		Find()

	if err != nil {
		return nil, err
	}

	var productCodes []string
	for _, result := range results {
		productCodes = append(productCodes, result.Code)
	}

	return productCodes, nil
}

func CheckBeforeChangeProductName(ctx context.Context, uid int64, code string, name string) (bool, error) {
	var (
		product     = query.WebProduct
		productUser = query.WebProductUserRelation
	)

	/*
		1,SIMPLE,web_product_user_relation,,ref,"uk_product_user,idx_user_id",idx_user_id,8,const,2,100,Using temporary; Using filesort
		1,SIMPLE,web_product,,eq_ref,PRIMARY,PRIMARY,8,playground.web_product_user_relation.product_id,1,7.69,Using where
	*/
	res, err := product.WithContext(ctx).Select(product.Code).
		Join(productUser, productUser.ProductID.EqCol(product.ID), productUser.UserID.Eq(uid)).
		Where(product.Name.Eq(name), product.DeletedAt.IsNull()). // No need explicit filter deleteAt actually when using GORM
		First()

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Printf("Record Not Found, Name is %v", name)
			return true, err
		}
		return false, err
	}

	return res.Code == code, nil
}

func ListProductCodesByUserRegion(ctx context.Context, region string) ([]string, error) {
	var (
		user        = query.User
		product     = query.WebProduct
		productUser = query.WebProductUserRelation
	)

	/*
		ON: filter Joined table
		Where: filter final result
		Predicate-pushdown: move filter closer to data source
	*/
	results, err := product.WithContext(ctx).Select(product.Code).
		Distinct().
		Join(productUser, productUser.ProductID.EqCol(product.ID)).
		Join(user, user.ID.EqCol(productUser.UserID)).
		Where(user.Region.Eq(region), user.Status.Eq(1), product.DeletedAt.IsNull()).
		Find()

	if err != nil {
		return nil, err
	}

	var productCodes []string
	for _, result := range results {
		productCodes = append(productCodes, result.Code)
	}

	return productCodes, nil
}

func DeleteUser(ctx context.Context, uid int64) (bool, error) {
	// TODO 硬删除 user 问题: 相关记录该如何处理？
	return false, nil
}

func ListAllProductWithTimeout(ctx context.Context, uid int64) ([]string, error) {
	// TODO 上有通过 ctx 设置超时 超时取消 mysql 查询
	return nil, nil
}

func UpsertUser(ctx context.Context, email string, name string) (bool, error) {
	// Solution:
	//
	// 1. `Clauses(clause.OnConflict)`:
	//   - Pros: Atomic, efficient, avoids race conditions, idiomatic GORM.
	//   - Cons: Requires database support for `ON CONFLICT` or `ON DUPLICATE KEY UPDATE`.
	//
	// 2. Find then Create/Update in Transaction:
	//   - Pros: Works on any database, explicit control.
	//   - Cons: Less efficient (multiple DB round trips), requires manual transaction management.
	//
	// 3. Raw SQL:
	//   - Pros: Full control over the SQL.
	//   - Cons: Less portable, bypasses GORM features, potential for SQL injection.
	//
	// We'll use `clause.OnConflict` as it's the best practice with GORM.
	user := &model.User{
		ID:    int64(uuid.New().ID()),
		Email: email,
		Name:  name,
		Code:  uuid.New().String(),
		Extra: "{}",
		Status: func() *int32 {
			var s int32 = 1
			return &s
		}(),
	}

	err := query.Q.User.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "email"}},
		DoUpdates: clause.AssignmentColumns([]string{"name"}),
	}).Create(user)

	if err != nil {
		return false, err
	}

	return true, nil
}

func SelectUserForUpdate(ctx context.Context) {
	// TODO
	return
}
