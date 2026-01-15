import { createFileRoute } from '@tanstack/react-router';
import { UploadPage } from '@/pages/upload';

export const Route = createFileRoute('/_protected/upload')({
  component: UploadPage,
});
