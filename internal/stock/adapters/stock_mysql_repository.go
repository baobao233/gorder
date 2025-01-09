package adapters

import (
	"context"
	"github.com/baobao233/gorder/stock/entity"
	"github.com/baobao233/gorder/stock/infrastructure/persistent"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	data, err := m.db.BatchGetStockByID(ctx, ids)
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

		var dest []*persistent.StockModel
		err = tx.Table("o_stock").
			Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
			Where("product_id IN ?", getIDFromEntities(data)).
			Find(&dest).Error
		if err != nil {
			return errors.Wrap(err, "failed to find data")
		}
		existing := m.unmarshalFromDatabase(dest)

		updated, err := updateFn(ctx, existing, data)
		if err != nil {
			return err
		}

		for _, upd := range updated {
			if err = tx.Table("o_stock").Where("product_id = ?", upd.ID).Update("quantity", upd.Quantity).Error; err != nil {
				return errors.Wrapf(err, "unable to update %s", upd.ID)
			}
		}
		return nil
	})
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
