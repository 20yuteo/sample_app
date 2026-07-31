import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  turbopack: {
    root: process.cwd()
  },
  images: {
    remotePatterns: [
      {
        protocol: "https",
        hostname: "placehold.co"
      },
      {
        protocol: "http",
        hostname: "localhost",
        port: "9000"
      },
      {
        protocol: "http",
        hostname: "127.0.0.1",
        port: "9000"
      }
    ]
  }
};

export default nextConfig;
