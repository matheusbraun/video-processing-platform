import { Link, useNavigate } from '@tanstack/react-router';
import { StatusBadge } from '@/components/status-badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useVideoDownload, useVideoStatus } from '@/hooks/use-videos';

interface VideoDetailPageProps {
  videoId: string;
}

type VideoStatus = 'PENDING' | 'PROCESSING' | 'COMPLETED' | 'FAILED';

function parseVideoStatus(status: string): VideoStatus {
  const validStatuses: VideoStatus[] = ['PENDING', 'PROCESSING', 'COMPLETED', 'FAILED'];
  return validStatuses.includes(status as VideoStatus) ? (status as VideoStatus) : 'PENDING';
}

export function VideoDetailPage({ videoId }: VideoDetailPageProps) {
  const navigate = useNavigate();
  const { data, isLoading, isError, error } = useVideoStatus(videoId);
  const downloadQuery = useVideoDownload(videoId);

  const handleDownload = async () => {
    try {
      const result = await downloadQuery.refetch();
      if (result.data?.download_url) {
        window.open(result.data.download_url, '_blank');
      }
    } catch (err) {
      console.error('Download failed:', err);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-center">
          <div className="mx-auto h-8 w-8 animate-spin rounded-full border-4 border-gray-300 border-t-blue-600" />
          <p className="mt-4 text-gray-600">Loading video details...</p>
        </div>
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div>
        <div className="mb-6">
          <Button variant="outline" onClick={() => navigate({ to: '/videos' })}>
            ← Back to Videos
          </Button>
        </div>
        <Alert variant="destructive">
          <AlertDescription>
            {error instanceof Error ? error.message : 'Failed to load video details'}
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  const formatDate = (date: string) => {
    return new Date(date).toLocaleString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const canDownload = data.status === 'COMPLETED';
  const isProcessing = data.status === 'PENDING' || data.status === 'PROCESSING';

  return (
    <div className="mx-auto max-w-3xl">
      <div className="mb-6">
        <Link to="/videos">
          <Button variant="outline">← Back to Videos</Button>
        </Link>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-start justify-between">
            <CardTitle className="text-2xl">{data.filename}</CardTitle>
            <StatusBadge status={parseVideoStatus(data.status)} />
          </div>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <p className="text-sm font-medium text-gray-500">Video ID</p>
              <p className="mt-1 font-mono text-sm">{data.video_id}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-500">Status</p>
              <p className="mt-1">{data.status}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-500">Uploaded</p>
              <p className="mt-1">{formatDate(data.created_at)}</p>
            </div>
            {data.completed_at && (
              <div>
                <p className="text-sm font-medium text-gray-500">Completed</p>
                <p className="mt-1">{formatDate(data.completed_at)}</p>
              </div>
            )}
            {data.frame_count && (
              <div>
                <p className="text-sm font-medium text-gray-500">Total Frames</p>
                <p className="mt-1">{data.frame_count} frames</p>
              </div>
            )}
          </div>

          {data.error_message && (
            <Alert variant="destructive">
              <AlertDescription>{data.error_message}</AlertDescription>
            </Alert>
          )}

          {isProcessing && (
            <Alert>
              <AlertDescription>
                Your video is being processed. This page will automatically update when processing
                is complete.
              </AlertDescription>
            </Alert>
          )}

          <div className="flex gap-4">
            <Button
              onClick={handleDownload}
              disabled={!canDownload || downloadQuery.isFetching}
              className="w-full"
            >
              {downloadQuery.isFetching ? 'Generating Download...' : 'Download Frames (ZIP)'}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
