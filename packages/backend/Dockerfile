FROM node:20-alpine

WORKDIR /app

# Установка pnpm
RUN npm install -g pnpm

# Копирование файлов
COPY package.json ./
COPY prisma ./prisma/

# Установка зависимостей
RUN pnpm install

# Копирование исходного кода
COPY . .

# Генерация Prisma клиента
RUN npx prisma generate

# Сборка приложения
RUN pnpm run build

# Создание директории для uploads
RUN mkdir -p /app/uploads

EXPOSE 4000

CMD ["sh", "-c", "npx prisma migrate deploy && node dist/main"]
