package tracing

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authorization"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/filtering"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// testSessionContextData implements the sessionContextData interface used by AttachSessionContextDataToSpan, so we avoid importing authentication/sessions (which would create an import cycle with tracing tests).
type testSessionContextData struct {
	servicePerms    authorization.ServiceRolePermissionChecker
	userID          string
	activeAccountID string
}

func (t *testSessionContextData) GetUserID() string          { return t.userID }
func (t *testSessionContextData) GetActiveAccountID() string { return t.activeAccountID }
func (t *testSessionContextData) GetServicePermissions() authorization.ServiceRolePermissionChecker {
	return t.servicePerms
}

type mockSessionData struct {
	mock.Mock
}

// GetUserID implements our interface.
func (m *mockSessionData) GetUserID() string {
	return m.Called().String(0)
}

// GetServicePermissions implements our interface.
func (m *mockSessionData) GetServicePermissions() authorization.ServiceRolePermissionChecker {
	return m.Called().Get(0).(authorization.ServiceRolePermissionChecker)
}

// GetActiveAccountID implements our interface.
func (m *mockSessionData) GetActiveAccountID() string {
	return m.Called().String(0)
}

func TestAttachSessionContextDataToSpan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		_, span := StartSpan(context.Background())

		AttachSessionContextDataToSpan(span, &testSessionContextData{
			servicePerms: authorization.NewServiceRolePermissionChecker(authorization.ServiceUserRole.String()),
		})
	})
}

func TestAttachRequestToSpan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx, span := StartSpan(context.Background())
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/", http.NoBody)
		req.Header.Set(t.Name(), "blah")
		require.NoError(t, err)

		AttachRequestToSpan(span, req)
	})
}

func TestAttachResponseToSpan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		_, span := StartSpan(context.Background())
		res := &http.Response{
			Header: map[string][]string{},
		}
		res.Header.Set(t.Name(), "blah")

		AttachResponseToSpan(span, res)
	})
}

func TestAttachErrorToSpan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		_, span := StartSpan(context.Background())

		AttachErrorToSpan(span, t.Name(), errors.New("blah"))
	})
}

func TestAttachQueryFilterToSpan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		_, span := StartSpan(context.Background())

		AttachQueryFilterToSpan(span, filtering.DefaultQueryFilter())
	})

	T.Run("with nil", func(t *testing.T) {
		t.Parallel()

		_, span := StartSpan(context.Background())

		AttachQueryFilterToSpan(span, nil)
	})
}

func TestAttachToSpan(T *testing.T) {
	T.Parallel()

	type testCase struct {
		x             any
		name          string
		attachmentKey string
	}
	tests := []testCase{

		{
			name:          "int",
			attachmentKey: T.Name(),
			x:             1,
		},
		{
			name:          "int slice",
			attachmentKey: T.Name(),
			x:             []int{1},
		},
		{
			name:          "int8",
			attachmentKey: T.Name(),
			x:             int8(1),
		},
		{
			name:          "int8 slice",
			attachmentKey: T.Name(),
			x:             []int8{int8(1)},
		},
		{
			name:          "int16",
			attachmentKey: T.Name(),
			x:             int16(1),
		},
		{
			name:          "int16 slice",
			attachmentKey: T.Name(),
			x:             []int16{int16(1)},
		},
		{
			name:          "int32",
			attachmentKey: T.Name(),
			x:             int32(1),
		},
		{
			name:          "int32 slice",
			attachmentKey: T.Name(),
			x:             []int32{int32(1)},
		},
		{
			name:          "int64",
			attachmentKey: T.Name(),
			x:             int64(1),
		},
		{
			name:          "int64 slice",
			attachmentKey: T.Name(),
			x:             []int64{int64(1)},
		},
		{
			name:          "uint",
			attachmentKey: T.Name(),
			x:             uint(1),
		},
		{
			name:          "uint slice",
			attachmentKey: T.Name(),
			x:             []uint{uint(1)},
		},
		{
			name:          "uint8",
			attachmentKey: T.Name(),
			x:             uint8(1),
		},
		{
			name:          "uint8 slice",
			attachmentKey: T.Name(),
			x:             []uint8{uint8(1)},
		},
		{
			name:          "uint16",
			attachmentKey: T.Name(),
			x:             uint16(1),
		},
		{
			name:          "uint16 slice",
			attachmentKey: T.Name(),
			x:             []uint16{uint16(1)},
		},
		{
			name:          "uint32",
			attachmentKey: T.Name(),
			x:             uint32(1),
		},
		{
			name:          "uint32 slice",
			attachmentKey: T.Name(),
			x:             []uint32{uint32(1)},
		},
		{
			name:          "uint64",
			attachmentKey: T.Name(),
			x:             uint64(1),
		},
		{
			name:          "uint64 slice",
			attachmentKey: T.Name(),
			x:             []uint64{uint64(1)},
		},
		{
			name:          "float32",
			attachmentKey: T.Name(),
			x:             float32(1),
		},
		{
			name:          "float32 slice",
			attachmentKey: T.Name(),
			x:             []float32{float32(1)},
		},
		{
			name:          "float64",
			attachmentKey: T.Name(),
			x:             float64(1),
		},
		{
			name:          "float64 slice",
			attachmentKey: T.Name(),
			x:             []float64{float64(1)},
		},
		{
			name:          "string",
			attachmentKey: T.Name(),
			x:             "test",
		},
		{
			name:          "string slice",
			attachmentKey: T.Name(),
			x:             []string{"test"},
		},
		{
			name:          "bool",
			attachmentKey: T.Name(),
			x:             true,
		},
		{
			name:          "bool slice",
			attachmentKey: T.Name(),
			x:             []bool{true},
		},
	}
	for _, tt := range tests {
		T.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, span := StartSpan(context.Background())
			AttachToSpan(span, tt.attachmentKey, tt.x)
		})
	}
}
