import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_protected/videos')({
  component: VideosPage,
});

function VideosPage() {
  return (
    <div>
      <h1 className="mb-4 text-2xl font-bold">My Videos</h1>
      <p className="text-gray-600">Video list page - to be implemented</p>
    </div>
  );
}
