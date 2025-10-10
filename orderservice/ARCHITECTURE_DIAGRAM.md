# OrderService Clean Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         PRESENTATION LAYER                               │
│                    (internal/adapter/http/)                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│  ┌──────────────────┐         ┌──────────────────┐                      │
│  │  UserHandler     │         │  OrderHandler    │                      │
│  ├──────────────────┤         ├──────────────────┤                      │
│  │ - CreateUser()   │         │ - CreateOrder()  │                      │
│  │ - GetUser()      │         │ - GetOrder()     │                      │
│  │ - handleError()  │         │ - handleError()  │                      │
│  └────────┬─────────┘         └────────┬─────────┘                      │
│           │                            │                                 │
│           │ Uses DTOs                  │ Uses DTOs                       │
│           │                            │                                 │
│  ┌────────▼────────────────────────────▼─────────┐                      │
│  │            dtos.go                             │                      │
│  ├────────────────────────────────────────────────┤                      │
│  │ UserResponse     (has json tags)               │                      │
│  │ OrderResponse    (has json tags)               │                      │
│  │ ItemResponse     (has json tags)               │                      │
│  │ ItemRequest      (has json tags)               │                      │
│  │                                                 │                      │
│  │ FromDomainUser(), FromDomainOrder()            │                      │
│  │ ItemRequest.ToDomain()                         │                      │
│  └─────────────────────────────────────────────────┘                      │
│                           │                                               │
└───────────────────────────┼───────────────────────────────────────────────┘
                            │ depends on ↓
┌───────────────────────────┼───────────────────────────────────────────────┐
│                           │      USE CASE LAYER                           │
│                           │  (internal/usecase/)                          │
├───────────────────────────┼───────────────────────────────────────────────┤
│                           │                                               │
│  ┌────────────────────────▼─────────────┐  ┌──────────────────────────┐ │
│  │      UserService                     │  │    OrderService          │ │
│  ├──────────────────────────────────────┤  ├──────────────────────────┤ │
│  │ - repo UserRepository                │  │ - repo OrderRepository   │ │
│  │                                      │  │                          │ │
│  │ - CreateUser()                       │  │ - CreateOrder()          │ │
│  │ - GetUser()                          │  │ - GetOrder()             │ │
│  │ - translateError()                   │  │ - translateError()       │ │
│  └──────────────────────────────────────┘  └──────────────────────────┘ │
│                                                                           │
│  ┌─────────────────────────────────────────────────────────────────────┐ │
│  │               repositories.go (interfaces)                          │ │
│  ├─────────────────────────────────────────────────────────────────────┤ │
│  │ type UserRepository interface {                                    │ │
│  │     Save(ctx, *User) error                                         │ │
│  │     FindByID(ctx, id) (*User, error)                               │ │
│  │ }                                                                   │ │
│  │                                                                     │ │
│  │ type OrderRepository interface {                                   │ │
│  │     Save(ctx, *Order) error                                        │ │
│  │     FindByID(ctx, id) (*Order, error)                              │ │
│  │ }                                                                   │ │
│  └─────────────────────────────────────────────────────────────────────┘ │
│                                                                           │
│  ┌─────────────────────────────────────────────────────────────────────┐ │
│  │               interfaces.go (errors)                                │ │
│  ├─────────────────────────────────────────────────────────────────────┤ │
│  │ ErrUnavailable  = errors.New("service unavailable")                │ │
│  │ ErrNotFound     = domain.ErrNotFound                               │ │
│  │ ErrInvalid      = domain.ErrInvalidInput                           │ │
│  │ ErrConflict     = domain.ErrConflict                               │ │
│  └─────────────────────────────────────────────────────────────────────┘ │
│                           │                                               │
└───────────────────────────┼───────────────────────────────────────────────┘
                            │ depends on ↓
