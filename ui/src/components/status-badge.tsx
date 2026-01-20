import { Badge } from '@/components/ui/badge';

type VideoStatus = 'PENDING' | 'PROCESSING' | 'COMPLETED' | 'FAILED';

interface StatusBadgeProps {
  status: VideoStatus;
}

export function StatusBadge({ status }: StatusBadgeProps) {
  const variants = {
    PENDING: { variant: 'secondary' as const, label: 'Pending', className: '' },
    PROCESSING: {
      variant: 'default' as const,
      label: 'Processing',
      className: '',
    },
    COMPLETED: {
      variant: 'default' as const,
      label: 'Completed',
      className: 'bg-green-600',
    },
    FAILED: { variant: 'destructive' as const, label: 'Failed', className: '' },
  };

  const config = variants[status];

  return (
    <Badge variant={config.variant} className={config.className}>
      {config.label}
    </Badge>
  );
}
