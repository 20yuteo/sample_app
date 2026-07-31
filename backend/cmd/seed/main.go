package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type scale struct {
	Products       int
	Customers      int
	Admins         int
	Orders         int
	OrderItems     int
	Reviews        int
	StatusEvents   int
	Shipments      int
	ShipmentEvents int
	AdminNotes     int
}

var scales = map[string]scale{
	"30k":    {Products: 30_000, Customers: 30_000, Admins: 30_000, Orders: 30_000, OrderItems: 30_000, Reviews: 30_000, StatusEvents: 30_000, Shipments: 30_000, ShipmentEvents: 30_000, AdminNotes: 30_000},
	"dev":    {Products: 20_000, Customers: 50_000, Admins: 200, Orders: 100_000, OrderItems: 300_000, Reviews: 80_000, StatusEvents: 450_000, Shipments: 90_000, ShipmentEvents: 270_000, AdminNotes: 40_000},
	"medium": {Products: 200_000, Customers: 500_000, Admins: 2_000, Orders: 1_000_000, OrderItems: 3_000_000, Reviews: 800_000, StatusEvents: 4_500_000, Shipments: 900_000, ShipmentEvents: 2_700_000, AdminNotes: 400_000},
	"large":  {Products: 1_000_000, Customers: 2_000_000, Admins: 10_000, Orders: 5_000_000, OrderItems: 15_000_000, Reviews: 4_000_000, StatusEvents: 22_500_000, Shipments: 4_500_000, ShipmentEvents: 13_500_000, AdminNotes: 2_000_000},
}

var (
	brandNames = []string{"Northline", "Kumo", "Aster", "Peakform", "Mori", "Orbit", "Nexa", "Harbor", "Luma", "Civic"}
	categories = []string{"Apparel", "Shoes", "Bags", "Home", "Kitchen", "Outdoor", "Gadgets", "Beauty", "Books", "Sports", "Toys", "Office"}
	adjectives = []string{"Essential", "Premium", "Lightweight", "Urban", "Classic", "Eco", "Studio", "Travel", "Pro", "Daily"}
	nouns      = []string{"Jacket", "Sneaker", "Backpack", "Lamp", "Bottle", "Chair", "Keyboard", "Serum", "Notebook", "Watch"}
	prefs      = []string{"Tokyo", "Osaka", "Kanagawa", "Aichi", "Fukuoka", "Hokkaido", "Kyoto", "Saitama", "Chiba", "Hyogo"}
)

func main() {
	ctx := context.Background()
	databaseURL := env("DATABASE_URL", "postgres://commerce:commerce@localhost:5432/commerce?sslmode=disable")
	scaleName := env("SEED_SCALE", "dev")
	selected, ok := scales[scaleName]
	if !ok {
		log.Fatal().Str("scale", scaleName).Msg("unknown seed scale")
	}
	if override := os.Getenv("SEED_PRODUCTS"); override != "" {
		selected.Products = mustAtoi(override)
	}
	if override := os.Getenv("SEED_CUSTOMERS"); override != "" {
		selected.Customers = mustAtoi(override)
	}
	if override := os.Getenv("SEED_ADMINS"); override != "" {
		selected.Admins = mustAtoi(override)
	}
	if override := os.Getenv("SEED_ORDERS"); override != "" {
		selected.Orders = mustAtoi(override)
	}
	if override := os.Getenv("SEED_ORDER_ITEMS"); override != "" {
		selected.OrderItems = mustAtoi(override)
	}
	if override := os.Getenv("SEED_REVIEWS"); override != "" {
		selected.Reviews = mustAtoi(override)
	}
	if override := os.Getenv("SEED_STATUS_EVENTS"); override != "" {
		selected.StatusEvents = mustAtoi(override)
	}
	if override := os.Getenv("SEED_SHIPMENTS"); override != "" {
		selected.Shipments = mustAtoi(override)
	}
	if override := os.Getenv("SEED_SHIPMENT_EVENTS"); override != "" {
		selected.ShipmentEvents = mustAtoi(override)
	}
	if override := os.Getenv("SEED_ADMIN_NOTES"); override != "" {
		selected.AdminNotes = mustAtoi(override)
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("connect database")
	}
	defer pool.Close()

	log.Info().Interface("scale", selected).Msg("resetting data")
	truncate(ctx, pool)
	seedStatic(ctx, pool)
	seedProducts(ctx, pool, selected.Products)
	seedCustomers(ctx, pool, selected.Customers)
	seedAdmins(ctx, pool, selected.Admins)
	seedAuthUsers(ctx, pool, selected)
	seedOrders(ctx, pool, selected)
	seedOrderManagement(ctx, pool, selected)
	seedReviews(ctx, pool, selected)
	refreshProductStats(ctx, pool)
	log.Info().Msg("seed complete")
}

