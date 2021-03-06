package models_test

import (
	"database/sql"
	"testing"

	"github.com/cloudfoundry-incubator/notifications/application"
	"github.com/cloudfoundry-incubator/notifications/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestModelsSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Models Suite")
}

var sqlDB *sql.DB

var _ = BeforeEach(func() {
	env := application.NewEnvironment()

	var err error
	sqlDB, err = sql.Open("mysql", env.DatabaseURL)
	Expect(err).NotTo(HaveOccurred())
})

func TruncateTables() {
	db := models.NewDatabase(sqlDB, models.Config{})
	env := application.NewEnvironment()
	db.Migrate(env.ModelMigrationsPath)
	db.Setup()

	connection := db.Connection().(*models.Connection)
	err := connection.TruncateTables()
	if err != nil {
		panic(err)
	}
}
