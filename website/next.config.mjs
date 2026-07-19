/** @type {import('next').NextConfig} */
const rawBase = process.env.BASE_PATH ?? process.env.NEXT_PUBLIC_BASE_PATH ?? "";
// Project Pages: https://sraphaz.github.io/arah-harness/
// Local/dev: leave BASE_PATH empty.
const basePath = rawBase.replace(/\/$/, "");

const nextConfig = {
  output: "export",
  trailingSlash: true,
  images: { unoptimized: true },
  ...(basePath
    ? {
        basePath,
        assetPrefix: basePath,
      }
    : {}),
  env: {
    NEXT_PUBLIC_BASE_PATH: basePath,
  },
};

export default nextConfig;
