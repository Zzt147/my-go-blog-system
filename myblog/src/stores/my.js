import { defineStore } from 'pinia'
import { reactive, ref } from 'vue'

export const useStore = defineStore(
  'my',
  () => {
    // 文章ID状态
    const articleId = ref(0)

    // 分页参数状态
    const page = reactive({ pageParams: null })

    // 首页页码状态
    const home = reactive({ page: 1 })

    // 用户状态
    // 使用 reactive 包裹 user 属性，是为了保持和你项目中 store.user.user 的调用结构一致
    const user = reactive({ 
      user: null, 
      token: ''  // <--- 这里的空字符串是占位符
      })

    // --- Actions / Functions ---

    // 登录：保存用户信息
    function login(userData, tokenStr) {
      // 注意：reactive 对象直接修改属性，不需要 .value
      user.user = userData
      user.token = tokenStr // <--- 关键：把 token 存进响应式对象里
    }

    // 注销：清空用户信息
    function logout() {
      // 清空 state
      user.user = null
      // 由于开启了 persist，pinia 插件会自动把 localStorage 里的数据也同步清空
      user.token = '' // <--- 记得一起清空
    }

    return { articleId, page, home, user, login, logout }
  },
  {
    // [新增] 开启持久化配置
    // 需先安装并注册 pinia-plugin-persistedstate 插件
    // 开启后，刷新浏览器页面，user 等状态会自动从 localStorage 恢复
    persist: true,
  },
)
