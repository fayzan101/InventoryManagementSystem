package router

import (
	"encoding/json"
	"net/http"
	"strings"

	"myapp/internal/auth"
	"myapp/internal/categories"
	"myapp/internal/customers"
	"myapp/internal/employees"
	"myapp/internal/inventory"
	"myapp/internal/orders"
	"myapp/internal/products"
	"myapp/internal/reports"
	"myapp/internal/suppliers"
	"myapp/internal/swagger"
	"myapp/internal/transfers"
	"myapp/internal/warehouses"
	"myapp/internal/websocket"
)

func Setup(openAPISpec []byte) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/auth/login", auth.Login)
	mux.Handle("/auth/register", secure(http.HandlerFunc(auth.Register)))
	mux.Handle("/auth/me", secure(http.HandlerFunc(auth.Me)))

	swagger.Register(mux, openAPISpec)

	registerProductRoutes(mux)
	registerWarehouseRoutes(mux)
	registerInventoryRoutes(mux)
	registerSupplierRoutes(mux)
	registerOrderRoutes(mux)
	registerCustomerRoutes(mux)
	registerCategoryRoutes(mux)
	registerEmployeeRoutes(mux)
	registerTransferRoutes(mux)
	registerReportRoutes(mux)

	mux.HandleFunc("/ws/inventory", websocket.HandleWebSocket)
	mux.HandleFunc("/ws/warehouses", websocket.HandleWebSocket)
	mux.HandleFunc("/ws/products", websocket.HandleWebSocket)
	mux.HandleFunc("/ws/suppliers", websocket.HandleWebSocket)

	return mux
}

func secure(h http.Handler) http.Handler {
	return auth.Middleware(h)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "IMS Backend API v1.2",
	})
}

func registerProductRoutes(mux *http.ServeMux) {
	mux.Handle("/products", secure(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			products.CreateProduct(w, r)
		case http.MethodGet:
			products.ListProducts(w, r)
		default:
			methodNotAllowed(w)
		}
	})))
	mux.Handle("/products/search", secure(http.HandlerFunc(products.SearchProducts)))
	mux.Handle("/products/", secure(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/search") {
			products.SearchProducts(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			products.GetProduct(w, r)
		case http.MethodPut:
			products.UpdateProduct(w, r)
		case http.MethodDelete:
			products.DeleteProduct(w, r)
		default:
			methodNotAllowed(w)
		}
	})))
}

func registerWarehouseRoutes(mux *http.ServeMux) {
	mux.Handle("/warehouses", secure(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			warehouses.CreateWarehouse(w, r)
		case http.MethodGet:
			warehouses.ListWarehouses(w, r)
		default:
			methodNotAllowed(w)
		}
	})))
	mux.Handle("/warehouses/", secure(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			warehouses.GetWarehouse(w, r)
		case http.MethodPut:
			warehouses.UpdateWarehouse(w, r)
		case http.MethodDelete:
			warehouses.DeleteWarehouse(w, r)
		default:
			methodNotAllowed(w)
		}
	})))
}

func registerInventoryRoutes(mux *http.ServeMux) {
	mux.Handle("/inventory/adjust", secure(http.HandlerFunc(inventory.AdjustInventory)))
	mux.Handle("/inventory/low-stock", secure(http.HandlerFunc(inventory.GetLowStock)))
	mux.Handle("/inventory/movements", secure(http.HandlerFunc(inventory.GetStockMovements)))
	mux.Handle("/inventory/", secure(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			inventory.GetProductInventory(w, r)
		} else {
			methodNotAllowed(w)
		}
	})))
	mux.Handle("/inventory", secure(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			inventory.GetInventory(w, r)
		} else {
			methodNotAllowed(w)
		}
	})))
}

func registerSupplierRoutes(mux *http.ServeMux) {
	mux.Handle("/suppliers", secure(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			suppliers.CreateSupplier(w, r)
		case http.MethodGet:
			suppliers.ListSuppliers(w, r)
		default:
			methodNotAllowed(w)
		}
	})))
	mux.Handle("/suppliers/", secure(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			suppliers.UpdateSupplier(w, r)
		} else {
			methodNotAllowed(w)
		}
	})))
}

