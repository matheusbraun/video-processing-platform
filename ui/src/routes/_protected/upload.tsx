import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_protected/upload')({
  component: UploadPage,
});

function UploadPage() {
  return (
    <div>
      <h1 className="mb-4 text-2xl font-bold">Upload Video</h1>
      <p className="text-gray-600">Upload page - to be implemented</p>
    </div>
  );
}
