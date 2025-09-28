// Package main Fluxio API Server
//
// Servidor de API para autenticación y gestión de usuarios
//
//	Schemes: http
//	Host: localhost:8080
//	BasePath: /
//	Version: 1.0.0
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Security:
//	- bearer
//
// swagger:meta
package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/Osminalx/fluxio/docs"
	"github.com/Osminalx/fluxio/internal/api"
	"github.com/Osminalx/fluxio/internal/auth"
	"github.com/Osminalx/fluxio/internal/db"
	"github.com/Osminalx/fluxio/internal/middleware"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
	"github.com/joho/godotenv"
)

// @title Fluxio API
// @version 1.0
// @description API de autenticación y gestión de usuarios con GORM y JWT
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http

// @securityDefinitions.apikey bearerAuth
// @in header
// @name Authorization
// @description Ingresa "Bearer" seguido de un espacio y el token JWT

// handleIncomeRoutes maneja el enrutamiento para los endpoints de income
func handleIncomeRoutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	
	switch {
	case path == "/api/v1/incomes":
		switch r.Method {
		case http.MethodGet:
			api.GetAllIncomesHandler(w, r)
		case http.MethodPost:
			api.CreateIncomeHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/incomes/active":
		if r.Method == http.MethodGet {
			api.GetActiveIncomesHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/incomes/deleted":
		if r.Method == http.MethodGet {
			api.GetDeletedIncomesHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/incomes/") && strings.HasSuffix(path, "/restore"):
		if r.Method == http.MethodPost {
			api.RestoreIncomeHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/incomes/") && strings.HasSuffix(path, "/status"):
		if r.Method == http.MethodPatch {
			api.ChangeIncomeStatusHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/incomes/"):
		// Endpoints con ID individual: /api/v1/incomes/{id}
		switch r.Method {
		case http.MethodGet:
			api.GetIncomeByIDHandler(w, r)
		case http.MethodPatch:
			api.UpdateIncomeHandler(w, r)
		case http.MethodDelete:
			api.DeleteIncomeHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// handleExpenseRoutes manages routing for expense endpoints
func handleExpenseRoutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	
	switch {
	case path == "/api/v1/expenses":
		switch r.Method {
		case http.MethodGet:
			api.GetAllExpensesHandler(w, r)
		case http.MethodPost:
			api.CreateExpenseHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/expenses/active":
		if r.Method == http.MethodGet {
			api.GetActiveExpensesHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/expenses/deleted":
		if r.Method == http.MethodGet {
			api.GetDeletedExpensesHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/expenses/date-range":
		if r.Method == http.MethodGet {
			api.GetExpensesByDateRangeHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/expenses/monthly":
		if r.Method == http.MethodGet {
			api.GetMonthlyExpensesHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/expenses/summary":
		if r.Method == http.MethodGet {
			api.GetExpensesSummaryHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/expenses/category/"):
		if r.Method == http.MethodGet {
			api.GetExpensesByCategoryHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/expenses/bank-account/"):
		if r.Method == http.MethodGet {
			api.GetExpensesByBankAccountHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/expenses/") && strings.HasSuffix(path, "/restore"):
		if r.Method == http.MethodPost {
			api.RestoreExpenseHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/expenses/") && strings.HasSuffix(path, "/status"):
		if r.Method == http.MethodPatch {
			api.ChangeExpenseStatusHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/expenses/"):
		switch r.Method {
		case http.MethodGet:
			api.GetExpenseByIDHandler(w, r)
		case http.MethodPatch:
			api.UpdateExpenseHandler(w, r)
		case http.MethodDelete:
			api.DeleteExpenseHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// handleBudgetRoutes manages routing for budget endpoints
func handleBudgetRoutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	
	switch {
	case path == "/api/v1/budgets":
		switch r.Method {
		case http.MethodGet:
			api.GetAllBudgetsHandler(w, r)
		case http.MethodPost:
			api.CreateBudgetHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/budgets/active":
		if r.Method == http.MethodGet {
			api.GetActiveBudgetsHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/budgets/deleted":
		if r.Method == http.MethodGet {
			api.GetDeletedBudgetsHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/budgets/by-month":
		if r.Method == http.MethodGet {
			api.GetBudgetByMonthYearHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/budgets/") && strings.HasSuffix(path, "/restore"):
		if r.Method == http.MethodPost {
			api.RestoreBudgetHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/budgets/") && strings.HasSuffix(path, "/status"):
		if r.Method == http.MethodPatch {
			api.ChangeBudgetStatusHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/budgets/"):
		switch r.Method {
		case http.MethodGet:
			api.GetBudgetByIDHandler(w, r)
		case http.MethodPatch:
			api.UpdateBudgetHandler(w, r)
		case http.MethodDelete:
			api.DeleteBudgetHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// handleBankAccountRoutes manages routing for bank account endpoints
func handleBankAccountRoutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	
	switch {
	case path == "/api/v1/bank-accounts":
		switch r.Method {
		case http.MethodGet:
			api.GetAllBankAccountsHandler(w, r)
		case http.MethodPost:
			api.CreateBankAccountHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/bank-accounts/active":
		if r.Method == http.MethodGet {
			api.GetActiveBankAccountsHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/bank-accounts/deleted":
		if r.Method == http.MethodGet {
			api.GetDeletedBankAccountsHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/bank-accounts/") && strings.HasSuffix(path, "/restore"):
		if r.Method == http.MethodPost {
			api.RestoreBankAccountHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/bank-accounts/") && strings.HasSuffix(path, "/status"):
		if r.Method == http.MethodPatch {
			api.ChangeBankAccountStatusHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/bank-accounts/"):
		switch r.Method {
		case http.MethodGet:
			api.GetBankAccountByIDHandler(w, r)
		case http.MethodPatch:
			api.UpdateBankAccountHandler(w, r)
		case http.MethodDelete:
			api.DeleteBankAccountHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// handleFixedExpenseRoutes manages routing for fixed expense endpoints
func handleFixedExpenseRoutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	
	switch {
	case path == "/api/v1/fixed-expenses":
		switch r.Method {
		case http.MethodGet:
			api.GetAllFixedExpensesHandler(w, r)
		case http.MethodPost:
			api.CreateFixedExpenseHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/fixed-expenses/"):
		switch r.Method {
		case http.MethodGet:
			api.GetFixedExpenseByIDHandler(w, r)
		case http.MethodPatch:
			api.UpdateFixedExpenseHandler(w, r)
		case http.MethodDelete:
			api.DeleteFixedExpenseHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// handleBudgetHistoryRoutes manages routing for budget history endpoints
func handleBudgetHistoryRoutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	
	switch {
	case path == "/api/v1/budget-history":
		if r.Method == http.MethodGet {
			api.GetAllBudgetHistory(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/budget-history/date-range":
		if r.Method == http.MethodGet {
			api.GetBudgetHistoryByDateRange(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/budget-history/reasons":
		if r.Method == http.MethodGet {
			api.GetBudgetHistoryWithReasons(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/budget-history/stats":
		if r.Method == http.MethodGet {
			api.GetBudgetHistoryStats(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/budget-history/patterns":
		if r.Method == http.MethodGet {
			api.AnalyzeBudgetPatterns(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/budgets/") && strings.HasSuffix(path, "/history"):
		if r.Method == http.MethodGet {
			api.GetBudgetHistoryByBudgetID(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/budget-history/"):
		if r.Method == http.MethodGet {
			api.GetBudgetHistoryByID(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// handleGoalRoutes manages routing for goal endpoints
func handleGoalRoutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	
	switch {
	case path == "/api/v1/goals":
		switch r.Method {
		case http.MethodGet:
			api.GetAllGoalsHandler(w, r)
		case http.MethodPost:
			api.CreateGoalHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/goals/active":
		if r.Method == http.MethodGet {
			api.GetActiveGoalsHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/goals/deleted":
		if r.Method == http.MethodGet {
			api.GetDeletedGoalsHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/goals/") && strings.HasSuffix(path, "/restore"):
		if r.Method == http.MethodPost {
			api.RestoreGoalHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/goals/") && strings.HasSuffix(path, "/status"):
		if r.Method == http.MethodPatch {
			api.ChangeGoalStatusHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/goals/"):
		switch r.Method {
		case http.MethodGet:
			api.GetGoalByIDHandler(w, r)
		case http.MethodPatch:
			api.UpdateGoalHandler(w, r)
		case http.MethodDelete:
			api.DeleteGoalHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// handleUserCategoryRoutes manages routing for user category endpoints
func handleUserCategoryRoutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	
	switch {
	case path == "/api/v1/user-categories":
		switch r.Method {
		case http.MethodGet:
			api.GetUserCategories(w, r)
		case http.MethodPost:
			api.CreateUserCategory(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/user-categories/grouped":
		if r.Method == http.MethodGet {
			api.GetUserCategoriesGroupedByType(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/user-categories/defaults":
		if r.Method == http.MethodPost {
			api.CreateDefaultUserCategories(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/user-categories/stats":
		if r.Method == http.MethodGet {
			api.GetUserCategoryStats(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/user-categories/expense-type/"):
		if r.Method == http.MethodGet {
			api.GetUserCategoriesByExpenseType(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/user-categories/expense-type-name/"):
		if r.Method == http.MethodGet {
			api.GetUserCategoriesByExpenseTypeName(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/user-categories/") && strings.HasSuffix(path, "/restore"):
		if r.Method == http.MethodPost {
			api.RestoreUserCategory(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/user-categories/"):
		switch r.Method {
		case http.MethodGet:
			api.GetUserCategoryByID(w, r)
		case http.MethodPut:
			api.UpdateUserCategory(w, r)
		case http.MethodDelete:
			api.SoftDeleteUserCategory(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// handleExpenseTypeRoutes manages routing for expense type endpoints
func handleExpenseTypeRoutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	
	switch {
	case path == "/api/v1/expense-types":
		if r.Method == http.MethodGet {
			api.GetAllExpenseTypes(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/expense-types/with-categories":
		if r.Method == http.MethodGet {
			api.GetExpenseTypesWithUserCategories(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/expense-types/name/"):
		if r.Method == http.MethodGet {
			api.GetExpenseTypeByName(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/expense-types/"):
		if r.Method == http.MethodGet {
			api.GetExpenseTypeByID(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// handleSetupRoutes manages routing for system setup endpoints
func handleSetupRoutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	
	switch {
	case path == "/api/v1/setup/initialize":
		if r.Method == http.MethodPost {
			api.InitializeExpenseSystem(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/setup/user":
		if r.Method == http.MethodPost {
			api.SetupNewUser(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/setup/overview":
		if r.Method == http.MethodGet {
			api.GetSystemOverview(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// handleReminderRoutes manages routing for reminder endpoints
func handleReminderRoutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	
	switch {
	case path == "/api/v1/reminders":
		switch r.Method {
		case http.MethodGet:
			api.GetAllRemindersHandler(w, r)
		case http.MethodPost:
			api.CreateReminderHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/reminders/overdue":
		if r.Method == http.MethodGet {
			api.GetOverdueRemindersHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/reminders/stats":
		if r.Method == http.MethodGet {
			api.GetReminderStatsHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/reminders/") && strings.HasSuffix(path, "/complete"):
		if r.Method == http.MethodPost {
			api.CompleteReminderHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/reminders/"):
		switch r.Method {
		case http.MethodGet:
			api.GetReminderByIDHandler(w, r)
		case http.MethodPatch:
			api.UpdateReminderHandler(w, r)
		case http.MethodDelete:
			api.DeleteReminderHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// handleTransferRoutes manages routing for transfer endpoints
func handleTransferRoutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	
	switch {
	case path == "/api/v1/transfers":
		switch r.Method {
		case http.MethodGet:
			api.GetAllTransfersHandler(w, r)
		case http.MethodPost:
			api.CreateTransferHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case path == "/api/v1/transfers/stats":
		if r.Method == http.MethodGet {
			api.GetTransferStatsHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/transfers/account/"):
		if r.Method == http.MethodGet {
			api.GetTransfersByAccountHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	case strings.HasPrefix(path, "/api/v1/transfers/"):
		switch r.Method {
		case http.MethodGet:
			api.GetTransferByIDHandler(w, r)
		case http.MethodPatch:
			api.UpdateTransferHandler(w, r)
		case http.MethodDelete:
			api.DeleteTransferHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize Swagger docs
	docs.SwaggerInfo.Title = "Fluxio API"
	docs.SwaggerInfo.Description = "API de autenticación y gestión de usuarios con GORM y JWT"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http"}

	// Connect to database
	logger.Info("🗄️  Conectando a la base de datos...")
	db.Connect()
	logger.Info("✅ Conectado a Postgres con GORM")

	// Create main router
	mux := http.NewServeMux()
	
	// We'll wrap the entire mux with logging middleware at the end

	// API v1 routes - PUBLIC (no authentication required)
	mux.HandleFunc("/api/v1/hello", api.HelloHandler)
	mux.HandleFunc("/api/v1/auth/login", api.LoginHandler)
	mux.HandleFunc("/api/v1/auth/register", api.RegisterHandler)
	mux.HandleFunc("/api/v1/auth/refresh", api.RefreshTokenHandler)
	mux.HandleFunc("/api/v1/auth/logout", api.LogoutHandler)
	mux.HandleFunc("/api/v1/auth/logout-all", api.LogoutAllHandler)
	
	
	
	// Expense Types endpoints - PUBLIC (read-only, no auth needed for basic info)
	mux.HandleFunc("/api/v1/expense-types", handleExpenseTypeRoutes)
	mux.HandleFunc("/api/v1/expense-types/", handleExpenseTypeRoutes)
	
	// Setup endpoints - PUBLIC (system initialization)
	mux.HandleFunc("/api/v1/setup/", handleSetupRoutes)


	// API v1 routes - PROTECTED (require authentication)
	protectedMux := http.NewServeMux()
	
	// Auth endpoints - PROTECTED
	protectedMux.HandleFunc("/api/v1/auth/me", api.MeHandler)
	
	// Income endpoints - PROTECTED
	protectedMux.HandleFunc("/api/v1/incomes", handleIncomeRoutes)
	protectedMux.HandleFunc("/api/v1/incomes/", handleIncomeRoutes)
	
	// Expense endpoints - PROTECTED
	protectedMux.HandleFunc("/api/v1/expenses", handleExpenseRoutes)
	protectedMux.HandleFunc("/api/v1/expenses/", handleExpenseRoutes)
	
	// Budget endpoints - PROTECTED
	protectedMux.HandleFunc("/api/v1/budgets", handleBudgetRoutes)
	protectedMux.HandleFunc("/api/v1/budgets/", handleBudgetRoutes)
	
	// Bank Account endpoints - PROTECTED
	protectedMux.HandleFunc("/api/v1/bank-accounts", handleBankAccountRoutes)
	protectedMux.HandleFunc("/api/v1/bank-accounts/", handleBankAccountRoutes)
	
	// Fixed Expense endpoints - PROTECTED
	protectedMux.HandleFunc("/api/v1/fixed-expenses", handleFixedExpenseRoutes)
	protectedMux.HandleFunc("/api/v1/fixed-expenses/", handleFixedExpenseRoutes)
	
	// Budget History endpoints - PROTECTED
	protectedMux.HandleFunc("/api/v1/budget-history", handleBudgetHistoryRoutes)
	protectedMux.HandleFunc("/api/v1/budget-history/", handleBudgetHistoryRoutes)
	
	// Goal endpoints - PROTECTED
	protectedMux.HandleFunc("/api/v1/goals", handleGoalRoutes)
	protectedMux.HandleFunc("/api/v1/goals/", handleGoalRoutes)
	
	// User Category endpoints - PROTECTED
	protectedMux.HandleFunc("/api/v1/user-categories", handleUserCategoryRoutes)
	protectedMux.HandleFunc("/api/v1/user-categories/", handleUserCategoryRoutes)
	
	// Reminder endpoints - PROTECTED
	protectedMux.HandleFunc("/api/v1/reminders", handleReminderRoutes)
	protectedMux.HandleFunc("/api/v1/reminders/", handleReminderRoutes)
	
	// Transfer endpoints - PROTECTED
	protectedMux.HandleFunc("/api/v1/transfers", handleTransferRoutes)
	protectedMux.HandleFunc("/api/v1/transfers/", handleTransferRoutes)
	
	// Expense Types endpoints - PROTECTED (for endpoints that need user context)
	protectedMux.HandleFunc("/api/v1/expense-types", handleExpenseTypeRoutes)
	protectedMux.HandleFunc("/api/v1/expense-types/", handleExpenseTypeRoutes)
	
	// Apply auth middleware to protected API v1 routes
	mux.Handle("/api/v1/protected/", auth.AuthMiddleware(protectedMux))
	mux.Handle("/api/v1/auth/me", auth.AuthMiddleware(protectedMux))
	mux.Handle("/api/v1/incomes", auth.AuthMiddleware(protectedMux))
	mux.Handle("/api/v1/incomes/", auth.AuthMiddleware(protectedMux))
	mux.Handle("/api/v1/expenses", auth.AuthMiddleware(protectedMux))
	mux.Handle("/api/v1/expenses/", auth.AuthMiddleware(protectedMux))
	mux.Handle("/api/v1/budgets", auth.AuthMiddleware(protectedMux))
	mux.Handle("/api/v1/budgets/", auth.AuthMiddleware(protectedMux))
	mux.Handle("/api/v1/bank-accounts", auth.AuthMiddleware(protectedMux))
	mux.Handle("/api/v1/bank-accounts/", auth.AuthMiddleware(protectedMux))
	mux.Handle("/api/v1/fixed-expenses", auth.AuthMiddleware(protectedMux))
	mux.Handle("/api/v1/fixed-expenses/", auth.AuthMiddleware(protectedMux))
	mux.Handle("/api/v1/budget-history", auth.AuthMiddleware(protectedMux))
	mux.Handle("/api/v1/budget-history/", auth.AuthMiddleware(protectedMux))
	mux.Handle("/api/v1/goals", auth.AuthMiddleware(protectedMux))
	mux.Handle("/api/v1/goals/", auth.AuthMiddleware(protectedMux))
	mux.Handle("/api/v1/user-categories", auth.AuthMiddleware(protectedMux))
	mux.Handle("/api/v1/user-categories/", auth.AuthMiddleware(protectedMux))
	mux.Handle("/api/v1/reminders", auth.AuthMiddleware(protectedMux))
	mux.Handle("/api/v1/reminders/", auth.AuthMiddleware(protectedMux))
	mux.Handle("/api/v1/transfers", auth.AuthMiddleware(protectedMux))
	mux.Handle("/api/v1/transfers/", auth.AuthMiddleware(protectedMux))
	mux.Handle("/api/v1/expense-types/with-categories", auth.AuthMiddleware(protectedMux))

	// Serve swagger.json file
	mux.HandleFunc("/docs/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.ServeFile(w, r, "docs/swagger.json")
	})

	// Scalar API Reference (public access - no versioning needed)
	mux.HandleFunc("/reference", func(w http.ResponseWriter, r *http.Request) {
		htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
			SpecURL: "http://localhost:8080/docs/swagger.json",
			CustomOptions: scalar.CustomOptions{
				PageTitle: "Fluxio API Documentation",
			},
			DarkMode: true,
		})

		if err != nil {
			http.Error(w, "Error generating API documentation", http.StatusInternalServerError)
			logger.Error("Error generating Scalar documentation: %v", err)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(htmlContent))
	})
	
	// Health check endpoint (no versioning)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","version":"1.0"}`))
	})

	logger.Info("🚀 Server started on port: 8080")
	logger.Info("  GET  /reference - Scalar API Documentation")

	// Apply CORS and logging middleware to all routes
	allowedOrigins := []string{
		"http://172.16.0.2:3000",
		"http://localhost:3000",
	}
	
	handler := middleware.RestrictedCORSMiddleware(allowedOrigins)(middleware.LoggingMiddleware(mux))
	
	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		logger.Fatal("Error al iniciar el servidor: %v", err)
	}
}