func registerOrderRoutes(mux *http.ServeMux) {
	mux.Handle("/purchase-orders", secure(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			orders.CreatePurchaseOrder(w, r)
		case http.MethodGet:
			orders.ListPurchaseOrders(w, r)
		default:
			methodNotAllowed(w)
		}
	})))
	mux.Handle("/purchase-orders/", secure(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/receive") && r.Method == http.MethodPut {
			orders.ReceivePurchaseOrder(w, r)
		} else {
			methodNotAllowed(w)
		}
	})))
	mux.Handle("/orders", secure(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			orders.CreateOrder(w, r)
		case http.MethodGet:
			orders.ListOrders(w, r)
		default:
			methodNotAllowed(w)
		}
	})))
	mux.Handle("/orders/", secure(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/status") && r.Method == http.MethodPut {
			orders.UpdateOrderStatus(w, r)
			return
		}
		if r.Method == http.MethodGet {
			orders.GetOrder(w, r)
			return
		}
		methodNotAllowed(w)
	})))
}

func registerCustomerRoutes(mux *http.ServeMux) {
	mux.Handle("/customers", secure(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			customers.CreateCustomer(w, r)
		case http.MethodGet:
			customers.ListCustomers(w, r)
		default:
			methodNotAllowed(w)
		}
	})))
	mux.Handle("/customers/", secure(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			customers.GetCustomer(w, r)
		case http.MethodPut:
			customers.UpdateCustomer(w, r)
		case http.MethodDelete:
			customers.DeleteCustomer(w, r)
		default:
			methodNotAllowed(w)
		}
	})))
}

func registerCategoryRoutes(mux *http.ServeMux) {
	mux.Handle("/categories", secure(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			categories.CreateCategory(w, r)
		case http.MethodGet:
			categories.ListCategories(w, r)
		default:
			methodNotAllowed(w)
		}
	})))
	mux.Handle("/categories/", secure(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			categories.GetCategory(w, r)
		case http.MethodPut:
			categories.UpdateCategory(w, r)
		case http.MethodDelete:
			categories.DeleteCategory(w, r)
		default:
			methodNotAllowed(w)
		}
	})))
}

func registerEmployeeRoutes(mux *http.ServeMux) {
	mux.Handle("/employees", secure(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			employees.CreateEmployee(w, r)
		case http.MethodGet:
			employees.ListEmployees(w, r)
		default:
			methodNotAllowed(w)
		}
	})))
	mux.Handle("/employees/", secure(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			employees.GetEmployee(w, r)
		case http.MethodPut:
			employees.UpdateEmployee(w, r)
		case http.MethodDelete:
			employees.DeleteEmployee(w, r)
		default:
			methodNotAllowed(w)
		}
	})))
}

func registerTransferRoutes(mux *http.ServeMux) {
	mux.Handle("/transfers", secure(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			transfers.CreateTransfer(w, r)
		case http.MethodGet:
			transfers.ListTransfers(w, r)
		default:
			methodNotAllowed(w)
		}
	})))
	mux.Handle("/transfers/", secure(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			transfers.GetTransfer(w, r)
		} else {
			methodNotAllowed(w)
		}
	})))
}

func registerReportRoutes(mux *http.ServeMux) {
	mux.Handle("/reports/stock-summary", secure(http.HandlerFunc(reports.GetStockSummary)))
	mux.Handle("/reports/sales-summary", secure(http.HandlerFunc(reports.GetSalesSummary)))
	mux.Handle("/reports/top-products", secure(http.HandlerFunc(reports.GetTopProducts)))
	mux.Handle("/audit-logs", secure(http.HandlerFunc(reports.GetAuditLogs)))
}

func methodNotAllowed(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	_, _ = w.Write([]byte(`{"status":"error","code":"method_not_allowed","message":"Method not allowed"}`))
}
