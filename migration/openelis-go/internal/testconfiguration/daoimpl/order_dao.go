package daoimpl

import (
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	"openelis-go/internal/common/audittrail"
)

// OrderDAO backs the three *Order screens: they move sort_order and write
// nothing else.
//
// They ARE audited, unlike the renames and the UOM writes — activity 'U', one
// history row per row moved, payload carrying the sort order being REPLACED.
// Measured on all three.
//
// The payload's FIELD NAME is the bean property, not the column, and the three
// do not agree: Panel and TestSection call setSortOrderInt and emit
// `<sortOrderInt>`, while TypeOfSample calls setSortOrder and emits
// `<sortOrder>`. One column, two names, decided by which entity is being
// written — the same declared-field rule e1 pinned for the delete payload.
type OrderDAO struct {
	DB    *gorm.DB
	Audit *audittrail.Service
}

// Reorder writes one sort order per id, in a single transaction — Java hands
// the whole list to updateAll, so either every row moves or none does.
func (d *OrderDAO) Reorder(table, auditTable, payloadField string, ids []string, orders []int, sysUserID int64) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		ts := time.Now().UTC().Truncate(time.Millisecond)
		for i, id := range ids {
			var old []string
			if err := tx.Raw(
				fmt.Sprintf(`SELECT COALESCE(sort_order::text, '') FROM %s WHERE id = ?`, table),
				id).Scan(&old).Error; err != nil {
				return err
			}
			if len(old) == 0 {
				// The list Java builds only holds entities its loader returned,
				// so an id naming nothing is simply not in it.
				continue
			}
			if old[0] == strconv.Itoa(orders[i]) {
				// Hibernate finds the entity clean and skips both the UPDATE and
				// the history row — the rule e1 measured for site_information.
				continue
			}
			if err := tx.Exec(
				fmt.Sprintf(`UPDATE %s SET sort_order = ?, lastupdated = ? WHERE id = ?`, table),
				orders[i], ts, id).Error; err != nil {
				return err
			}
			changes := audittrail.Field(payloadField, old[0])
			if err := d.Audit.Write(tx, auditTable, id, sysUserID,
				audittrail.ActivityUpdate, &changes, ts); err != nil {
				return err
			}
		}
		return nil
	})
}
