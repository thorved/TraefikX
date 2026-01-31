"use client";

import { useEffect, useState, Suspense, useRef } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useAuth } from "@/contexts/AuthContext";
import { authApi } from "@/lib/api";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { AlertCircle, CheckCircle2, Loader2 } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

function OIDCCallbackContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [status, setStatus] = useState<"loading" | "success" | "error">(
    "loading",
  );
  const [errorMessage, setErrorMessage] = useState("");

  const { refreshUser } = useAuth();
  const processedRef = useRef(false);

  useEffect(() => {
    if (processedRef.current) return;

    // Check if we have parameters to process
    const code = searchParams.get("code");
    const state = searchParams.get("state");
    const error = searchParams.get("error");

    // If no params, don't mark as processed yet (wait for hydration?) 
    // Actually searchParams should be available.
    if (!code && !state && !error) return;

    processedRef.current = true;

    if (error) {
      setStatus("error");
      setErrorMessage(
        searchParams.get("error_description") || "Authentication failed",
      );
      setTimeout(() => router.push("/login"), 3000);
      return;
    }

    if (!code || !state) {
      setStatus("error");
      setErrorMessage("Invalid callback parameters");
      setTimeout(() => router.push("/login"), 3000);
      return;
    }

    const handleCallback = async () => {
      try {
        const response = await authApi.oidcCallback(code, state);

        const { access_token, refresh_token } = response.data;
        localStorage.setItem("access_token", access_token);
        localStorage.setItem("refresh_token", refresh_token);

        // Update auth context
        await refreshUser();

        setStatus("success");
        setTimeout(() => router.push("/dashboard"), 1500);
      } catch (err: any) {
        console.error("OIDC Callback error:", err);
        setStatus("error");
        setErrorMessage(err.response?.data?.error || "Authentication failed");
        setTimeout(() => router.push("/login"), 3000);
      }
    };

    handleCallback();
  }, [searchParams, router, refreshUser]);

  return (
    <Card className="w-full max-w-md">
      <CardHeader className="text-center">
        <CardTitle className="text-2xl">OIDC Authentication</CardTitle>
        <CardDescription>
          {status === "loading" && "Processing your authentication..."}
          {status === "success" && "Authentication successful!"}
          {status === "error" && "Authentication failed"}
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col items-center space-y-4">
        {status === "loading" && (
          <>
            <Loader2 className="h-12 w-12 animate-spin text-primary" />
            <p className="text-sm text-muted-foreground">
              Please wait while we verify your credentials...
            </p>
          </>
        )}

        {status === "success" && (
          <>
            <CheckCircle2 className="h-12 w-12 text-green-600" />
            <p className="text-sm text-muted-foreground">
              Redirecting to dashboard...
            </p>
          </>
        )}

        {status === "error" && (
          <>
            <AlertCircle className="h-12 w-12 text-red-600" />
            <Alert variant="destructive">
              <AlertTitle>Error</AlertTitle>
              <AlertDescription>{errorMessage}</AlertDescription>
            </Alert>
            <p className="text-sm text-muted-foreground">
              Redirecting to login page...
            </p>
          </>
        )}
      </CardContent>
    </Card>
  );
}

export default function OIDCCallbackPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-background via-background to-muted p-4">
      <Suspense
        fallback={
          <Card className="w-full max-w-md">
            <CardHeader className="text-center">
              <CardTitle className="text-2xl">OIDC Authentication</CardTitle>
              <CardDescription>Loading...</CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col items-center space-y-4">
              <Loader2 className="h-12 w-12 animate-spin text-primary" />
            </CardContent>
          </Card>
        }
      >
        <OIDCCallbackContent />
      </Suspense>
    </div>
  );
}
