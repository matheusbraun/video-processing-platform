import { createFileRoute, Outlet } from '@tanstack/react-router';

export const Route = createFileRoute('/_protected/videos')({
  component: () => <Outlet />,
});
