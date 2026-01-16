<script setup>
import { reactive, ref, inject, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { User, Lock, Message, Key, Picture, Back } from '@element-plus/icons-vue'
import backImg from '@/assets/back.jpg'

const router = useRouter()
const axios = inject('axios')

const isLoading = ref(false)
const formSize = ref('large')

const registerForm = reactive({
  username: '',
  password: '',
  confirmPassword: '',
  email: '',
  code: '',
  captcha: '',
  captchaKey: ''
})

const captchaUrl = ref('')
const countdown = ref(0)
let timer = null

const refreshCaptcha = () => {
  const key = new Date().getTime().toString()
  registerForm.captchaKey = key
  captchaUrl.value = `/api/user/captcha?key=${key}`
}

onMounted(() => refreshCaptcha())

const sendEmailCode = async () => {
  if (!registerForm.email) return ElMessage.warning('请输入邮箱')
  if (!registerForm.captcha) return ElMessage.warning('请输入图形验证码')

  try {
    const params = new URLSearchParams()
    params.append('email', registerForm.email)
    params.append('captcha', registerForm.captcha)
    params.append('captchaKey', registerForm.captchaKey)
    params.append('type', 'register')

    const res = await axios.post('/api/user/sendEmailCode', params)

    if (res.data.success) {
      ElMessage.success('验证码已发送')
      countdown.value = 60
      timer = setInterval(() => {
        countdown.value--
        if (countdown.value <= 0) clearInterval(timer)
      }, 1000)
    } else {
      ElMessage.error(res.data.msg)
      refreshCaptcha()
    }
  } catch (e) {
    ElMessage.error('发送失败')
  }
}

const handleRegister = async () => {
  if (registerForm.password !== registerForm.confirmPassword) return ElMessage.warning('两次密码不一致')
  if (!registerForm.code) return ElMessage.warning('请输入邮箱验证码')

  isLoading.value = true
  try {
    const res = await axios.post('/api/user/register', {
      username: registerForm.username,
      password: registerForm.password,
      email: registerForm.email,
      code: registerForm.code
    })

    if (res.data.success) {
      ElMessage.success('注册成功，请登录')
      router.push('/login')
    } else {
      ElMessage.error(res.data.msg)
    }
  } catch (err) {
    ElMessage.error('注册失败')
  } finally {
    isLoading.value = false
  }
}

const goToLogin = () => router.push('/login')
</script>

<template>
  <div class="flat-container" :style="{ backgroundImage: `url(${backImg})` }">
    <div class="flat-card">
      <div class="header-row">
        <el-button link :icon="Back" @click="goToLogin" class="back-btn">登录</el-button>
        <h2 class="card-title" style="margin:0; flex:1">注册新账号</h2>
      </div>

      <el-form :model="registerForm" :size="formSize" class="register-scroll">
        <el-form-item>
          <el-input v-model="registerForm.username" placeholder="用户名 (2-20字符)" :prefix-icon="User" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="registerForm.email" placeholder="电子邮箱" :prefix-icon="Message" />
        </el-form-item>

        <div class="code-group">
          <el-input v-model="registerForm.captcha" placeholder="图形验证码" :prefix-icon="Picture" style="width: 60%" />
          <img :src="captchaUrl" @click="refreshCaptcha" class="captcha-img" title="点击刷新" />
        </div>

        <div class="code-group">
          <el-input v-model="registerForm.code" placeholder="邮箱验证码" :prefix-icon="Key" style="width: 60%" />
          <el-button class="code-btn" :disabled="countdown > 0" @click="sendEmailCode">
            {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
          </el-button>
        </div>

        <el-form-item>
          <el-input v-model="registerForm.password" type="password" placeholder="设置密码" :prefix-icon="Lock"
            show-password />
        </el-form-item>
        <el-form-item>
          <el-input v-model="registerForm.confirmPassword" type="password" placeholder="确认密码" :prefix-icon="Lock" />
        </el-form-item>

        <el-button type="success" class="action-btn" :loading="isLoading" @click="handleRegister">
          立即注册
        </el-button>
      </el-form>
    </div>

    <div class="footer-text">
      <p>2022 © Powered By <span style="color: #0e90d2">CrazyStone</span></p>
    </div>
  </div>
</template>

<style scoped>
/* 复用与 Login.vue 一致的样式 */
.flat-container {
  width: 100vw;
  height: 100vh;
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  position: relative;
}

.flat-card {
  width: 400px;
  /* 注册页内容多，稍微高一点 */
  min-height: 500px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(5px);
  border-radius: 8px;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.15);
  padding: 30px 35px;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  justify-content: center;
  margin-bottom: 20px;
}

.card-title {
  text-align: center;
  color: #333;
  font-weight: 500;
  font-size: 1.8rem;
  letter-spacing: 1px;
}

.action-btn {
  width: 100%;
  margin-top: 20px;
  height: 44px;
  font-size: 16px;
  letter-spacing: 2px;
}

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

.header-row {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
  position: relative;
}

.back-btn {
  position: absolute;
  left: 0;
  font-size: 14px;
}

.footer-text {
  position: absolute;
  bottom: 40px;
  text-align: center;
  color: #fff;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.5);
  font-size: 14px;
}

/* 注册内容如果过多可以滚动 */
.register-scroll {
  max-height: 550px;
  overflow-y: auto;
  padding-right: 5px;
}

.register-scroll::-webkit-scrollbar {
  width: 0;
}
</style>