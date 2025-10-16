package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gostratum/core/logx"
	"github.com/gostratum/httpx/responsex"

	"github.com/gostratum/examples/orderservice/internal/domain"
	"github.com/gostratum/examples/orderservice/internal/usecase"
)

// MockUserRepo implements the usecase.UserRepository interface for testing
type MockUserRepo struct {
	users       map[string]*domain.User
	saveError   error
	findError   error
	updateError error
}

func NewMockUserRepo() *MockUserRepo {
	return &MockUserRepo{
		users: make(map[string]*domain.User),
	}
}

func (m *MockUserRepo) Save(ctx context.Context, u *domain.User) error {
	if m.saveError != nil {
		return m.saveError
	}
	m.users[u.ID] = u
	return nil
}

func (m *MockUserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	if m.findError != nil {
		return nil, m.findError
	}
	user, exists := m.users[id]
	if !exists {
		return nil, usecase.ErrNotFound
	}
	return user, nil
}

func (m *MockUserRepo) Update(ctx context.Context, u *domain.User) error {
	if m.updateError != nil {
		return m.updateError
	}
	if _, exists := m.users[u.ID]; !exists {
		return usecase.ErrNotFound
	}
	m.users[u.ID] = u
	return nil
}

func (m *MockUserRepo) SetSaveError(err error) {
	m.saveError = err
}

func (m *MockUserRepo) SetFindError(err error) {
	m.findError = err
}

func (m *MockUserRepo) SetUpdateError(err error) {
	m.updateError = err
}

func TestUserHandler_CreateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		setupRepoError error
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid user creation",
			requestBody: map[string]string{
				"name":  "John Doe",
				"email": "john@example.com",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid request body",
			requestBody: map[string]string{
				"invalid": "data",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request payload",
		},
		{
			name: "empty name",
			requestBody: map[string]string{
				"name":  "",
				"email": "john@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request payload",
		},
		{
			name: "repository unavailable",
			requestBody: map[string]string{
				"name":  "John Doe",
				"email": "john@example.com",
			},
			setupRepoError: errors.New("database connection failed"),
			expectedStatus: http.StatusServiceUnavailable,
			expectedError:  "service temporarily unavailable",
		},
	}

	logger := logx.NewNoopLogger()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockUserRepo()
			if tt.setupRepoError != nil {
				repo.SetSaveError(tt.setupRepoError)
			}

			service := usecase.NewUserService(repo)
			handler := NewUserHandler(service, nil, logger)

			// Create request
			var body bytes.Buffer
			json.NewEncoder(&body).Encode(tt.requestBody)

			req, _ := http.NewRequest(http.MethodPost, "/users", &body)
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Call handler
			handler.CreateUser(c)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("CreateUser() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			// Check response body for errors
			if tt.expectedError != "" {
				var envelope responsex.Envelope[any]
				json.Unmarshal(w.Body.Bytes(), &envelope)
				if !envelope.Ok {
					if envelope.Error == nil || envelope.Error.Message != tt.expectedError {
						actualMsg := ""
						if envelope.Error != nil {
							actualMsg = envelope.Error.Message
						}
						t.Errorf("CreateUser() error message = %v, want %v", actualMsg, tt.expectedError)
					}
				}
			}

			// Check successful response
			if tt.expectedStatus == http.StatusCreated {
				var envelope responsex.Envelope[domain.User]
				err := json.Unmarshal(w.Body.Bytes(), &envelope)
				if err != nil {
					t.Errorf("CreateUser() failed to unmarshal response: %v", err)
				}
				if !envelope.Ok {
					t.Errorf("CreateUser() envelope.Ok should be true")
				}
				if envelope.Data.ID == "" {
					t.Errorf("CreateUser() user.ID should not be empty")
				}
				if envelope.Data.Name == "" {
					t.Errorf("CreateUser() user.Name should not be empty")
				}
				if envelope.Data.Email == "" {
					t.Errorf("CreateUser() user.Email should not be empty")
				}
			}
		})
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUser := &domain.User{
		ID:        "test-user-id",
		Name:      "John Doe",
		Email:     "john@example.com",
		CreatedAt: time.Now(),
	}

	tests := []struct {
		name           string
		userID         string
		setupUser      *domain.User
		setupRepoError error
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "existing user",
			userID:         "test-user-id",
			setupUser:      testUser,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty user ID",
			userID:         "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "user id is required",
		},
		{
			name:           "non-existing user",
			userID:         "non-existing",
			setupRepoError: usecase.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedError:  "user not found",
		},
		{
			name:           "repository unavailable",
			userID:         "test-user-id",
			setupRepoError: usecase.ErrUnavailable,
			expectedStatus: http.StatusServiceUnavailable,
			expectedError:  "service temporarily unavailable",
		},
	}

	logger := logx.NewNoopLogger()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockUserRepo()

			if tt.setupUser != nil {
				repo.users[tt.setupUser.ID] = tt.setupUser
			}

			if tt.setupRepoError != nil {
				repo.SetFindError(tt.setupRepoError)
			}

			service := usecase.NewUserService(repo)
			handler := NewUserHandler(service, nil, logger)

			// Create request
			req, _ := http.NewRequest(http.MethodGet, "/users/"+tt.userID, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = []gin.Param{{Key: "id", Value: tt.userID}}

			// Call handler
			handler.GetUser(c)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("GetUser() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			// Check response body for errors
			if tt.expectedError != "" {
				var envelope responsex.Envelope[any]
				json.Unmarshal(w.Body.Bytes(), &envelope)
				if !envelope.Ok {
					if envelope.Error == nil || envelope.Error.Message != tt.expectedError {
						actualMsg := ""
						if envelope.Error != nil {
							actualMsg = envelope.Error.Message
						}
						t.Errorf("GetUser() error message = %v, want %v", actualMsg, tt.expectedError)
					}
				}
			}

			// Check successful response
			if tt.expectedStatus == http.StatusOK {
				var envelope responsex.Envelope[domain.User]
				err := json.Unmarshal(w.Body.Bytes(), &envelope)
				if err != nil {
					t.Errorf("GetUser() failed to unmarshal response: %v", err)
				}
				if !envelope.Ok {
					t.Errorf("GetUser() envelope.Ok should be true")
				}
				if envelope.Data.ID != tt.setupUser.ID {
					t.Errorf("GetUser() user.ID = %v, want %v", envelope.Data.ID, tt.setupUser.ID)
				}
				if envelope.Data.Name != tt.setupUser.Name {
					t.Errorf("GetUser() user.Name = %v, want %v", envelope.Data.Name, tt.setupUser.Name)
				}
				if envelope.Data.Email != tt.setupUser.Email {
					t.Errorf("GetUser() user.Email = %v, want %v", envelope.Data.Email, tt.setupUser.Email)
				}
			}
		})
	}
}
