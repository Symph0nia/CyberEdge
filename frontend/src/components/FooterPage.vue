<template>
  <footer
    class="bg-gray-800/90 backdrop-blur-sm text-center py-6 border-t border-gray-700/30 shadow-inner"
  >
    <div class="container mx-auto px-4">
      <div class="flex flex-col md:flex-row items-center justify-between">
        <!-- 版权信息 -->
        <p
          class="tracking-wide text-gray-200 text-sm font-light mb-2 md:mb-0 flex items-center"
        >
          <span class="flex items-center">
            <i class="ri-shield-keyhole-line mr-2 text-gray-400"></i>
            &copy; 2025 CyberEdge 综合扫描器
          </span>
        </p>

        <!-- 分隔线 - 仅在移动视图显示 -->
        <div class="w-16 h-px bg-gray-700/50 my-2 md:hidden"></div>

        <!-- 标语 -->
        <p
          class="tracking-wide text-gray-400 text-xs font-light flex items-center"
        >
          <span>致力于保护您的网络安全与隐私</span>
          <i
            class="ri-shield-check-line ml-2 transform hover:scale-110 transition-all duration-200 hover:text-cyan-400"
          ></i>
        </p>
      </div>

      <!-- 底部链接 -->
      <div
        class="mt-4 flex items-center justify-center space-x-4 text-xs text-gray-500"
      >
        <a href="#" class="hover:text-gray-300 transition-colors duration-200"
          >隐私政策</a
        >
        <span class="text-gray-600">|</span>
        <a href="#" class="hover:text-gray-300 transition-colors duration-200"
          >使用条款</a
        >
        <span class="text-gray-600">|</span>
        <a href="#" class="hover:text-gray-300 transition-colors duration-200"
          >关于我们</a
        >
      </div>
    </div>
  </footer>
</template>

<script>
export default {
  name: "FooterPage",
};
</script>

<style scoped>
footer {
  min-height: 80px;
}

/* 链接悬停效果 */
a {
  position: relative;
}

a::after {
  content: "";
  position: absolute;
  width: 0;
  height: 1px;
  bottom: -2px;
  left: 0;
  background-color: rgba(156, 163, 175, 0.5);
  transition: width 0.3s ease;
}

a:hover::after {
  width: 100%;
}
</style>
