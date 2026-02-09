// datastore.go persists imported model payloads through datastore seeding.
// internal/features/ai/model/adapters/seeder/datastore.go
package seeder

import (
	"database/sql"

	"github.com/MadeByDoug/wls-chatbot/internal/core/datastore"
	modelports "github.com/MadeByDoug/wls-chatbot/internal/features/ai/model/ports"
)

// DatastoreSeeder imports model payloads via datastore seed operations.
type DatastoreSeeder struct{}

var _ modelports.ModelSeeder = (*DatastoreSeeder)(nil)

// NewDatastoreSeeder creates a datastore-backed model seeder adapter.
func NewDatastoreSeeder() *DatastoreSeeder {

	return &DatastoreSeeder{}
}

// SeedModels imports model payload bytes into the configured datastore.
func (*DatastoreSeeder) SeedModels(db *sql.DB, payload []byte) error {

	return datastore.SeedModels(db, payload)
}
