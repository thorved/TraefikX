import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: 'export',
  reactCompiler: true,
  
  // Add trailing slash for static export
  trailingSlash: true,
  
  // Proxy API requests to backend during development
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://localhost:8080/api/:path*',
      },
    ];
  },
  // Disable image optimization for static export
  images: {
    unoptimized: true,
  },
};

export default nextConfig;
