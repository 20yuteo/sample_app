import { Catalog } from "@/components/catalog";
import { QueryProvider } from "@/components/query-provider";

export default function Home() {
  return (
    <QueryProvider>
      <Catalog />
    </QueryProvider>
  );
}

