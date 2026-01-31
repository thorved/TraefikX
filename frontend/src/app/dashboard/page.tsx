"use client";

import { ProtectedLayout } from "@/components/layout/protected-layout";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth } from "@/contexts/AuthContext";
import { useProxies } from "@/hooks/use-proxies";
import { 
  Shield, 
  Globe, 
  ArrowRightLeft, 
  Radio, 
  AlertTriangle,
  Activity 
} from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";

export default function DashboardPage() {
  const { user, checkAdmin } = useAuth();
  const { proxies, isLoading } = useProxies();

  // Calculate stats
  const proxyCount = proxies.length;
  const onlineCount = proxies.filter((p) => p.status === "online").length;
  const offlineCount = proxies.filter((p) => p.status === "offline").length;
  const sslCount = proxies.filter((p) => p.ssl).length;

  const isAdmin = checkAdmin();

  const stats = [
    {
      name: isAdmin ? "All Proxy Hosts" : "My Proxy Hosts",
      value: proxyCount,
      icon: Globe,
      color: "text-green-600",
      bgColor: "bg-green-100",
      description: isAdmin ? "Total active proxies" : "Your active proxies",
      href: "/proxies",
    },
    {
      name: "SSL Enabled",
      value: sslCount,
      icon: Shield,
      color: "text-blue-600",
      bgColor: "bg-blue-100",
      description: "Hosts with HTTPS",
      href: "/proxies",
    },
    {
      name: "Online",
      value: onlineCount,
      icon: Activity,
      color: "text-emerald-600",
      bgColor: "bg-emerald-100",
      description: "Running services",
      href: "/proxies",
    },
    {
      name: "Offline",
      value: offlineCount,
      icon: AlertTriangle,
      color: "text-red-600",
      bgColor: "bg-red-100",
      description: "Inactive hosts",
      href: "/proxies",
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

        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {stats.map((stat) => (
            <Link key={stat.name} href={stat.href}>
              <Card className="cursor-pointer hover:shadow-md transition-shadow">
                <CardContent className="p-6">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-4">
                      <div className={`${stat.bgColor} p-3 rounded-lg`}>
                        <stat.icon className={`h-6 w-6 ${stat.color}`} />
                      </div>
                      <div>
                        <p className="text-sm font-medium text-muted-foreground">
                          {stat.name}
                        </p>
                        <p className="text-2xl font-bold">
                          {isLoading ? "-" : stat.value}
                        </p>
                      </div>
                    </div>
                  </div>
                  <p className="text-xs text-muted-foreground mt-2">
                    {stat.description}
                  </p>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle>Getting Started</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <p className="text-sm text-muted-foreground">
                TraefikX provides a modern interface for managing Traefik reverse
                proxies. Create proxy hosts to route traffic to your services.
              </p>

              <div className="flex gap-2">
                <Button asChild>
                  <Link href="/proxies">
                    <Globe className="mr-2 h-4 w-4" />
                    Manage Proxy Hosts
                  </Link>
                </Button>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Quick Tips</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex items-start gap-3">
                <Shield className="h-5 w-5 text-green-600 mt-0.5" />
                <div>
                  <p className="text-sm font-medium">SSL Certificates</p>
                  <p className="text-xs text-muted-foreground">
                    Enable SSL to automatically obtain Let&apos;s Encrypt certificates
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <Globe className="h-5 w-5 text-blue-600 mt-0.5" />
                <div>
                  <p className="text-sm font-medium">Multiple Domains</p>
                  <p className="text-xs text-muted-foreground">
                    Add multiple domain names to route them to the same service
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </ProtectedLayout>
  );
}
