import { createRouter, createWebHistory } from "vue-router";
import store from "../store";
import LoginPage from "@/components/Login/LoginPage.vue";
import UserManagement from "@/components/User/UserManagement.vue";
import ProfilePage from "@/components/Profile/ProfilePage.vue";
import SettingsPage from "@/components/Settings/SettingsPage.vue";

const routes = [
  {
    path: "/",
    name: "Home",
    redirect: "/user-management"
  },
  {
    path: "/login",
    name: "LoginPage",
    component: LoginPage,
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
