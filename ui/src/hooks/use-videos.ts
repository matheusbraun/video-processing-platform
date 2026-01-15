import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type {
	ApiResponse,
	DownloadResponse,
	UploadResponse,
	VideoListResponse,
	VideoStatusResponse,
} from "@/lib/api.types";

export function useVideos(limit = 20, offset = 0) {
	return useQuery({
		queryKey: ["videos", limit, offset],
		queryFn: async () => {
			const response = await api
				.get("api/v1/videos", { searchParams: { limit, offset } })
				.json<ApiResponse<VideoListResponse>>();
			return response.data;
		},
	});
}

export function useVideoStatus(videoId: string) {
	return useQuery({
		queryKey: ["video", videoId],
		queryFn: async () => {
			const response = await api
				.get(`api/v1/videos/${videoId}/status`)
				.json<ApiResponse<VideoStatusResponse>>();
			return response.data;
		},
		refetchInterval: (query) => {
			const status = query.state.data?.status;
			return status === "PENDING" || status === "PROCESSING" ? 3000 : false;
		},
	});
}

export function useVideoDownload(videoId: string) {
	return useQuery({
		queryKey: ["video", videoId, "download"],
		queryFn: async () => {
			const response = await api
				.get(`api/v1/videos/${videoId}/download`)
				.json<ApiResponse<DownloadResponse>>();
			return response.data;
		},
		enabled: false,
	});
}

export function useUploadVideo() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: async (file: File) => {
			const formData = new FormData();
			formData.append("video", file);

			const response = await api
				.post("api/v1/videos/upload", { body: formData })
				.json<ApiResponse<UploadResponse>>();
			return response.data;
		},
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["videos"] });
		},
	});
}
