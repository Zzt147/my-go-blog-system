<script setup>
import { reactive, ref, inject } from 'vue'
import { ElMessage, ElNotification } from 'element-plus'
import { useRouter } from 'vue-router'
import { useStore } from '@/stores/my.js'
import qs from 'qs'
import { User, Lock, Message, Key, Picture, Back } from '@element-plus/icons-vue'
import backImg from '@/assets/back.jpg' // 导入背景图

const router = useRouter()
const store = useStore()
const axios = inject('axios')

// === 状态控制 ===
const isResetMode = ref(false) // false=登录模式, true=重置密码模式
const isLoading = ref(false)
const formSize = ref('large')

// === 登录数据 ===
const loginForm = reactive({
  username: '',
  password: ''
})

// === 重置密码数据 ===
const resetForm = reactive({
  username: '',
  email: '',
  code: '',
  newPassword: '',
  captcha: '',
  captchaKey: ''
})
const resetCaptchaUrl = ref('')
const countdown = ref(0)
let timer = null
const isSending = ref(false)

// === 方法 ===

// 切换模式
const toggleMode = () => {
  isResetMode.value = !isResetMode.value
  if (isResetMode.value) {
    refreshResetCaptcha()
    // 清空重置表单
    resetForm.username = ''
    resetForm.email = ''
    resetForm.code = ''
    resetForm.newPassword = ''
    resetForm.captcha = ''
  }
}

// 刷新验证码
// [MODIFY] 修改验证码刷新逻辑，适配 Go 后端的 JSON 响应
// [MODIFY] 替换原有的 refreshResetCaptcha 函数
const refreshResetCaptcha = () => {
  const key = new Date().getTime().toString()
  resetForm.captchaKey = key
  
  // [核心修改] 使用 axios 获取 JSON 格式的验证码
  axios.get(`/api/user/captcha?key=${key}`).then(res => {
    if (res.data && res.data.img) {
      // Go 后端返回的是 { img: "base64...", key: "..." }
      resetCaptchaUrl.value = res.data.img
    }
  }).catch(err => {
    console.error(err)
    ElMessage.error("验证码加载失败")
  })
}

// 发送邮件验证码
// [MODIFY] 2. 发送邮件验证码 (严格保留你的函数名 sendResetCode)
const sendResetCode = async () => {
  if (!resetForm.username) return ElMessage.warning('请输入用户名')
  if (!resetForm.email) return ElMessage.warning('请输入邮箱')
  if (!resetForm.captcha) return ElMessage.warning('请输入图形验证码')

  try {
    // 核心修改：Go 后端 SendEmailCode 接收的是 PostForm (x-www-form-urlencoded)
    // 所以必须用 URLSearchParams 包装，不能传 JSON 对象
    const params = new URLSearchParams()
    params.append('email', resetForm.email)
    params.append('captcha', resetForm.captcha)
    params.append('captchaKey', resetForm.captchaKey)
    params.append('type', 'reset') 
    params.append('username', resetForm.username)

    const res = await axios.post('/api/user/sendEmailCode', params)
    
    if (res.data.success) {
      ElMessage.success(res.data.map.msg || '验证码已发送')
      // 倒计时逻辑
      isSending.value = true
      countdown.value = 60
      timer = setInterval(() => {
        countdown.value--
        if (countdown.value <= 0) {
          clearInterval(timer)
          isSending.value = false
        }
      }, 1000)
    } else {
      ElMessage.error(res.data.msg || '发送失败')
      refreshResetCaptcha() // 失败自动刷新验证码
    }
  } catch (error) {
    console.error(error)
    ElMessage.error('发送请求失败')
  }
}

// 提交重置密码
// [MODIFY] 3. 提交重置密码 (严格保留你的函数名 handleResetPassword)
const handleResetPassword = async () => {
  // 1. 前端校验
  if (!resetForm.username || !resetForm.email || !resetForm.code || !resetForm.newPassword) {
    return ElMessage.warning('请补全信息')
  }

  try {
    // 2. 发送请求 (Go 后端 ResetPassword 接口接收的是 JSON)
    const res = await axios.post('/api/user/resetPassword', {
      username: resetForm.username,
      email: resetForm.email,
      code: resetForm.code,
      password: resetForm.newPassword // 注意：后端 DTO 里的字段叫 password，前端这里对应 newPassword
    })

    // 3. 处理结果
    if (res.data.success) {
      ElMessage.success('密码重置成功，请登录')
      toggleMode() // 切回登录模式
      // 贴心优化：自动填入用户名
      loginForm.username = resetForm.username
    } else {
      ElMessage.error(res.data.msg || '重置失败')
    }
  } catch (err) {
    console.error(err)
    ElMessage.error('重置请求失败')
  }
}

// 登录
const handleLogin = async () => {
  if (!loginForm.username || !loginForm.password) return ElMessage.warning('请输入账号密码')

  isLoading.value = true
  try {
    const res = await axios({
      method: 'post',
      url: '/api/login',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      data: qs.stringify(loginForm)
    })

    if (res.data.success) {
      store.user.user = res.data.map.user
      ElNotification.success(`欢迎回来，${res.data.map.user.username}`)

      const role = res.data.map.user.authorities[0]
      router.push(role === 'ROLE_admin' ? '/admin_Main' : '/')
    } else {
      ElMessage.error(res.data.msg)
    }
  } catch (err) {
    ElMessage.error("系统错误")
  } finally {
    isLoading.value = false
  }
}

