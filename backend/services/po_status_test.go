package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateStatusTransition_DraftToSent_Valid(t *testing.T) {
	err := ValidatePOStatusTransition("draft", "sent")
	assert.NoError(t, err)
}

func TestValidateStatusTransition_DraftToCancelled_Valid(t *testing.T) {
	err := ValidatePOStatusTransition("draft", "cancelled")
	assert.NoError(t, err)
}

func TestValidateStatusTransition_SentToCancelled_Valid(t *testing.T) {
	err := ValidatePOStatusTransition("sent", "cancelled")
	assert.NoError(t, err)
}

func TestValidateStatusTransition_ReceivedToCompleted_Valid(t *testing.T) {
	err := ValidatePOStatusTransition("received", "completed")
	assert.NoError(t, err)
}

func TestValidateStatusTransition_CompletedToAnything_Invalid(t *testing.T) {
	for _, next := range []string{"draft", "sent", "received", "cancelled"} {
		err := ValidatePOStatusTransition("completed", next)
		assert.Error(t, err, "completed -> %s should be invalid", next)
	}
}

func TestValidateStatusTransition_CancelledToAnything_Invalid(t *testing.T) {
	for _, next := range []string{"draft", "sent", "received", "completed"} {
		err := ValidatePOStatusTransition("cancelled", next)
		assert.Error(t, err, "cancelled -> %s should be invalid", next)
	}
}

func TestValidateStatusTransition_ReceivedToDraft_Invalid(t *testing.T) {
	err := ValidatePOStatusTransition("received", "draft")
	assert.Error(t, err)
}

func TestValidateStatusTransition_SentToDraft_Invalid(t *testing.T) {
	err := ValidatePOStatusTransition("sent", "draft")
	assert.Error(t, err)
}
