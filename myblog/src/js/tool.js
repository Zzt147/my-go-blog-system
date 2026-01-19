import axios from 'axios'

// --- 1. 定义 Axios 实例 ---
const request = axios.create({
  // 注意：这里假设你在 vite.config.js 里配置了 /api 代理转发到后端
  // 如果没配置代理，可以直接写 'http://localhost:8080'
  baseURL: '',
  timeout: 5000,
})

// 请求拦截器（可以在这里自动带上 token）
request.interceptors.request.use(
  (config) => {
    // ✅ [NEW] 只有当 Content-Type 既没被组件设置，且数据不是表单类型时，才给个默认 JSON (其实 Axios 默认也会给，这里为了稳妥可保留判断)
    if (!config.headers['Content-Type'] && !(config.data instanceof FormData || config.data instanceof URLSearchParams)) {
      config.headers['Content-Type'] = 'application/json;charset=utf-8'
    }
    // [FIX] 从 localStorage 获取用户信息 (假设你的 Store 持久化存储名为 'user' 或其他)
    // 注意：你需要确认一下你的 Pinia 或 localStorage 里到底存的 Key 是什么
    // 通常如果是 Pinia 持久化，可能是 'my-store' 或者 'user'

    // 假设你的 Login.vue 把数据存到了 localStorage 的 'user' 字段里
    // 1. 从 localStorage 获取 Pinia 的持久化数据
    // 你的 Store ID 是 'my'，所以 Key 是 'my'
    const userJson = localStorage.getItem('my')
    if (userJson) {
      try {
        // 这里解析出来的是整个 Store 的状态数据
        const storeData = JSON.parse(userJson)
        // 确保 user 对象里有 token
        // 根据之前的 Login 代码，后端返回结构是 { user: {...}, token: "..." }
        // 但前端 Store 可能存的是整个 user 对象，或者 store.user 结构

        // ⚠️ 关键修正：
        // 1. 获取 token (假设在 user.token 或 user.user.token，请根据你实际存储结构调整)
        //    如果不确定，可以在控制台 Application -> Local Storage 里看一眼
        // 2. 根据 Store 结构获取 Token
        // Store 结构是: user: { user: {...}, token: "..." }
        // 所以路径是 myStore.user.token
        // ⚠️ [FIX] 关键修正：修复之前引用不存在的 myStore 变量的问题
        // 正确的变量应该是上面的 storeData (即解析后的 userJson)
        // 路径：storeData (全量数据) -> .user (State中的user对象) -> .token
        const token = storeData.user?.token

        if (token) {
          // 2. 必须使用 'Authorization' 字段，并加上 'Bearer ' 前缀
          //这是国际通用的 JWT 标准格式，也是我们 Go 后端中间件里写死的解析方式
          config.headers['Authorization'] = 'Bearer ' + token
        }
      } catch (e) {
        console.error('Token 解析失败', e)
      }
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  },
)

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    // 可以在这里统一处理后端返回的错误码
    return response
  },
  (error) => {
    console.log('请求出错：' + error)
    return Promise.reject(error)
  },
)

// --- 2. 原有的工具函数 (保持不变) ---
function undefine(i) {
  if ('undefined' == typeof i) {
    return true
  } else {
    return false
  }
}

function nullZeroBlank(i) {
  if (i == null) return true
  else if (typeof i == 'string') {
    let str = i.replace(/\[\]/g, ' ').trim()
    if (str.length == 0) return true
  } else if (i == 0) return true

  return false
}

function notNullZeroBlank(i) {
  return !nullZeroBlank(i)
}

function dateFormat(dateString, format) {
  try {
    let date = new Date(dateString)
    // 补零函数
    const pad = (n) => (n < 10 ? '0' + n : n)

    if ('yyyy-MM-dd' == format) {
      let dateFormat = date.getFullYear() + '-'
      dateFormat += date.getMonth() + 1 + '-'
      dateFormat += date.getDate()
      return dateFormat
    }
    // 【新增】支持时分秒格式
    else if ('yyyy-MM-dd HH:mm:ss' == format) {
      let str = date.getFullYear() + '-'
      str += pad(date.getMonth() + 1) + '-'
      str += pad(date.getDate()) + ' '
      str += pad(date.getHours()) + ':'
      str += pad(date.getMinutes()) + ':'
      str += pad(date.getSeconds())
      return str
    } else {
      return '无此格式！'
    }
  } catch (e) {
    return '格式转换错误！'
  }
}

// --- 3. 导出 ---
// 【关键】默认导出 request，解决 "does not provide an export named default" 报错
export default request

// 同时也保留命名导出，不影响其他地方引用工具函数
export { undefine, nullZeroBlank, notNullZeroBlank, dateFormat }
