// Package form ports the response shapes of the sample batch-entry form load.
// Folder layout mirrors the Java source during migration.
package form

import (
	commonform "openelis-go/internal/common/form"
	"openelis-go/internal/common/util"
)

// ProjectDataDTO mirrors org.openelisglobal.sample.form.ProjectData.
//
// Every member below is a FIELD DEFAULT on a freshly constructed bean — the
// batch-entry form load builds two of these and populates neither, so the
// response is the bean's initial state. Two booleans initialise to TRUE
// (abbottOrRocheAnalysis and preservCytTaken); the other 49 are false. The three
// `new ArrayList()` members serialize as [] rather than being dropped.
//
// hivStatusList is the exception: it is populated, with the FULL dictionary
// entities of the HIVResult category — the heavy shape, not {id,value}.
//
// projectDataEID and projectDataVL are two separate instances of this bean with
// identical contents. Measured: every key holds the same value in both.
type ProjectDataDTO struct {
	AbbottOrRocheAnalysis        bool `json:"abbottOrRocheAnalysis"` // defaults TRUE on the bean
	AsanteTest                   bool `json:"asanteTest"`
	BasoTest                     bool `json:"basoTest"`
	CcmhTest                     bool `json:"ccmhTest"`
	Cd3CountTest                 bool `json:"cd3CountTest"`
	Cd4CountTest                 bool `json:"cd4CountTest"`
	Cd4cd8Test                   bool `json:"cd4cd8Test"`
	CollectionDoneByHealthWorker bool `json:"collectionDoneByHealthWorker"`
	CreatinineTest               bool `json:"creatinineTest"`
	DBSTaken                     bool `json:"dbsTaken"`
	DBSvlTaken                   bool `json:"dbsvlTaken"`
	DNAPCR                       bool `json:"dnaPCR"`
	DryTubeTaken                 bool `json:"dryTubeTaken"`
	EdtaTubeTaken                bool `json:"edtaTubeTaken"`
	EoTest                       bool `json:"eoTest"`
	GbTest                       bool `json:"gbTest"`
	GeneXpertAnalysis            bool `json:"geneXpertAnalysis"`
	GenieII100Test               bool `json:"genieII100Test"`
	GenieII10Test                bool `json:"genieII10Test"`
	GenieIITest                  bool `json:"genieIITest"`
	GenotypingTest               bool `json:"genotypingTest"`
	GenscreenTest                bool `json:"genscreenTest"`
	GlycemiaTest                 bool `json:"glycemiaTest"`
	GrTest                       bool `json:"grTest"`
	HbTest                       bool `json:"hbTest"`
	HctTest                      bool `json:"hctTest"`
	HpvTest                      bool `json:"hpvTest"`
	InnoliaTest                  bool `json:"innoliaTest"`
	IntegralTest                 bool `json:"integralTest"`
	LymphTest                    bool `json:"lymphTest"`
	MonoTest                     bool `json:"monoTest"`
	MurexTest                    bool `json:"murexTest"`
	NeutTest                     bool `json:"neutTest"`
	NfsTest                      bool `json:"nfsTest"`
	P24AgTest                    bool `json:"p24AgTest"`
	PlasmaTaken                  bool `json:"plasmaTaken"`
	PlqTest                      bool `json:"plqTest"`
	PreservCytTaken              bool `json:"preservCytTaken"` // defaults TRUE on the bean
	PscvlTaken                   bool `json:"pscvlTaken"`
	SelfCollection               bool `json:"selfCollection"`
	SerologyHIVTest              bool `json:"serologyHIVTest"`
	SerumTaken                   bool `json:"serumTaken"`
	TcmhTest                     bool `json:"tcmhTest"`
	TransaminaseALTLTest         bool `json:"transaminaseALTLTest"`
	TransaminaseASTLTest         bool `json:"transaminaseASTLTest"`
	TransaminaseTest             bool `json:"transaminaseTest"`
	VgmTest                      bool `json:"vgmTest"`
	ViralLoadTest                bool `json:"viralLoadTest"`
	VironostikaTest              bool `json:"vironostikaTest"`
	Wb1Test                      bool `json:"wb1Test"`
	Wb2Test                      bool `json:"wb2Test"`

	// `new ArrayList()` on the bean, never filled on this path -> [].
	EidWhichPCRList          []any `json:"eidWhichPCRList"`
	EidSecondPCRReasonList   []any `json:"eidSecondPCRReasonList"`
	IsUnderInvestigationList []any `json:"isUnderInvestigationList"`

	// Populated from the HIVResult dictionary category, as full entities.
	HivStatusList []commonform.DictionaryEntityDTO `json:"hivStatusList"`
}

