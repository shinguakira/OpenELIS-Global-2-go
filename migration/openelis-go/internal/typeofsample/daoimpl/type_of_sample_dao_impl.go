// Package daoimpl ports org.openelisglobal.typeofsample.daoimpl.
// Folder layout mirrors the Java source during migration.
package daoimpl

import (
	"gorm.io/gorm"

	"openelis-go/internal/typeofsample/valueholder"
)

// TypeOfSampleDAOImpl ports TypeOfSampleDAOImpl — reads clinlims.type_of_sample
// via GORM's query builder.
type TypeOfSampleDAOImpl struct {
	DB *gorm.DB
}

// GetAllSortOrdered mirrors TypeOfSampleDAOImpl.getAllTypeOfSamplesSortOrdered():
// all rows ordered by sort_order ascending.
// Name = description when non-blank, else local_abbrev (mirrors getLocalizedName()).
func (d *TypeOfSampleDAOImpl) GetAllSortOrdered() ([]valueholder.TypeOfSample, error) {
	var samples []valueholder.TypeOfSample
	result := d.DB.
		Select(`id,
		        CASE WHEN description IS NOT NULL AND TRIM(description) != ''
		             THEN description
		             ELSE COALESCE(local_abbrev, '') END AS name,
		        COALESCE(sort_order, 0) AS sort_order`).
		Order("sort_order").
		Find(&samples)
	return samples, result.Error
}
