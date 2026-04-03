package repository

import (
	"context"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type PromptRecord struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Type string `json:"type"` // "decision", "analysis", etc.
	Pair string `json:"pair"`

	// Raw prompt/response for debugging
	RawPrompt string `gorm:"type:text" json:"raw_prompt"`
	RawAnswer string `gorm:"type:text" json:"raw_answer"`

	// Parsed decision
	Answer     string  `gorm:"type:text" json:"answer"`
	Action     string  `json:"action"`
	SizePct    float64 `json:"size_pct"`
	Confidence float64 `json:"confidence"`
	Success    bool    `json:"success"`
}

func (PromptRecord) TableName() string {
	return "prompts"
}

type Subscription struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Symbol   string    `json:"symbol"`
	Channel  string    `json:"channel"`
	IsActive bool      `json:"is_active"`
	LastData time.Time `json:"last_data"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}

type Repository interface {
	SavePrompt(ctx context.Context, p PromptRecord) error
	GetPromptsList(ctx context.Context, limit int) ([]PromptRecord, error)

	SaveSubscription(ctx context.Context, s Subscription) error
	GetActiveSubscriptions(ctx context.Context) ([]Subscription, error)
	DeleteSubscription(ctx context.Context, symbol string) error

	Close() error
}

type sqliteRepo struct {
	db *gorm.DB
}

func NewSQLiteRepository(dbPath string) (Repository, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&PromptRecord{}, &Subscription{}); err != nil {
		return nil, err
	}

	return &sqliteRepo{db: db}, nil
}

func (r *sqliteRepo) SavePrompt(ctx context.Context, p PromptRecord) error {
	return r.db.WithContext(ctx).Create(&p).Error
}

func (r *sqliteRepo) GetPromptsList(ctx context.Context, limit int) ([]PromptRecord, error) {
	var records []PromptRecord
	err := r.db.WithContext(ctx).Order("created_at DESC").Limit(limit).Find(&records).Error
	return records, err
}

func (r *sqliteRepo) SaveSubscription(ctx context.Context, s Subscription) error {
	var existing Subscription
	err := r.db.WithContext(ctx).Where("symbol = ?", s.Symbol).First(&existing).Error
	if err == nil {
		s.ID = existing.ID
		return r.db.WithContext(ctx).Updates(&s).Error
	}
	return r.db.WithContext(ctx).Create(&s).Error
}

func (r *sqliteRepo) GetActiveSubscriptions(ctx context.Context) ([]Subscription, error) {
	var subs []Subscription
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&subs).Error
	return subs, err
}

func (r *sqliteRepo) DeleteSubscription(ctx context.Context, symbol string) error {
	return r.db.WithContext(ctx).Where("symbol = ?", symbol).Delete(&Subscription{}).Error
}

func (r *sqliteRepo) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
