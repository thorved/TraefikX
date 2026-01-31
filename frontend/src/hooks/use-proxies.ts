import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { proxyApi } from "@/lib/api";
import { ProxyHost, CreateProxyHostRequest, UpdateProxyHostRequest } from "@/types";
import { toast } from "sonner";

export function useProxies() {
  const queryClient = useQueryClient();

  // Fetch all proxy hosts
  const proxiesQuery = useQuery({
    queryKey: ["proxies"],
    queryFn: async () => {
      const response = await proxyApi.listProxies();
      return response.data.proxies;
    },
  });

  // Create proxy host
  const createProxy = useMutation({
    mutationFn: (data: CreateProxyHostRequest) => proxyApi.createProxy(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["proxies"] });
      toast.success("Proxy host created successfully");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to create proxy host");
    },
  });

  // Update proxy host
  const updateProxy = useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateProxyHostRequest }) =>
      proxyApi.updateProxy(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["proxies"] });
      toast.success("Proxy host updated successfully");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to update proxy host");
    },
  });

  // Delete proxy host
  const deleteProxy = useMutation({
    mutationFn: (id: number) => proxyApi.deleteProxy(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["proxies"] });
      toast.success("Proxy host deleted successfully");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to delete proxy host");
    },
  });

  return {
    proxies: proxiesQuery.data || [],
    isLoading: proxiesQuery.isLoading,
    isError: proxiesQuery.isError,
    error: proxiesQuery.error,
    createProxy,
    updateProxy,
    deleteProxy,
    refetch: proxiesQuery.refetch,
  };
}
