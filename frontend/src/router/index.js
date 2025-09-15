import { createRouter, createWebHistory } from "vue-router";
import store from "../store";
import Home from "@/components/HomePage.vue";
import LoginPage from "@/components/Login/LoginPage.vue";
import SystemConfiguration from "@/components/Config/SystemConfiguration.vue";
import UserManagement from "@/components/User/UserManagement.vue";
import WAFDashboard from "@/components/Dashboard.vue";
import GoogleAuthQRCode from "@/components/Login/GoogleAuthQRCode.vue";
import TaskManagement from "@/components/Task/TaskManagement.vue";
import PortScanResults from "@/components/Port/PortScanResults.vue";
import SubdomainScanResults from "@/components/Subdomain/SubdomainScanResults.vue";
import PathScanResults from "@/components/Path/PathScanResults.vue";
import ScanResultDetail from "@/components/Results/ScanResultDetail.vue";
import TargetManagement from "@/components/Target/TargetManagement.vue";
// TargetDetail组件已删除，使用通用组件替代
import UnderDevelopment from "@/components/UnderDevelopment.vue";
import ToolConfiguration from "@/components/Config/ToolConfiguration.vue";

const routes = [
  {
    path: "/",
    name: "Home",
    component: Home,
  },
  {
    path: "/login",
    name: "LoginPage",
    component: LoginPage,
  },
  {
    path: "/system-configuration",
    name: "SystemConfiguration",
    component: SystemConfiguration,
  },
  {
    path: "/tool-configuration",
    name: "ToolConfiguration",
    component: ToolConfiguration,
  },
  {
    path: "/user-management",
    name: "UserManagement",
    component: UserManagement,
  },
  {
    path: "/setup-2fa",
    name: "Setup2FA",
    component: GoogleAuthQRCode,
  },
  {
    path: "/dashboard",
    name: "WAFDashboard",
    component: WAFDashboard,
  },
  {
    path: "/task-management", // 新增的任务管理路由
    name: "TaskManagement",
    component: TaskManagement, // 确保已导入 TaskManagement 组件
  },
  {
    path: "/port-scan-results",
    name: "PortScanResults",
    component: PortScanResults,
  },
  {
    path: "/port-scan-results/:id",
    name: "PortScanDetail",
    component: ScanResultDetail,
    props: true,
  },
  {
    path: "/subdomain-scan-results",
    name: "SubdomainScanResults",
    component: SubdomainScanResults,
  },
  {
    path: "/subdomain-scan-results/:id",
    name: "SubdomainScanDetail",
    component: ScanResultDetail,
    props: true,
  },
  {
    path: "/path-scan-results",
    name: "PathScanResults",
    component: PathScanResults, // 确保导入了 PathScanResults 组件
  },
  {
    path: "/path-scan-results/:id",
    name: "PathScanDetail",
    component: ScanResultDetail,
    props: true,
  },
  {
    path: "/task-management",
    name: "TaskManagement",
    component: TaskManagement,
  },
  {
    path: "/target-management",
    name: "TargetManagement",
    component: TargetManagement,
  },
  {
    path: "/target-management/:id",
    name: "TargetDetail",
    component: TargetDetail,
  },
  {
    path: "/under-development",
    name: "UnderDevelopment",
    component: UnderDevelopment,
  },
];

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes,
});

router.beforeEach(async (to, from, next) => {
  await store.dispatch("checkAuth"); // 确保最新的认证状态

  const isAuthenticated = store.state.isAuthenticated;
  if (
    !isAuthenticated &&
    to.name !== "LoginPage" &&
    to.name !== "Home" &&
    to.name !== "Setup2FA"
  ) {
    next({ name: "Home" });
  } else {
    next();
  }
});

export default router;
