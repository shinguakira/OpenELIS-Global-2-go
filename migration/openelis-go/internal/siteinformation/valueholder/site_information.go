// Package valueholder ports org.openelisglobal.siteinformation.valueholder.
// Folder layout mirrors the Java source during migration.
package valueholder

import "time"

// SiteInformation is one clinlims.site_information row.
//
// The table is the application's configuration: every write through these
// endpoints runs ConfigurationProperties.loadDBValuesIntoConfiguration(), so a
// change here takes effect immediately and process-wide. That is why the e2e
// spec for this wave creates and destroys its OWN row rather than editing a
// shipped one.
type SiteInformation struct {
	ID          int64      `gorm:"column:id;primaryKey"`
	Name        string     `gorm:"column:name"`
	Lastupdated *time.Time `gorm:"column:lastupdated"`
	Description *string    `gorm:"column:description"`
	Value       *string    `gorm:"column:value"`
	Encrypted   *bool      `gorm:"column:encrypted"`
	DomainID    *int64     `gorm:"column:domain_id"`
	ValueType   string     `gorm:"column:value_type"`
	// InstructionKey and DescriptionKey are message keys, not text. The form
	// endpoint resolves them through the bundle; the menu endpoint ships the
	// keys raw and leaves the resolving to the client.
	InstructionKey       *string `gorm:"column:instruction_key"`
	Group                int     `gorm:"column:group"`
	ScheduleID           *int64  `gorm:"column:schedule_id"`
	Tag                  *string `gorm:"column:tag"`
	DictionaryCategoryID *int64  `gorm:"column:dictionary_category_id"`
	DescriptionKey       *string `gorm:"column:description_key"`
	NameKey              *string `gorm:"column:name_key"`

	// Joined from site_information_domain — the menu nests it as an object.
	DomainName        *string `gorm:"column:domain_name"`
	DomainDescription *string `gorm:"column:domain_description"`
}

// TableName is the schema-qualified table.
func (SiteInformation) TableName() string { return "clinlims.site_information" }
