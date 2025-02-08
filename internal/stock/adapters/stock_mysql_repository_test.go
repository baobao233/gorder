package adapters

import (
	"context"
	"fmt"
	"github.com/baobao233/gorder/common/entity"
	"github.com/baobao233/gorder/stock/infrastructure/persistent"
	"github.com/baobao233/gorder/stock/infrastructure/persistent/builder"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"sync"
	"testing"
	"time"
)

func setupTestDBt(t *testing.T) *persistent.MySQL {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		viper.GetString("mysql.user"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetString("mysql.port"),
		"",
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)

	testDB := viper.GetString("mysql.dbname") + "_shadow"
	assert.NoError(t, db.Exec("DROP DATABASE IF EXISTS "+testDB).Error)
	assert.NoError(t, db.Exec("CREATE DATABASE IF NOT EXISTS "+testDB).Error)

	dsn = fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		viper.GetString("mysql.user"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetString("mysql.port"),
		testDB,
	)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	})
	assert.NoError(t, err)
	assert.NoError(t, db.AutoMigrate(&persistent.StockModel{}))

	return persistent.NewMySQLWithDB(db)
}

// 测试竞争
func TestMySQLStockRepository_UpdateStock_Race(t *testing.T) {
	t.Parallel()
	db := setupTestDBt(t)

	var (
		ctx          = context.Background()
		testItem     = "test-race-item"
		initialStock = 100
	)
	err := db.Create(ctx, nil, &persistent.StockModel{ProductID: testItem, Quantity: int32(initialStock)})
	assert.NoError(t, err)

	repo := NewMySQLStockRepository(db)
	var wg sync.WaitGroup
	goroutines := 10
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := repo.UpdateStock(
				ctx,
				[]*entity.ItemWithQuantity{
					{ID: testItem, Quantity: 1},
				},
				func(ctx context.Context, existing []*entity.ItemWithQuantity, query []*entity.ItemWithQuantity) ([]*entity.ItemWithQuantity, error) {
					var newItems []*entity.ItemWithQuantity
					for _, e := range existing {
						for _, q := range query {
							if e.ID == q.ID {
								newItems = append(newItems, &entity.ItemWithQuantity{
									ID:       e.ID,
									Quantity: e.Quantity - q.Quantity,
								})
							}
						}
					}
					return newItems, nil
				},
			)
			assert.NoError(t, err)
		}()
	}
	wg.Wait()

	res, err := repo.db.BatchGetStockByID(ctx, builder.NewStock().ProductIDs(testItem))
	assert.NoError(t, err)
	assert.NotEmpty(t, res, "res can not be empty")

	expected := initialStock - goroutines
	assert.Equal(t, int32(expected), res[0].Quantity)
}

// 测试超卖
func TestMySQLStockRepository_UpdateStock_Oversell(t *testing.T) {
	t.Parallel()
	db := setupTestDBt(t)

	var (
		ctx          = context.Background()
		testItem     = "test-race-item"
		initialStock = 5
	)
	err := db.Create(ctx, nil, &persistent.StockModel{ProductID: testItem, Quantity: int32(initialStock)})
	assert.NoError(t, err)

	repo := NewMySQLStockRepository(db)
	var wg sync.WaitGroup
	goroutines := 100
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := repo.UpdateStock(
				ctx,
				[]*entity.ItemWithQuantity{
					{ID: testItem, Quantity: 1},
				},
				func(ctx context.Context, existing []*entity.ItemWithQuantity, query []*entity.ItemWithQuantity) ([]*entity.ItemWithQuantity, error) {
					var newItems []*entity.ItemWithQuantity
					for _, e := range existing {
						for _, q := range query {
							if e.ID == q.ID {
								newItems = append(newItems, &entity.ItemWithQuantity{
									ID:       e.ID,
									Quantity: e.Quantity - q.Quantity,
								})
							}
						}
					}
					return newItems, nil
				},
			)
			assert.NoError(t, err)
		}()
		time.Sleep(20 * time.Millisecond)
	}
	wg.Wait()

	res, err := repo.db.BatchGetStockByID(ctx, builder.NewStock().ProductIDs(testItem))
	assert.NoError(t, err)
	assert.NotEmpty(t, res, "res can not be empty")

	// assert.Equal(t, 0, res[0].Quantity)
	assert.GreaterOrEqual(t, res[0].Quantity, int32(0))
}
