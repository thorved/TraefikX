"use client";

import { ProtectedLayout } from "@/components/layout/protected-layout";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useAuth } from "@/contexts/AuthContext";
import { LayoutDashboard, Users, Shield, Activity } from "lucide-react";

export default function DashboardPage() {
  const { user, checkAdmin } = useAuth();

  const stats = [
    {
      name: "Role",
      value: checkAdmin() ? "Administrator" : "User",
      icon: Shield,
      description: "Your account type",
    },
    {
      name: "Email",
      value: user?.email || "Unknown",
      icon: Users,
      description: "Your email address",
    },
    {
      name: "Status",
      value: user?.is_active ? "Active" : "Inactive",
      icon: Activity,
      description: "Account status",
    },
  ];

  return (
    <ProtectedLayout>
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
          <p className="text-muted-foreground">
            Welcome back, {user?.email?.split("@")[0] || "User"}!
          </p>
        </div>

        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {stats.map((stat) => (
            <Card key={stat.name}>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">
                  {stat.name}
                </CardTitle>
                <stat.icon className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stat.value}</div>
                <p className="text-xs text-muted-foreground">
                  {stat.description}
                </p>
              </CardContent>
            </Card>
          ))}
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Getting Started</CardTitle>
            <CardDescription>Manage your TraefikX installation</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm text-muted-foreground">
              TraefikX provides a modern interface for managing Traefik reverse
              proxies. Use the sidebar to navigate between different sections.
            </p>

            <div className="grid gap-4 md:grid-cols-2">
              <div className="rounded-lg border p-3">
                <h4 className="font-semibold mb-2">User Management</h4>
                <p className="text-sm text-muted-foreground">
                  {checkAdmin()
                    ? "As an administrator, you can manage users, create new accounts, and configure authentication settings."
                    : "View and update your profile settings, change your password, and manage account security."}
                </p>
              </div>

              <div className="rounded-lg border p-3">
                <h4 className="font-semibold mb-2">Authentication</h4>
                <p className="text-sm text-muted-foreground">
                  Secure your account with strong passwords and optional OIDC
                  integration for single sign-on.
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </ProtectedLayout>
  );
}
