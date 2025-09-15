import { createApp } from "vue";
import App from "./App.vue";
import "./main.css";
import "remixicon/fonts/remixicon.css";
import router from "./router";
import store from "./store";

createApp(App).use(router).use(store).mount("#app");
