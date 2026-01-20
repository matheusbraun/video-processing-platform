import { useCallback, useState } from 'react';
import { Card } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { useUploadVideo } from '@/hooks/use-videos';

interface UploadZoneProps {
  onUploadComplete?: (videoId: string) => void;
}

export function UploadZone({ onUploadComplete }: UploadZoneProps) {
  const [isDragging, setIsDragging] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const uploadVideo = useUploadVideo();

  const handleUpload = useCallback(
    async (file: File) => {
      try {
        setUploadProgress(0);
        const result = await uploadVideo.mutateAsync(file);
        setUploadProgress(100);
        onUploadComplete?.(result?.video_id ?? '');
      } catch (error) {
        console.error('Upload failed:', error);
        setUploadProgress(0);
      }
    },
    [uploadVideo, onUploadComplete],
  );

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
  }, []);

  const handleDrop = useCallback(
    async (e: React.DragEvent) => {
      e.preventDefault();
      setIsDragging(false);

      const files = Array.from(e.dataTransfer.files);
      const videoFile = files.find((file) => file.type.startsWith('video/'));

      if (videoFile) {
        await handleUpload(videoFile);
      }
    },
    [handleUpload],
  );

  const handleFileInput = useCallback(
    async (e: React.ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      if (file) {
        await handleUpload(file);
      }
    },
    [handleUpload],
  );

  return (
    <Card
      className={`relative border-2 border-dashed p-12 text-center transition-colors ${
        isDragging ? 'border-blue-500 bg-blue-50' : 'border-gray-300'
      }`}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
    >
      <input
        type="file"
        accept="video/*"
        onChange={handleFileInput}
        className="absolute inset-0 cursor-pointer opacity-0"
        disabled={uploadVideo.isPending}
      />
      <div className="space-y-4">
        <div className="mx-auto h-12 w-12 text-gray-400">
          <svg
            className="h-full w-full"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            role="img"
            aria-label="Upload icon"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
            />
          </svg>
        </div>
        <div>
          <p className="text-lg font-medium">
            {uploadVideo.isPending ? 'Uploading...' : 'Drop your video here'}
          </p>
          <p className="text-sm text-gray-500">or click to browse</p>
        </div>
        {uploadVideo.isPending && uploadProgress > 0 && (
          <div className="mx-auto w-full max-w-xs">
            <Progress value={uploadProgress} />
            <p className="mt-2 text-sm text-gray-600">{uploadProgress}%</p>
          </div>
        )}
        {uploadVideo.isError && (
          <p className="text-sm text-red-600">
            {uploadVideo.error instanceof Error ? uploadVideo.error.message : 'Upload failed'}
          </p>
        )}
      </div>
    </Card>
  );
}
