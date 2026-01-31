import axios, {
  AxiosError,
  AxiosInstance,
  InternalAxiosRequestConfig,
} from "axios";
import { AuthResponse, User, ApiError, Service, Middleware, Router, CreateServiceRequest, UpdateServiceRequest, CreateMiddlewareRequest, UpdateMiddlewareRequest, CreateRouterRequest, UpdateRouterRequest, ProxyHost, CreateProxyHostRequest, UpdateProxyHostRequest, HTTPProvider, CreateHTTPProviderRequest, UpdateHTTPProviderRequest, MergedTraefikConfig } from "@/types";

// Determine the base URL based on environment
// During development with Next.js dev server: use relative URLs (proxied via rewrites)
// For static export/production: use full backend URL
const getBaseURL = () => {
  // In browser, check if we're running from static export
  if (typeof window !== "undefined") {
    // If running from file:// protocol or without Next.js dev server
    // we need to use the full backend URL
    const isStaticExport =
      process.env.NEXT_PUBLIC_API_URL ||
      (window.location.protocol === "file:" ? "http://localhost:8080" : "");
    return isStaticExport;
  }
  return "";
};

const API_BASE_URL = getBaseURL();

// Create axios instance
const api: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
});

// Request interceptor - add auth token
api.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem("access_token");
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error),
);

// Response interceptor - handle token refresh
api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<ApiError>) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & {
      _retry?: boolean;
    };

    // If 401 and not already retrying
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      const refreshToken = localStorage.getItem("refresh_token");
      if (refreshToken) {
        try {
          const baseURL = getBaseURL();
          const response = await axios.post<AuthResponse>(
            `${baseURL}/api/auth/refresh`,
            { refresh_token: refreshToken },
          );

          const { access_token, refresh_token } = response.data;
          localStorage.setItem("access_token", access_token);
          localStorage.setItem("refresh_token", refresh_token);

          // Retry original request
          if (originalRequest.headers) {
            originalRequest.headers.Authorization = `Bearer ${access_token}`;
          }
          return api(originalRequest);
        } catch (refreshError) {
          // Refresh failed, clear tokens and redirect to login
          localStorage.removeItem("access_token");
          localStorage.removeItem("refresh_token");
          window.location.href = "/login";
          return Promise.reject(refreshError);
        }
      }
    }

    return Promise.reject(error);
  },
);

// Auth API
export const authApi = {
  login: (email: string, password: string) =>
    api.post<AuthResponse>("/api/auth/login", { email, password }),

  logout: () => api.post("/api/auth/logout"),

  refresh: (refreshToken: string) =>
    api.post<AuthResponse>("/api/auth/refresh", {
      refresh_token: refreshToken,
    }),

  getMe: () => api.get<User>("/api/auth/me"),

  changePassword: (currentPassword: string | undefined, newPassword: string) =>
    api.put("/api/auth/password", {
      current_password: currentPassword,
      new_password: newPassword,
    }),

  togglePasswordLogin: (enabled: boolean) =>
    api.post("/api/auth/password/toggle", { enabled }),

  removePassword: () => api.delete("/api/auth/password"),

  // OIDC
  getOIDCStatus: () =>
    api.get<{ enabled: boolean; provider_name?: string }>(
      "/api/auth/oidc/status",
    ),

  initiateOIDC: () =>
    api.get<{ auth_url: string; state: string }>("/api/auth/oidc"),

  initiateOIDCLink: () =>
    api.post<{ auth_url: string; state: string }>("/api/auth/oidc/link"),

  unlinkOIDC: () => api.delete("/api/auth/oidc/link"),

  oidcCallback: (code: string, state: string) =>
    api.get<AuthResponse>("/api/auth/oidc/callback", {
      params: { code, state },
    }),
};

// Users API
export const usersApi = {
  listUsers: () => api.get<{ users: User[] }>("/api/users"),

  getUser: (id: number | string) => api.get<User>(`/api/users/${id}`),

  createUser: (data: {
    email: string;
    password?: string;
    role: "admin" | "user";
    oidc_enabled: boolean;
  }) => api.post<User>("/api/users", data),

  updateUser: (
    id: number,
    data: {
      email?: string;
      role?: "admin" | "user";
      is_active?: boolean;
      oidc_enabled?: boolean;
    },
  ) => api.put<User>(`/api/users/${id}`, data),

  deleteUser: (id: number) => api.delete(`/api/users/${id}`),

  resetPassword: (id: number, newPassword: string) =>
    api.post(`/api/users/${id}/reset-password`, { new_password: newPassword }),

  toggleUserPasswordLogin: (id: number, enabled: boolean) =>
    api.post(`/api/users/${id}/password/toggle`, { enabled }),

  toggleUserOIDC: (id: number, enabled: boolean) =>
    api.post(`/api/users/${id}/oidc/toggle`, { enabled }),
};

