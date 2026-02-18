package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockUser struct {
	mock.Mock
}

func NewMockUser() *MockUser {
	return &MockUser{}
}

// FullName implements the User interface.
func (m *MockUser) FullName() string {
	return m.Called().String(0)
}

// GetID implements the User interface.
func (m *MockUser) GetID() string {
	return m.Called().String(0)
}

// GetUsername implements the User interface.
func (m *MockUser) GetUsername() string {
	return m.Called().String(0)
}

// GetEmail implements the User interface.
func (m *MockUser) GetEmail() string {
	return m.Called().String(0)
}

// GetFirstName implements the User interface.
func (m *MockUser) GetFirstName() string {
	return m.Called().String(0)
}

// GetLastName implements the User interface.
func (m *MockUser) GetLastName() string {
	return m.Called().String(0)
}
