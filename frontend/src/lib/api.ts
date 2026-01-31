import axios, {
  AxiosError,
  AxiosInstance,
  InternalAxiosRequestConfig,
} from "axios";
import { AuthResponse, User, ApiError } from "@/types";

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

export default api;