┌───────────────────────────┼───────────────────────────────────────────────┐
│                           │       DOMAIN LAYER                            │
│                           │   (internal/domain/)                          │
├───────────────────────────┼───────────────────────────────────────────────┤
│                           │                                               │
│  ┌────────────────────────▼──────┐   ┌────────────────────────────────┐ │
│  │         User                  │   │         Order                  │ │
│  ├───────────────────────────────┤   ├────────────────────────────────┤ │
│  │ ID        string              │   │ ID        string               │ │
│  │ Name      string              │   │ UserID    string               │ │
│  │ Email     string              │   │ Items     []Item               │ │
│  │ CreatedAt time.Time           │   │ Status    string               │ │
│  │                               │   │ Total     float64              │ │
│  │ NewUser()                     │   │ CreatedAt time.Time            │ │
│  │ Validate()                    │   │                                │ │
│  └───────────────────────────────┘   │ NewOrder()                     │ │
│                                       │ AddItem()                      │ │
│  ┌───────────────────────────────┐   │ Validate()                     │ │
│  │         Item                  │   └────────────────────────────────┘ │
│  ├───────────────────────────────┤                                       │
│  │ ID      uint                  │                                       │
│  │ OrderID string                │                                       │
│  │ SKU     string                │   ┌────────────────────────────────┐ │
│  │ Qty     int                   │   │      errors.go                 │ │
│  │ Price   float64               │   ├────────────────────────────────┤ │
│  └───────────────────────────────┘   │ ErrNotFound                    │ │
│                                       │ ErrInvalidInput                │ │
│  ⚠️  NO JSON TAGS                     │ ErrConflict                    │ │
│  ⚠️  NO INFRASTRUCTURE DEPS           └────────────────────────────────┘ │
│  ✅  PURE BUSINESS LOGIC                                                 │
│                                                                           │
└───────────────────────────┬───────────────────────────────────────────────┘
                            ↑ implements interfaces
┌───────────────────────────┼───────────────────────────────────────────────┐
│                           │    INFRASTRUCTURE LAYER                       │
│                           │  (internal/adapter/repo/)                     │
├───────────────────────────┼───────────────────────────────────────────────┤
│                           │                                               │
│  ┌────────────────────────┴──────┐   ┌────────────────────────────────┐ │
│  │      UserRepo                 │   │       OrderRepo                │ │
│  ├───────────────────────────────┤   ├────────────────────────────────┤ │
│  │ - db *gorm.DB                 │   │ - db *gorm.DB                  │ │
│  │                               │   │                                │ │
│  │ implements UserRepository     │   │ implements OrderRepository     │ │
│  │                               │   │                                │ │
│  │ Save(ctx, user)               │   │ Save(ctx, order)               │ │
│  │ FindByID(ctx, id)             │   │ FindByID(ctx, id)              │ │
│  │                               │   │                                │ │
│  │ Returns:                      │   │ Returns:                       │ │
│  │ - domain.ErrNotFound          │   │ - domain.ErrNotFound           │ │
│  │ - domain.ErrConflict          │   │ - raw errors                   │ │
│  │ - raw errors                  │   │                                │ │
│  └───────────────────────────────┘   └────────────────────────────────┘ │
│                                                                           │
│  ┌─────────────────────────────────────────────────────────────────────┐ │
│  │               entities.go (GORM models)                             │ │
│  ├─────────────────────────────────────────────────────────────────────┤ │
│  │ UserEntity   - has gorm tags                                        │ │
│  │ OrderEntity  - has gorm tags                                        │ │
│  │ ItemEntity   - has gorm tags                                        │ │
│  │                                                                     │ │
│  │ FromDomain() - domain → entity                                      │ │
│  │ ToDomain()   - entity → domain                                      │ │
│  └─────────────────────────────────────────────────────────────────────┘ │
│                                                                           │
│  ⚠️  NO VALIDATION (belongs in use case)                                 │
│  ⚠️  NO IMPORTS FROM USECASE                                             │
│  ✅  ONLY PERSISTENCE LOGIC                                              │
│                                                                           │
└───────────────────────────────────────────────────────────────────────────┘
```

## Dependency Rule Compliance

```
HTTP Handlers
    │
    ├─> Use Case Services
    │       │
    │       ├─> Domain Entities
    │       │
    │       └─> Repository Interfaces (defined here)
    │
    └─> DTOs (for HTTP serialization)

Repository Implementations
    │
    ├─> Repository Interfaces (from use case)
    │
    └─> Domain Entities (for data conversion)
```

### Key Points:

1. **Dependencies point INWARD only**
   - HTTP → Use Case → Domain
   - Repo → Domain (implements interfaces from use case)

2. **Domain layer is 100% independent**
   - No imports from any other layer
   - Pure business logic
   - No framework dependencies

3. **Use case layer defines its needs**
   - Defines repository interfaces
   - Wraps domain errors
   - Orchestrates business logic

4. **Adapters implement interfaces**
   - HTTP adapters use DTOs
   - Repo adapters use GORM entities
   - Both convert to/from domain models

5. **Error flow is clean**
   - Domain errors are pure (no dependencies)
   - Repos return domain errors or raw errors
   - Use cases translate all errors
   - HTTP layer maps to status codes
