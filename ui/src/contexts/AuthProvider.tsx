import { useEffect, useState } from 'react';
import { jwtDecode } from 'jwt-decode';
import apiClient from '@/api/client';
import { AuthContext, type AuthUser } from '@/contexts/AuthContext';

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setToken] = useState<string | null>(
    localStorage.getItem('authToken')
  );
  const [user, setUser] = useState<AuthUser | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    if (token) {
      try {
        const decoded: { user_id: number; role: string; exp: number } =
          jwtDecode(token);
        // Check if token is expired
        if (decoded.exp * 1000 > Date.now()) {
          setUser({ userId: decoded.user_id, role: decoded.role });
          // Set the auth token for all future axios requests
          apiClient.defaults.headers.common[
            'Authorization'
          ] = `Bearer ${token}`;
        } else {
          // Token is expired
          logout();
        }
      } catch (error) {
        console.error('Invalid token:', error);
        logout();
      }
    }
    setIsLoading(false);
  }, [token]);

  const login = (newToken: string) => {
    localStorage.setItem('authToken', newToken);
    setToken(newToken);
  };

  const logout = () => {
    localStorage.removeItem('authToken');
    setToken(null);
    setUser(null);
    delete apiClient.defaults.headers.common['Authorization'];
  };

  const isAuthenticated = !!token && !!user;

  return (
    <AuthContext.Provider
      value={{ token, user, login, logout, isAuthenticated, isLoading }}
    >
      {children}
    </AuthContext.Provider>
  );
}
