import {createApp} from 'vue'
import App from './App.vue'
import router from './router'
import './style.css';

// naive-ui 组件由 unplugin-vue-components 按需自动引入
// 无需全局注册，模板中直接使用 NButton 等组件即可

// 创建应用并注册路由
createApp(App).use(router).mount('#app')
