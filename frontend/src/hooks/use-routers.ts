import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { routersApi } from "@/lib/api";
import { Router, CreateRouterRequest, UpdateRouterRequest } from "@/types";
import { toast } from "sonner";

export function useRouters() {
  const queryClient = useQueryClient();

  // Fetch all routers
  const routersQuery = useQuery({
    queryKey: ["routers"],
    queryFn: async () => {
      const response = await routersApi.listRouters();
      return response.data.routers;
    },
  });

  // Create router
  const createRouter = useMutation({
    mutationFn: (data: CreateRouterRequest) => routersApi.createRouter(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["routers"] });
      toast.success("Router created successfully");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to create router");
    },
  });

  // Update router
  const updateRouter = useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateRouterRequest }) =>
      routersApi.updateRouter(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["routers"] });
      toast.success("Router updated successfully");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to update router");
    },
  });

  // Delete router
  const deleteRouter = useMutation({
    mutationFn: (id: number) => routersApi.deleteRouter(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["routers"] });
      toast.success("Router deleted successfully");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to delete router");
    },
  });

  return {
    routers: routersQuery.data || [],
    isLoading: routersQuery.isLoading,
    isError: routersQuery.isError,
    error: routersQuery.error,
    createRouter,
    updateRouter,
    deleteRouter,
    refetch: routersQuery.refetch,
  };
}

export function useRouter(id: number) {
  const queryClient = useQueryClient();

  const routerQuery = useQuery({
    queryKey: ["routers", id],
    queryFn: async () => {
      const response = await routersApi.getRouter(id);
      return response.data;
    },
    enabled: !!id,
  });

  const updateRouter = useMutation({
    mutationFn: (data: UpdateRouterRequest) => routersApi.updateRouter(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["routers", id] });
      queryClient.invalidateQueries({ queryKey: ["routers"] });
      toast.success("Router updated successfully");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to update router");
    },
  });

  const deleteRouter = useMutation({
    mutationFn: () => routersApi.deleteRouter(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["routers"] });
      toast.success("Router deleted successfully");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to delete router");
    },
  });

  return {
    router: routerQuery.data,
    isLoading: routerQuery.isLoading,
    isError: routerQuery.isError,
    updateRouter,
    deleteRouter,
  };
}
