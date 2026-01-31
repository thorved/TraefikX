"use client";

import { ProtectedLayout } from "@/components/layout/protected-layout";
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
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useRouters } from "@/hooks/use-routers";
import { useServices } from "@/hooks/use-services";
import { useState } from "react";
import { Plus, Pencil, Trash2, Network, Check, X } from "lucide-react";
import { CreateRouterRequest } from "@/types";

export default function RoutersPage() {
  const { routers, isLoading, createRouter, deleteRouter } = useRouters();
  const { services } = useServices();
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [selectedRouter, setSelectedRouter] = useState<any>(null);

  const [formData, setFormData] = useState<CreateRouterRequest>({
    name: "",
    hostnames: [""],
    service_id: 0,
    tls_enabled: true,
    redirect_https: true,
    entry_points: ["web", "websecure"],
  });

  const handleCreate = () => {
    createRouter.mutate(formData, {
      onSuccess: () => {
        setIsCreateDialogOpen(false);
        setFormData({
          name: "",
          hostnames: [""],
          service_id: 0,
          tls_enabled: true,
          redirect_https: true,
          entry_points: ["web", "websecure"],
        });
      },
    });
  };

  const handleDelete = () => {
    if (!selectedRouter) return;
    deleteRouter.mutate(selectedRouter.id, {
      onSuccess: () => {
        setIsDeleteDialogOpen(false);
        setSelectedRouter(null);
      },
    });
  };

  const openDeleteDialog = (router: any) => {
    setSelectedRouter(router);
    setIsDeleteDialogOpen(true);
  };

  const addHostnameField = () => {
    setFormData({ ...formData, hostnames: [...formData.hostnames, ""] });
  };

  const removeHostnameField = (index: number) => {
    setFormData({
      ...formData,
      hostnames: formData.hostnames.filter((_, i) => i !== index),
    });
  };

  const updateHostname = (index: number, value: string) => {
    const newHostnames = [...formData.hostnames];
    newHostnames[index] = value;
    setFormData({ ...formData, hostnames: newHostnames });
  };

  if (isLoading) {
    return (
      <ProtectedLayout>
        <div className="flex items-center justify-center h-64">Loading routers...</div>
      </ProtectedLayout>
    );
  }

  return (
    <ProtectedLayout>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Routers</h1>
            <p className="text-muted-foreground">
              Manage Traefik routers and routing rules
            </p>
          </div>
          <Button onClick={() => setIsCreateDialogOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Add Router
          </Button>
        </div>

        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {routers.map((router) => (
            <Card key={router.id}>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">{router.name}</CardTitle>
                <Network className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="space-y-2">
                  <div className="text-xs text-muted-foreground">
                    Service: {router.service_name}
                  </div>
                  <div className="text-xs space-y-1">
                    {router.hostnames.map((hostname, idx) => (
                      <div key={idx} className="truncate">
                        {hostname}
                      </div>
                    ))}
                  </div>
                  <div className="flex gap-2 mt-2">
                    {router.tls_enabled && (
                      <span className="text-xs bg-green-100 text-green-800 px-2 py-1 rounded">
                        TLS
                      </span>
                    )}
                    {router.redirect_https && (
                      <span className="text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded">
                        HTTPS Redirect
                      </span>
                    )}
                  </div>
                </div>
                <div className="mt-4 flex gap-2">
                  <Button
                    variant="destructive"
                    size="sm"
                    onClick={() => openDeleteDialog(router)}
                  >
                    <Trash2 className="mr-2 h-4 w-4" />
                    Delete
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>

        {routers.length === 0 && (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-10">
              <Network className="h-12 w-12 text-muted-foreground mb-4" />
              <CardTitle className="mb-2">No routers yet</CardTitle>
              <CardDescription className="text-center mb-4">
                Create your first router to start routing traffic
              </CardDescription>
              <Button onClick={() => setIsCreateDialogOpen(true)}>
                <Plus className="mr-2 h-4 w-4" />
                Add Router
              </Button>
            </CardContent>
          </Card>
        )}

        {/* Create Dialog */}
        <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
          <DialogContent className="max-w-lg">
            <DialogHeader>
              <DialogTitle>Create Router</DialogTitle>
              <DialogDescription>
                Add a new router to route traffic to your services
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="name">Router Name</Label>
                <Input
                  id="name"
                  value={formData.name}
                  onChange={(e) =>
                    setFormData({ ...formData, name: e.target.value })
                  }
                  placeholder="my-router"
                />
              </div>
              <div className="space-y-2">
                <Label>Hostnames</Label>
                {formData.hostnames.map((hostname, index) => (
                  <div key={index} className="flex gap-2">
                    <Input
                      value={hostname}
                      onChange={(e) => updateHostname(index, e.target.value)}
                      placeholder="app.example.com"
                    />
                    {formData.hostnames.length > 1 && (
                      <Button
                        type="button"
                        variant="outline"
                        size="icon"
                        onClick={() => removeHostnameField(index)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    )}
                  </div>
                ))}
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={addHostnameField}
                >
                  <Plus className="mr-2 h-4 w-4" />
                  Add Hostname
                </Button>
              </div>
              <div className="space-y-2">
                <Label htmlFor="service">Service</Label>
                <select
                  id="service"
                  value={formData.service_id}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      service_id: parseInt(e.target.value),
                    })
                  }
                  className="w-full p-2 border rounded-md"
                >
                  <option value={0}>Select a service</option>
                  {services.map((service) => (
                    <option key={service.id} value={service.id}>
                      {service.name}
                    </option>
                  ))}
                </select>
              </div>
              <div className="flex items-center gap-4">
                <label className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={formData.tls_enabled}
                    onChange={(e) =>
                      setFormData({ ...formData, tls_enabled: e.target.checked })
                    }
                  />
                  Enable TLS
                </label>
                <label className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={formData.redirect_https}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        redirect_https: e.target.checked,
                      })
                    }
                  />
                  Redirect HTTP to HTTPS
                </label>
              </div>
            </div>
            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => setIsCreateDialogOpen(false)}
              >
                Cancel
              </Button>
              <Button
                onClick={handleCreate}
                disabled={
                  createRouter.isPending ||
                  !formData.name ||
                  !formData.service_id
                }
              >
                {createRouter.isPending ? "Creating..." : "Create"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Delete Dialog */}
        <Dialog open={isDeleteDialogOpen} onOpenChange={setIsDeleteDialogOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Delete Router</DialogTitle>
              <DialogDescription>
                Are you sure you want to delete &quot;{selectedRouter?.name}&quot;? This
                action cannot be undone.
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => setIsDeleteDialogOpen(false)}
              >
                Cancel
              </Button>
              <Button
                variant="destructive"
                onClick={handleDelete}
                disabled={deleteRouter.isPending}
              >
                {deleteRouter.isPending ? "Deleting..." : "Delete"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </ProtectedLayout>
  );
}
