import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { servicesApi } from "@/lib/api";
import { Service, CreateServiceRequest, UpdateServiceRequest } from "@/types";
import { toast } from "sonner";

export function useServices() {
  const queryClient = useQueryClient();

  // Fetch all services
  const servicesQuery = useQuery({
    queryKey: ["services"],
    queryFn: async () => {
      const response = await servicesApi.listServices();
      return response.data.services;
    },
  });

  // Create service
  const createService = useMutation({
    mutationFn: (data: CreateServiceRequest) => servicesApi.createService(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["services"] });
      toast.success("Service created successfully");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to create service");
    },
  });

  // Update service
  const updateService = useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateServiceRequest }) =>
      servicesApi.updateService(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["services"] });
      toast.success("Service updated successfully");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to update service");
    },
  });

  // Delete service
  const deleteService = useMutation({
    mutationFn: (id: number) => servicesApi.deleteService(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["services"] });
      toast.success("Service deleted successfully");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to delete service");
    },
  });

  return {
    services: servicesQuery.data || [],
    isLoading: servicesQuery.isLoading,
    isError: servicesQuery.isError,
    error: servicesQuery.error,
    createService,
    updateService,
    deleteService,
    refetch: servicesQuery.refetch,
  };
}

export function useService(id: number) {
  const queryClient = useQueryClient();

  const serviceQuery = useQuery({
    queryKey: ["services", id],
    queryFn: async () => {
      const response = await servicesApi.getService(id);
      return response.data;
    },
    enabled: !!id,
  });

  const updateService = useMutation({
    mutationFn: (data: UpdateServiceRequest) => servicesApi.updateService(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["services", id] });
      queryClient.invalidateQueries({ queryKey: ["services"] });
      toast.success("Service updated successfully");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to update service");
    },
  });

  const deleteService = useMutation({
    mutationFn: () => servicesApi.deleteService(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["services"] });
      toast.success("Service deleted successfully");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to delete service");
    },
  });

  return {
    service: serviceQuery.data,
    isLoading: serviceQuery.isLoading,
    isError: serviceQuery.isError,
    updateService,
    deleteService,
  };
}
