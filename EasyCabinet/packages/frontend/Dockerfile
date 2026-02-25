FROM node:20-alpine AS builder

WORKDIR /app

# Установка pnpm
RUN npm install -g pnpm

# Копирование package.json
COPY package.json ./

# Установка зависимостей
RUN pnpm install

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN pnpm run build

# Production образ с nginx
FROM nginx:alpine

# Копирование собранного приложения
COPY --from=builder /app/dist /usr/share/nginx/html

# Копирование конфигурации nginx
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 3000

CMD ["nginx", "-g", "daemon off;"]
