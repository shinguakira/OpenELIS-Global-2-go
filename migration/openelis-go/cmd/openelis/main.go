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

	// auth layers (P0 Foundations)
	authrest "openelis-go/internal/auth/controller/rest"
	authdaoimpl "openelis-go/internal/auth/daoimpl"
	authmiddleware "openelis-go/internal/auth/middleware"
	authservice "openelis-go/internal/auth/service"
	authsession "openelis-go/internal/auth/session"

	commondaoimpl "openelis-go/internal/common/daoimpl"
	"openelis-go/internal/common/db"
	"openelis-go/internal/common/i18n"
	commonrest "openelis-go/internal/common/rest"
	commonservices "openelis-go/internal/common/services"
	"openelis-go/internal/common/web"

	// dictionarycategory layers
	dictcatrest "openelis-go/internal/dictionarycategory/controller/rest"
	dictcatdaoimpl "openelis-go/internal/dictionarycategory/daoimpl"
	dictcatservice "openelis-go/internal/dictionarycategory/service"

	// localization layers (a2)
	localizationrest "openelis-go/internal/localization/controller/rest"
	localizationdao "openelis-go/internal/localization/daoimpl"
	localizationservice "openelis-go/internal/localization/service"

	// organization layers (b2)
	orgrest "openelis-go/internal/organization/controller/rest"
	orgdaoimpl "openelis-go/internal/organization/daoimpl"
	orgservice "openelis-go/internal/organization/service"

	// panel layers
	paneldaoimpl "openelis-go/internal/panel/daoimpl"
	panelservice "openelis-go/internal/panel/service"

	// patient layers (c1)
	patientrest "openelis-go/internal/patient/controller/rest"
	patientdaoimpl "openelis-go/internal/patient/daoimpl"
	patientservice "openelis-go/internal/patient/service"

	// sample layers (c2)
	samplerest "openelis-go/internal/sample/controller/rest"
	sampledaoimpl "openelis-go/internal/sample/daoimpl"
	sampleservice "openelis-go/internal/sample/service"

	// provider layers (b2)
	providerrest "openelis-go/internal/provider/controller/rest"
	providerdaoimpl "openelis-go/internal/provider/daoimpl"
	providerservice "openelis-go/internal/provider/service"

	// system (a1)
	systemrest "openelis-go/internal/system/controller/rest"

	// test domain layers (TestSection)
	testdaoimpl "openelis-go/internal/test/daoimpl"
	testservice "openelis-go/internal/test/service"

	// testcalculated (a2)
	calculatedrest "openelis-go/internal/testcalculated/controller/rest"

	// testcatalog editor controller
	testcatalogrest "openelis-go/internal/testcatalog/controller/rest"
	testcatalogservice "openelis-go/internal/testcatalog/service"

	// testconfiguration layers (TestCatalog)
	testconfigrest "openelis-go/internal/testconfiguration/controller/rest"
	testconfigdaoimpl "openelis-go/internal/testconfiguration/daoimpl"
	testconfigservice "openelis-go/internal/testconfiguration/service"

	// typeofsample layers
	tosdaoimpl "openelis-go/internal/typeofsample/daoimpl"
	tosservice "openelis-go/internal/typeofsample/service"

	// unitofmeasure layers
	uomrest "openelis-go/internal/unitofmeasure/controller/rest"
	uomdaoimpl "openelis-go/internal/unitofmeasure/daoimpl"
	uomservice "openelis-go/internal/unitofmeasure/service"
)

