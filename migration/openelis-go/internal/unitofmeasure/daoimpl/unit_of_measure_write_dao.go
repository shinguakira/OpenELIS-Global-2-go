package daoimpl

import (
	"errors"

	"gorm.io/gorm"

	"openelis-go/internal/unitofmeasure/valueholder"
)

// The write half of UnitOfMeasureDAOImpl, for the UomCreate / UomRenameEntry
// screens.
//
// NO AUDIT ROW, and that is measured rather than assumed. reference_tables
// carries UNIT_OF_MEASURE with keep_history = 'Y', and
// UnitOfMeasureServiceImpl extends AuditableBaseObjectServiceImpl — everything
// a reader needs to conclude these writes are recorded. They are not: the
// service never sets auditTrailLog = true, so the mechanism stays off, and
// creating a UOM through Java leaves clinlims.history untouched. A port that
// wrote the history row would be more correct than the thing it ports, which
// is the failure mode this project treats as a defect.

// UomRow is a UOM as the write screens see it — name and description are
// separate columns and the rename path moves only one of them.
type UomRow struct {
	ID          int64  `gorm:"column:id"`
	Name        string `gorm:"column:name"`
	Description string `gorm:"column:description"`
}

// ErrNoSuchUom is returned when a rename names an id that does not exist.
//
// The controller does NOT surface it: updateUomNames guards with
// `if (unitOfMeasure != null)` and silently skips the whole block, so an
// unknown id answers 200 having written nothing. The error exists so the DAO
// does not have to pretend the update happened.
var ErrNoSuchUom = errors.New("unit_of_measure: no such row")

// GetAllForNames returns every UOM with both columns, for the list and
// name-string builders.
//
// No ORDER BY: createUnitOfMeasureList calls getAll(), which is
// BaseDAOImpl.getAll() with no ordering, so the row order is whatever the plan
// yields — id order in practice, on the primary-key scan.
func (d *UnitOfMeasureDAOImpl) GetAllForNames() ([]UomRow, error) {
	rows := []UomRow{}
	err := d.DB.Table("clinlims.unit_of_measure").
		Select(`id, COALESCE(name, '') AS name, COALESCE(description, '') AS description`).
		Order("id").
		Scan(&rows).Error
	return rows, err
}

// GetByID loads one UOM, or nil when it does not exist.
func (d *UnitOfMeasureDAOImpl) GetByID(id string) (*UomRow, error) {
	var row UomRow
	err := d.DB.Table("clinlims.unit_of_measure").
		Select(`id, COALESCE(name, '') AS name, COALESCE(description, '') AS description`).
		Where("id = ?", id).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// Insert ports createUnitOfMeasure + unitOfMeasureService.insert.
//
// name and description both take the SAME submitted string —
// createUnitOfMeasure calls setDescription and setUnitOfMeasureName with the
// one argument it was given. is_active is left to the column default 'Y'.
//
// The userId createUnitOfMeasure accepts is dropped on the floor: it is never
// applied to the entity. With no audit row there is nowhere for it to show, so
// nothing here takes it either.
func (d *UnitOfMeasureDAOImpl) Insert(name string) (*valueholder.UnitOfMeasure, error) {
	var row valueholder.UnitOfMeasure
	err := d.DB.Raw(`
		INSERT INTO clinlims.unit_of_measure (id, name, description, lastupdated)
		VALUES (nextval('clinlims.unit_of_measure_seq'), ?, ?, now())
		RETURNING id, name`, name, name).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// UpdateName ports updateUomNames, and the name is the whole point: it writes
// the NAME column and nothing else.
//
// setUnitOfMeasureName(nameEnglish.trim()) is the only setter called on the
// loaded entity. description keeps whatever the create put there, so the two
// columns agree when a UOM is created and disagree from the first rename
// onward. Trimming happens here because Java does it before the setter, not in
// the database.
func (d *UnitOfMeasureDAOImpl) UpdateName(id, name string) error {
	res := d.DB.Exec(
		`UPDATE clinlims.unit_of_measure SET name = ?, lastupdated = now() WHERE id = ?`,
		name, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNoSuchUom
	}
	return nil
}
