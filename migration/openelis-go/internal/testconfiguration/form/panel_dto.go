package form

import locform "openelis-go/internal/localization/form"

// SampleTypePanelDTO is one entry of PanelCreate's existingPanelList /
// inactivePanelList.
//
// Panels is a POINTER because the key's presence is observable and carries
// meaning. createTypeOfSamplePanelMap creates a map entry the moment ANY
// sampletype_panel row for that sample type is seen, before the active filter
// runs — so:
//
//	no join rows at all          → absent from the map → `panels` key ABSENT
//	join rows, none matching     → empty list in the map → `panels: []`
//	join rows that match         → `panels: [...]`
//
// Measured: of the 14 active sample types, 4 carry the key, and the same 4
// carry it in the inactive list — with `[]` there, because those panels are
// active.
type SampleTypePanelDTO struct {
	TypeOfSampleName string      `json:"typeOfSampleName"`
	Panels           *[]PanelDTO `json:"panels,omitempty"`
}

// PanelDTO is the Panel ENTITY as Jackson serialises it — not a projection.
//
// The screen hands back the whole persistence object, nested Localization and
// all, which is why this carries fields no form needs. Field order is the
// bean's declaration order: EnumValueItemImpl's lastupdated and isActive
// first, then Panel's own.
type PanelDTO struct {
	// Lastupdated is epoch MILLISECONDS, the shape Jackson gives a
	// java.sql.Timestamp on this codebase.
	Lastupdated *int64 `json:"lastupdated,omitempty"`
	IsActive    string `json:"isActive"`
	ID          string `json:"id"`
	PanelName   string `json:"panelName"`
	Description string `json:"description"`
	// Loinc is dropped when null under Include.NON_NULL — and it is null for
	// every shipped panel, because PanelCreate's own form never binds
	// panelLoinc.
	Loinc        *string                  `json:"loinc,omitempty"`
	SortOrderInt int                      `json:"sortOrderInt"`
	Localization *locform.LocalizationDTO `json:"localization,omitempty"`
}
