package training_test

import (
	"testing"

	"github.com/MaksimDzhangirov/three-dots/part13/internal/trainings/domain/training"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTraining_Cancel(t *testing.T) {
	tr := newExampleTraining(t)
	// всегда хорошая идея проверить предусловия перед выполнением теста ;-)
	assert.False(t, tr.IsCanceled())

	err := tr.Cancel()
	require.NoError(t, err)
	assert.True(t, tr.IsCanceled())
}

func TestTraining_Cancel_already_cenceled(t *testing.T) {
	tr := newCanceledTraining(t)

	assert.EqualError(t, tr.Cancel(), training.ErrTrainingAlreadyCanceled.Error())
}
