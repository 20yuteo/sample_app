package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Product struct {
	ID             int64     `json:"id"`
	BrandName      string    `json:"brandName"`
	CategoryName   string    `json:"categoryName"`
	SKU            string    `json:"sku"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	PriceCents     int32     `json:"priceCents"`
	CompareAtCents *int32    `json:"compareAtCents,omitempty"`
	Rating         string    `json:"rating"`
	ReviewCount    int32     `json:"reviewCount"`
	ImageURL       string    `json:"imageUrl"`
	StockQuantity  int32     `json:"stockQuantity"`
	CreatedAt      time.Time `json:"createdAt"`
}

type ProductFilters struct {
	Query      string
	CategoryID int64
	MinPrice   int32
	MaxPrice   int32
	Limit      int32
	Offset     int32
}

func (s *Store) ListCategories(ctx context.Context) ([]Category, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, slug
		FROM categories
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]Category, 0)
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name, &category.Slug); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, rows.Err()
}

func (s *Store) ListProducts(ctx context.Context, filters ProductFilters) ([]Product, error) {
	if filters.Limit <= 0 || filters.Limit > 100 {
		filters.Limit = 24
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			p.id,
			b.name,
			c.name,
			p.sku,
			p.name,
			p.description,
			p.price_cents,
			p.compare_at_cents,
			p.rating::text,
			p.review_count,
			COALESCE(pi.url, ''),
			i.quantity,
			p.created_at
		FROM products p
		JOIN brands b ON b.id = p.brand_id
		JOIN categories c ON c.id = p.category_id
		JOIN inventory i ON i.product_id = p.id
		LEFT JOIN LATERAL (
			SELECT url
			FROM product_images
			WHERE product_id = p.id
			ORDER BY sort_order
			LIMIT 1
		) pi ON true
		WHERE p.status = 'active'
			AND ($1 = '' OR p.search_vector @@ plainto_tsquery('simple', $1))
			AND ($2::bigint = 0 OR p.category_id = $2)
			AND ($3::integer = 0 OR p.price_cents >= $3)
			AND ($4::integer = 0 OR p.price_cents <= $4)
		ORDER BY p.created_at DESC, p.id DESC
		LIMIT $5 OFFSET $6
	`, filters.Query, filters.CategoryID, filters.MinPrice, filters.MaxPrice, filters.Limit, filters.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]Product, 0, filters.Limit)
	for rows.Next() {
		var product Product
		if err := rows.Scan(
			&product.ID,
			&product.BrandName,
			&product.CategoryName,
			&product.SKU,
			&product.Name,
			&product.Description,
			&product.PriceCents,
			&product.CompareAtCents,
			&product.Rating,
			&product.ReviewCount,
			&product.ImageURL,
			&product.StockQuantity,
			&product.CreatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, rows.Err()
}
