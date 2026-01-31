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
import { useServices } from "@/hooks/use-services";
import { useState } from "react";
import { Plus, Pencil, Trash2, Server, ArrowLeft } from "lucide-react";
import Link from "next/link";
import { CreateServiceRequest, UpdateServiceRequest } from "@/types";

export default function ServicesPage() {
  const { services, isLoading, createService, updateService, deleteService } = useServices();
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [selectedService, setSelectedService] = useState<any>(null);

  const [formData, setFormData] = useState<CreateServiceRequest>({
    name: "",
    servers: [""],
    pass_host_header: true,
  });

  const handleCreate = () => {
    createService.mutate(formData, {
      onSuccess: () => {
        setIsCreateDialogOpen(false);
        setFormData({ name: "", servers: [""], pass_host_header: true });
      },
    });
  };

  const handleEdit = () => {
    if (!selectedService) return;
    const updateData: UpdateServiceRequest = {
      name: formData.name,
      servers: formData.servers.filter((s) => s.trim() !== ""),
      pass_host_header: formData.pass_host_header,
    };
    updateService.mutate(
      { id: selectedService.id, data: updateData },
      {
        onSuccess: () => {
          setIsEditDialogOpen(false);
          setSelectedService(null);
        },
      }
    );
  };

  const handleDelete = () => {
    if (!selectedService) return;
    deleteService.mutate(selectedService.id, {
      onSuccess: () => {
        setIsDeleteDialogOpen(false);
        setSelectedService(null);
      },
    });
  };

  const openEditDialog = (service: any) => {
    setSelectedService(service);
    setFormData({
      name: service.name,
      servers: service.servers.map((s: any) => s.url),
      pass_host_header: service.pass_host_header,
    });
    setIsEditDialogOpen(true);
  };

  const openDeleteDialog = (service: any) => {
    setSelectedService(service);
    setIsDeleteDialogOpen(true);
  };

  const addServerField = () => {
    setFormData({ ...formData, servers: [...formData.servers, ""] });
  };

  const removeServerField = (index: number) => {
    setFormData({
      ...formData,
      servers: formData.servers.filter((_, i) => i !== index),
    });
  };

  const updateServer = (index: number, value: string) => {
    const newServers = [...formData.servers];
    newServers[index] = value;
    setFormData({ ...formData, servers: newServers });
  };

  if (isLoading) {
    return (
      <ProtectedLayout>
        <div className="flex items-center justify-center h-64">Loading services...</div>
      </ProtectedLayout>
    );
  }

  return (
    <ProtectedLayout>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Services</h1>
            <p className="text-muted-foreground">
              Manage backend services for your Traefik routers
            </p>
          </div>
          <Button onClick={() => setIsCreateDialogOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Add Service
          </Button>
        </div>

        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {services.map((service) => (
            <Card key={service.id}>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">{service.name}</CardTitle>
                <Server className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="space-y-2">
                  <div className="text-xs text-muted-foreground">
                    {service.servers.length} server{service.servers.length !== 1 ? "s" : ""}
                  </div>
                  <div className="text-xs space-y-1">
                    {service.servers.slice(0, 2).map((server, idx) => (
                      <div key={idx} className="truncate">
                        {server.url}
                      </div>
                    ))}
                    {service.servers.length > 2 && (
                      <div className="text-muted-foreground">
                        +{service.servers.length - 2} more
                      </div>
                    )}
                  </div>
                </div>
                <div className="mt-4 flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => openEditDialog(service)}
                  >
                    <Pencil className="mr-2 h-4 w-4" />
                    Edit
                  </Button>
                  <Button
                    variant="destructive"
                    size="sm"
                    onClick={() => openDeleteDialog(service)}
                  >
                    <Trash2 className="mr-2 h-4 w-4" />
                    Delete
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>

        {services.length === 0 && (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-10">
              <Server className="h-12 w-12 text-muted-foreground mb-4" />
              <CardTitle className="mb-2">No services yet</CardTitle>
              <CardDescription className="text-center mb-4">
                Create your first service to start routing traffic
              </CardDescription>
              <Button onClick={() => setIsCreateDialogOpen(true)}>
                <Plus className="mr-2 h-4 w-4" />
                Add Service
              </Button>
            </CardContent>
          </Card>
        )}

        {/* Create Dialog */}
        <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
          <DialogContent className="max-w-lg">
            <DialogHeader>
              <DialogTitle>Create Service</DialogTitle>
              <DialogDescription>
                Add a new backend service for your routers
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="name">Service Name</Label>
                <Input
                  id="name"
                  value={formData.name}
                  onChange={(e) =>
                    setFormData({ ...formData, name: e.target.value })
                  }
                  placeholder="my-service"
                />
              </div>
              <div className="space-y-2">
                <Label>Servers</Label>
                {formData.servers.map((server, index) => (
                  <div key={index} className="flex gap-2">
                    <Input
                      value={server}
                      onChange={(e) => updateServer(index, e.target.value)}
                      placeholder="http://localhost:8080"
                    />
                    {formData.servers.length > 1 && (
                      <Button
                        type="button"
                        variant="outline"
                        size="icon"
                        onClick={() => removeServerField(index)}
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
                  onClick={addServerField}
                >
                  <Plus className="mr-2 h-4 w-4" />
                  Add Server
                </Button>
              </div>
            </div>
            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => setIsCreateDialogOpen(false)}
              >
                Cancel
              </Button>
              <Button onClick={handleCreate} disabled={createService.isPending}>
                {createService.isPending ? "Creating..." : "Create"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Edit Dialog */}
        <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
          <DialogContent className="max-w-lg">
            <DialogHeader>
              <DialogTitle>Edit Service</DialogTitle>
              <DialogDescription>Update service configuration</DialogDescription>
            </DialogHeader>
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="edit-name">Service Name</Label>
                <Input
                  id="edit-name"
                  value={formData.name}
                  onChange={(e) =>
                    setFormData({ ...formData, name: e.target.value })
                  }
                />
              </div>
              <div className="space-y-2">
                <Label>Servers</Label>
                {formData.servers.map((server, index) => (
                  <div key={index} className="flex gap-2">
                    <Input
                      value={server}
                      onChange={(e) => updateServer(index, e.target.value)}
                      placeholder="http://localhost:8080"
                    />
                    {formData.servers.length > 1 && (
                      <Button
                        type="button"
                        variant="outline"
                        size="icon"
                        onClick={() => removeServerField(index)}
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
                  onClick={addServerField}
                >
                  <Plus className="mr-2 h-4 w-4" />
                  Add Server
                </Button>
              </div>
            </div>
            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => setIsEditDialogOpen(false)}
              >
                Cancel
              </Button>
              <Button onClick={handleEdit} disabled={updateService.isPending}>
                {updateService.isPending ? "Updating..." : "Update"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Delete Dialog */}
        <Dialog open={isDeleteDialogOpen} onOpenChange={setIsDeleteDialogOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Delete Service</DialogTitle>
              <DialogDescription>
                Are you sure you want to delete &quot;{selectedService?.name}&quot;? This
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
                disabled={deleteService.isPending}
              >
                {deleteService.isPending ? "Deleting..." : "Delete"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </ProtectedLayout>
  );
}
