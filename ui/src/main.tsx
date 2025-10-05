import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import './index.css';
import { ThemeProvider } from '@/contexts/ThemeProvider.tsx';
import App from '@/App';
import {
  createBrowserRouter,
  Navigate,
  RouterProvider,
} from 'react-router-dom';
import { LoginPage } from '@/pages/LoginPage';
import { ProtectedRoute } from '@/components/ProtectedRoute';
import { DashboardPage } from '@/pages/DashboardPage';
import { AuthProvider } from '@/contexts/AuthProvider';
import { StaffLayout } from '@/pages/staff/StaffLayout';
import { OnboardingPage } from '@/pages/staff/OnboardingPage';

// Define routes
const router = createBrowserRouter([
  {
    path: '/',
    element: <App />,
    children: [
      {
        index: true, // Add this to redirect from "/"
        element: <Navigate to="/dashboard" replace />,
      },
      {
        path: 'login',
        element: <LoginPage />,
      },
      {
        element: <ProtectedRoute />,
        children: [
          {
            element: <StaffLayout />, // Wrap protected pages in the layout
            children: [
              {
                path: 'dashboard',
                element: <DashboardPage />,
              },
              {
                path: 'onboarding',
                element: <OnboardingPage />,
              },
              {
                path: 'tables',
                element: <div>Table Management Page Placeholder</div>,
              },
            ],
          },
        ],
      },
    ],
  },
]);

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ThemeProvider defaultTheme="dark" storageKey="restaurant-ui-theme">
      <AuthProvider>
        <RouterProvider router={router} />
      </AuthProvider>
    </ThemeProvider>
  </StrictMode>
);
