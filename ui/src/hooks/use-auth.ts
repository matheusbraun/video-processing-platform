import { useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { ApiResponse, AuthResponse, LoginRequest, RegisterRequest } from '@/lib/api.types';

export function useLogin() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (credentials: LoginRequest) => {
      const response = await api
        .post('api/v1/auth/login', { json: credentials })
        .json<ApiResponse<AuthResponse>>();
      return response.data;
    },
    onSuccess: (data) => {
      localStorage.setItem('access_token', data?.access_token ?? '');
      localStorage.setItem('refresh_token', data?.refresh_token ?? '');
      localStorage.setItem('user', JSON.stringify(data?.user));
      queryClient.invalidateQueries();
    },
  });
}

export function useRegister() {
  return useMutation({
    mutationFn: async (credentials: RegisterRequest) => {
      const response = await api
        .post('api/v1/auth/register', { json: credentials })
        .json<ApiResponse<{ user_id: number; username: string; email: string }>>();
      return response.data;
    },
  });
}

export function useLogout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      const refreshToken = localStorage.getItem('refresh_token');
      if (refreshToken) {
        await api.post('api/v1/auth/logout', {
          json: { refresh_token: refreshToken },
        });
      }
    },
    onSuccess: () => {
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
      localStorage.removeItem('user');
      queryClient.clear();
    },
  });
}

export function isAuthenticated(): boolean {
  return !!localStorage.getItem('access_token');
}

export function getCurrentUser() {
  const userStr = localStorage.getItem('user');
  return userStr ? JSON.parse(userStr) : null;
}
