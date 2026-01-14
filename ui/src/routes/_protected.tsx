import { createFileRoute, Link, Outlet, redirect, useNavigate } from '@tanstack/react-router';
import { Button } from '@/components/ui/button';
import { getCurrentUser, isAuthenticated, useLogout } from '@/hooks/use-auth';

export const Route = createFileRoute('/_protected')({
  beforeLoad: async () => {
    if (!isAuthenticated()) {
      throw redirect({ to: '/login' });
    }
  },
  component: ProtectedLayout,
});

function ProtectedLayout() {
  const logout = useLogout();
  const user = getCurrentUser();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout.mutateAsync();
    navigate({ to: '/login' });
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="border-b bg-white">
        <div className="container mx-auto flex h-16 items-center justify-between px-4">
          <div className="flex items-center gap-8">
            <h1 className="text-xl font-bold">Video Processing Platform</h1>
            <nav className="flex gap-4">
              <Link to="/videos" className="text-sm font-medium hover:underline">
                Videos
              </Link>
              <Link to="/upload" className="text-sm font-medium hover:underline">
                Upload
              </Link>
            </nav>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-sm text-gray-600">{user?.email}</span>
            <Button variant="outline" size="sm" onClick={handleLogout} disabled={logout.isPending}>
              {logout.isPending ? 'Logging out...' : 'Logout'}
            </Button>
          </div>
        </div>
      </header>
      <main className="container mx-auto px-4 py-8">
        <Outlet />
      </main>
    </div>
  );
}