// Services API (under /traefik)
export const servicesApi = {
  listServices: () => api.get<{ services: Service[] }>("/api/traefik/services"),

  getService: (id: number) => api.get<Service>(`/api/traefik/services/${id}`),

  createService: (data: CreateServiceRequest) =>
    api.post<Service>("/api/traefik/services", data),

  updateService: (id: number, data: UpdateServiceRequest) =>
    api.put<Service>(`/api/traefik/services/${id}`, data),

  deleteService: (id: number) => api.delete(`/api/traefik/services/${id}`),
};

// Middlewares API (under /traefik)
export const middlewaresApi = {
  listMiddlewares: () => api.get<{ middlewares: Middleware[] }>("/api/traefik/middlewares"),

  getMiddleware: (id: number) => api.get<Middleware>(`/api/traefik/middlewares/${id}`),

  createMiddleware: (data: CreateMiddlewareRequest) =>
    api.post<Middleware>("/api/traefik/middlewares", data),

  updateMiddleware: (id: number, data: UpdateMiddlewareRequest) =>
    api.put<Middleware>(`/api/traefik/middlewares/${id}`, data),

  deleteMiddleware: (id: number) => api.delete(`/api/traefik/middlewares/${id}`),
};

// Routers API (under /traefik)
export const routersApi = {
  listRouters: () => api.get<{ routers: Router[] }>("/api/traefik/routers"),

  getRouter: (id: number) => api.get<Router>(`/api/traefik/routers/${id}`),

  createRouter: (data: CreateRouterRequest) =>
    api.post<Router>("/api/traefik/routers", data),

  updateRouter: (id: number, data: UpdateRouterRequest) =>
    api.put<Router>(`/api/traefik/routers/${id}`, data),

  deleteRouter: (id: number) => api.delete(`/api/traefik/routers/${id}`),
};

// Traefik Provider API (for viewing the generated config)
export const providerApi = {
  getConfig: (token: string) =>
    api.get("/api/traefik/provider/config", { params: { token } }),
};

// Proxy Hosts API (NPM-style combined router + service)
export const proxyApi = {
  listProxies: () => api.get<{ proxies: ProxyHost[] }>("/api/traefik/proxies"),

  getProxy: (id: number) => api.get<ProxyHost>(`/api/traefik/proxies/${id}`),

  createProxy: (data: CreateProxyHostRequest) =>
    api.post<ProxyHost>("/api/traefik/proxies", data),

  updateProxy: (id: number, data: UpdateProxyHostRequest) =>
    api.put<ProxyHost>(`/api/traefik/proxies/${id}`, data),

  deleteProxy: (id: number) => api.delete(`/api/traefik/proxies/${id}`),
};

// HTTP Providers API (for aggregating external Traefik HTTP providers)
export const httpProvidersApi = {
  listProviders: () =>
    api.get<{ providers: HTTPProvider[] }>("/api/traefik/http-providers"),

  getProvider: (id: number) =>
    api.get<HTTPProvider>(`/api/traefik/http-providers/${id}`),

  createProvider: (data: CreateHTTPProviderRequest) =>
    api.post<HTTPProvider>("/api/traefik/http-providers", data),

  updateProvider: (id: number, data: UpdateHTTPProviderRequest) =>
    api.put<HTTPProvider>(`/api/traefik/http-providers/${id}`, data),

  deleteProvider: (id: number) =>
    api.delete(`/api/traefik/http-providers/${id}`),

  refreshProvider: (id: number) =>
    api.post(`/api/traefik/http-providers/${id}/refresh`),

  testProvider: (id: number) =>
    api.post<HTTPProvider>(`/api/traefik/http-providers/${id}/test`),

  getMergedConfig: () =>
    api.get<MergedTraefikConfig>("/api/traefik/merged-config"),
};

export default api;
