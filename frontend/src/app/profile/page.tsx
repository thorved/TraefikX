"use client";

import { useState, useEffect } from "react";
import { ProtectedLayout } from "@/components/layout/protected-layout";
import { useAuth } from "@/contexts/AuthContext";
import { authApi } from "@/lib/api";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Separator } from "@/components/ui/separator";
import { Badge } from "@/components/ui/badge";
import {
  User,
  Mail,
  Shield,
  Key,
  Link as LinkIcon,
  Unlink,
  Check,
  AlertCircle,
} from "lucide-react";
import { toast } from "sonner";

export default function ProfilePage() {
  const { user, refreshUser } = useAuth();
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [oidcStatus, setOidcStatus] = useState<{
    enabled: boolean;
    provider_name?: string;
  }>({
    enabled: false,
  });

  useEffect(() => {
    const fetchOIDCStatus = async () => {
      try {
        const response = await authApi.getOIDCStatus();
        setOidcStatus(response.data);
      } catch (error) {
        console.error("Failed to fetch OIDC status:", error);
      }
    };
    fetchOIDCStatus();
  }, []);

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();

    if (newPassword !== confirmPassword) {
      toast.error("Passwords do not match");
      return;
    }

    if (newPassword.length < 12) {
      toast.error("Password must be at least 12 characters");
      return;
    }

    setIsLoading(true);
    try {
      await authApi.changePassword(
        user?.password_enabled ? currentPassword : undefined,
        newPassword,
      );
      toast.success("Password changed successfully");
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
      refreshUser();
    } catch (error: any) {
      toast.error(error.response?.data?.error || "Failed to change password");
    } finally {
      setIsLoading(false);
    }
  };

  const handleTogglePassword = async (enabled: boolean) => {
    try {
      await authApi.togglePasswordLogin(enabled);
      toast.success(`Password login ${enabled ? "enabled" : "disabled"}`);
      refreshUser();
    } catch (error: any) {
      toast.error(error.response?.data?.error || "Failed to update setting");
    }
  };

  const handleLinkOIDC = async () => {
    try {
      const response = await authApi.initiateOIDCLink();
      window.location.href = response.data.auth_url;
    } catch (error: any) {
      toast.error(
        error.response?.data?.error || "Failed to initiate OIDC linking",
      );
    }
  };

  const handleUnlinkOIDC = async () => {
    if (!confirm("Are you sure you want to unlink your OIDC account?")) return;

    try {
      await authApi.unlinkOIDC();
      toast.success("OIDC account unlinked successfully");
      refreshUser();
    } catch (error: any) {
      toast.error(
        error.response?.data?.error || "Failed to unlink OIDC account",
      );
    }
  };

  return (
    <ProtectedLayout>
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Profile</h1>
          <p className="text-muted-foreground">
            Manage your account settings and security
          </p>
        </div>

        <div className="grid gap-6 md:grid-cols-2">
          {/* Account Info */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <User className="h-5 w-5" />
                Account Information
              </CardTitle>
              <CardDescription>Your account details and status</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-1">
                <Label className="text-muted-foreground">Email</Label>
                <div className="flex items-center gap-2">
                  <Mail className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">{user?.email}</span>
                </div>
              </div>

              <Separator />

              <div className="space-y-1">
                <Label className="text-muted-foreground">Role</Label>
                <div className="flex items-center gap-2">
                  <Shield className="h-4 w-4 text-muted-foreground" />
                  <Badge
                    variant={user?.role === "admin" ? "default" : "secondary"}
                  >
                    {user?.role === "admin" ? "Administrator" : "User"}
                  </Badge>
                </div>
              </div>

              <Separator />

              <div className="space-y-1">
                <Label className="text-muted-foreground">Status</Label>
                <div className="flex items-center gap-2">
                  <Check className="h-4 w-4 text-green-600" />
                  <span className="text-green-600 font-medium">Active</span>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Authentication Methods */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Key className="h-5 w-5" />
                Authentication Methods
              </CardTitle>
              <CardDescription>Manage how you sign in</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              {/* Password Login */}
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label>Password Login</Label>
                  <p className="text-sm text-muted-foreground">
                    {user?.password_enabled
                      ? "Password authentication is enabled"
                      : "Password authentication is disabled"}
                  </p>
                </div>
                <Switch
                  checked={user?.password_enabled}
                  onCheckedChange={handleTogglePassword}
                />
              </div>

              <Separator />

              {/* OIDC */}
              {oidcStatus.enabled && (
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <div className="space-y-0.5">
                      <Label>{oidcStatus.provider_name || "OIDC"} Login</Label>
                      <p className="text-sm text-muted-foreground">
                        {user?.is_linked_to_oidc
                          ? `Linked to ${oidcStatus.provider_name || "OIDC"}`
                          : "Not linked"}
                      </p>
                    </div>
                    {user?.is_linked_to_oidc ? (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={handleUnlinkOIDC}
                        disabled={!user?.password_enabled}
                      >
                        <Unlink className="mr-2 h-4 w-4" />
                        Unlink
                      </Button>
                    ) : (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={handleLinkOIDC}
                      >
                        <LinkIcon className="mr-2 h-4 w-4" />
                        Link Account
                      </Button>
                    )}
                  </div>

                  {!user?.password_enabled && user?.is_linked_to_oidc && (
                    <div className="flex items-start gap-2 rounded-md bg-yellow-50 p-3 text-sm text-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-200">
                      <AlertCircle className="h-4 w-4 shrink-0 mt-0.5" />
                      <p>
                        You cannot unlink OIDC while password login is disabled.
                        Enable password login first.
                      </p>
                    </div>
                  )}
                </div>
              )}
            </CardContent>
          </Card>

          {/* Change Password */}
          <Card className="md:col-span-2">
            <CardHeader>
              <CardTitle>Change Password</CardTitle>
              <CardDescription>
                Update your password to keep your account secure
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form
                onSubmit={handleChangePassword}
                className="space-y-4 max-w-md"
              >
                {user?.password_enabled && (
                  <div className="space-y-2">
                    <Label htmlFor="current">Current Password</Label>
                    <Input
                      id="current"
                      type="password"
                      value={currentPassword}
                      onChange={(e) => setCurrentPassword(e.target.value)}
                      required={user?.password_enabled}
                    />
                  </div>
                )}

                <div className="space-y-2">
                  <Label htmlFor="new">New Password</Label>
                  <Input
                    id="new"
                    type="password"
                    value={newPassword}
                    onChange={(e) => setNewPassword(e.target.value)}
                    required
                    minLength={12}
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="confirm">Confirm New Password</Label>
                  <Input
                    id="confirm"
                    type="password"
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                    required
                  />
                </div>

                <p className="text-sm text-muted-foreground">
                  Password must be at least 12 characters with uppercase,
                  lowercase, number, and special character.
                </p>

                <Button type="submit" disabled={isLoading}>
                  {isLoading ? "Updating..." : "Update Password"}
                </Button>
              </form>
            </CardContent>
          </Card>
        </div>
      </div>
    </ProtectedLayout>
  );
}
