package repo

import (
	"time"

	"github.com/google/uuid"
	"github.com/gostratum/examples/orderservice/internal/domain"
	"gorm.io/gorm"
)

// UserEntity represents the GORM model for user table
type UserEntity struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name      string    `gorm:"not null"`
	Email     string    `gorm:"uniqueIndex;not null"`
	AvatarURL string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// TableName specifies the table name for UserEntity
func (UserEntity) TableName() string {
	return "users"
}

// BeforeCreate generates UUID for new users
func (u *UserEntity) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

// ToDomain converts UserEntity to domain.User
func (u *UserEntity) ToDomain() *domain.User {
	return &domain.User{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		AvatarURL: u.AvatarURL,
		CreatedAt: u.CreatedAt,
	}
}

// FromDomain creates UserEntity from domain.User
func (u *UserEntity) FromDomain(user *domain.User) {
	u.ID = user.ID
	u.Name = user.Name
	u.Email = user.Email
	u.AvatarURL = user.AvatarURL
	u.CreatedAt = user.CreatedAt
}

// ItemEntity represents the GORM model for item table
type ItemEntity struct {
	ID      uint    `gorm:"primaryKey"`
	OrderID string  `gorm:"type:uuid;not null;index"`
	SKU     string  `gorm:"not null"`
	Qty     int     `gorm:"not null"`
	Price   float64 `gorm:"not null"`
}

// TableName specifies the table name for ItemEntity
func (ItemEntity) TableName() string {
	return "items"
}

// ToDomain converts ItemEntity to domain.Item
func (i *ItemEntity) ToDomain() domain.Item {
	return domain.Item{
		ID:      i.ID,
		OrderID: i.OrderID,
		SKU:     i.SKU,
		Qty:     i.Qty,
		Price:   i.Price,
	}
}

// FromDomain creates ItemEntity from domain.Item
func (i *ItemEntity) FromDomain(item domain.Item) {
	i.ID = item.ID
	i.OrderID = item.OrderID
	i.SKU = item.SKU
	i.Qty = item.Qty
	i.Price = item.Price
}

// OrderEntity represents the GORM model for order table
type OrderEntity struct {
	ID        string       `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID    string       `gorm:"type:uuid;not null;index"`
	Items     []ItemEntity `gorm:"foreignKey:OrderID"`
	Status    string       `gorm:"default:'pending'"`
	Total     float64      `gorm:"not null"`
	CreatedAt time.Time    `gorm:"autoCreateTime"`
}

// TableName specifies the table name for OrderEntity
func (OrderEntity) TableName() string {
	return "orders"
}

// BeforeCreate generates UUID for new orders
func (o *OrderEntity) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	return nil
}

// ToDomain converts OrderEntity to domain.Order
func (o *OrderEntity) ToDomain() *domain.Order {
	items := make([]domain.Item, len(o.Items))
	for i, item := range o.Items {
		items[i] = item.ToDomain()
	}

	return &domain.Order{
		ID:        o.ID,
		UserID:    o.UserID,
		Items:     items,
		Status:    o.Status,
		Total:     o.Total,
		CreatedAt: o.CreatedAt,
	}
}

// FromDomain creates OrderEntity from domain.Order
func (o *OrderEntity) FromDomain(order *domain.Order) {
	o.ID = order.ID
	o.UserID = order.UserID
	o.Status = order.Status
	o.Total = order.Total
	o.CreatedAt = order.CreatedAt

	items := make([]ItemEntity, len(order.Items))
	for i, item := range order.Items {
		items[i].FromDomain(item)
	}
	o.Items = items
}
