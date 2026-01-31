"use client";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { useProxies } from "@/hooks/use-proxies";
import { ProxyHost } from "@/types";
import { AlertTriangle } from "lucide-react";

interface DeleteProxyDialogProps {
  proxy: ProxyHost | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function DeleteProxyDialog({
  proxy,
  open,
  onOpenChange,
}: DeleteProxyDialogProps) {
  const { deleteProxy } = useProxies();

  const handleDelete = () => {
    if (!proxy) return;

    deleteProxy.mutate(proxy.id, {
      onSuccess: () => {
        onOpenChange(false);
      },
    });
  };

  if (!proxy) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <div className="flex items-center gap-2">
            <AlertTriangle className="h-6 w-6 text-red-500" />
            <DialogTitle>Delete Proxy Host</DialogTitle>
          </div>
          <DialogDescription className="pt-2">
            Are you sure you want to delete this proxy host?
            <br />
            <strong>{proxy.domain_names.join(", ")}</strong>
            <br />
            <br />
            This action cannot be undone. The domain will stop working
            immediately.
          </DialogDescription>
        </DialogHeader>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={handleDelete}
            disabled={deleteProxy.isPending}
          >
            {deleteProxy.isPending ? "Deleting..." : "Delete"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
