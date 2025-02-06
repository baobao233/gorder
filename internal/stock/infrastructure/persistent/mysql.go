package persistent

import (
	"context"
	"fmt"
	_ "github.com/baobao233/gorder/common/config"
	"github.com/baobao233/gorder/common/logging"
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
	CreatedAt time.Time `gorm:"column:created_at autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at autoUpdateTime"`
}

// TableName 需要为 StockModel 指定 TableName
func (StockModel) TableName() string {
	return "o_stock"
}

func (m *StockModel) BeforeCreate(tx *gorm.DB) (err error) {
	m.UpdatedAt = time.Now()
	return nil
}

func (d MySQL) StartTransaction(f func(tx *gorm.DB) error) error {
	return d.db.Transaction(f)
}

// BatchGetStockByID 根据 ids 查询 stock 中还有多少库存
func (d MySQL) BatchGetStockByID(ctx context.Context, productIDs []string) ([]StockModel, error) {
	_, deferLog := logging.WhenMySQL(ctx, "BatchGetStockByID", productIDs)
	var res []StockModel
	tx := d.db.WithContext(ctx).Clauses(clause.Returning{}).Where("product_id IN ?", productIDs).Find(&res)
	defer deferLog(res, &tx.Error)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return res, nil
}

func (d MySQL) Create(ctx context.Context, create *StockModel) error {
	_, deferLog := logging.WhenMySQL(ctx, "Create", create)
	var returning StockModel
	err := d.db.WithContext(ctx).Model(&returning).Clauses(clause.Returning{}).Create(create).Error
	defer deferLog(returning, &err)
	return err
}
