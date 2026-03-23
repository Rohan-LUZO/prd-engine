package domain

import "time"

// ProductSurface represents where a module belongs
// Keep this explicit — strings are safer than iota here
type ProductSurface string

const (
	SurfaceCustomerApp    ProductSurface = "customer_app"
	SurfaceWebsite        ProductSurface = "website"
	SurfaceAdminDashboard ProductSurface = "admin_dashboard"
	SurfacePartnerApp     ProductSurface = "partner_app"
	SurfaceLeadsApp       ProductSurface = "leads_app"
)

// Module is the core domain entity
// It represents ONE VERSION of a module
type Module struct {
	ID      string
	Version int

	Title    string
	Order    int
	Surfaces []ProductSurface

	Content string

	CreatedBy string
	CreatedAt time.Time
	UpdatedBy string
	UpdatedAt time.Time
}
