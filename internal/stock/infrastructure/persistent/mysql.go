package persistent

import (
	"context"
	"fmt"
	_ "github.com/baobao233/gorder/common/config"
	"github.com/baobao233/gorder/common/logging"
	"github.com/baobao233/gorder/stock/infrastructure/persistent/builder"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type MySQL struct {
	db *gorm.DB
}

func NewMySQL() *MySQL {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		viper.GetString("mysql.user"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetString("mysql.port"),
		viper.GetString("mysql.dbname"),
	)
	logrus.Debug("dsn = ", dsn)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logrus.Panicf("connect to mysql failed, err = %v", err)
	}
	//db.Callback().Create().Before("gorm:create").Register("set_create_time", func(d *gorm.DB) {
	//	d.Statement.SetColumn("CreatedAt", time.Now().Format(time.DateTime))
	//})
	return &MySQL{db: db}
}

func NewMySQLWithDB(db *gorm.DB) *MySQL {
	return &MySQL{
		db: db,
	}
}

type StockModel struct {
	ID        int64     `gorm:"column:id"`
	ProductID string    `gorm:"column:product_id"`
	Quantity  int32     `gorm:"column:quantity"`
	Version   int64     `gorm:"column:version"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName 需要为 StockModel 指定 TableName
func (StockModel) TableName() string {
	return "o_stock"
}

func (m *StockModel) BeforeCreate(tx *gorm.DB) (err error) {
	m.UpdatedAt = time.Now()
	return nil
}

func (d MySQL) UseTransaction(tx *gorm.DB) *gorm.DB {
	if tx == nil {
		return d.db
	}
	return tx
}

func (d MySQL) StartTransaction(f func(tx *gorm.DB) error) error {
	return d.db.Transaction(f)
}

func (d MySQL) GetStockByID(ctx context.Context, query *builder.Stock) (result *StockModel, err error) {
	_, deferLog := logging.WhenMySQL(ctx, "GetStockByID", query) // log
	defer deferLog(result, &err)
	err = query.Fill(d.db.WithContext(ctx)).First(&result).Error // builder 模式去封装 sql 语句的链式调用，做 where 之类的填充
	if err != nil {
		return nil, err
	}
	return result, nil
}

// BatchGetStockByID 根据 ids 查询 stock 中还有多少库存
func (d MySQL) BatchGetStockByID(ctx context.Context, query *builder.Stock) (result []StockModel, err error) {
	_, deferLog := logging.WhenMySQL(ctx, "BatchGetStockByID", query) // log
	defer deferLog(result, &err)

	err = query.Fill(d.db.WithContext(ctx).Find(&result)).Error // builder 模式去封装 sql 语句的链式调用，做 where 之类的填充
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (d MySQL) Update(ctx context.Context, tx *gorm.DB, cond *builder.Stock, update map[string]any) (err error) {
	var returning StockModel
	_, deferLog := logging.WhenMySQL(ctx, "Update", cond)
	defer deferLog(returning, &err)

	res := cond.Fill(d.UseTransaction(tx).WithContext(ctx).Model(&returning).Clauses(clause.Returning{})).Updates(update) // 链式调用
	return res.Error
}

func (d MySQL) Create(ctx context.Context, tx *gorm.DB, create *StockModel) (err error) {
	var returning StockModel
	_, deferLog := logging.WhenMySQL(ctx, "Create", create)
	defer deferLog(returning, &err)

	return d.UseTransaction(tx).WithContext(ctx).Model(&returning).Clauses(clause.Returning{}).Create(create).Error
}