// BatchEntrySetupDTO mirrors the SampleBatchEntrySetup form load.
//
// Shares many keys with SamplePatientEntry but is NOT the same object:
//   - sampleTypes here is every ACTIVE human type_of_sample; SamplePatientEntry
//     role-filters the same key through getUserSampleTypes.
//   - currentTime exists only here (SamplePatientEntry puts a receivedTime
//     inside sampleOrderItems instead).
//   - project, projectDataEID and projectDataVL exist only here.
//   - there is no patientProperties / patientSearch / referralReasons /
//     referralOrganizations / rejectReasonList.
type BatchEntrySetupDTO struct {
	FormName   string `json:"formName"`
	FormMethod string `json:"formMethod"`

	CancelAction   string `json:"cancelAction"`
	CancelMethod   string `json:"cancelMethod"`
	SubmitOnCancel bool   `json:"submitOnCancel"`

	CurrentDate string `json:"currentDate"`
	CurrentTime string `json:"currentTime"`

	CustomNotificationLogic bool `json:"customNotificationLogic"`
	FacilityIDCheck         bool `json:"facilityIDCheck"`
	LocalDBOnly             bool `json:"localDBOnly"`
	OrderEntryOnly          bool `json:"orderEntryOnly"`
	PatientInfoCheck        bool `json:"patientInfoCheck"`
	UseReferral             bool `json:"useReferral"`
	Warning                 bool `json:"warning"`

	PatientUpdateStatus string `json:"patientUpdateStatus"`
	Project             string `json:"project"`
	SampleXML           string `json:"sampleXML"`

	InitialSampleConditionList []util.IdValuePair `json:"initialSampleConditionList"`
	SampleTypes                []util.IdValuePair `json:"sampleTypes"`
	TestSectionList            []util.IdValuePair `json:"testSectionList"`

	Projects []commonform.ProjectEntityDTO `json:"projects"`

	SampleOrderItems *SampleOrderFormDTO `json:"sampleOrderItems"`

	ProjectDataEID *ProjectDataDTO `json:"projectDataEID"`
	ProjectDataVL  *ProjectDataDTO `json:"projectDataVL"`
}

// SampleOrderFormDTO is the BLANK-FORM variant of sampleOrderItems, shared by
// SampleBatchEntrySetup and SamplePatientEntry.
//
// Distinguished from SampleEdit's variant by what is ABSENT: there is no
// labNo, sampleId, collectionDate or priority, because no sample exists yet.
// It carries a requestDate that SampleEdit's does not.
type SampleOrderFormDTO struct {
	RequestDate            string `json:"requestDate"`
	ReceivedDateForDisplay string `json:"receivedDateForDisplay"`
	ReceivedTime           string `json:"receivedTime"`

	PaymentOptions       []util.IdValuePair `json:"paymentOptions"`
	PriorityList         []util.IdValuePair `json:"priorityList"`
	ProgramList          []util.IdValuePair `json:"programList"`
	ProvidersList        []util.IdValuePair `json:"providersList"`
	ReferringSiteList    []util.IdValuePair `json:"referringSiteList"`
	TestLocationCodeList []util.IdValuePair `json:"testLocationCodeList"`

	EnvironmentalFields map[string]any `json:"environmentalFields"`
	IsEQASample         bool           `json:"isEQASample"`
	Modified            bool           `json:"modified"`
	ReadOnly            bool           `json:"readOnly"`
}
