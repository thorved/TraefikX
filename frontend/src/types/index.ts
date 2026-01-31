export type UserRole = "admin" | "user";

export interface User {
  id: number;
  email: string;
  role: UserRole;
  is_active: boolean;
  password_enabled: boolean;
  oidc_provider?: string;
  oidc_enabled: boolean;
  is_linked_to_oidc: boolean;
  created_at: string;
  updated_at: string;
  last_login_at?: string;
}

export interface UserResponse {
  id: number;
  email: string;
  role: UserRole;
  is_active: boolean;
  password_enabled: boolean;
  oidc_provider?: string;
  oidc_enabled: boolean;
  is_linked_to_oidc: boolean;
  created_at: string;
  updated_at: string;
  last_login_at?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
  user: User;
}

export interface CreateUserRequest {
  email: string;
  password?: string;
  role: UserRole;
  oidc_enabled: boolean;
}

export interface UpdateUserRequest {
  email?: string;
  role?: UserRole;
  is_active?: boolean;
  oidc_enabled?: boolean;
}

export interface ChangePasswordRequest {
  current_password?: string;
  new_password: string;
}

export interface OIDCStatus {
  enabled: boolean;
  provider_name?: string;
}

export interface OIDCAuthResponse {
  auth_url: string;
  state: string;
}

export interface ApiError {
  error: string;
}

export interface UsersListResponse {
  users: UserResponse[];
}

// Traefik Types
export interface Service {
  id: number;
  name: string;
  type: string;
  servers: Server[];
  load_balancer_type: string;
  pass_host_header: boolean;
  health_check_enabled: boolean;
  health_check_path?: string;
  health_check_interval?: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Server {
  id: number;
  url: string;
  weight: number;
}

export interface Middleware {
  id: number;
  name: string;
  type: "redirectScheme" | "headers" | "stripPrefix" | "addPrefix";
  config?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Router {
  id: number;
  name: string;
  hostnames: string[];
  service_id: number;
  service_name: string;
  tls_enabled: boolean;
  tls_cert_resolver: string;
  redirect_https: boolean;
  entry_points: string[];
  middlewares: MiddlewareInfo[];
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface MiddlewareInfo {
  id: number;
  name: string;
  type: string;
  priority: number;
}

// Request Types
export interface CreateServiceRequest {
  name: string;
  servers: string[];
  load_balancer_type?: "wrr" | "drr";
  pass_host_header?: boolean;
  health_check_enabled?: boolean;
  health_check_path?: string;
  health_check_interval?: number;
}

export interface UpdateServiceRequest {
  name?: string;
  servers?: string[];
  load_balancer_type?: "wrr" | "drr";
  pass_host_header?: boolean;
  health_check_enabled?: boolean;
  health_check_path?: string;
  health_check_interval?: number;
  is_active?: boolean;
}

export interface MiddlewareConfig {
  scheme?: string;
  permanent?: boolean;
  port?: string;
  customRequestHeaders?: Record<string, string>;
  customResponseHeaders?: Record<string, string>;
  sslRedirect?: boolean;
  prefixes?: string[];
  forceSlash?: boolean;
  prefix?: string;
}

export interface CreateMiddlewareRequest {
  name: string;
  type: "redirectScheme" | "headers" | "stripPrefix" | "addPrefix";
  config: MiddlewareConfig;
}

export interface UpdateMiddlewareRequest {
  name?: string;
  type?: "redirectScheme" | "headers" | "stripPrefix" | "addPrefix";
  config?: MiddlewareConfig;
  is_active?: boolean;
}

export interface CreateRouterRequest {
  name: string;
  hostnames: string[];
  service_id: number;
  tls_enabled?: boolean;
  tls_cert_resolver?: string;
  redirect_https?: boolean;
  entry_points?: string[];
  middleware_ids?: number[];
}

export interface UpdateRouterRequest {
  name?: string;
  hostnames?: string[];
  service_id?: number;
  tls_enabled?: boolean;
  tls_cert_resolver?: string;
  redirect_https?: boolean;
  entry_points?: string[];
  middleware_ids?: number[];
  is_active?: boolean;
}

// NPM-style Proxy Host Types
export interface ProxyHost {
  id: number;
  domain_names: string[];
  forward_scheme: string;
  forward_host: string;
  forward_port: number;
  ssl: boolean;
  ssl_provider?: string;
  access: string;
  status: "online" | "offline";
  created_at: string;
}

export interface CreateProxyHostRequest {
  domain_names: string[];
  forward_scheme: "http" | "https";
  forward_host: string;
  forward_port: number;
  ssl?: boolean;
  ssl_provider?: string;
  access: "public" | "private";
}

export interface UpdateProxyHostRequest {
  domain_names?: string[];
  forward_scheme?: "http" | "https";
  forward_host?: string;
  forward_port?: number;
  ssl?: boolean;
  ssl_provider?: string;
  access?: "public" | "private";
}

export interface DashboardStats {
  proxy_hosts: number;
  redirection_hosts: number;
  streams: number;
  offline_hosts: number;
}

// HTTP Provider Types
export interface HTTPProvider {
  id: number;
  name: string;
  url: string;
  priority: number;
  is_active: boolean;
  refresh_interval: number;
  last_fetched: string | null;
  last_error: string | null;
  router_count: number;
  service_count: number;
  middleware_count: number;
  created_at: string;
  updated_at: string;
}

export interface CreateHTTPProviderRequest {
  name: string;
  url: string;
  priority: number;
  refresh_interval: number;
  is_active: boolean;
}

export interface UpdateHTTPProviderRequest {
  name?: string;
  url?: string;
  priority?: number;
  refresh_interval?: number;
  is_active?: boolean;
}

export interface ConflictInfo {
  type: "router" | "service" | "middleware";
  name: string;
  source: string;
  overridden_by: string;
  source_priority: number;
}

export interface ProviderSourceInfo {
  name: string;
  priority: number;
  status: "healthy" | "degraded" | "unhealthy" | "inactive";
  last_fetched?: string;
  last_error?: string;
  router_count: number;
  service_count: number;
  middleware_count: number;
}

export interface MergedTraefikConfig {
  config: {
    http: {
      routers?: Record<string, unknown>;
      services?: Record<string, unknown>;
      middlewares?: Record<string, unknown>;
    };
  };
  conflicts: ConflictInfo[];
  sources: ProviderSourceInfo[];
}
