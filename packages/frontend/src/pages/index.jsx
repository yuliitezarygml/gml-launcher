import { For } from "solid-js";
import { FaBrandsWindows, FaBrandsLinux, FaBrandsApple } from "solid-icons/fa";
import { FiDownload, FiUser, FiShield } from "solid-icons/fi";

export default function Index() {
  const links = [
    {
      title: "Windows",
      icon: <FaBrandsWindows class="w-16 h-16" />,
      link: "#",
    },
    {
      title: "Linux",
      icon: <FaBrandsLinux class="w-16 h-16" />,
      link: "#",
    },
    {
      title: "MacOS",
      icon: <FaBrandsApple class="w-16 h-16" />,
      link: "#",
    },
  ];

  const features = [
    {
      icon: <FiUser class="w-8 h-8" />,
      title: "Персонализация",
      description: "Загружайте свои скины и плащи"
    },
    {
      icon: <FiShield class="w-8 h-8" />,
      title: "Безопасность",
      description: "Защищенная авторизация с JWT"
    },
    {
      icon: <FiDownload class="w-8 h-8" />,
      title: "Простота",
      description: "Скачайте лаунчер и начните играть"
    }
  ];

  return (
    <div class="min-h-screen flex flex-col">
      {/* Hero Section */}
      <div class="flex-1 flex flex-col items-center justify-center px-4 py-16">
        <div class="max-w-4xl w-full text-center">
          {/* Animated gradient title */}
          <h1 class="text-6xl md:text-7xl font-bold mb-6 bg-gradient-to-r from-blue-400 via-purple-500 to-pink-500 bg-clip-text text-transparent animate-gradient">
            Easy Cabinet
          </h1>
          
          <p class="text-xl md:text-2xl text-gray-300 mb-12 font-light">
            Ваш личный кабинет для управления профилем игрока
          </p>

          {/* Download Section */}
          <div class="mb-16">
            <h2 class="text-2xl md:text-3xl mb-8 text-gray-200 font-light">
              Скачать лаунчер
            </h2>
            <div class="flex flex-wrap gap-6 items-center justify-center">
              <For each={links}>{({ link, title, icon }) => (
                <a
                  href={link}
                  download
                  class="group relative bg-gradient-to-br from-neutral-800 to-neutral-900 hover:from-neutral-700 hover:to-neutral-800 rounded-2xl px-8 py-6 transition-all duration-300 hover:scale-105 hover:shadow-2xl hover:shadow-blue-500/20 border border-neutral-700 hover:border-blue-500/50"
                >
                  <div class="flex flex-col items-center gap-3">
                    <div class="text-blue-400 group-hover:text-blue-300 transition-colors">
                      {icon}
                    </div>
                    <span class="text-lg font-medium">{title}</span>
                  </div>
                  <div class="absolute inset-0 rounded-2xl bg-gradient-to-br from-blue-500/0 to-purple-500/0 group-hover:from-blue-500/10 group-hover:to-purple-500/10 transition-all duration-300" />
                </a>
              )}</For>
            </div>
          </div>

          {/* Features Section */}
          <div class="grid md:grid-cols-3 gap-8 mt-16">
            <For each={features}>{({ icon, title, description }) => (
              <div class="bg-neutral-800/50 backdrop-blur-sm rounded-xl p-6 border border-neutral-700/50 hover:border-blue-500/30 transition-all duration-300 hover:transform hover:-translate-y-1">
                <div class="text-blue-400 mb-4 flex justify-center">
                  {icon}
                </div>
                <h3 class="text-xl font-medium mb-2">{title}</h3>
                <p class="text-gray-400 text-sm">{description}</p>
              </div>
            )}</For>
          </div>

          {/* CTA Buttons */}
          <div class="flex flex-wrap gap-4 justify-center mt-16">
            <a
              href="/register"
              class="px-8 py-4 bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 rounded-xl font-medium transition-all duration-300 hover:scale-105 hover:shadow-lg hover:shadow-purple-500/50"
            >
              Создать аккаунт
            </a>
            <a
              href="/login"
              class="px-8 py-4 bg-neutral-800 hover:bg-neutral-700 rounded-xl font-medium border border-neutral-600 hover:border-blue-500/50 transition-all duration-300"
            >
              Войти
            </a>
          </div>
        </div>
      </div>

      {/* Footer */}
      <footer class="py-8 text-center text-gray-500 text-sm border-t border-neutral-800">
        <p>© 2024 Easy Cabinet. Личный кабинет для игроков Minecraft</p>
      </footer>
    </div>
  );
}
