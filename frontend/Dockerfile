FROM node:25-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:25-alpine
WORKDIR /app
COPY --from=builder /app/build build
COPY --from=builder /app/package.json .
COPY --from=builder /app/node_modules node_modules

# Add a non-root user with home directory disabled (-D)
RUN adduser -D appuser

# Ensure the /app directory is owned by the non-root user
RUN chown -R appuser /app

EXPOSE 3000
ENV NODE_ENV=production

# Switch to the non-root user
USER appuser

CMD ["node", "build"]
