<template>
  <footer
    
  >
    <div >
      <div class="md:">
        <!-- 版权信息 -->
        <p
          
        >
          <span >
            <i class="ri-shield-keyhole-line"></i>
            &copy; 2025 CyberEdge 综合扫描器
          </span>
        </p>

        <!-- 分隔线 - 仅在移动视图显示 -->
        <div class="my-2"></div>

        <!-- 标语 -->
        <p
          
        >
          <span>致力于保护您的网络安全与隐私</span>
          <i
            class="ri-shield-check-line duration-200 hover:"
          ></i>
        </p>
      </div>

      <!-- 底部链接 -->
      <div
        
      >
        <a href="#" class="hover: duration-200"
          >隐私政策</a
        >
        <span >|</span>
        <a href="#" class="hover: duration-200"
          >使用条款</a
        >
        <span >|</span>
        <a href="#" class="hover: duration-200"
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
