package baseproject

import (
	"embed"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/zentooling/golang-web-server/admin"
	"github.com/zentooling/golang-web-server/infra"
	"github.com/zentooling/golang-web-server/login"
	"github.com/zentooling/golang-web-server/middleware"
	"github.com/zentooling/golang-web-server/routes"
)

// staticFS is an embedded file system
//
//go:embed web/*
var staticFS embed.FS

// Run is the main function that runs the entire package and starts the webserver, this is called by /cmd/base/main.go
func Run() {

	// We load environment variables, these are only read when the application launches
	conf := infra.LoadEnvVariables()

	// Init Logging and save leg level var - configurable at runtime on config page
	logLevelVar := infra.InitLogging(conf)

	// Load Translations
	bundle := infra.LoadLanguageBundles()

	// We connect to the database using the configuration generated from the environment variables.
	db, err := infra.ConnectToDatabase(conf)
	if err != nil {
		slog.Error("Run", "error", err)
		os.Exit(2)
	}

	// set global app-wide cfg parms
	infra.InitLair(db, conf, bundle, logLevelVar)

	// Once a database connection is established we run any needed migrations
	err = infra.MigrateDatabase(db)
	if err != nil {
		slog.Error("Run", "error", err)
		os.Exit(3)
	}
	// t will hold all our html templates used to render pages
	var t *template.Template

	// We parse and load the html files into our t variable
	t, err = loadTemplates()
	if err != nil {
		slog.Error("Run", "error", err)
		os.Exit(4)
	}

	// A gin Engine instance with the default configuration
	r := gin.Default()
	// proxies below should be correct for most internal networks
	err = r.SetTrustedProxies([]string{"127.0.0.1", "192.168.1.1/24", "10.0.0.0/8"})
	if err != nil {
		slog.Error("Run", "error", err)
		os.Exit(5)
	}

	// We create a new cookie store with a key used to secure cookies with HMAC
	store := cookie.NewStore([]byte(conf.CookieSecret))

	// We define our session middleware to be used globally on all routes
	r.Use(sessions.Sessions("golang_base_project_session", store))

	//// log request/responses to see info about bots hitting us
	//r.Use(middleware.RequestLogger())
	//r.Use(middleware.ResponseLogger())

	// We pass our template variable t to the gin engine so it can be used to render html pages
	r.SetHTMLTemplate(t)

	// our assets are only located in a section of our file system. so we create a sub file system.
	subFS, err := fs.Sub(staticFS, "dist/assets")
	if err != nil {
		slog.Error("Run", "error", err)
		os.Exit(6)
	}

	// All static assets are under the /assets path so we make this its own group called assets
	assets := r.Group("/assets")

	// This middleware sets the Cache-Control header and is applied to the assets group only
	assets.Use(middleware.Cache(conf.CacheMaxAge))

	// All requests to /assets will use the sub fil system which contains all our static assets
	assets.StaticFS("/", http.FS(subFS))

	// Session middleware is applied to all groups after this point.
	r.Use(middleware.Session(db))

	// A General middleware is defined to add default headers to improve site security
	r.Use(middleware.General())

	ctx := infra.LairInstance()
	loginSvc := login.NewService(ctx)
	adminSvc := admin.NewService(ctx)
	routeSvc := routes.NewService(ctx)

	// Any request to / will call controller.Index
	r.GET("/", routeSvc.Index)

	// We want to handle both POST and GET requests on the /search route. We define both but use the same function to handle the requests.
	r.GET("/search", routeSvc.Search)
	r.POST("/search", routeSvc.Search)
	r.Any("/search/:page", routeSvc.Search)
	r.Any("/search/:page/:query", routeSvc.Search)

	// We define our 404 handler for when a page can not be found
	r.NoRoute(routeSvc.NoRoute)

	// noAuth is a group for routes which should only be accessed if the user is not authenticated
	noAuth := r.Group("/")
	noAuth.Use(middleware.NoAuth())

	// instantiate login service with lair interface instance - can
	// use different implementation for testing

	noAuth.GET("/login", loginSvc.Login)
	noAuth.GET("/register", loginSvc.Register)
	noAuth.GET("/activate/resend", loginSvc.ResendActivation)
	noAuth.GET("/activate/:token", loginSvc.Activate)
	noAuth.GET("/user/password/forgot", loginSvc.ForgotPassword)
	noAuth.GET("/user/password/reset/:token", loginSvc.ResetPassword)

	// We make a separate group for our post requests on the same endpoints so that we can define our throttling middleware on POST requests only.
	noAuthPost := noAuth.Group("/")
	noAuthPost.Use(middleware.Throttle(conf.RequestsPerMinute))

	noAuthPost.POST("/loglevel", adminSvc.LoggingRouteHandlerPost)
	noAuthPost.POST("/login", loginSvc.LoginPost)
	noAuthPost.POST("/register", loginSvc.RegisterPost)
	noAuthPost.POST("/activate/resend", loginSvc.ResendActivationPost)
	noAuthPost.POST("/user/password/forgot", loginSvc.ForgotPasswordPost)
	noAuthPost.POST("/user/password/reset/:token", loginSvc.ResetPasswordPost)

	// the adminGroup group handles routes that should only be accessible to authenticated Admin users
	adminGroup := r.Group("/")
	adminGroup.Use(middleware.Admin())
	adminGroup.Use(middleware.Sensitive())

	adminGroup.GET("/config", adminSvc.ConfigRouteHandler)
	adminGroup.POST("/config", adminSvc.ConfigRouteHandlerPost)
	adminGroup.GET("/admin", adminSvc.Admin)
	// We need to handle post from the login redirect
	adminGroup.POST("/admin", adminSvc.Admin)

	// this group is for the main application which does not require admin privs
	authGroup := r.Group("/")
	authGroup.Use(middleware.Auth())
	authGroup.Use(middleware.Sensitive())
	authGroup.GET("/logout", loginSvc.Logout)

	// This starts our webserver, our application will not stop running or go past this point unless
	// an error occurs or the web server is stopped for some reason. It is designed to run forever.
	err = r.Run(":" + conf.Port)
	if err != nil {
		slog.Error("Run", "error", err)
		os.Exit(7)
	}
}