func truncate(ctx context.Context, pool *pgxpool.Pool) {
	_, err := pool.Exec(ctx, `
		TRUNCATE reviews, order_admin_notes, shipment_events, shipments, order_status_events, order_items, orders, addresses, auth_users, admin_users, customers, inventory, product_images, products, categories, brands
		RESTART IDENTITY CASCADE
	`)
	must(err)
}

func seedStatic(ctx context.Context, pool *pgxpool.Pool) {
	rows := make([][]any, 0, len(brandNames))
	for i, name := range brandNames {
		rows = append(rows, []any{name, slug(name, i)})
	}
	_, err := pool.CopyFrom(ctx, pgx.Identifier{"brands"}, []string{"name", "slug"}, pgx.CopyFromRows(rows))
	must(err)

	rows = rows[:0]
	for i, name := range categories {
		rows = append(rows, []any{nil, name, slug(name, i)})
	}
	_, err = pool.CopyFrom(ctx, pgx.Identifier{"categories"}, []string{"parent_id", "name", "slug"}, pgx.CopyFromRows(rows))
	must(err)
}

func seedProducts(ctx context.Context, pool *pgxpool.Pool, count int) {
	copyBatches(ctx, count, 25_000, func(start, end int) error {
		productRows := make([][]any, 0, end-start)
		imageRows := make([][]any, 0, end-start)
		inventoryRows := make([][]any, 0, end-start)
		for i := start; i < end; i++ {
			brandID := int64((i % len(brandNames)) + 1)
			categoryID := int64((i % len(categories)) + 1)
			price := 800 + (i*37)%120_000
			name := fmt.Sprintf("%s %s %04d", adjectives[i%len(adjectives)], nouns[(i/len(adjectives))%len(nouns)], i)
			productRows = append(productRows, []any{
				brandID,
				categoryID,
				fmt.Sprintf("SKU-%09d", i+1),
				name,
				fmt.Sprintf("%s for daily commerce workloads and catalog search testing.", name),
				price,
				price + 1200 + (i % 20_000),
				"active",
				time.Now().Add(-time.Duration(i%730) * 24 * time.Hour),
				time.Now(),
			})
			productID := int64(i + 1)
			imageRows = append(imageRows, []any{productID, imageURL(i), name, 0})
			inventoryRows = append(inventoryRows, []any{productID, (i * 13) % 500, 0, time.Now()})
		}
		if _, err := pool.CopyFrom(ctx, pgx.Identifier{"products"}, []string{"brand_id", "category_id", "sku", "name", "description", "price_cents", "compare_at_cents", "status", "created_at", "updated_at"}, pgx.CopyFromRows(productRows)); err != nil {
			return err
		}
		if _, err := pool.CopyFrom(ctx, pgx.Identifier{"product_images"}, []string{"product_id", "url", "alt", "sort_order"}, pgx.CopyFromRows(imageRows)); err != nil {
			return err
		}
		_, err := pool.CopyFrom(ctx, pgx.Identifier{"inventory"}, []string{"product_id", "quantity", "reserved_quantity", "updated_at"}, pgx.CopyFromRows(inventoryRows))
		return err
	})
}

