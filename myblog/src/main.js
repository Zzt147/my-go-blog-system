import { createApp } from 'vue'
import { createPinia } from 'pinia'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate' // 引入持久化插件

import App from './App.vue'
import router from './router'

/* import the fontawesome core */
import { library } from '@fortawesome/fontawesome-svg-core'
/* import font awesome icon component */
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

// 引入图标
import { faWeibo, faGithub } from '@fortawesome/free-brands-svg-icons'
library.add(faWeibo)
library.add(faGithub)

const app = createApp(App)

// 1. 创建 Pinia 实例
const pinia = createPinia()
// 2. 将插件添加到这个实例中
pinia.use(piniaPluginPersistedstate)

app.component('font-awesome-icon', FontAwesomeIcon)

// 3. 使用配置好的 pinia
app.use(pinia)
app.use(router)

// --- 合并 ElementPlus 的引入 ---
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

app.use(ElementPlus, {
  locale: zhCn,
})

for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

// -------------------------------------------------------------------------
// [FIX] 核心修复：引入 tool.js 中配置好拦截器的 request 实例
// -------------------------------------------------------------------------
// ❌ 删除或注释掉原来直接引入 axios 的代码
// import axios from 'axios' 

// ✅ 引入你封装的 tool.js (注意路径)
import request from './js/tool' 
import VueAxios from 'vue-axios'

// 使用配置好拦截器的 request 实例来注册 VueAxios
app.use(VueAxios, request)

// 覆盖注入 'axios'，这样组件里 inject('axios') 拿到的就是带拦截器的版本了
app.provide('axios', request)  
// -------------------------------------------------------------------------

import vueCropper from 'vue-cropper'
import 'vue-cropper/dist/index.css'
app.use(vueCropper)

app.mount('#app')