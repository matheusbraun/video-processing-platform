import { createFileRoute, redirect } from '@tanstack/react-router';
import { isAuthenticated } from '@/hooks/use-auth';

export const Route = createFileRoute('/')({
  beforeLoad: () => {
    if (isAuthenticated()) {
      throw redirect({ to: '/videos' });
    } else {
      throw redirect({ to: '/login' });
    }
  },
});
