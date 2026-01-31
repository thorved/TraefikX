import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { middlewaresApi } from "@/lib/api";
import { Middleware, CreateMiddlewareRequest, UpdateMiddlewareRequest } from "@/types";
import { toast } from "sonner";

export function useMiddlewares() {
  const queryClient = useQueryClient();

  // Fetch all middlewares
  const middlewaresQuery = useQuery({
    queryKey: ["middlewares"],
    queryFn: async () => {
      const response = await middlewaresApi.listMiddlewares();
      return response.data.middlewares;
    },
  });

  // Create middleware
  const createMiddleware = useMutation({
    mutationFn: (data: CreateMiddlewareRequest) => middlewaresApi.createMiddleware(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["middlewares"] });
      toast.success("Middleware created successfully");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to create middleware");
    },
  });

  // Update middleware
  const updateMiddleware = useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateMiddlewareRequest }) =>
      middlewaresApi.updateMiddleware(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["middlewares"] });
      toast.success("Middleware updated successfully");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to update middleware");
    },
  });

  // Delete middleware
  const deleteMiddleware = useMutation({
    mutationFn: (id: number) => middlewaresApi.deleteMiddleware(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["middlewares"] });
      toast.success("Middleware deleted successfully");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to delete middleware");
    },
  });

  return {
    middlewares: middlewaresQuery.data || [],
    isLoading: middlewaresQuery.isLoading,
    isError: middlewaresQuery.isError,
    error: middlewaresQuery.error,
    createMiddleware,
    updateMiddleware,
    deleteMiddleware,
    refetch: middlewaresQuery.refetch,
  };
}

export function useMiddleware(id: number) {
  const queryClient = useQueryClient();

  const middlewareQuery = useQuery({
    queryKey: ["middlewares", id],
    queryFn: async () => {
      const response = await middlewaresApi.getMiddleware(id);
      return response.data;
    },
    enabled: !!id,
  });

  const updateMiddleware = useMutation({
    mutationFn: (data: UpdateMiddlewareRequest) => middlewaresApi.updateMiddleware(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["middlewares", id] });
      queryClient.invalidateQueries({ queryKey: ["middlewares"] });
      toast.success("Middleware updated successfully");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to update middleware");
    },
  });

  const deleteMiddleware = useMutation({
    mutationFn: () => middlewaresApi.deleteMiddleware(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["middlewares"] });
      toast.success("Middleware deleted successfully");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to delete middleware");
    },
  });

  return {
    middleware: middlewareQuery.data,
    isLoading: middlewareQuery.isLoading,
    isError: middlewareQuery.isError,
    updateMiddleware,
    deleteMiddleware,
  };
}