const goToRegister = () => router.push('/register')
</script>

<template>
  <div class="flat-container" :style="{ backgroundImage: `url(${backImg})` }">
    <div class="flat-card">
      <transition name="fade-slide" mode="out-in">

        <div v-if="!isResetMode" key="login" class="form-content">
          <h2 class="card-title">欢迎登录博客</h2>

          <el-form :model="loginForm" :size="formSize" @keyup.enter="handleLogin">
            <el-form-item>
              <el-input v-model="loginForm.username" placeholder="用户名" :prefix-icon="User" />
            </el-form-item>
            <el-form-item>
              <el-input v-model="loginForm.password" type="password" placeholder="密码" :prefix-icon="Lock"
                show-password />
            </el-form-item>

            <div class="links-row">
              <el-button type="primary" link @click="goToRegister">注册账号</el-button>
              <el-button type="warning" link @click="toggleMode">忘记密码?</el-button>
            </div>

            <el-button type="primary" class="action-btn" :loading="isLoading" @click="handleLogin">
              立即登录
            </el-button>
          </el-form>
        </div>

        <div v-else key="reset" class="form-content">
          <div class="header-row">
            <el-button link :icon="Back" @click="toggleMode" class="back-btn">返回</el-button>
            <h2 class="card-title" style="margin:0; flex:1">重置密码</h2>
          </div>

          <el-form :model="resetForm" :size="formSize" class="reset-form">
            <el-form-item>
              <el-input v-model="resetForm.username" placeholder="用户名" :prefix-icon="User" />
            </el-form-item>
            <el-form-item>
              <el-input v-model="resetForm.email" placeholder="注册邮箱" :prefix-icon="Message" />
            </el-form-item>

            <div class="code-group">
              <el-input v-model="resetForm.captcha" placeholder="图形码" :prefix-icon="Picture" style="width: 60%" />
              <img :src="resetCaptchaUrl" @click="refreshResetCaptcha" class="captcha-img" title="点击刷新" />
            </div>

            <div class="code-group">
              <el-input v-model="resetForm.code" placeholder="邮件码" :prefix-icon="Key" style="width: 60%" />
              <el-button class="code-btn" :disabled="countdown > 0 || isSending" @click="sendResetCode">
                {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
              </el-button>
            </div>

            <el-form-item>
              <el-input v-model="resetForm.newPassword" type="password" placeholder="新密码" :prefix-icon="Lock"
                show-password />
            </el-form-item>

            <el-button type="warning" class="action-btn" @click="handleResetPassword">
              确认修改
            </el-button>
          </el-form>
        </div>

      </transition>
    </div>

    <div class="footer-text">
      <p>2022 © Powered By <span style="color: #0e90d2">CrazyStone</span></p>
    </div>
  </div>
</template>

<style scoped>
/* 扁平化风格样式 */
.flat-container {
  width: 100vw;
  height: 100vh;
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
  display: flex;
  flex-direction: column;
  justify-content: center;
  /* 垂直居中 */
  align-items: center;
  /* 水平居中 */
  position: relative;
}

/* 卡片样式 */
.flat-card {
  width: 400px;
  min-height: 420px;
  background: rgba(255, 255, 255, 0.95);
  /* 轻微透明的白底 */
  backdrop-filter: blur(5px);
  border-radius: 8px;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.15);
  /* 柔和阴影 */
  padding: 40px 35px;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  justify-content: center;
  margin-bottom: 20px;
}

.card-title {
  text-align: center;
  color: #333;
  margin-bottom: 30px;
  font-weight: 500;
  font-size: 1.8rem;
  letter-spacing: 1px;
}

/* 表单控件微调 */
.action-btn {
  width: 100%;
  margin-top: 25px;
  height: 44px;
  font-size: 16px;
  letter-spacing: 2px;
}

.links-row {
  display: flex;
  justify-content: space-between;
  margin-top: 5px;
  padding: 0 5px;
}

/* 验证码组合 */
.code-group {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 18px;
}

.captcha-img {
  width: 38%;
  height: 40px;
  border-radius: 4px;
  cursor: pointer;
  border: 1px solid #dcdfe6;
}

.code-btn {
  width: 38%;
  height: 40px;
}

/* 顶部返回栏 */
.header-row {
  display: flex;
  align-items: center;
  margin-bottom: 25px;
  position: relative;
}

.back-btn {
  position: absolute;
  left: 0;
  font-size: 14px;
}

/* 底部文字 */
.footer-text {
  position: absolute;
  bottom: 40px;
  text-align: center;
  color: #fff;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.5);
  /* 增加阴影防止背景太亮看不清 */
  font-size: 14px;
}

/* 过渡动画 */
.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: all 0.3s ease;
}

.fade-slide-enter-from {
  opacity: 0;
  transform: translateX(20px);
}

.fade-slide-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}
</style>