func seedCustomers(ctx context.Context, pool *pgxpool.Pool, count int) {
	copyBatches(ctx, count, 50_000, func(start, end int) error {
		customerRows := make([][]any, 0, end-start)
		addressRows := make([][]any, 0, end-start)
		for i := start; i < end; i++ {
			segment := []string{"new", "regular", "vip", "wholesale"}[i%4]
			customerRows = append(customerRows, []any{
				fmt.Sprintf("customer%09d@example.test", i+1),
				fmt.Sprintf("Customer %09d", i+1),
				segment,
				time.Now().Add(-time.Duration(i%1_200) * 24 * time.Hour),
			})
			addressRows = append(addressRows, []any{
				int64(i + 1),
				fmt.Sprintf("%03d-%04d", i%999, (i*17)%9999),
				prefs[i%len(prefs)],
				fmt.Sprintf("City %03d", i%300),
				fmt.Sprintf("%d-%d-%d", i%100, (i/10)%100, (i/100)%100),
				"",
			})
		}
		if _, err := pool.CopyFrom(ctx, pgx.Identifier{"customers"}, []string{"email", "name", "segment", "created_at"}, pgx.CopyFromRows(customerRows)); err != nil {
			return err
		}
		_, err := pool.CopyFrom(ctx, pgx.Identifier{"addresses"}, []string{"customer_id", "postal_code", "prefecture", "city", "line1", "line2"}, pgx.CopyFromRows(addressRows))
		return err
	})
}

func seedAdmins(ctx context.Context, pool *pgxpool.Pool, count int) {
	copyBatches(ctx, count, 10_000, func(start, end int) error {
		rows := make([][]any, 0, end-start)
		for i := start; i < end; i++ {
			role := []string{"operator", "manager", "super_admin"}[i%3]
			rows = append(rows, []any{
				fmt.Sprintf("admin%07d@example.test", i+1),
				fmt.Sprintf("Admin User %07d", i+1),
				role,
				i%97 != 0,
				time.Now().Add(-time.Duration(i%900) * 24 * time.Hour),
			})
		}
		_, err := pool.CopyFrom(ctx, pgx.Identifier{"admin_users"}, []string{"email", "name", "role", "active", "created_at"}, pgx.CopyFromRows(rows))
		return err
	})
}

func seedAuthUsers(ctx context.Context, pool *pgxpool.Pool, s scale) {
	copyBatches(ctx, s.Customers, 50_000, func(start, end int) error {
		rows := make([][]any, 0, end-start)
		for i := start; i < end; i++ {
			rows = append(rows, []any{
				deterministicUUID("customer", i+1),
				fmt.Sprintf("customer%09d@example.test", i+1),
				"customer",
				int64(i + 1),
				nil,
				i%101 != 0,
				time.Now().Add(-time.Duration(i%30) * 24 * time.Hour),
				time.Now().Add(-time.Duration(i%1_200) * 24 * time.Hour),
			})
		}
		_, err := pool.CopyFrom(ctx, pgx.Identifier{"auth_users"}, []string{"keycloak_subject", "email", "account_type", "customer_id", "admin_user_id", "enabled", "last_login_at", "created_at"}, pgx.CopyFromRows(rows))
		return err
	})

	copyBatches(ctx, s.Admins, 10_000, func(start, end int) error {
		rows := make([][]any, 0, end-start)
		for i := start; i < end; i++ {
			rows = append(rows, []any{
				deterministicUUID("admin", i+1),
				fmt.Sprintf("admin%07d@example.test", i+1),
				"ec_admin",
				nil,
				int64(i + 1),
				i%97 != 0,
				time.Now().Add(-time.Duration(i%14) * 24 * time.Hour),
				time.Now().Add(-time.Duration(i%900) * 24 * time.Hour),
			})
		}
		_, err := pool.CopyFrom(ctx, pgx.Identifier{"auth_users"}, []string{"keycloak_subject", "email", "account_type", "customer_id", "admin_user_id", "enabled", "last_login_at", "created_at"}, pgx.CopyFromRows(rows))
		return err
	})
}

