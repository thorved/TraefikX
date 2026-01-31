import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { httpProvidersApi } from "@/lib/api";
import {
  HTTPProvider,
  CreateHTTPProviderRequest,
  UpdateHTTPProviderRequest,
  MergedTraefikConfig,
} from "@/types";
import { toast } from "sonner";

export function useHTTPProviders() {
  const queryClient = useQueryClient();

  // Fetch all providers
  const providersQuery = useQuery({
    queryKey: ["http-providers"],
    queryFn: async () => {
      const response = await httpProvidersApi.listProviders();
      return response.data.providers;
    },
  });

  // Create provider
  const createProvider = useMutation({
    mutationFn: (data: CreateHTTPProviderRequest) =>
      httpProvidersApi.createProvider(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["http-providers"] });
      toast.success("HTTP Provider created successfully");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to create HTTP Provider");
    },
  });

  // Update provider
  const updateProvider = useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: number;
      data: UpdateHTTPProviderRequest;
    }) => httpProvidersApi.updateProvider(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["http-providers"] });
      toast.success("HTTP Provider updated successfully");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to update HTTP Provider");
    },
  });

  // Delete provider
  const deleteProvider = useMutation({
    mutationFn: (id: number) => httpProvidersApi.deleteProvider(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["http-providers"] });
      toast.success("HTTP Provider deleted successfully");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to delete HTTP Provider");
    },
  });

  // Refresh provider
  const refreshProvider = useMutation({
    mutationFn: (id: number) => httpProvidersApi.refreshProvider(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["http-providers"] });
      toast.success("HTTP Provider refresh triggered");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to refresh HTTP Provider");
    },
  });

  // Test provider
  const testProvider = useMutation({
    mutationFn: (id: number) => httpProvidersApi.testProvider(id),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["http-providers"] });
      const provider = data.data;
      if (provider.last_error) {
        toast.warning(`HTTP Provider test completed with error: ${provider.last_error}`);
      } else {
        toast.success("HTTP Provider is healthy");
      }
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to test HTTP Provider");
    },
  });

  return {
    providers: providersQuery.data || [],
    isLoading: providersQuery.isLoading,
    isError: providersQuery.isError,
    error: providersQuery.error,
    createProvider,
    updateProvider,
    deleteProvider,
    refreshProvider,
    testProvider,
    refetch: providersQuery.refetch,
  };
}

export function useHTTPProvider(id: number) {
  const queryClient = useQueryClient();

  const providerQuery = useQuery({
    queryKey: ["http-providers", id],
    queryFn: async () => {
      const response = await httpProvidersApi.getProvider(id);
      return response.data;
    },
    enabled: !!id,
  });

  const updateProvider = useMutation({
    mutationFn: (data: UpdateHTTPProviderRequest) =>
      httpProvidersApi.updateProvider(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["http-providers", id] });
      queryClient.invalidateQueries({ queryKey: ["http-providers"] });
      toast.success("HTTP Provider updated successfully");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to update HTTP Provider");
    },
  });

  const deleteProvider = useMutation({
    mutationFn: () => httpProvidersApi.deleteProvider(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["http-providers"] });
      toast.success("HTTP Provider deleted successfully");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to delete HTTP Provider");
    },
  });

  const refreshProvider = useMutation({
    mutationFn: () => httpProvidersApi.refreshProvider(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["http-providers", id] });
      queryClient.invalidateQueries({ queryKey: ["http-providers"] });
      toast.success("HTTP Provider refresh triggered");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to refresh HTTP Provider");
    },
  });

  const testProvider = useMutation({
    mutationFn: () => httpProvidersApi.testProvider(id),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["http-providers", id] });
      queryClient.invalidateQueries({ queryKey: ["http-providers"] });
      const provider = data.data;
      if (provider.last_error) {
        toast.warning(`HTTP Provider test completed with error: ${provider.last_error}`);
      } else {
        toast.success("HTTP Provider is healthy");
      }
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to test HTTP Provider");
    },
  });

  return {
    provider: providerQuery.data,
    isLoading: providerQuery.isLoading,
    isError: providerQuery.isError,
    updateProvider,
    deleteProvider,
    refreshProvider,
    testProvider,
  };
}

export function useMergedConfig(refetchInterval = 30000) {
  const queryClient = useQueryClient();

  const mergedConfigQuery = useQuery({
    queryKey: ["traefik-merged-config"],
    queryFn: async () => {
      const response = await httpProvidersApi.getMergedConfig();
      return response.data;
    },
    refetchInterval,
  });

  const refetch = () => {
    queryClient.invalidateQueries({ queryKey: ["traefik-merged-config"] });
  };

  return {
    mergedConfig: mergedConfigQuery.data,
    isLoading: mergedConfigQuery.isLoading,
    isError: mergedConfigQuery.isError,
    error: mergedConfigQuery.error,
    refetch,
  };
}