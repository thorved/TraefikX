"use client";

import { useState } from "react";
import { ProtectedLayout } from "@/components/layout/protected-layout";
import { useHTTPProviders, useMergedConfig } from "@/hooks/use-http-providers";
import { HTTPProvider, CreateHTTPProviderRequest } from "@/types";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  RefreshCw,
  Plus,
  Trash2,
  Edit3,
  Activity,
  AlertCircle,
  CheckCircle,
  WifiOff,
  Server,
  Globe,
  Layers,
} from "lucide-react";
import { formatDistanceToNow } from "@/lib/utils";

export default function HTTPProvidersPage() {
  const { providers, isLoading, createProvider, updateProvider, deleteProvider, refreshProvider, testProvider } =
    useHTTPProviders();
  const { mergedConfig, isLoading: isLoadingConfig, refetch: refetchConfig } = useMergedConfig(30000);
  
  // Get local stats from merged config
  const localSource = mergedConfig?.sources?.find(s => s.name === "local");
  const localStats = localSource ? {
    router_count: localSource.router_count,
    service_count: localSource.service_count,
    middleware_count: localSource.middleware_count,
  } : { router_count: 0, service_count: 0, middleware_count: 0 };

  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [editingProvider, setEditingProvider] = useState<HTTPProvider | null>(null);
  const [formData, setFormData] = useState<CreateHTTPProviderRequest>({
    name: "",
    url: "",
    priority: 0,
    refresh_interval: 30,
    is_active: true,
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (editingProvider) {
      await updateProvider.mutateAsync({ id: editingProvider.id, data: formData });
      setEditingProvider(null);
    } else {
      await createProvider.mutateAsync(formData);
    }
    setIsAddDialogOpen(false);
    resetForm();
  };

  const resetForm = () => {
    setFormData({
      name: "",
      url: "",
      priority: 0,
      refresh_interval: 30,
      is_active: true,
    });
  };

  const handleEdit = (provider: HTTPProvider) => {
    setEditingProvider(provider);
    setFormData({
      name: provider.name,
      url: provider.url,
      priority: provider.priority,
      refresh_interval: provider.refresh_interval,
      is_active: provider.is_active,
    });
    setIsAddDialogOpen(true);
  };

  const handleDelete = async (id: number) => {
    if (confirm("Are you sure you want to delete this HTTP Provider?")) {
      await deleteProvider.mutateAsync(id);
    }
  };

  const getStatusBadge = (provider: HTTPProvider) => {
    if (!provider.is_active) {
      return (
        <Badge variant="secondary" className="gap-1">
          <WifiOff className="h-3 w-3" />
          Inactive
        </Badge>
      );
    }
    if (provider.last_error) {
      return (
        <Badge variant="destructive" className="gap-1">
          <AlertCircle className="h-3 w-3" />
          Error
        </Badge>
      );
    }
    if (provider.last_fetched) {
      return (
        <Badge variant="default" className="gap-1 bg-green-600">
          <CheckCircle className="h-3 w-3" />
          Healthy
        </Badge>
      );
    }
    return (
      <Badge variant="outline" className="gap-1">
        <Activity className="h-3 w-3" />
        Pending
      </Badge>
    );
  };

  return (
    <ProtectedLayout requireAdmin>
      <div className="container mx-auto py-6 space-y-6">
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-3xl font-bold">HTTP Providers</h1>
            <p className="text-muted-foreground">
              Manage external Traefik HTTP Providers and aggregate their configurations
            </p>
          </div>
          <Dialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen}>
            <DialogTrigger asChild>
              <Button onClick={() => { setEditingProvider(null); resetForm(); }}>
                <Plus className="mr-2 h-4 w-4" />
                Add HTTP Provider
              </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[500px]">
              <DialogHeader>
                <DialogTitle>
                  {editingProvider ? "Edit HTTP Provider" : "Add New HTTP Provider"}
                </DialogTitle>
                <DialogDescription>
                  Configure an external Traefik HTTP Provider to aggregate
                </DialogDescription>
              </DialogHeader>
              <form onSubmit={handleSubmit} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="name">Name</Label>
                  <Input
                    id="name"
                    value={formData.name}
                    onChange={(e) =>
                      setFormData({ ...formData, name: e.target.value })
                    }
                    placeholder="e.g., production-cluster"
                    required
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="url">URL</Label>
                  <Input
                    id="url"
                    value={formData.url}
                    onChange={(e) =>
                      setFormData({ ...formData, url: e.target.value })
                    }
                    placeholder="e.g., http://100.111.54.25:8080/traefik/http"
                    required
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="priority">Priority</Label>
                    <Input
                      id="priority"
                      type="number"
                      value={formData.priority}
                      onChange={(e) =>
                        setFormData({ ...formData, priority: parseInt(e.target.value) || 0 })
                      }
                      min={0}
                      placeholder="Higher = more priority"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="refresh">Refresh Interval (s)</Label>
                    <Input
                      id="refresh"
                      type="number"
                      value={formData.refresh_interval}
                      onChange={(e) =>
                        setFormData({ ...formData, refresh_interval: parseInt(e.target.value) || 30 })
                      }
                      min={5}
                    />
                  </div>
                </div>
                <div className="flex items-center space-x-2">
                  <Switch
                    id="active"
                    checked={formData.is_active}
                    onCheckedChange={(checked) =>
                      setFormData({ ...formData, is_active: checked })
                    }
                  />
                  <Label htmlFor="active">Active</Label>
                </div>
                <DialogFooter>
                  <Button type="submit">
                    {editingProvider ? "Update" : "Create"}
                  </Button>
                </DialogFooter>
              </form>
            </DialogContent>
          </Dialog>
        </div>

        <Tabs defaultValue="providers" className="space-y-4">
          <TabsList>
            <TabsTrigger value="providers">HTTP Providers</TabsTrigger>
            <TabsTrigger value="config">Merged Config</TabsTrigger>
          </TabsList>

          <TabsContent value="providers" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>External HTTP Providers</CardTitle>
                <CardDescription>
                  External Traefik HTTP Providers aggregated into your configuration.
                  Priority: Local (highest) → Providers (by priority value)
                </CardDescription>
              </CardHeader>
              <CardContent>
                {isLoading ? (
                  <div className="space-y-2">
                    <Skeleton className="h-12 w-full" />
                    <Skeleton className="h-12 w-full" />
                    <Skeleton className="h-12 w-full" />
                  </div>
                ) : (
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Name</TableHead>
                        <TableHead>URL</TableHead>
                        <TableHead>Priority</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead>Last Fetch</TableHead>
                        <TableHead>Stats</TableHead>
                        <TableHead className="text-right">Actions</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      <TableRow className="bg-muted/50">
                        <TableCell className="font-medium">
                          <div className="flex items-center gap-2">
                            <Server className="h-4 w-4 text-primary" />
                            local
                          </div>
                        </TableCell>
                        <TableCell className="text-muted-foreground">This instance</TableCell>
                        <TableCell>
                          <Badge variant="outline">∞ (Highest)</Badge>
                        </TableCell>
                        <TableCell>
                          <Badge variant="default" className="gap-1 bg-green-600">
                            <CheckCircle className="h-3 w-3" />
                            Active
                          </Badge>
                        </TableCell>
                        <TableCell>-</TableCell>
                        <TableCell>
                          <div className="flex gap-2 text-xs text-muted-foreground">
                            <span className="flex items-center gap-1">
                              <Globe className="h-3 w-3" />
                              {localStats.router_count}
                            </span>
                            <span className="flex items-center gap-1">
                              <Server className="h-3 w-3" />
                              {localStats.service_count}
                            </span>
                            <span className="flex items-center gap-1">
                              <Layers className="h-3 w-3" />
                              {localStats.middleware_count}
                            </span>
                          </div>
                        </TableCell>
                        <TableCell className="text-right">-</TableCell>
                      </TableRow>
                      {providers.length === 0 ? (
                        <TableRow>
                          <TableCell colSpan={7} className="text-center py-8 text-muted-foreground">
                            No external HTTP Providers configured.
                            <br />
                            Add a provider to aggregate external Traefik configurations.
                          </TableCell>
                        </TableRow>
                      ) : (
                        providers.map((provider) => (
                          <TableRow key={provider.id}>
                            <TableCell className="font-medium">{provider.name}</TableCell>
                            <TableCell className="font-mono text-sm max-w-xs truncate">
                              {provider.url}
                            </TableCell>
                            <TableCell>
                              <Badge variant="outline">{provider.priority}</Badge>
                            </TableCell>
                            <TableCell>{getStatusBadge(provider)}</TableCell>
                            <TableCell>
                              {provider.last_fetched
                                ? formatDistanceToNow(new Date(provider.last_fetched), { addSuffix: true })
                                : "Never"}
                            </TableCell>
                            <TableCell>
                              <div className="flex gap-2 text-xs text-muted-foreground">
                                <span className="flex items-center gap-1">
                                  <Globe className="h-3 w-3" />
                                  {provider.router_count}
                                </span>
                                <span className="flex items-center gap-1">
                                  <Server className="h-3 w-3" />
                                  {provider.service_count}
                                </span>
                                <span className="flex items-center gap-1">
                                  <Layers className="h-3 w-3" />
                                  {provider.middleware_count}
                                </span>
                              </div>
                            </TableCell>
                            <TableCell className="text-right">
                              <div className="flex justify-end gap-2">
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  onClick={() => refreshProvider.mutate(provider.id)}
                                  disabled={refreshProvider.isPending}
                                  title="Refresh"
                                >
                                  <RefreshCw className="h-4 w-4" />
                                </Button>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  onClick={() => testProvider.mutate(provider.id)}
                                  disabled={testProvider.isPending}
                                  title="Test Connection"
                                >
                                  <Activity className="h-4 w-4" />
                                </Button>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  onClick={() => handleEdit(provider)}
                                  title="Edit"
                                >
                                  <Edit3 className="h-4 w-4" />
                                </Button>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  onClick={() => handleDelete(provider.id)}
                                  disabled={deleteProvider.isPending}
                                  title="Delete"
                                >
                                  <Trash2 className="h-4 w-4" />
                                </Button>
                              </div>
                            </TableCell>
                          </TableRow>
                        ))
                      )}
                    </TableBody>
                  </Table>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="config" className="space-y-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <div>
                  <CardTitle>Merged Configuration</CardTitle>
                  <CardDescription>
                    Combined configuration from all sources with priority resolution
                  </CardDescription>
                </div>
                <Button variant="outline" size="sm" onClick={refetchConfig}>
                  <RefreshCw className="mr-2 h-4 w-4" />
                  Refresh
                </Button>
              </CardHeader>
              <CardContent className="space-y-4">
                {isLoadingConfig ? (
                  <div className="space-y-2">
                    <Skeleton className="h-32 w-full" />
                  </div>
                ) : mergedConfig ? (
                  <>
                    {/* Sources Overview */}
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                      {mergedConfig.sources?.map((source) => (
                        <Card key={source.name} className="bg-muted/50">
                          <CardContent className="p-4">
                            <div className="flex items-center justify-between mb-2">
                              <span className="font-semibold">{source.name}</span>
                              <Badge
                                variant={
                                  source.status === "healthy"
                                    ? "default"
                                    : source.status === "degraded"
                                    ? "secondary"
                                    : source.status === "inactive"
                                    ? "outline"
                                    : "destructive"
                                }
                              >
                                {source.status}
                              </Badge>
                            </div>
                            <div className="text-xs text-muted-foreground space-y-1">
                              <div>Priority: {source.priority}</div>
                              <div className="flex gap-3">
                                <span>{source.router_count} routers</span>
                                <span>{source.service_count} services</span>
                                <span>{source.middleware_count} middlewares</span>
                              </div>
                            </div>
                          </CardContent>
                        </Card>
                      ))}
                    </div>

                    {/* Conflicts Warning */}
                    {mergedConfig.conflicts && mergedConfig.conflicts.length > 0 && (
                      <Card className="border-yellow-500/50 bg-yellow-500/10">
                        <CardHeader>
                          <CardTitle className="text-yellow-600 flex items-center gap-2">
                            <AlertCircle className="h-5 w-5" />
                            Configuration Conflicts ({mergedConfig.conflicts.length})
                          </CardTitle>
                        </CardHeader>
                        <CardContent>
                          <div className="h-48 overflow-auto">
                            <div className="space-y-2">
                              {mergedConfig.conflicts.map((conflict, idx) => (
                                <div
                                  key={idx}
                                  className="text-sm p-2 rounded bg-yellow-500/20"
                                >
                                  <span className="font-medium capitalize">
                                    {conflict.type}
                                  </span>{" "}
                                  <code className="bg-background px-1 rounded">
                                    {conflict.name}
                                  </code>{" "}
                                  from <strong>{conflict.source}</strong> was overridden by{" "}
                                  <strong>{conflict.overridden_by}</strong> (priority: {" "}
                                  {conflict.source_priority})
                                </div>
                              ))}
                            </div>
                          </div>
                        </CardContent>
                      </Card>
                    )}

                    {/* Merged Config JSON */}
                    <Card>
                      <CardHeader>
                        <CardTitle>Generated Configuration</CardTitle>
                        <CardDescription>
                          This is the final configuration served to Traefik
                        </CardDescription>
                      </CardHeader>
                      <CardContent>
                        <div className="h-96 overflow-auto">
                          <pre className="text-xs bg-muted p-4 rounded-lg">
                            {JSON.stringify(mergedConfig.config, null, 2)}
                          </pre>
                        </div>
                      </CardContent>
                    </Card>
                  </>
                ) : null}
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </ProtectedLayout>
  );
}