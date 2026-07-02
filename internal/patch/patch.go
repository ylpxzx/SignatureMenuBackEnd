package patch

import (
	"fmt"

	"signature-menu-backend/internal/store"
)

func Run(name string, dataStore *store.Store) error {
	switch name {
	case "run_mock_data":
		return runMockData(dataStore)
	default:
		return fmt.Errorf("unknown patch %q", name)
	}
}
