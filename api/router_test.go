package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sinhashubham95/go-example-project/api"
	"github.com/stretchr/testify/assert"
)

func testAPI(t *testing.T, request *http.Request, expectedStatus int) {
	router := api.GetRouter()
	w := httptest.NewRecorder()
	router.ServeHTTP(w, request)
	assert.Equal(t, expectedStatus, w.Code)
}

func TestPing(t *testing.T) {
	request, err := http.NewRequest(http.MethodGet, "/actuator/ping", nil)
	assert.NoError(t, err)
	testAPI(t, request, http.StatusOK)
}
