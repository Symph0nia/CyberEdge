import { createApp } from "vue";
import App from "./App.vue";
import "./main.css";
import "remixicon/fonts/remixicon.css";
import router from "./router";
import store from "./store";

// Ant Design Vue
import Antd from 'ant-design-vue';
import 'ant-design-vue/dist/reset.css';

const app = createApp(App);

app.use(router)
   .use(store)
   .use(Antd)
   .mount("#app");
