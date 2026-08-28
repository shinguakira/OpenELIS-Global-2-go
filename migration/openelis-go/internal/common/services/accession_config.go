package services

import "strings"

// AccessionConfig is the accession-number configuration SampleEdit reports.
//
// Every value here is DERIVED from site_information, not hardcoded. An earlier
// version of this port carried them as constants ("SITEYEARNUM", 20, 15, 5)
// with a note claiming they lived in a properties file the Go service cannot
// read. That was wrong for three of the four: the format is a real DB row, and
// the three lengths all fall out of the accession prefix.
//
// The SITEYEARNUM rules (SiteYearAccessionValidator):
//
//	getSiteEndIndex()      = len(prefix)                 -> invariant part
//	getMaxAccessionLength()= len(prefix) + 15
//	getChangeableLength()  = max - invariant             = 15
//
// so a deployment with a different prefix reports different lengths, and
// hardcoding 20/15/5 silently breaks there.
type AccessionConfig struct {
	Format               string
	IDSeparator          string
	MaxAccessionLength   int
	EditableAccession    int
	NonEditableAccession int
}

// siteYearChangeableLength is the 15 in `getSiteEndIndex() + 15`. It is a
// literal in SiteYearAccessionValidator, not configuration.
const siteYearChangeableLength = 15

// defaultIDSeparator is `default.idSeparator`. Unlike the other three this one
// really is a properties-file value with no site_information row, so it stays a
// constant — verified by searching site_information for it and finding nothing.
const defaultIDSeparator = ";"

// AccessionConfiguration reads the configuration behind SampleEdit's accession
// fields.
func (s *DisplayListService) AccessionConfiguration() (AccessionConfig, error) {
	// Property.AccessionFormat -> site_information row `acessionFormat`. The
	// missing 'c' is Java's, and matching it is required to find the row.
	format, err := s.DAO.SiteInformation("acessionFormat")
	if err != nil {
		return AccessionConfig{}, err
	}
	prefix, err := s.DAO.SiteInformation("Accession number prefix")
	if err != nil {
		return AccessionConfig{}, err
	}

	invariant := len(strings.TrimSpace(prefix))
	return AccessionConfig{
		Format:               format,
		IDSeparator:          defaultIDSeparator,
		MaxAccessionLength:   invariant + siteYearChangeableLength,
		EditableAccession:    siteYearChangeableLength,
		NonEditableAccession: invariant,
	}, nil
}
