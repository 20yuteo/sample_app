"use client";

import Image from "next/image";
import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Search, SlidersHorizontal, ShoppingCart, Star } from "lucide-react";

import { AuthStatus } from "@/components/auth-status";
import { Category, getCategories, getProducts, yen } from "@/lib/api";

const priceRanges = [
  { label: "すべて", min: 0, max: 0 },
  { label: "1万円未満", min: 0, max: 9_999 },
  { label: "1万-5万円", min: 10_000, max: 50_000 },
  { label: "5万円以上", min: 50_001, max: 0 }
];

export function Catalog() {
  const [query, setQuery] = useState("");
  const [categoryId, setCategoryId] = useState(0);
  const [priceIndex, setPriceIndex] = useState(0);

  const categories = useQuery({
    queryKey: ["categories"],
    queryFn: getCategories
  });

  const priceRange = priceRanges[priceIndex];
  const products = useQuery({
    queryKey: ["products", query, categoryId, priceIndex],
    queryFn: () =>
      getProducts({
        q: query,
        categoryId,
        minPrice: priceRange.min,
        maxPrice: priceRange.max,
        limit: 36
      })
  });

  const selectedCategory = useMemo(
    () => categories.data?.find((category) => category.id === categoryId),
    [categories.data, categoryId]
  );

  return (
    <main className="min-h-screen bg-[#f7f8f6] text-ink">
      <header className="border-b border-[#dfe5df] bg-white">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-4 sm:px-6 lg:px-8">
          <div>
            <p className="text-xs font-semibold uppercase tracking-wide text-moss">Commerce Lab</p>
            <h1 className="text-2xl font-semibold">商品カタログ</h1>
          </div>
          <div className="flex items-center gap-2">
            <AuthStatus />
            <button className="inline-flex h-10 w-10 items-center justify-center rounded-md border border-[#d6ddd7] bg-white text-ink shadow-sm">
              <ShoppingCart size={20} aria-label="Cart" />
            </button>
          </div>
        </div>
      </header>

      <section className="mx-auto grid max-w-7xl gap-5 px-4 py-6 sm:px-6 lg:grid-cols-[280px_1fr] lg:px-8">
        <aside className="h-fit rounded-md border border-[#dfe5df] bg-white p-4">
          <div className="mb-4 flex items-center gap-2 text-sm font-semibold">
            <SlidersHorizontal size={17} />
            フィルター
          </div>

          <label className="mb-2 block text-xs font-medium text-[#5d6a63]">検索</label>
          <div className="mb-5 flex h-10 items-center gap-2 rounded-md border border-[#d6ddd7] px-3">
            <Search size={17} className="text-[#758177]" />
            <input
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              className="min-w-0 flex-1 bg-transparent outline-none"
              placeholder="SKU / 商品名"
            />
          </div>

          <label className="mb-2 block text-xs font-medium text-[#5d6a63]">カテゴリ</label>
          <select
            value={categoryId}
            onChange={(event) => setCategoryId(Number(event.target.value))}
            className="mb-5 h-10 w-full rounded-md border border-[#d6ddd7] bg-white px-3"
          >
            <option value={0}>すべて</option>
            {categories.data?.map((category: Category) => (
              <option key={category.id} value={category.id}>
                {category.name}
              </option>
            ))}
          </select>

          <label className="mb-2 block text-xs font-medium text-[#5d6a63]">価格帯</label>
          <div className="grid gap-2">
            {priceRanges.map((range, index) => (
              <button
                key={range.label}
                onClick={() => setPriceIndex(index)}
                className={`h-10 rounded-md border px-3 text-left text-sm ${
                  priceIndex === index
                    ? "border-moss bg-[#eef4ec] font-semibold text-moss"
                    : "border-[#d6ddd7] bg-white"
                }`}
              >
                {range.label}
              </button>
            ))}
          </div>
        </aside>

        <div>
          <div className="mb-4 flex flex-wrap items-end justify-between gap-3">
            <div>
              <h2 className="text-xl font-semibold">
                {selectedCategory ? selectedCategory.name : "すべての商品"}
              </h2>
              <p className="text-sm text-[#5d6a63]">PostgreSQL の大量 seed データを検索できます。</p>
            </div>
            <p className="text-sm text-[#5d6a63]">{products.data?.length ?? 0} 件表示</p>
          </div>

          {products.isLoading ? (
            <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
              {Array.from({ length: 9 }).map((_, index) => (
                <div key={index} className="h-80 animate-pulse rounded-md bg-[#e6ebe5]" />
              ))}
            </div>
          ) : products.isError ? (
            <div className="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-700">
              API に接続できません。バックエンドと seed 済み DB を起動してください。
            </div>
          ) : (
            <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
              {products.data?.map((product) => (
                <article key={product.id} className="overflow-hidden rounded-md border border-[#dfe5df] bg-white">
                  <div className="relative aspect-square bg-mist">
                    {product.imageUrl ? (
                      <Image
                        src={product.imageUrl}
                        alt={product.name}
                        fill
                        unoptimized
                        className="object-cover"
                        sizes="33vw"
                      />
                    ) : null}
                  </div>
                  <div className="p-4">
                    <div className="mb-2 flex items-center justify-between gap-2 text-xs text-[#66736b]">
                      <span>{product.brandName}</span>
                      <span>{product.sku}</span>
                    </div>
                    <h3 className="line-clamp-2 min-h-12 text-base font-semibold">{product.name}</h3>
                    <p className="mt-2 line-clamp-2 min-h-10 text-sm text-[#5d6a63]">{product.description}</p>
                    <div className="mt-4 flex items-center justify-between">
                      <div>
                        <p className="text-lg font-semibold">{yen(product.priceCents)}</p>
                        <p className="text-xs text-[#66736b]">在庫 {product.stockQuantity}</p>
                      </div>
                      <div className="flex items-center gap-1 text-sm">
                        <Star size={16} className="fill-clay text-clay" />
                        {product.rating} ({product.reviewCount})
                      </div>
                    </div>
                  </div>
                </article>
              ))}
            </div>
          )}
        </div>
      </section>
    </main>
  );
}
