import { NavLink, Outlet } from 'react-router-dom';
import { Home, UserPlus, Table, UtensilsCrossed } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { useAuth } from '@/contexts/AuthContext';

// A simple sidebar component
function Sidebar() {
  const { logout } = useAuth();
  const navLinkClass = ({ isActive }: { isActive: boolean }) =>
    `flex items-center gap-3 rounded-lg px-3 py-2 transition-all ${
      isActive
        ? 'bg-muted text-primary'
        : 'text-muted-foreground hover:text-primary'
    }`;

  return (
    <div className="hidden border-r bg-muted/40 md:block">
      <div className="flex h-full max-h-screen flex-col gap-2">
        <div className="flex h-14 items-center border-b px-4 lg:h-[60px] lg:px-6">
          <NavLink
            to="/dashboard"
            className="flex items-center gap-2 font-semibold"
          >
            <UtensilsCrossed className="h-6 w-6" />
            <span>Restaurant Admin</span>
          </NavLink>
        </div>
        <div className="flex-1">
          <nav className="grid items-start px-2 text-sm font-medium lg:px-4">
            <NavLink to="/dashboard" className={navLinkClass}>
              <Home className="h-4 w-4" />
              Dashboard
            </NavLink>
            <NavLink to="/tables" className={navLinkClass}>
              <Table className="h-4 w-4" />
              Table Management
            </NavLink>
            <NavLink to="/onboarding" className={navLinkClass}>
              <UserPlus className="h-4 w-4" />
              Customer Onboarding
            </NavLink>
          </nav>
        </div>
        <div className="mt-auto p-4">
          <Button size="sm" className="w-full" onClick={logout}>
            Logout
          </Button>
        </div>
      </div>
    </div>
  );
}

export function StaffLayout() {
  return (
    <div className="grid min-h-screen w-full md:grid-cols-[220px_1fr] lg:grid-cols-[280px_1fr]">
      <Sidebar />
      <div className="flex flex-col">
        <main className="flex flex-1 flex-col gap-4 p-4 lg:gap-6 lg:p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
