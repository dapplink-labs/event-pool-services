# Database Backend Models

This directory contains your business data models.

## How to Add Models

1. Create your model struct files here (e.g., `user.go`, `product.go`)
2. Define your GORM models with proper tags
3. Add migration files in the `migrations/` directory
4. Register models in `database/db.go` for auto-migration

## Example Model

```go
package backend

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Username  string    `gorm:"index" json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
```

## Database Interface Pattern

For each model, consider creating a DB interface:

```go
type UserDB interface {
	Create(user *User) error
	FindByID(id uint) (*User, error)
	Update(user *User) error
	Delete(id uint) error
}
```

Then implement it in your database layer.
