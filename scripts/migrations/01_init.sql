-- Create schemas
CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS videos;
CREATE SCHEMA IF NOT EXISTS notifications;

-- Auth schema tables
CREATE TABLE IF NOT EXISTS auth.users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS auth.refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES auth.users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Videos schema tables
CREATE TABLE IF NOT EXISTS videos.videos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    original_path TEXT NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED')),
    fps INTEGER DEFAULT 1,
    frame_count INTEGER,
    zip_path TEXT,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    expires_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP + INTERVAL '15 days')
);

CREATE INDEX IF NOT EXISTS idx_videos_user_status ON videos.videos(user_id, status);
CREATE INDEX IF NOT EXISTS idx_videos_created_at ON videos.videos(created_at);
CREATE INDEX IF NOT EXISTS idx_videos_expires_at ON videos.videos(expires_at) WHERE status = 'COMPLETED';

-- Notifications schema tables
CREATE TABLE IF NOT EXISTS notifications.notification_log (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    video_id UUID REFERENCES videos.videos(id) ON DELETE SET NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('EMAIL', 'WEBHOOK')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('PENDING', 'SENT', 'FAILED')),
    recipient TEXT NOT NULL,
    subject TEXT,
    error_message TEXT,
    sent_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications.notification_log(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_video_id ON notifications.notification_log(video_id);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications.notification_log(status);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION auth.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for users table
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON auth.users
    FOR EACH ROW
    EXECUTE FUNCTION auth.update_updated_at_column();

-- Grant permissions
GRANT USAGE ON SCHEMA auth TO videoadmin;
GRANT USAGE ON SCHEMA videos TO videoadmin;
GRANT USAGE ON SCHEMA notifications TO videoadmin;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA auth TO videoadmin;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA videos TO videoadmin;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA notifications TO videoadmin;

GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA auth TO videoadmin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA videos TO videoadmin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA notifications TO videoadmin;
