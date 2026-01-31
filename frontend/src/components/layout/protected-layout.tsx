'use client';

import { useEffect } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { useAuth } from '@/contexts/AuthContext';
import { Sidebar } from './sidebar';
import { Header } from './header';
import { Toaster } from '@/components/ui/sonner';

interface ProtectedLayoutProps {
  children: React.ReactNode;
  requireAdmin?: boolean;
}

export function ProtectedLayout({ children, requireAdmin = false }: ProtectedLayoutProps) {
  const { isAuthenticated, isLoading, checkAdmin } = useAuth();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (!isLoading) {
      if (!isAuthenticated) {
        router.push('/login');
      } else if (requireAdmin && !checkAdmin()) {
        router.push('/dashboard');
      }
    }
  }, [isAuthenticated, isLoading, requireAdmin, checkAdmin, router, pathname]);

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent"></div>
      </div>
    );
  }

  if (!isAuthenticated || (requireAdmin && !checkAdmin())) {
    return null;
  }

  return (
    <div>
      <Sidebar />

      <div className="lg:pl-72">
        <Header />

        <main className="py-10">
          <div className="px-4 sm:px-6 lg:px-8">{children}</div>
        </main>
      </div>

      <Toaster />
    </div>
  );
}
