package fakes

import (
	"github.com/cloudfoundry-incubator/notifications/gobble"
	"github.com/cloudfoundry-incubator/notifications/models"
)

type PersistenceProvider struct {
	database       *Database
	gobbleDatabase *GobbleDatabase
}

func NewPersistenceProvider(database *Database, gobbleDatabase *GobbleDatabase) *PersistenceProvider {
	return &PersistenceProvider{
		database:       database,
		gobbleDatabase: gobbleDatabase,
	}
}

func (pp *PersistenceProvider) Database() models.DatabaseInterface {
	return pp.database
}

func (pp *PersistenceProvider) GobbleDatabase() gobble.DatabaseInterface {
	return pp.gobbleDatabase
}

type GobbleDatabase struct {
	MigrateWasCalled bool
	MigrationsDir    string
}

func (gd *GobbleDatabase) Migrate(migrationsDir string) {
	gd.MigrateWasCalled = true
	gd.MigrationsDir = migrationsDir
}
