import { Link } from "@tanstack/react-router";
import { StatusBadge } from "@/components/status-badge";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardFooter,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import type { Video } from "@/lib/api.types";

interface VideoCardProps {
	video: Video;
}

export function VideoCard({ video }: VideoCardProps) {
	const formatDate = (date: string) => {
		return new Date(date).toLocaleDateString("en-US", {
			year: "numeric",
			month: "short",
			day: "numeric",
			hour: "2-digit",
			minute: "2-digit",
		});
	};

	return (
		<Card>
			<CardHeader>
				<div className="flex items-start justify-between">
					<CardTitle className="text-lg">{video.filename}</CardTitle>
					<StatusBadge status={video.status} />
				</div>
			</CardHeader>
			<CardContent>
				<div className="space-y-2 text-sm text-gray-600">
					<div>
						<span className="font-medium">Uploaded:</span>{" "}
						{formatDate(video.created_at)}
					</div>
					{video.frame_count && (
						<div>
							<span className="font-medium">Frames:</span> {video.frame_count}
						</div>
					)}
					{video.error_message && (
						<div className="text-red-600">
							<span className="font-medium">Error:</span> {video.error_message}
						</div>
					)}
				</div>
			</CardContent>
			<CardFooter>
				<Link
					to="/videos/$videoId"
					params={{ videoId: video.id }}
					className="w-full"
				>
					<Button variant="outline" className="w-full">
						View Details
					</Button>
				</Link>
			</CardFooter>
		</Card>
	);
}
