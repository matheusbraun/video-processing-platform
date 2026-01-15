import { Link } from '@tanstack/react-router';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { VideoCard } from '@/components/video-card';
import { useVideos } from '@/hooks/use-videos';

export function VideosPage() {
  const { data, isLoading, isError, error } = useVideos();

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-center">
          <div className="mx-auto h-8 w-8 animate-spin rounded-full border-4 border-gray-300 border-t-blue-600" />
          <p className="mt-4 text-gray-600">Loading videos...</p>
        </div>
      </div>
    );
  }

  if (isError) {
    return (
      <div>
        <h1 className="mb-6 text-3xl font-bold">My Videos</h1>
        <Alert variant="destructive">
          <AlertDescription>
            {error instanceof Error ? error.message : 'Failed to load videos'}
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-3xl font-bold">My Videos</h1>
        <Link to="/upload">
          <Button>Upload New Video</Button>
        </Link>
      </div>

      {data && data.videos.length === 0 ? (
        <Alert>
          <AlertDescription>
            You haven't uploaded any videos yet.{' '}
            <Link to="/upload" className="font-medium text-blue-600 hover:underline">
              Upload your first video
            </Link>
          </AlertDescription>
        </Alert>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {data?.videos.map((video) => (
            <VideoCard key={video.id} video={video} />
          ))}
        </div>
      )}

      {data && data.total > data.videos.length && (
        <div className="mt-6 text-center">
          <p className="text-sm text-gray-500">
            Showing {data.videos.length} of {data.total} videos
          </p>
        </div>
      )}
    </div>
  );
}
