package adapters

import (
	"context"
	"github.com/baobao233/gorder/stock/entity"
	"github.com/baobao233/gorder/stock/infrastructure/persistent"
	"github.com/baobao233/gorder/stock/infrastructure/persistent/builder"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type MySQLStockRepository struct {
	db *persistent.MySQL
}

func NewMySQLStockRepository(db *persistent.MySQL) *MySQLStockRepository {
	return &MySQLStockRepository{db: db}
}

func (m MySQLStockRepository) GetItems(ctx context.Context, ids []string) ([]*entity.Item, error) {
	//TODO implement me
	panic("implement me")
}

func (m MySQLStockRepository) GetStock(ctx context.Context, ids []string) ([]*entity.ItemWithQuantity, error) {
	query := builder.NewStock().ProductIDs(ids...) // builder 模式
	data, err := m.db.BatchGetStockByID(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "BatchGetStockByID error")
	}
	var res []*entity.ItemWithQuantity
	for _, d := range data {
		res = append(res, &entity.ItemWithQuantity{
			ID:       d.ProductID,
			Quantity: d.Quantity,
		})
	}
	return res, nil
}

// UpdateStock 悲观锁or乐观锁更新库存
func (m MySQLStockRepository) UpdateStock(
	ctx context.Context,
	data []*entity.ItemWithQuantity,
	updateFn func(
		ctx context.Context,
		existing []*entity.ItemWithQuantity,
		query []*entity.ItemWithQuantity,
	) ([]*entity.ItemWithQuantity, error),
) error {
	// 开启事务，并且执行数据库更新操作
	return m.db.StartTransaction(func(tx *gorm.DB) (err error) {
		defer func() {
			if err != nil {
				logrus.Warnf("update stock transaction err, err=%v", err)
			}
		}()

		// 加悲观锁更新库存 or 加乐观锁更新库存
		err = m.updatePessimistic(ctx, tx, data, updateFn)
		//err = m.updateOptimistic(ctx, tx, data, updateFn)
		return err
	})
}

func (m MySQLStockRepository) updatePessimistic(ctx context.Context,
	tx *gorm.DB,
	data []*entity.ItemWithQuantity,
	updateFn func(ctx context.Context, existing []*entity.ItemWithQuantity, query []*entity.ItemWithQuantity,
	) ([]*entity.ItemWithQuantity, error)) error {
	var dest []*persistent.StockModel
	queryIDs := getIDFromEntities(data)
	if err := builder.NewStock().ProductIDs(queryIDs...).ForUpdate().
		Fill(tx.Model(&persistent.StockModel{})).
		Find(&dest).Error; err != nil {
		return errors.Wrap(err, "failed to find data")
	}

	existing := m.unmarshalFromDatabase(dest)

	updated, err := updateFn(ctx, existing, data)
	if err != nil {
		return err
	}

	for _, upd := range updated {
		for _, query := range data {
			if err = builder.NewStock().ProductIDs(upd.ID).QuantityGT(query.Quantity).
				Fill(tx.Model(&persistent.StockModel{})).
				Update("quantity", gorm.Expr("quantity - ?", query.Quantity)).Error; err != nil {
				return errors.Wrapf(err, "unable to update %s", upd.ID)
			}
		}
	}
	return nil
}

func (m MySQLStockRepository) updateOptimistic(ctx context.Context,
	tx *gorm.DB,
	data []*entity.ItemWithQuantity,
	updateFn func(ctx context.Context, existing []*entity.ItemWithQuantity, query []*entity.ItemWithQuantity,
	) ([]*entity.ItemWithQuantity, error)) error {
	var dest []*persistent.StockModel
	queryIDs := getIDFromEntities(data)
	if err := builder.NewStock().ProductIDs(queryIDs...).
		Fill(tx.Model(&persistent.StockModel{})).Find(&dest).Error; err != nil {
		return errors.Wrap(err, "failed to find data")
	}

	for _, queryData := range data {
		// 查询最新版本
		var newestRecord persistent.StockModel
		if err := builder.NewStock().ProductIDs(queryData.ID).
			Fill(tx.Model(&persistent.StockModel{})).
			First(&newestRecord).Error; err != nil {
			return err
		}

		// 更新再查最新版本与再查版本是否一致，一致时则更新；不重试，因为 version 一定是递增

		if err := builder.NewStock().ProductIDs(queryData.ID).Versions(newestRecord.Version).QuantityGT(queryData.Quantity).
			Fill(tx.Model(&persistent.StockModel{})).Updates(map[string]any{
			"quantity": gorm.Expr("quantity - ?", queryData.Quantity),
			"version":  newestRecord.Version + 1,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}

func (m MySQLStockRepository) unmarshalFromDatabase(dest []*persistent.StockModel) []*entity.ItemWithQuantity {
	var result []*entity.ItemWithQuantity
	for _, d := range dest {
		result = append(result, &entity.ItemWithQuantity{
			ID:       d.ProductID,
			Quantity: d.Quantity,
		})
	}
	return result
}

func getIDFromEntities(data []*entity.ItemWithQuantity) []string {
	var ids []string
	for _, d := range data {
		ids = append(ids, d.ID)
	}
	return ids
}
