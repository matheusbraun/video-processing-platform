export interface ApiResponse<T> {
  data?: T;
  message?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: {
    user_id: number;
    username: string;
    email: string;
  };
}

export interface Video {
  id: string;
  filename: string;
  status: 'PENDING' | 'PROCESSING' | 'COMPLETED' | 'FAILED';
  frame_count?: number;
  created_at: string;
  completed_at?: string;
}

export interface VideoListResponse {
  videos: Video[];
  total: number;
  limit: number;
  offset: number;
  has_more: boolean;
}

export interface VideoStatusResponse {
  video_id: string;
  filename: string;
  status: string;
  frame_count?: number;
  error_message?: string;
  created_at: string;
  started_at?: string;
  completed_at?: string;
}

export interface DownloadResponse {
  download_url: string;
  filename: string;
  expires_in: number;
}

export interface UploadResponse {
  video_id: string;
  filename: string;
  status: string;
}
