"use client";

import { ProtectedLayout } from "@/components/layout/protected-layout";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useProxies } from "@/hooks/use-proxies";
import { useAuth } from "@/contexts/AuthContext";
import { useState } from "react";
import { 
  Plus, 
  MoreVertical, 
  Globe, 
  Shield, 
  CheckCircle2, 
  XCircle,
  Search,
  Trash2,
  Pencil
} from "lucide-react";
import { Input } from "@/components/ui/input";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { AddProxyDialog } from "./add-proxy-dialog";
import { EditProxyDialog } from "./edit-proxy-dialog";
import { DeleteProxyDialog } from "./delete-proxy-dialog";
import { ProxyHost } from "@/types";

export default function ProxiesPage() {
  const { checkAdmin } = useAuth();
  const isAdmin = checkAdmin();
  const { proxies, isLoading, refetch } = useProxies();
  const [searchQuery, setSearchQuery] = useState("");
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [editingProxy, setEditingProxy] = useState<ProxyHost | null>(null);
  const [deletingProxy, setDeletingProxy] = useState<ProxyHost | null>(null);

  // Filter proxies based on search
  const filteredProxies = proxies.filter((proxy) => {
    const query = searchQuery.toLowerCase();
    return (
      proxy.domain_names.some((d) => d.toLowerCase().includes(query)) ||
      proxy.forward_host.toLowerCase().includes(query)
    );
  });

  return (
    <ProtectedLayout>
      <div className="space-y-6">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
            <div>
              <CardTitle className="text-2xl">
                {isAdmin ? "All Proxy Hosts" : "My Proxy Hosts"}
              </CardTitle>
              <p className="text-sm text-muted-foreground mt-1">
                {isAdmin 
                  ? "Manage all reverse proxy hosts" 
                  : "Manage your reverse proxy hosts"}
              </p>
            </div>
            <Button onClick={() => setIsAddDialogOpen(true)}>
              <Plus className="mr-2 h-4 w-4" />
              Add Proxy Host
            </Button>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-4 mb-4">
              <div className="relative flex-1 max-w-sm">
                <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search domains..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-8"
                />
              </div>
            </div>

            <div className="border rounded-md">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[50px]"></TableHead>
                    <TableHead>Source</TableHead>
                    <TableHead>Destination</TableHead>
                    <TableHead>SSL</TableHead>
                    <TableHead>Access</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="w-[50px]"></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {isLoading ? (
                    <TableRow>
                      <TableCell colSpan={7} className="text-center py-8">
                        Loading...
                      </TableCell>
                    </TableRow>
                  ) : filteredProxies.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={7} className="text-center py-8">
                        <div className="flex flex-col items-center gap-2">
                          <Globe className="h-8 w-8 text-muted-foreground" />
                          <p className="text-muted-foreground">
                            {searchQuery
                              ? "No matching proxy hosts found"
                              : "No proxy hosts yet. Create your first one!"}
                          </p>
                          {!searchQuery && (
                            <Button
                              variant="outline"
                              onClick={() => setIsAddDialogOpen(true)}
                            >
                              <Plus className="mr-2 h-4 w-4" />
                              Add Proxy Host
                            </Button>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  ) : (
                    filteredProxies.map((proxy) => (
                      <TableRow key={proxy.id}>
                        <TableCell>
                          <div className="w-8 h-8 rounded-full bg-gray-100 flex items-center justify-center">
                            <Globe className="h-4 w-4 text-gray-600" />
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="space-y-1">
                            {proxy.domain_names.map((domain, idx) => (
                              <div key={`${proxy.id}-domain-${idx}`} className="font-medium">
                                {domain}
                              </div>
                            ))}
                            <div className="text-xs text-muted-foreground">
                              Created: {proxy.created_at}
                            </div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="font-mono text-sm">
                            {proxy.forward_scheme}://{proxy.forward_host}:
                            {proxy.forward_port}
                          </div>
                        </TableCell>
                        <TableCell>
                          {proxy.ssl ? (
                            <div className="flex items-center gap-1 text-green-600">
                              <Shield className="h-4 w-4" />
                              <span className="text-xs">
                                {proxy.ssl_provider || "Let's Encrypt"}
                              </span>
                            </div>
                          ) : (
                            <span className="text-xs text-muted-foreground">-</span>
                          )}
                        </TableCell>
                        <TableCell>
                          <span
                            className={`text-xs px-2 py-1 rounded-full ${
                              proxy.access === "public"
                                ? "bg-green-100 text-green-700"
                                : "bg-yellow-100 text-yellow-700"
                            }`}
                          >
                            {proxy.access === "public" ? "Public" : "Private"}
                          </span>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-1">
                            {proxy.status === "online" ? (
                              <>
                                <CheckCircle2 className="h-4 w-4 text-green-500" />
                                <span className="text-sm text-green-600">Online</span>
                              </>
                            ) : (
                              <>
                                <XCircle className="h-4 w-4 text-red-500" />
                                <span className="text-sm text-red-600">Offline</span>
                              </>
                            )}
                          </div>
                        </TableCell>
                        <TableCell>
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" size="icon">
                                <MoreVertical className="h-4 w-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuItem
                                onClick={() => setEditingProxy(proxy)}
                              >
                                <Pencil className="mr-2 h-4 w-4" />
                                Edit
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                className="text-red-600"
                                onClick={() => setDeletingProxy(proxy)}
                              >
                                <Trash2 className="mr-2 h-4 w-4" />
                                Delete
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>

        <AddProxyDialog
          open={isAddDialogOpen}
          onOpenChange={setIsAddDialogOpen}
        />

        <EditProxyDialog
          proxy={editingProxy}
          open={!!editingProxy}
          onOpenChange={(open) => !open && setEditingProxy(null)}
        />

        <DeleteProxyDialog
          proxy={deletingProxy}
          open={!!deletingProxy}
          onOpenChange={(open) => !open && setDeletingProxy(null)}
        />
      </div>
    </ProtectedLayout>
  );
}
