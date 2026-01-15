import { createFileRoute } from '@tanstack/react-router';
import { VideoDetailPage } from '@/pages/video-detail';

export const Route = createFileRoute('/_protected/videos/$videoId')({
  component: VideoDetailComponent,
});

function VideoDetailComponent() {
  const { videoId } = Route.useParams();
  return <VideoDetailPage videoId={videoId} />;
}
