# Architecture Analysis - Clean Architecture Review

## Current Architecture Assessment

### ✅ What's Good (Following Clean Architecture)

1. **Clear Layer Separation**
   ```
   internal/
   ├── domain/          ✅ Core business entities (innermost layer)
   ├── usecase/         ✅ Business logic (application layer)
   ├── ports/           ✅ Interfaces/contracts
   └── adapter/         ✅ External implementations (outermost layer)
       ├── http/        - Web delivery
       └── repo/        - Data persistence
   ```

2. **Dependency Rule is Respected**
   - ✅ `domain/` has NO dependencies on other layers
   - ✅ `usecase/` depends only on `domain/` and `ports/`
   - ✅ `adapter/` depends on `domain/`, `ports/`, and `usecase/`
   - ✅ Dependencies point inward (toward domain)

3. **Interface-Driven Design**
   - ✅ Repository interfaces defined in `ports/`
   - ✅ Use cases depend on interfaces, not concrete implementations
   - ✅ Enables testing with mocks
   - ✅ Dependency injection works properly

4. **Domain-Centric**
   - ✅ Pure business logic in `domain/` (User, Order, Item)
   - ✅ Domain entities have no infrastructure concerns
   - ✅ Validation logic lives in domain layer

## ⚠️ Issues Found (Not Fully Clean Architecture)

### 1. **Ports Location** - CRITICAL ❌

**Current:**
```
internal/
├── ports/              ❌ At same level as domain
│   └── repositories.go
```

**Problem:**
- Ports (interfaces) should be owned by their consumers (use cases)
- Currently, `ports/` is a separate package, creating an extra dependency

**Clean Architecture says:**
- "The use case defines the input/output ports"
- Interfaces should be in the layer that USES them, not separately

**Recommended Fix:**
```
internal/
├── domain/
│   ├── user.go
│   └── order.go
├── usecase/
│   ├── user_service.go
│   ├── order_service.go
│   ├── repositories.go    ✅ MOVE HERE - owned by use cases
│   └── interfaces.go       ✅ Use case errors
└── adapter/
    ├── http/
    └── repo/
```

### 2. **Entity Duplication** - MODERATE ⚠️

**Current Issue:**
```go
// internal/domain/user.go
type User struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

// internal/adapter/repo/entities.go
type UserEntity struct {
    ID        string    `gorm:"primaryKey;..."`
    Name      string    `gorm:"not null"`
    Email     string    `gorm:"uniqueIndex;not null"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
}
```

**Analysis:**
- ✅ This is actually **CORRECT** in Clean Architecture!
- Domain entities should NOT have infrastructure tags (`gorm`, `json`)
- Adapter entities should be separate with framework-specific tags
- Conversion methods (`ToDomain()`, `FromDomain()`) are the right approach

**Why it's good:**
1. **Separation of Concerns**: Domain doesn't know about GORM
2. **Independence**: Can change database without touching domain
3. **Testability**: Domain entities are pure Go structs
4. **Flexibility**: Different adapters can have different entity structures

**BUT** - Minor improvement needed:
```go
// Domain should NOT have JSON tags!
type User struct {
    ID        string    // ❌ Remove `json:"id"`
    Name      string    // ❌ Remove `json:"name"`
    Email     string    // ❌ Remove `json:"email"`
    CreatedAt time.Time // ❌ Remove `json:"created_at"`
}
```

JSON tags are infrastructure concerns (HTTP adapter), not domain concerns!

### 3. **Adapter Importing Usecase** - CRITICAL ❌

**Problem Found:**
```go
// internal/adapter/repo/user_repo.go
import (
    "github.com/gostratum/examples/orderservice/internal/usecase"  // ❌ BAD!
)

func (r *UserRepo) Save(ctx context.Context, u *domain.User) error {
    // ...
    return usecase.ErrUnavailable  // ❌ Adapter using usecase error
}
```

**Why it's wrong:**
- Adapters should NOT import use cases
- Creates circular dependency risk
- Violates the dependency rule

**Correct approach:**
```go
// Option 1: Define errors in domain layer
// internal/domain/errors.go
var (
    ErrNotFound = errors.New("not found")
    ErrConflict = errors.New("conflict")
)

// Option 2: Adapter returns raw errors, usecase translates
func (r *UserRepo) Save(ctx context.Context, u *domain.User) error {
    err := r.db.Create(entity).Error
    return err  // Return raw error
}

// Usecase layer handles error translation
func (s *UserService) CreateUser(...) (*domain.User, error) {
    err := s.repo.Save(ctx, user)
    if err != nil {
        if errors.Is(err, gorm.ErrDuplicateKey) {
            return nil, ErrInvalid
        }
        return nil, ErrUnavailable
    }
    return user, nil
}
```

### 4. **HTTP Handler Structure** - MINOR ⚠️

**Current:**
```go
type UserHandler struct {
    service *usecase.UserService  // ⚠️ Concrete type
    log     *zap.Logger
}
```

**Better approach:**
```go
// Define interface in adapter/http package
type UserService interface {
    CreateUser(ctx context.Context, name, email string) (*domain.User, error)
    GetUser(ctx context.Context, id string) (*domain.User, error)
}