func seedOrders(ctx context.Context, pool *pgxpool.Pool, s scale) {
	rng := rand.New(rand.NewSource(42))
	copyBatches(ctx, s.Orders, 25_000, func(start, end int) error {
		orderRows := make([][]any, 0, end-start)
		for i := start; i < end; i++ {
			subtotal := 800 + (i*37)%120_000
			tax := subtotal / 10
			shipping := 0
			if subtotal < 8_000 {
				shipping = 600
			}
			status := []string{"pending", "paid", "shipped", "delivered", "cancelled"}[rng.Intn(5)]
			orderRows = append(orderRows, []any{
				int64(rng.Intn(s.Customers) + 1),
				fmt.Sprintf("ORD-%012d", i+1),
				status,
				subtotal,
				tax,
				shipping,
				subtotal + tax + shipping,
				time.Now().Add(-time.Duration(rng.Intn(730*24)) * time.Hour),
			})
		}
		_, err := pool.CopyFrom(ctx, pgx.Identifier{"orders"}, []string{"customer_id", "order_number", "status", "subtotal_cents", "tax_cents", "shipping_cents", "total_cents", "ordered_at"}, pgx.CopyFromRows(orderRows))
		return err
	})

	copyBatches(ctx, s.OrderItems, 50_000, func(start, end int) error {
		itemRows := make([][]any, 0, end-start)
		for i := start; i < end; i++ {
			productID := int64((i % s.Products) + 1)
			qty := 1 + (i % 4)
			unit := 800 + int(productID*37)%120_000
			total := qty * unit
			itemRows = append(itemRows, []any{
				int64((i % s.Orders) + 1),
				productID,
				qty,
				unit,
				total,
			})
		}
		_, err := pool.CopyFrom(ctx, pgx.Identifier{"order_items"}, []string{"order_id", "product_id", "quantity", "unit_price_cents", "total_cents"}, pgx.CopyFromRows(itemRows))
		return err
	})
}

func seedOrderManagement(ctx context.Context, pool *pgxpool.Pool, s scale) {
	statuses := []string{"pending", "paid", "picking", "packed", "shipped", "delivered", "cancelled", "returned"}
	actors := []string{"system", "customer", "ec_admin"}
	copyBatches(ctx, s.StatusEvents, 50_000, func(start, end int) error {
		rows := make([][]any, 0, end-start)
		for i := start; i < end; i++ {
			orderID := int64((i % s.Orders) + 1)
			actorType := actors[i%len(actors)]
			var actorID any
			switch actorType {
			case "customer":
				actorID = int64(((i % s.Customers) + 1))
			case "ec_admin":
				actorID = int64(s.Customers + ((i % s.Admins) + 1))
			default:
				actorID = nil
			}
			rows = append(rows, []any{
				orderID,
				statuses[i%len(statuses)],
				actorType,
				actorID,
				fmt.Sprintf("Order workflow event %09d", i+1),
				time.Now().Add(-time.Duration((s.StatusEvents-i)%17_520) * time.Hour),
			})
		}
		_, err := pool.CopyFrom(ctx, pgx.Identifier{"order_status_events"}, []string{"order_id", "status", "actor_type", "actor_auth_user_id", "note", "occurred_at"}, pgx.CopyFromRows(rows))
		return err
	})

	carriers := []string{"Yamato", "Sagawa", "JapanPost", "DHL", "FedEx"}
	shipmentStatuses := []string{"label_created", "picked_up", "in_transit", "out_for_delivery", "delivered", "failed", "returned"}
	copyBatches(ctx, s.Shipments, 50_000, func(start, end int) error {
		shipmentRows := make([][]any, 0, end-start)
		for i := start; i < end; i++ {
			orderID := int64((i % s.Orders) + 1)
			createdAt := time.Now().Add(-time.Duration(i%8_760) * time.Hour)
			status := shipmentStatuses[i%len(shipmentStatuses)]
			var deliveredAt any
			if status == "delivered" {
				deliveredAt = createdAt.Add(72 * time.Hour)
			}
			shipmentRows = append(shipmentRows, []any{
				orderID,
				carriers[i%len(carriers)],
				fmt.Sprintf("TRK%014d", i+1),
				status,
				createdAt.Add(6 * time.Hour),
				createdAt.Add(96 * time.Hour),
				deliveredAt,
				createdAt,
			})
		}
		_, err := pool.CopyFrom(ctx, pgx.Identifier{"shipments"}, []string{"order_id", "carrier", "tracking_number", "status", "shipped_at", "estimated_delivery_at", "delivered_at", "created_at"}, pgx.CopyFromRows(shipmentRows))
		return err
	})

	copyBatches(ctx, s.ShipmentEvents, 50_000, func(start, end int) error {
		eventRows := make([][]any, 0, end-start)
		for i := start; i < end; i++ {
			createdAt := time.Now().Add(-time.Duration(i%8_760) * time.Hour)
			eventRows = append(eventRows, []any{
				int64((i % s.Shipments) + 1),
				shipmentStatuses[i%len(shipmentStatuses)],
				fmt.Sprintf("%s Hub %03d", prefs[i%len(prefs)], i%300),
				fmt.Sprintf("Shipment event %09d", i+1),
				createdAt,
			})
		}
		_, err := pool.CopyFrom(ctx, pgx.Identifier{"shipment_events"}, []string{"shipment_id", "status", "location", "message", "occurred_at"}, pgx.CopyFromRows(eventRows))
		return err
	})

	noteTypes := []string{"fraud_check", "customer_support", "shipping", "refund", "internal"}
	copyBatches(ctx, s.AdminNotes, 50_000, func(start, end int) error {
		rows := make([][]any, 0, end-start)
		for i := start; i < end; i++ {
			rows = append(rows, []any{
				int64((i % s.Orders) + 1),
				int64((i % s.Admins) + 1),
				noteTypes[i%len(noteTypes)],
				fmt.Sprintf("Operational note for order management workload %09d", i+1),
				time.Now().Add(-time.Duration(i%4_380) * time.Hour),
			})
		}
		_, err := pool.CopyFrom(ctx, pgx.Identifier{"order_admin_notes"}, []string{"order_id", "admin_user_id", "note_type", "body", "created_at"}, pgx.CopyFromRows(rows))
		return err
	})
}

