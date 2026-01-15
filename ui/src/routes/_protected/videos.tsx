import { createFileRoute } from '@tanstack/react-router';
import { VideosPage } from '@/pages/videos';

export const Route = createFileRoute('/_protected/videos')({
  component: VideosPage,
});
