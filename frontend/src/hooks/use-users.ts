import { useState, useEffect } from 'react';
import { usersApi } from '@/lib/api';
import { User } from '@/types';

export function useUsers() {
  const [users, setUsers] = useState<User[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchUsers = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const response = await usersApi.listUsers();
      setUsers(response.data.users);
    } catch (err) {
      setError('Failed to fetch users');
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  };

  const createUser = async (data: {
    email: string;
    password?: string;
    role: 'admin' | 'user';
    oidc_enabled: boolean;
  }) => {
    const response = await usersApi.createUser(data);
    setUsers([...users, response.data]);
    return response.data;
  };

  const updateUser = async (
    id: number,
    data: {
      email?: string;
      role?: 'admin' | 'user';
      is_active?: boolean;
      oidc_enabled?: boolean;
    }
  ) => {
    const response = await usersApi.updateUser(id, data);
    setUsers(users.map((u) => (u.id === id ? response.data : u)));
    return response.data;
  };

  const deleteUser = async (id: number) => {
    await usersApi.deleteUser(id);
    setUsers(users.filter((u) => u.id !== id));
  };

  useEffect(() => {
    fetchUsers();
  }, []);

  return {
    users,
    isLoading,
    error,
    refetch: fetchUsers,
    createUser,
    updateUser,
    deleteUser,
  };
}
