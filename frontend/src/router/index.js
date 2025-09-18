import { createRouter, createWebHistory } from "vue-router";
import store from "../store";
import LoginPage from "@/components/Login/LoginPage.vue";
import UserManagement from "@/components/User/UserManagement.vue";
import ProfilePage from "@/components/Profile/ProfilePage.vue";
import SettingsPage from "@/components/Settings/SettingsPage.vue";
import ProjectList from "@/components/Project/ProjectList.vue";
import ProjectDetail from "@/components/Project/ProjectDetail.vue";
import VulnerabilityList from "@/components/Vulnerability/VulnerabilityList.vue";

const routes = [
  {
    path: "/",
    name: "Home",
    redirect: "/projects"
  },
  {
    path: "/login",
    name: "LoginPage",
    component: LoginPage,
  },
  {
    path: "/projects",
    name: "ProjectList",
    component: ProjectList,
  },
  {
    path: "/projects/:id",
    name: "ProjectDetail",
    component: ProjectDetail,
  },
  {
    path: "/vulnerabilities",
    name: "VulnerabilityList",
    component: VulnerabilityList,
  },
  {
    path: "/vulnerabilities/:projectId",
    name: "ProjectVulnerabilities",
    component: VulnerabilityList,
  },
  {
    path: "/user-management",
    name: "UserManagement",
    component: UserManagement,
  },
  {
    path: "/profile",
    name: "ProfilePage",
    component: ProfilePage,
  },
  {
    path: "/settings",
    name: "SettingsPage",
    component: SettingsPage,
  },
];

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes,
});

router.beforeEach(async (to, from, next) => {
  await store.dispatch("checkAuth");

  const isAuthenticated = store.state.isAuthenticated;
  if (
    !isAuthenticated &&
    to.name !== "LoginPage"
  ) {
    next({ name: "LoginPage" });
  } else {
    next();
  }
});

export default router;
