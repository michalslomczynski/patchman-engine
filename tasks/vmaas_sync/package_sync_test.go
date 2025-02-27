package vmaas_sync //nolint:revive,stylecheck

import (
	"app/base/core"
	"app/base/database"
	"app/base/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPkgListSyncPackages(t *testing.T) {
	utils.SkipWithoutDB(t)
	core.SetupTestEnvironment()
	configure()

	err := syncPackages(time.Now(), nil)
	assert.NoError(t, err)

	database.CheckPackagesNamesInDB(t, "", "bash", "curl")
	database.CheckPackagesNamesInDB(t, "summary like '% newest summary'", "bash", "curl")
	database.CheckEVRAsInDBSynced(t, 4, true,
		"77.0.1-1.fc31.src", "77.0.1-1.fc31.x86_64", // added firefox versions
		"5.7.13-200.fc31.src", "5.7.13-200.fc31.x86_64") // added kernel versions
	database.DeleteNewlyAddedPackages(t)
}
