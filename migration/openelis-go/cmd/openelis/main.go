// Command openelis is the minimal Go re-implementation of the OpenELIS backend.
// It serves ported REST endpoints beside the Java WAR; the nginx proxy routes
// individual paths here as each endpoint passes parity (strangler-fig).
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"gorm.io/gorm"

	"openelis-go/internal/common/db"
	"openelis-go/internal/common/i18n"
	commonrest "openelis-go/internal/common/rest"
	commonservices "openelis-go/internal/common/services"
	"openelis-go/internal/common/web"

	// dictionarycategory layers
	dictcatdaoimpl "openelis-go/internal/dictionarycategory/daoimpl"
	dictcatrest "openelis-go/internal/dictionarycategory/controller/rest"
	dictcatservice "openelis-go/internal/dictionarycategory/service"

	// localization layers (a2)
	localizationrest "openelis-go/internal/localization/controller/rest"
	localizationdao "openelis-go/internal/localization/daoimpl"
	localizationservice "openelis-go/internal/localization/service"

	// panel layers
	paneldaoimpl "openelis-go/internal/panel/daoimpl"
	panelservice "openelis-go/internal/panel/service"

	// system (a1)
	systemrest "openelis-go/internal/system/controller/rest"

	// test domain layers (TestSection)
	testdaoimpl "openelis-go/internal/test/daoimpl"
	testservice "openelis-go/internal/test/service"

	// testcalculated (a2)
	calculatedrest "openelis-go/internal/testcalculated/controller/rest"

	// testcatalog editor controller
	testcatalogrest "openelis-go/internal/testcatalog/controller/rest"

	// testconfiguration layers (TestCatalog)
	testconfigrest "openelis-go/internal/testconfiguration/controller/rest"
	testconfigservice "openelis-go/internal/testconfiguration/service"

	// typeofsample layers
	tosdaoimpl "openelis-go/internal/typeofsample/daoimpl"
	tosservice "openelis-go/internal/typeofsample/service"

	// unitofmeasure layers
	uomdaoimpl "openelis-go/internal/unitofmeasure/daoimpl"
	uomrest "openelis-go/internal/unitofmeasure/controller/rest"
	uomservice "openelis-go/internal/unitofmeasure/service"
)

func main() {
	addr := os.Getenv("OE_GO_ADDR")
	if addr == "" {
		addr = ":8090"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		web.WriteJSON(w, http.StatusOK, map[string]string{"status": "UP"})
	})

	// Register each domain's REST routes (mirrors Spring auto-discovering
	// @RestController beans). One line per ported domain.
	systemrest.Routes(mux)     // a1: rest/server-time
	calculatedrest.Routes(mux) // a2: rest/math-functions
	commonrest.Routes(mux)     // a2: rest/sample-item-status-types

	// Connect to Postgres via GORM. Retry so a slow Compose startup (where
	// the DB container starts before Postgres is ready) does not permanently
	// disable DB-backed routes. On exhaustion we log.Fatal so the restart
	// policy (restart: unless-stopped) retries the whole process.
	const maxAttempts = 10
	const retryDelay = 3 * time.Second

	var gormDB *gorm.DB
	for i := 1; i <= maxAttempts; i++ {
		conn, err := db.OpenGORM()
		if err == nil {
			gormDB = conn
			break
		}
		if i == maxAttempts {
			log.Fatalf("DB unavailable after %d attempts (%v); exiting so restart policy can retry", maxAttempts, err)
		}
		log.Printf("DB not ready (attempt %d/%d): %v — retrying in %s", i, maxAttempts, err, retryDelay)
		time.Sleep(retryDelay)
	}

	// a2 domains (localization, status-types) use *sql.DB — extract from GORM
	// to share the single underlying connection pool.
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("failed to extract *sql.DB from GORM: %v", err)
	}

	// a2: rest/supportedlocales{,/active,/fallback}
	svc := &localizationservice.SupportedLocaleService{
		DAO: &localizationdao.SupportedLocaleDAO{DB: sqlDB},
	}
	localizationrest.Routes(mux, svc)
	log.Printf("DB-backed routes enabled (supportedlocales)")

	msgs := i18n.Messages()
	if statusSvc, err := commonservices.NewStatusService(sqlDB, msgs); err != nil {
		log.Printf("WARN: status service init failed (%v); status-type routes disabled", err)
	} else {
		commonrest.StatusRoutes(mux, statusSvc)
		log.Printf("DB-backed routes enabled (status-types)")
	}

	// -----------------------------------------------------------------------
	// b1: dictionary + test-catalog reference reads.
	// All b1 DAOs receive *gorm.DB.
	// Wire each domain: DAO → service → controller, then register routes.
	// -----------------------------------------------------------------------

	// dictionarycategory
	dictcatDAO := &dictcatdaoimpl.DictionaryCategoryDAOImpl{DB: gormDB}
	dictcatSvc := &dictcatservice.DictionaryCategoryService{DAO: dictcatDAO}
	dictcatrest.Routes(mux, &dictcatrest.DictionaryMenuRestController{Service: dictcatSvc})

	// unitofmeasure
	uomDAO := &uomdaoimpl.UnitOfMeasureDAOImpl{DB: gormDB}
	uomTypeMapDAO := &uomdaoimpl.UomTypeMapDAOImpl{DB: gormDB}
	uomSvc := &uomservice.UnitOfMeasureService{UomDAO: uomDAO, UomTypeMapDAO: uomTypeMapDAO}
	uomrest.Routes(mux, &uomrest.UnitOfMeasureRestController{Service: uomSvc})

	// test domain: TestSection (used by TestCatalogEditor + TestCatalog)
	testSectionDAO := &testdaoimpl.TestSectionDAOImpl{DB: gormDB}
	testSectionSvc := &testservice.TestSectionService{DAO: testSectionDAO}

	// typeofsample
	tosDAO := &tosdaoimpl.TypeOfSampleDAOImpl{DB: gormDB}
	tosSvc := &tosservice.TypeOfSampleService{DAO: tosDAO}

	// panel
	panelDAO := &paneldaoimpl.PanelDAOImpl{DB: gormDB}
	panelSvc := &panelservice.PanelService{DAO: panelDAO}

	// testcatalog editor (lab-units, sample-types, panels)
	testcatalogrest.Routes(mux, &testcatalogrest.TestCatalogEditorRestController{
		TestSectionService:  testSectionSvc,
		TypeOfSampleService: tosSvc,
		PanelService:        panelSvc,
	})

	// testconfiguration: TestCatalog (full catalog read)
	testconfigSvc := &testconfigservice.TestCatalogService{DB: gormDB}
	testconfigrest.Routes(mux, &testconfigrest.TestCatalogRestController{Service: testconfigSvc})

	log.Printf("DB-backed routes enabled (b1: dictionary-categories, uom, test-catalog, TestCatalog)")

	srv := &http.Server{Addr: addr, Handler: mux}
	log.Printf("openelis-go listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
