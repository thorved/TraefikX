'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAuth } from '@/contexts/AuthContext';
import { authApi } from '@/lib/api';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Shield, KeyRound, Github } from 'lucide-react';
import { toast } from 'sonner';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [oidcEnabled, setOidcEnabled] = useState(false);
  const [oidcProvider, setOidcProvider] = useState('');
  const { login, isAuthenticated } = useAuth();
  const router = useRouter();

  // Check if already authenticated
  useEffect(() => {
    if (isAuthenticated) {
      router.push('/dashboard');
    }
  }, [isAuthenticated, router]);

  // Check OIDC status
  useEffect(() => {
    const checkOIDC = async () => {
      try {
        const response = await authApi.getOIDCStatus();
        setOidcEnabled(response.data.enabled);
        if (response.data.enabled) {
          setOidcProvider(response.data.provider_name || 'OIDC');
        }
      } catch (error) {
        console.error('Failed to check OIDC status:', error);
      }
    };
    checkOIDC();
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      await login(email, password);
      toast.success('Login successful!');
      router.push('/dashboard');
    } catch (error: any) {
      const message = error.response?.data?.error || 'Login failed';
      toast.error(message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleOIDCLogin = async () => {
    setIsLoading(true);
    try {
      const response = await authApi.initiateOIDC();
      // Redirect to OIDC provider
      window.location.href = response.data.auth_url;
    } catch (error: any) {
      const message = error.response?.data?.error || 'Failed to initiate OIDC login';
      toast.error(message);
      setIsLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-background via-background to-muted p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          <div className="flex items-center justify-center gap-2 mb-4">
            <div className="p-2 bg-primary rounded-lg">
              <Shield className="h-8 w-8 text-primary-foreground" />
            </div>
          </div>
          <CardTitle className="text-2xl text-center">Welcome back</CardTitle>
          <CardDescription className="text-center">
            Sign in to your TraefikX account
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                placeholder="Enter your email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                placeholder="Enter your password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </div>
            <Button
              type="submit"
              className="w-full"
              disabled={isLoading}
            >
              <KeyRound className="mr-2 h-4 w-4" />
              {isLoading ? 'Signing in...' : 'Sign in with Password'}
            </Button>
          </form>

          {oidcEnabled && (
            <>
              <div className="relative">
                <div className="absolute inset-0 flex items-center">
                  <span className="w-full border-t" />
                </div>
                <div className="relative flex justify-center text-xs uppercase">
                  <span className="bg-card px-2 text-muted-foreground">Or continue with</span>
                </div>
              </div>

              <Button
                variant="outline"
                className="w-full"
                onClick={handleOIDCLogin}
                disabled={isLoading}
              >
                <Github className="mr-2 h-4 w-4" />
                Sign in with {oidcProvider}
              </Button>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
