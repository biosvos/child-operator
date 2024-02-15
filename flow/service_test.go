package flow_test

import (
	"testing"

	"github.com/biosvos/child-operator/flow"
	"github.com/stretchr/testify/require"
)

func TestService_Work(t *testing.T) {
	t.Parallel()
	t.Run("unhappy", func(t *testing.T) {
		t.Parallel()
		service := flow.NewService()

		err := service.Work()

		require.Error(t, err)
	})
}
