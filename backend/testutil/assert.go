package testutil

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertStatus asserts the HTTP response status code.
func AssertStatus(t *testing.T, rr *httptest.ResponseRecorder, expected int) {
	t.Helper()
	assert.Equal(t, expected, rr.Code, "unexpected status code")
}

// AssertJSONResponse parses JSON response body and returns it.
func AssertJSONResponse(t *testing.T, rr *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err, "failed to parse JSON response")

	return response
}

// AssertErrorResponse asserts error response with expected error message substring.
func AssertErrorResponse(t *testing.T, rr *httptest.ResponseRecorder, expectedStatus int, expectedMessage string) {
	t.Helper()

	// Assert status code
	AssertStatus(t, rr, expectedStatus)

	// Parse JSON response
	response := AssertJSONResponse(t, rr)

	// Assert error field exists and contains expected message
	errorMsg, ok := response["error"].(string)
	require.True(t, ok, "response should have 'error' field as string")
	assert.True(t, strings.Contains(errorMsg, expectedMessage),
		"expected error message to contain '%s', got '%s'", expectedMessage, errorMsg)
}

// AssertSuccessResponse asserts success response and returns "data" field.
func AssertSuccessResponse(t *testing.T, rr *httptest.ResponseRecorder, expectedStatus int) map[string]interface{} {
	t.Helper()

	// Assert status code
	AssertStatus(t, rr, expectedStatus)

	// Parse JSON response
	response := AssertJSONResponse(t, rr)

	// Assert data field exists
	data, ok := response["data"].(map[string]interface{})
	require.True(t, ok, "response should have 'data' field as object")

	return data
}
