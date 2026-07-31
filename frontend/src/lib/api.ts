import { z } from "zod";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

export const categorySchema = z.object({
  id: z.number(),
  name: z.string(),
  slug: z.string()
});

export const productSchema = z.object({
  id: z.number(),
  brandName: z.string(),
  categoryName: z.string(),
  sku: z.string(),
  name: z.string(),
  description: z.string(),
  priceCents: z.number(),
  compareAtCents: z.number().optional(),
  rating: z.string(),
  reviewCount: z.number(),
  imageUrl: z.string(),
  stockQuantity: z.number(),
  createdAt: z.string()
});

export type Category = z.infer<typeof categorySchema>;
export type Product = z.infer<typeof productSchema>;

export type ProductParams = {
  q?: string;
  categoryId?: number;
  minPrice?: number;
  maxPrice?: number;
  limit?: number;
  offset?: number;
};

export async function getCategories(): Promise<Category[]> {
  const response = await fetch(`${API_BASE_URL}/api/categories`, { cache: "no-store" });
  if (!response.ok) {
    throw new Error("カテゴリを取得できませんでした");
  }
  return z.array(categorySchema).parse(await response.json());
}

export async function getProducts(params: ProductParams): Promise<Product[]> {
  const url = new URL(`${API_BASE_URL}/api/products`);
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== "" && value !== 0) {
      url.searchParams.set(key, String(value));
    }
  });

  const response = await fetch(url, { cache: "no-store" });
  if (!response.ok) {
    throw new Error("商品を取得できませんでした");
  }
  return z.array(productSchema).parse(await response.json());
}

export function yen(cents: number): string {
  return new Intl.NumberFormat("ja-JP", {
    style: "currency",
    currency: "JPY",
    maximumFractionDigits: 0
  }).format(cents);
}