type UserHandler struct {
    service UserService  // ✅ Interface
    log     *zap.Logger
}
```

**Benefits:**
- Easier to mock for testing
- Handler doesn't depend on concrete usecase implementation
- More flexible for future changes

## 📊 Clean Architecture Score

| Aspect | Score | Status |
|--------|-------|--------|
| **Layer Separation** | 9/10 | ✅ Excellent |
| **Dependency Rule** | 7/10 | ⚠️ Some violations |
| **Interface Usage** | 8/10 | ✅ Good |
| **Entity Separation** | 9/10 | ✅ Excellent |
| **Testability** | 8/10 | ✅ Good |
| **Overall** | **8.2/10** | ✅ **GOOD** |

## 🔧 Recommended Improvements

### Priority 1 - CRITICAL (Do Now)

1. **Move ports/ into usecase/**
   ```bash
   # Move repository interfaces
   mv internal/ports/repositories.go internal/usecase/repositories.go
   rm -rf internal/ports/
   ```

2. **Remove usecase imports from adapters**
   ```go
   // internal/adapter/repo/user_repo.go
   // REMOVE: import "...internal/usecase"
   // Instead, return raw errors or use domain errors
   ```

3. **Remove JSON tags from domain entities**
   ```go
   // internal/domain/user.go
   type User struct {
       ID        string     // Remove json tags
       Name      string
       Email     string
       CreatedAt time.Time
   }
   ```

### Priority 2 - MODERATE (Should Do)

4. **Create HTTP DTOs (Data Transfer Objects)**
   ```go
   // internal/adapter/http/dtos.go
   type UserResponse struct {
       ID        string    `json:"id"`
       Name      string    `json:"name"`
       Email     string    `json:"email"`
       CreatedAt time.Time `json:"created_at"`
   }

   func (r *UserResponse) FromDomain(u *domain.User) {
       r.ID = u.ID
       r.Name = u.Name
       r.Email = u.Email
       r.CreatedAt = u.CreatedAt
   }
   ```

5. **Use interfaces in HTTP handlers**
   ```go
   // Define interface in http package for dependency inversion
   ```

### Priority 3 - NICE TO HAVE (Can Do Later)

6. **Move domain errors to domain package**
   ```go
   // internal/domain/errors.go
   var (
       ErrUserNotFound = errors.New("user not found")
       ErrInvalidUser  = errors.New("invalid user")
   )
   ```

7. **Add application-level DTOs**
   ```go
   // internal/usecase/dtos.go
   type CreateUserInput struct {
       Name  string
       Email string
   }
   ```

## 📋 Proposed Clean Architecture Structure

```
internal/
├── domain/                      # Enterprise Business Rules
│   ├── user.go                  # Pure entities (no tags!)
│   ├── order.go
│   ├── errors.go                # Domain errors
│   └── domain_test.go
│
├── usecase/                     # Application Business Rules
│   ├── user_service.go          # Use case implementations
│   ├── order_service.go
│   ├── repositories.go          # ✅ MOVED FROM ports/
│   ├── errors.go                # Application errors (ErrNotFound, etc.)
│   └── *_test.go
│
└── adapter/                     # Interface Adapters
    ├── http/                    # Web delivery
    │   ├── user_handler.go
    │   ├── order_handler.go
    │   ├── routes.go
    │   ├── dtos.go              # ✅ NEW - HTTP DTOs with json tags
    │   └── *_test.go
    │
    └── repo/                    # Data persistence
        ├── user_repo.go
        ├── order_repo.go
        ├── entities.go          # ✅ GORM entities (correct!)
        └── *_test.go
```

## 🎯 Summary

### Good News ✅
- Your architecture is **80% clean**!
- Layer separation is excellent
- Entity duplication is **intentional and correct**
- Dependency injection works well
- Tests are well-structured

### What Needs Fixing ❌
1. Move `ports/` into `usecase/` (critical)
2. Remove `usecase` imports from adapters (critical)
3. Remove infrastructure tags from domain entities (moderate)

### Key Insight 💡
The "duplicate entities" you noticed is actually **a feature, not a bug**! 

Clean Architecture requires:
- **Domain entities**: Pure business objects (no framework tags)
- **Adapter entities**: Framework-specific objects (GORM, JSON, etc.)
- **Conversion layer**: Maps between them

This separation allows you to:
- Change databases without touching domain
- Change HTTP framework without touching domain
- Test domain logic without any infrastructure

## 🚀 Next Steps

1. **Quick Win**: Move ports to usecase (15 minutes)
2. **Important**: Remove usecase imports from repo (30 minutes)
3. **Polish**: Remove JSON tags from domain, add HTTP DTOs (1 hour)

After these changes, you'll have a **9/10 Clean Architecture** implementation!
