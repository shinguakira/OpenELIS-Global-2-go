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
	if statusSvc, err := commonservices.NewStatusService(statusDAO, msgs); err != nil {
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

	srv := &http.Server{Addr: addr, Handler: mux}
	log.Printf("openelis-go listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
