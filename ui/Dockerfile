# Stage 1: Build the application
FROM oven/bun:1 AS builder

WORKDIR /app

# Copy package files
COPY ui/package.json ui/bun.lock ./

# Install dependencies
RUN bun install --frozen-lockfile

# Copy source code
COPY ui/ .

# Build the application
RUN bun run build

# Stage 2: Serve with Nginx
FROM nginx:alpine

# Copy custom nginx configuration
COPY ui/nginx.conf /etc/nginx/conf.d/default.conf

# Copy built assets from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Expose port 80
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
