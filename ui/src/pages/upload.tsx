import { useNavigate } from '@tanstack/react-router';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { UploadZone } from '@/components/upload-zone';

export function UploadPage() {
  const navigate = useNavigate();

  const handleUploadComplete = (videoId: string) => {
    navigate({ to: '/videos/$videoId', params: { videoId } });
  };

  return (
    <div className="mx-auto max-w-2xl">
      <h1 className="mb-6 text-3xl font-bold">Upload Video</h1>

      <Alert className="mb-6">
        <AlertTitle>Processing Information</AlertTitle>
        <AlertDescription>
          Your video will be processed to extract frames at 1 FPS. Once completed, you'll be able to
          download a ZIP file containing all extracted frames.
        </AlertDescription>
      </Alert>

      <UploadZone onUploadComplete={handleUploadComplete} />

      <div className="mt-6 text-sm text-gray-500">
        <p className="font-medium">Supported formats:</p>
        <ul className="ml-4 mt-2 list-disc space-y-1">
          <li>MP4, AVI, MOV, MKV, and other common video formats</li>
          <li>Maximum file size: 500MB</li>
        </ul>
      </div>
    </div>
  );
}