func seedReviews(ctx context.Context, pool *pgxpool.Pool, s scale) {
	rng := rand.New(rand.NewSource(84))
	copyBatches(ctx, s.Reviews, 50_000, func(start, end int) error {
		rows := make([][]any, 0, end-start)
		for i := start; i < end; i++ {
			rows = append(rows, []any{
				int64(rng.Intn(s.Products) + 1),
				int64(rng.Intn(s.Customers) + 1),
				1 + rng.Intn(5),
				fmt.Sprintf("Review %09d", i+1),
				"Realistic enough text for review list queries and aggregate workloads.",
				time.Now().Add(-time.Duration(rng.Intn(600*24)) * time.Hour),
			})
		}
		_, err := pool.CopyFrom(ctx, pgx.Identifier{"reviews"}, []string{"product_id", "customer_id", "rating", "title", "body", "created_at"}, pgx.CopyFromRows(rows))
		return err
	})
}

func refreshProductStats(ctx context.Context, pool *pgxpool.Pool) {
	_, err := pool.Exec(ctx, `
		UPDATE products p
		SET rating = COALESCE(stats.rating, 0),
			review_count = COALESCE(stats.review_count, 0),
			updated_at = now()
		FROM (
			SELECT product_id, round(avg(rating)::numeric, 2) AS rating, count(*)::integer AS review_count
			FROM reviews
			GROUP BY product_id
		) stats
		WHERE stats.product_id = p.id
	`)
	must(err)
	_, err = pool.Exec(ctx, "ANALYZE")
	must(err)
}

func copyBatches(ctx context.Context, count, batchSize int, fn func(start, end int) error) {
	for start := 0; start < count; start += batchSize {
		end := start + batchSize
		if end > count {
			end = count
		}
		log.Info().Int("start", start).Int("end", end).Msg("copy batch")
		must(fn(start, end))
		select {
		case <-ctx.Done():
			must(ctx.Err())
		default:
		}
	}
}

func imageURL(i int) string {
	storageURL := strings.TrimRight(env("STORAGE_PUBLIC_URL", "http://localhost:9000"), "/")
	bucket := env("PRODUCT_IMAGE_BUCKET", "commerce-images")
	return fmt.Sprintf("%s/%s/products/product-%d.svg", storageURL, bucket, i%6)
}

func slug(value string, index int) string {
	return strings.ToLower(strings.ReplaceAll(value, " ", "-")) + "-" + strconv.Itoa(index+1)
}

func deterministicUUID(namespace string, id int) string {
	var prefix string
	switch namespace {
	case "admin":
		prefix = "00000002"
	default:
		prefix = "00000001"
	}
	return fmt.Sprintf("%s-0000-4000-8000-%012d", prefix, id)
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func mustAtoi(value string) int {
	parsed, err := strconv.Atoi(value)
	must(err)
	return parsed
}

func must(err error) {
	if err != nil {
		log.Fatal().Err(err).Msg("seed failed")
	}
}