func main() {
	addr := os.Getenv("OE_GO_ADDR")
	if addr == "" {
		addr = ":8090"
	}

	mux := http.NewServeMux()

	// ANONYMOUS by Java's rule: "/health/**" is listed in
	// SecurityConfig.OPEN_PAGES. Registered straight on the mux (not through
	// web.Register) because it is not a /rest path and needs no context-path
	// alias.
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

	// -----------------------------------------------------------------------
	// P0 Foundations: authentication.
	//
	// This block must run before the service starts listening, and its failure
	// must be fatal: web.Register is DEFAULT-DENY and refuses every protected
	// route until a Protector is installed, so a process that reached
	// ListenAndServe without this would serve nothing but 500s. Failing loudly
	// here is the point — the alternative (degrading to anonymous access) is
	// how PHI leaks.
	//
	// See migration/auth-adoption-plan.md. Java's equivalent is
	// SecurityConfig.defaultSecurityConfigurationFilterChain, which has no
	// securityMatcher and ends in anyRequest().authenticated().
	// -----------------------------------------------------------------------
	sessionStore := authsession.NewMemoryStore()
	// Expire sessions on a timer, independently of any request. GET /session is
	// public and creates a session on every cookie-less call, so without this an
	// anonymous caller grows the store without bound. Tomcat does the same via
	// its container background process; the port is not free to skip it.
	stopReaper := sessionStore.StartReaper(authsession.DefaultReapInterval)
	defer stopReaper()

	authModuleDAO := &authdaoimpl.ModuleDAOImpl{DB: gormDB}
	authSvc := &authservice.AuthService{
		LoginDAO:  &authdaoimpl.LoginDAOImpl{DB: gormDB},
		RoleDAO:   &authdaoimpl.RoleDAOImpl{DB: gormDB},
		ModuleDAO: authModuleDAO,
	}

	// Refuse to start under an authorization model this service does not
	// implement, rather than silently applying the Role rules to a deployment
	// Java would evaluate differently. See EffectivePermissionsAgent for the
	// resolution order and for why the DB row alone is not a complete check.
	agentOverride, _, err := authModuleDAO.PermissionsAgentOverride()
	if err != nil {
		log.Fatalf("cannot read the permissions.agent configuration: %v", err)
	}
	agent, err := authservice.EffectivePermissionsAgent(os.Getenv("OE_PERMISSIONS_AGENT"), agentOverride)
	if err != nil {
		log.Fatalf("SECURITY: %v", err)
	}
	log.Printf("authorization model: permissions.agent=%s", agent)

	// AuthzService ports ModuleAuthenticationInterceptor. It is wired into the
	// Guard rather than onto individual routes because Java registers that
	// interceptor on /** — so a future ported endpoint that HAS a
	// system_module_url row gets checked without anyone remembering to.
	authzSvc := &authservice.AuthzService{ModuleDAO: authModuleDAO}
	web.UseProtector(&authmiddleware.Guard{Store: sessionStore, Authz: authzSvc})
	authrest.Routes(mux, &authrest.LoginRestController{
		Service: authSvc,
		Store:   sessionStore,
	})
	log.Printf("auth enabled: default-deny on every registered route " +
		"(POST ValidateLogin, GET session, POST Logout are open per Java LOGIN_PAGES)")

	// a2: rest/supportedlocales{,/active,/fallback}
	svc := &localizationservice.SupportedLocaleService{
		DAO: &localizationdao.SupportedLocaleDAO{DB: gormDB},
	}
	localizationrest.Routes(mux, svc)
	log.Printf("DB-backed routes enabled (supportedlocales)")

	msgs := i18n.Messages()
	statusDAO := &commondaoimpl.StatusDAOImpl{DB: gormDB}
	// Hoisted out of the if/else below because c1's merge/details also needs it
	// (to resolve the analysis statuses excluded from totalResults). Stays nil
	// if construction fails, and every consumer must handle that.
	var statusSvc *commonservices.StatusService
	if svc, err := commonservices.NewStatusService(statusDAO, msgs); err != nil {
		log.Printf("WARN: status service init failed (%v); status-type routes disabled", err)
	} else {
		statusSvc = svc
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

	// testcatalog editor (lab-units, sample-types, panels) — aggregates the
	// three services above; see internal/testcatalog/service for why this is
	// its own service layer rather than the controller calling them directly.
	testcatalogEditorSvc := &testcatalogservice.TestCatalogEditorService{
		TestSectionService:  testSectionSvc,
		TypeOfSampleService: tosSvc,
		PanelService:        panelSvc,
	}
	testcatalogrest.Routes(mux, &testcatalogrest.TestCatalogEditorRestController{Service: testcatalogEditorSvc})

	// testconfiguration: TestCatalog (full catalog read)
	testconfigDAO := &testconfigdaoimpl.TestCatalogDAOImpl{DB: gormDB}
	testconfigSvc := &testconfigservice.TestCatalogService{DAO: testconfigDAO}
	testconfigrest.Routes(mux, &testconfigrest.TestCatalogRestController{Service: testconfigSvc})

	log.Printf("DB-backed routes enabled (b1: dictionary-categories, uom, test-catalog, TestCatalog)")

	// -----------------------------------------------------------------------
	// b2: organization + provider reference reads.
	// -----------------------------------------------------------------------

	// organization
	orgDAO := &orgdaoimpl.OrganizationDAOImpl{DB: gormDB}
	orgSvc := &orgservice.OrganizationService{DAO: orgDAO}
	orgrest.Routes(mux, &orgrest.OrganizationRestController{Service: orgSvc})

	// provider
	providerDAO := &providerdaoimpl.ProviderDAOImpl{DB: gormDB}
	providerSvc := &providerservice.ProviderService{DAO: providerDAO}
	providerrest.Routes(mux, &providerrest.ProviderRestController{Service: providerSvc})

	log.Printf("DB-backed routes enabled (b2: organization, provider)")

	// -----------------------------------------------------------------------
	// c1: patient reads — the first wave that serves PHI (names, birth dates,
	// national IDs, addresses, phones, email).
	//
	// These are protected by construction: every route below goes through
	// web.Register, which is DEFAULT-DENY since P0 auth landed, so an
	// unauthenticated caller gets Java's 302 to /LoginPage and no data. The
	// additional "Reception" role gate Java applies to merge/details is
	// declared at the route itself — see internal/patient/controller/rest.
	//
	// The service nonetheless stays bound to loopback in docker-compose.go.yml:
	// that is now about session sharing during strangler coexistence
	// (auth-adoption-plan.md §8.1 — Java and Go do not share sessions), not
	// about the absence of an auth layer.
	// -----------------------------------------------------------------------
	patientDAO := &patientdaoimpl.PatientDAOImpl{DB: gormDB}
	patientSvc := &patientservice.PatientService{DAO: patientDAO}
	// merge/details needs the status service to resolve which analysis
	// statuses are excluded from dataSummary.totalResults, exactly as Java's
	// countResultsForPatient does via IStatusService. Without it the count
	// includes every analysis and diverges from Java — measured at 28 vs 0 on
	// the dev dataset, because every analysis there is "Not Tested", an
	// excluded status.
	//
	// FATAL, not a warning. An earlier revision logged and carried on with a
	// nil resolver, which meant a transient failure reading status_of_sample
	// turned into a 200 carrying a knowingly WRONG patient summary — for as
	// long as the process lived, with nothing but a startup log line to say so.
	// Wrong clinical data served confidently is worse than an outage, and it is
	// the same fail-closed rule the rest of this service already follows
	// (web.Register refuses without a Protector; startup refuses an unsupported
	// permissions.agent). Java has no degraded mode here either: its
	// IStatusService is a Spring bean, so a failure to build it fails the whole
	// context.
	if statusSvc == nil {
		log.Fatalf("status service unavailable; refusing to serve c1 merge/details," +
			" whose totalResults depends on it (see the WARN above for the cause)")
	}
	patientMergeSvc := &patientservice.PatientMergeService{DAO: patientDAO, Status: statusSvc}
	patientrest.Routes(mux, &patientrest.PatientRestController{
		Service:      patientSvc,
		MergeService: patientMergeSvc,
	})
	log.Printf("DB-backed routes enabled (c1: patient reads — PHI, authenticated; merge/details additionally requires the Reception role)")

	// -----------------------------------------------------------------------
	// c2: sample + order reads.
	// -----------------------------------------------------------------------
	sampleSvc := &sampleservice.SampleService{
		DAO:    &sampledaoimpl.SampleDAOImpl{DB: gormDB},
		Status: statusSvc,
	}
	samplerest.Routes(mux, &samplerest.SampleRestController{Service: sampleSvc})
	samplerest.PendingAnalysisRoutes(mux, &samplerest.PendingAnalysisRestController{
		Service: &sampleservice.PendingAnalysisService{
			DAO:    &sampledaoimpl.SampleDAOImpl{DB: gormDB},
			Status: statusSvc,
		},
	})
	samplerest.OrderAttachmentRoutes(mux, &samplerest.OrderAttachmentRestController{Service: sampleSvc})
	samplerest.UnassignedSampleRoutes(mux, &samplerest.UnassignedSampleRestController{
		Service: &sampleservice.UnassignedSampleService{DAO: &sampledaoimpl.UnassignedSampleDAOImpl{DB: gormDB}},
	})
	samplerest.OrderDashboardRoutes(mux, &samplerest.OrderDashboardRestController{
		Service: &sampleservice.OrderDashboardService{DAO: &sampledaoimpl.SampleDAOImpl{DB: gormDB}},
	})
	log.Printf("DB-backed routes enabled (c2: sample reads)")

	srv := &http.Server{Addr: addr, Handler: mux}
	log.Printf("openelis-go listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
