package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gostratum/examples/orderservice/internal/domain"
)

// MockUserRepository implements ports.UserRepository for testing
type MockUserRepository struct {
	users       map[string]*domain.User
	saveError   error
	findError   error
	updateError error
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*domain.User),
	}
}

func (m *MockUserRepository) Save(ctx context.Context, u *domain.User) error {
	if m.saveError != nil {
		return m.saveError
	}
	m.users[u.ID] = u
	return nil
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	if m.findError != nil {
		return nil, m.findError
	}
	user, exists := m.users[id]
	if !exists {
		return nil, errors.New("not found")
	}
	return user, nil
}

func (m *MockUserRepository) Update(ctx context.Context, u *domain.User) error {
	if m.updateError != nil {
		return m.updateError
	}
	if _, exists := m.users[u.ID]; !exists {
		return errors.New("not found")
	}
	m.users[u.ID] = u
	return nil
}

func (m *MockUserRepository) SetSaveError(err error) {
	m.saveError = err
}

func (m *MockUserRepository) SetFindError(err error) {
	m.findError = err
}

func (m *MockUserRepository) SetUpdateError(err error) {
	m.updateError = err
}

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name      string
		userName  string
		email     string
		saveError error
		wantErr   error
	}{
		{
			name:     "valid user creation",
			userName: "John Doe",
			email:    "john@example.com",
			wantErr:  nil,
		},
		{
			name:     "empty name should return invalid error",
			userName: "",
			email:    "john@example.com",
			wantErr:  ErrInvalid,
		},
		{
			name:     "empty email should return invalid error",
			userName: "John Doe",
			email:    "",
			wantErr:  ErrInvalid,
		},
		{
			name:     "invalid email should return invalid error",
			userName: "John Doe",
			email:    "invalid-email",
			wantErr:  ErrInvalid,
		},
		{
			name:      "repository error should return unavailable error",
			userName:  "John Doe",
			email:     "john@example.com",
			saveError: errors.New("database connection failed"),
			wantErr:   ErrUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockUserRepository()
			if tt.saveError != nil {
				repo.SetSaveError(tt.saveError)
			}

			ctx := context.Background()
			service := NewUserService(repo)
			user, err := service.CreateUser(ctx, tt.userName, tt.email)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
				}
				if user != nil {
					t.Errorf("CreateUser() should return nil user on error, got %v", user)
				}
			} else {
				if err != nil {
					t.Errorf("CreateUser() unexpected error = %v", err)
				}
				if user == nil {
					t.Errorf("CreateUser() should return user on success")
				} else {
					if user.Name != tt.userName {
						t.Errorf("CreateUser() user.Name = %v, want %v", user.Name, tt.userName)
					}
					if user.Email != tt.email {
						t.Errorf("CreateUser() user.Email = %v, want %v", user.Email, tt.email)
					}
					if user.ID == "" {
						t.Errorf("CreateUser() user.ID should not be empty")
					}
				}
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		setupUser *domain.User
		findError error
		wantErr   error
	}{
		{
			name:   "existing user should be returned",
			userID: "test-id",
			setupUser: &domain.User{
				ID:        "test-id",
				Name:      "John Doe",
				Email:     "john@example.com",
				CreatedAt: time.Now(),
			},
			wantErr: nil,
		},
		{
			name:      "non-existing user should return not found error",
			userID:    "non-existing",
			findError: ErrNotFound,
			wantErr:   ErrNotFound,
		},
		{
			name:      "repository error should return unavailable error",
			userID:    "test-id",
			findError: ErrUnavailable,
			wantErr:   ErrUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockUserRepository()

			if tt.setupUser != nil {
				repo.users[tt.setupUser.ID] = tt.setupUser
			}

			if tt.findError != nil {
				repo.SetFindError(tt.findError)
			}

			ctx := context.Background()
			service := NewUserService(repo)
			user, err := service.GetUser(ctx, tt.userID)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("GetUser() error = %v, wantErr %v", err, tt.wantErr)
				}
				if user != nil {
					t.Errorf("GetUser() should return nil user on error, got %v", user)
				}
			} else {
				if err != nil {
					t.Errorf("GetUser() unexpected error = %v", err)
				}
				if user == nil {
					t.Errorf("GetUser() should return user on success")
				} else {
					if user.ID != tt.setupUser.ID {
						t.Errorf("GetUser() user.ID = %v, want %v", user.ID, tt.setupUser.ID)
					}
					if user.Name != tt.setupUser.Name {
						t.Errorf("GetUser() user.Name = %v, want %v", user.Name, tt.setupUser.Name)
					}
					if user.Email != tt.setupUser.Email {
						t.Errorf("GetUser() user.Email = %v, want %v", user.Email, tt.setupUser.Email)
					}
				}
			}
		})
	}
}
