# Stage 1: Build the React TypeScript application
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package.json, package-lock.json and tsconfig files first
COPY package.json package-lock.json tsconfig*.json vite.config.ts index.html ./

# Install dependencies
RUN npm install --frozen-lockfile

# Copy the source code
COPY public ./public
COPY src ./src
COPY .env ./.env

# Build the React app for production
RUN npm run build

# Stage 2: Serve the application using 'serve' (a simple Node.js static server)
FROM node:20-alpine

WORKDIR /app

# Install 'serve' globally
RUN npm install -g serve

# Copy the built React app from the builder stage
COPY --from=builder /app/dist ./dist

# Exposed port for the 'serve' application
EXPOSE 3000

# Command to serve the built application
# 'serve -s dist' serves static files from the 'dist' directory
# '-l 3000' listens on port 3000
CMD ["serve", "-s", "dist", "-l", "3000"]