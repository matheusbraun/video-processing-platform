import { createFileRoute, Outlet, redirect } from '@tanstack/react-router';
import { isAuthenticated } from '@/hooks/use-auth';

export const Route = createFileRoute('/_auth')({
  beforeLoad: async () => {
    if (isAuthenticated()) {
      throw redirect({ to: '/videos' });
    }
  },
  component: AuthLayout,
});

function AuthLayout() {
  return (
    <div className="min-h-screen bg-gray-50">
      <Outlet />
    </div>
  );
}
