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
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useState } from "react";
import { useProxies } from "@/hooks/use-proxies";
import { Plus, X, AlertCircle } from "lucide-react";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

interface AddProxyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

// Domain validation regex
const DOMAIN_REGEX = /^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$/;

export function AddProxyDialog({ open, onOpenChange }: AddProxyDialogProps) {
  const { createProxy } = useProxies();
  const [domainNames, setDomainNames] = useState<string[]>([""]);
  const [domainErrors, setDomainErrors] = useState<(string | null)[]>([null]);
  const [forwardScheme, setForwardScheme] = useState<"http" | "https">("http");
  const [forwardHost, setForwardHost] = useState("");
  const [forwardPort, setForwardPort] = useState("80");
  const [ssl, setSsl] = useState(false);
  const [access, setAccess] = useState<"public" | "private">("public");

  const validateDomain = (domain: string): string | null => {
    if (!domain.trim()) return null; // Empty is ok while typing
    if (!DOMAIN_REGEX.test(domain)) {
      return "Invalid domain format";
    }
    if (!domain.includes(".")) {
      return "Domain must contain a dot (e.g., example.com)";
    }
    return null;
  };

  const handleAddDomain = () => {
    setDomainNames([...domainNames, ""]);
    setDomainErrors([...domainErrors, null]);
  };

  const handleRemoveDomain = (index: number) => {
    setDomainNames(domainNames.filter((_, i) => i !== index));
    setDomainErrors(domainErrors.filter((_, i) => i !== index));
  };

  const handleDomainChange = (index: number, value: string) => {
    const newDomains = [...domainNames];
    newDomains[index] = value;
    setDomainNames(newDomains);

    // Validate domain
    const newErrors = [...domainErrors];
    newErrors[index] = validateDomain(value);
    setDomainErrors(newErrors);
  };

  const handleSubmit = () => {
    const validDomains = domainNames.filter((d) => d.trim() !== "");
    if (validDomains.length === 0 || !forwardHost || !forwardPort) return;

    // Validate all domains
    const errors = domainNames.map((d) => d.trim() ? validateDomain(d) : null);
    setDomainErrors(errors);

    if (errors.some((e) => e !== null)) {
      return;
    }

    createProxy.mutate(
      {
        domain_names: validDomains,
        forward_scheme: forwardScheme,
        forward_host: forwardHost,
        forward_port: parseInt(forwardPort),
        ssl,
        access,
      },
      {
        onSuccess: () => {
          onOpenChange(false);
          resetForm();
        },
      }
    );
  };

  const resetForm = () => {
    setDomainNames([""]);
    setDomainErrors([null]);
    setForwardScheme("http");
    setForwardHost("");
    setForwardPort("80");
    setSsl(false);
    setAccess("public");
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Add Proxy Host</DialogTitle>
          <DialogDescription>
            Create a new reverse proxy to forward traffic to your service
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <Label>Domain Names</Label>
            {domainNames.map((domain, index) => (
              <div key={index} className="space-y-1">
                <div className="flex gap-2">
                  <Input
                    placeholder="example.com"
                    value={domain}
                    onChange={(e) => handleDomainChange(index, e.target.value)}
                    className={domainErrors[index] ? "border-red-500" : ""}
                  />
                  {domainNames.length > 1 && (
                    <Button
                      type="button"
                      variant="outline"
                      size="icon"
                      onClick={() => handleRemoveDomain(index)}
                    >
                      <X className="h-4 w-4" />
                    </Button>
                  )}
                </div>
                {domainErrors[index] && (
                  <div className="flex items-center gap-1 text-red-500 text-sm">
                    <AlertCircle className="h-4 w-4" />
                    {domainErrors[index]}
                  </div>
                )}
              </div>
            ))}
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={handleAddDomain}
            >
              <Plus className="mr-2 h-4 w-4" />
              Add Domain
            </Button>
          </div>

          <div className="space-y-2">
            <Label>Forward Target</Label>
            <div className="grid grid-cols-12 gap-2">
              <div className="col-span-3">
                <Select
                  value={forwardScheme}
                  onValueChange={(v: "http" | "https") => setForwardScheme(v)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Scheme" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="http">http://</SelectItem>
                    <SelectItem value="https">https://</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="col-span-6">
                <Input
                  placeholder="Hostname or IP"
                  value={forwardHost}
                  onChange={(e) => setForwardHost(e.target.value)}
                />
              </div>

              <div className="col-span-3">
                <Input
                  type="number"
                  placeholder="Port"
                  value={forwardPort}
                  onChange={(e) => setForwardPort(e.target.value)}
                />
              </div>
            </div>
            <p className="text-xs text-muted-foreground">
              Example: http://192.168.1.100:8080 or https://container-name:3000
            </p>
          </div>

          <div className="space-y-2">
            <Label>Access</Label>
            <Select
              value={access}
              onValueChange={(v: "public" | "private") => setAccess(v)}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="public">Publicly Accessible</SelectItem>
                <SelectItem value="private">Private (Authentication Required)</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex items-center justify-between rounded-lg border p-4">
            <div className="space-y-0.5">
              <Label className="text-base">SSL Certificate</Label>
              <p className="text-sm text-muted-foreground">
                Enable HTTPS with automatic Let&apos;s Encrypt certificate
              </p>
            </div>
            <Switch checked={ssl} onCheckedChange={setSsl} />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={
              createProxy.isPending ||
              domainNames.filter((d) => d.trim() !== "").length === 0 ||
              !forwardHost ||
              !forwardPort ||
              domainErrors.some((e) => e !== null)
            }
          >
            {createProxy.isPending ? "Creating..." : "Save"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
