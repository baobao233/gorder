package builder

import (
	"github.com/baobao233/gorder/common/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Stock struct {
	ID        []int64  `json:"ID,omitempty"`
	ProductID []string `json:"product_id,omitempty"`
	Quantity  []int32  `json:"quantity,omitempty"`
	Version   []int64  `json:"version,omitempty"`

	// extend
	OrderBy       string `json:"order_by,omitempty"` // 是否排序
	ForUpdateLock bool   `json:"for_update,omitempty"`
}

// NewStock 由上层去调用 stock 需要填充的信息，比如 IDs, ProductIDs, Versions 函数等
func NewStock() *Stock {
	return &Stock{}
}

func (s *Stock) FormatArg() (string, error) {
	return util.MarshallString(s)
}

func (s *Stock) Fill(db *gorm.DB) *gorm.DB {
	db = s.fillWhere(db)
	if s.OrderBy != "" {
		db.Order(s.OrderBy)
	}
	return db
}

func (s *Stock) fillWhere(db *gorm.DB) *gorm.DB {
	if len(s.ID) > 0 {
		db = db.Where("ID IN (?)", s.ID)
	}
	if len(s.ProductID) > 0 {
		db = db.Where("product_id IN (?)", s.ProductID)
	}
	if len(s.Version) > 0 {
		db = db.Where("Version IN (?)", s.Version)
	}
	if len(s.Quantity) > 0 {
		db = s.fillQuantityGT(db)
	}
	if s.ForUpdateLock {
		db = db.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate})
	}
	return db
}

func (s *Stock) fillQuantityGT(db *gorm.DB) *gorm.DB {
	db.Where("Quantity >= ?", s.Quantity)
	return db
}

func (s *Stock) IDs(v ...int64) *Stock {
	s.ID = v
	return s
}

func (s *Stock) ProductIDs(v ...string) *Stock {
	s.ProductID = v
	return s
}

func (s *Stock) Versions(v ...int64) *Stock {
	s.Version = v
	return s
}

func (s *Stock) Order(v string) *Stock {
	s.OrderBy = v
	return s
}

func (s *Stock) ForUpdate() *Stock {
	s.ForUpdateLock = true
	return s
}

func (s *Stock) QuantityGT(v ...int32) *Stock {
	s.Quantity = v
	return s
}
