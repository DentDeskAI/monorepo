package handlers

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/dentdesk/dentdesk/internal/auth"
)

func TestResourceHandler_CreateDoctor_ForbiddenForOperator(t *testing.T) {
	handler := &ResourceHandler{}
	c, w := createAdminContext(
		http.MethodPost,
		"/api/doctors",
		[]byte(`{"name":"Doctor"}`),
		&auth.Claims{UserID: uuid.New(), ClinicID: uuid.New(), Role: "operator"},
	)

	handler.CreateDoctor(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.JSONEq(t, `{"error":"forbidden"}`, w.Body.String())
}

func TestResourceHandler_CreateDoctor_InvalidBody(t *testing.T) {
	handler := &ResourceHandler{}
	c, w := createAdminContext(
		http.MethodPost,
		"/api/doctors",
		[]byte(`{"specialty":"therapist"}`),
		&auth.Claims{UserID: uuid.New(), ClinicID: uuid.New(), Role: "owner"},
	)

	handler.CreateDoctor(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"invalid body"}`, w.Body.String())
}

func TestResourceHandler_UpdateDoctor_ForbiddenForOperator(t *testing.T) {
	handler := &ResourceHandler{}
	c, w := createAdminContext(
		http.MethodPut,
		"/api/doctors/"+uuid.NewString(),
		[]byte(`{"name":"Doctor","active":true}`),
		&auth.Claims{UserID: uuid.New(), ClinicID: uuid.New(), Role: "operator"},
	)
	c.Params = []gin.Param{{Key: "id", Value: uuid.NewString()}}

	handler.UpdateDoctor(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.JSONEq(t, `{"error":"forbidden"}`, w.Body.String())
}

func TestResourceHandler_UpdateDoctor_BadID(t *testing.T) {
	handler := &ResourceHandler{}
	c, w := createAdminContext(
		http.MethodPut,
		"/api/doctors/not-a-uuid",
		[]byte(`{"name":"Doctor","active":true}`),
		&auth.Claims{UserID: uuid.New(), ClinicID: uuid.New(), Role: "admin"},
	)
	c.Params = []gin.Param{{Key: "id", Value: "not-a-uuid"}}

	handler.UpdateDoctor(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"bad id"}`, w.Body.String())
}

func TestResourceHandler_CreateChair_InvalidBody(t *testing.T) {
	handler := &ResourceHandler{}
	c, w := createAdminContext(
		http.MethodPost,
		"/api/chairs",
		[]byte(`{}`),
		&auth.Claims{UserID: uuid.New(), ClinicID: uuid.New(), Role: "owner"},
	)

	handler.CreateChair(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"invalid body"}`, w.Body.String())
}

func TestResourceHandler_UpdateChair_BadID(t *testing.T) {
	handler := &ResourceHandler{}
	c, w := createAdminContext(
		http.MethodPut,
		"/api/chairs/not-a-uuid",
		[]byte(`{"name":"Chair 1"}`),
		&auth.Claims{UserID: uuid.New(), ClinicID: uuid.New(), Role: "admin"},
	)
	c.Params = []gin.Param{{Key: "id", Value: "not-a-uuid"}}

	handler.UpdateChair(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"bad id"}`, w.Body.String())
}

func TestResourceHandler_CreatePatient_InvalidBody(t *testing.T) {
	handler := &ResourceHandler{}
	c, w := createAdminContext(
		http.MethodPost,
		"/api/patients",
		[]byte(`{"name":"Patient"}`),
		&auth.Claims{UserID: uuid.New(), ClinicID: uuid.New(), Role: "operator"},
	)

	handler.CreatePatient(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"invalid body"}`, w.Body.String())
}

func TestResourceHandler_GetPatient_BadID(t *testing.T) {
	handler := &ResourceHandler{}
	c, w := createAdminContext(
		http.MethodGet,
		"/api/patients/not-a-uuid",
		nil,
		&auth.Claims{UserID: uuid.New(), ClinicID: uuid.New(), Role: "operator"},
	)
	c.Params = []gin.Param{{Key: "id", Value: "not-a-uuid"}}

	handler.GetPatient(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"bad id"}`, w.Body.String())
}

func TestResourceHandler_UpdatePatient_BadID(t *testing.T) {
	handler := &ResourceHandler{}
	c, w := createAdminContext(
		http.MethodPut,
		"/api/patients/not-a-uuid",
		[]byte(`{"language":"ru"}`),
		&auth.Claims{UserID: uuid.New(), ClinicID: uuid.New(), Role: "operator"},
	)
	c.Params = []gin.Param{{Key: "id", Value: "not-a-uuid"}}

	handler.UpdatePatient(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"bad id"}`, w.Body.String())
}
