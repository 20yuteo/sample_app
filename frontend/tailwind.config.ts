import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./src/**/*.{js,ts,jsx,tsx,mdx}"],
  theme: {
    extend: {
      colors: {
        ink: "#172026",
        mist: "#eef2f4",
        moss: "#4f6f52",
        clay: "#b45f43",
        skyline: "#2f6f9f"
      }
    }
  },
  plugins: []
};

export default config;

