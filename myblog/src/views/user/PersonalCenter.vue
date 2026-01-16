<script setup>
import { ref, reactive, onMounted, inject } from 'vue'
import { useStore } from '@/stores/my'
import { ElMessage } from 'element-plus'
import Top from '@/components/Top.vue'
import { User, Timer, Edit, Document, ChatLineRound, Plus } from '@element-plus/icons-vue'

const store = useStore()
const axios = inject('axios')

const activeTab = ref('timeline')
const activities = ref([]) // 足迹
const myArticles = ref([]) // 我的文章
const myComments = ref([]) // 我的评论
const loading = ref(false)

const likedArticles = ref([])
const likedComments = ref([])

const getLikes = () => {
  if (!store.user.user) return

  // 1. 获取点赞的文章 (修改传参方式)
  // 错误写法: axios.post(url, { userId: ... }) -> 发送的是 JSON Body
  // 正确写法: 拼接 URL 或使用 params
  axios.post('/api/article/getMyLikedArticles?userId=' + store.user.user.id).then(res => {
    if (res.data.success) {
      likedArticles.value = res.data.map.articles
    }
  })

  // 2. 获取点赞的评论 (修改传参方式)
  axios.post('/api/comment/getMyLikedComments?userId=' + store.user.user.id).then(res => {
    if (res.data.success) {
      likedComments.value = res.data.map.comments
    }
  })
}
// 表单数据
const userInfoForm = reactive({
  id: '',
  username: '',
  name: '',
  email: '',
  avatar: '', // 头像
  code: '' // 【新增】验证码
})

// --- 核心功能 1: 头像上传 ---
// 上传成功的回调
function handleAvatarSuccess(response, uploadFile) {
  // response 是后端 FileController 返回的 Result 对象
  if (response.success) {
    // response.map.url 就是 /api/file/images/xxx.jpg
    userInfoForm.avatar = response.map.url
    ElMessage.success('头像上传成功，请记得点击“保存修改”按钮！')
  } else {
    ElMessage.error('上传失败: ' + response.msg)
  }
}
function beforeAvatarUpload(rawFile) {
  if (rawFile.type !== 'image/jpeg' && rawFile.type !== 'image/png') {
    ElMessage.error('头像必须是 JPG 或 PNG 格式!')
    return false
  } else if (rawFile.size / 1024 / 1024 > 2) {
    ElMessage.error('头像大小不能超过 2MB!')
    return false
  }
  return true
}

// 【新增】发送验证码
const sending = ref(false)
const timer = ref(0)
function sendEmailCode() {
  if (!userInfoForm.email) return ElMessage.warning('请先填写邮箱')
  sending.value = true
  axios.post('/api/user/sendEmailCode?email=' + userInfoForm.email).then(res => {
    if (res.data.success) {
      ElMessage.success('验证码已发送')
      timer.value = 60
      const interval = setInterval(() => {
        timer.value--
        if (timer.value <= 0) clearInterval(interval)
      }, 1000)
    } else {
      ElMessage.error(res.data.msg)
    }
    sending.value = false
  })
}

// --- 数据加载 ---
function loadAllData() {
  if (!store.user.user) return
  const u = store.user.user

  // 1. 初始化表单 (修复 Bug: 确保显示正确的数据)
  userInfoForm.id = u.id
  userInfoForm.username = u.username
  userInfoForm.name = u.name || u.username
  userInfoForm.email = u.email
  userInfoForm.avatar = u.avatar

  // 2. 加载日志
  axios.get('/api/oplog/getMyLogs?userId=' + u.id).then(res => {
    if (res.data.success) activities.value = res.data.map.logs
  })

  // 3. 加载我的文章 (功能 4)
  axios.post('/api/article/getMyArticles?userId=' + u.id).then(res => {
    if (res.data.success) myArticles.value = res.data.map.articles
  })

  // 4. 加载我的评论 (功能 5) - 注意这里传的是用户名，因为评论表存的是author名
  axios.post('/api/comment/getMyComments?username=' + (u.username)).then(res => {
    if (res.data.success) myComments.value = res.data.map.comments
  })
}

// --- 提交修改 ---
function submitUpdate() {
  axios.post('/api/user/updateInfo', userInfoForm).then(res => {
    if (res.data.success) {
      ElMessage.success('修改成功')
      store.login(res.data.map.user) // 更新 Store
      userInfoForm.code = '' // 清空验证码
      // 关键：更新 Pinia 里的用户信息，解决刷新前显示旧数据的Bug
      store.login(res.data.map.user)
    } else {
      ElMessage.error(res.data.msg)
    }
  })
}

onMounted(() => {
  loadAllData()
  getLikes() // <--- 别忘了调用这个方法！原代码中可能漏了调用，或者在 tab 切换时调用
})

// 工具：日期格式化 (处理 T)
const fmtDate = (str) => str ? str.replace('T', ' ') : ''
</script>

<template>
  <el-affix>
    <Top />
  </el-affix>

  <div class="center-container">
    <el-row :gutter="20">
      <el-col :xs="24" :sm="8">
        <el-card class="box-card" shadow="hover">
          <div class="user-header">
            <el-avatar :size="100" :src="store.user.user?.avatar || '/api/images/default.png'" />
            <h3 class="username">{{ store.user.user?.username }}</h3>
            <p class="reg-time">
              注册于: {{ fmtDate(store.user.user?.created) }}
            </p>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="16">
        <el-card shadow="hover" class="right-card">
          <el-tabs v-model="activeTab">

            <el-tab-pane name="timeline" label="我的足迹">
              <el-scrollbar max-height="500px">
                <el-timeline v-if="activities.length > 0" style="padding-top: 10px;">
                  <el-timeline-item v-for="(act, i) in activities" :key="i" :timestamp="fmtDate(act.created)"
                    placement="top" :color="act.type === 'COMMENT' ? '#409EFF' : '#909399'">
                    <span v-if="act.targetId && (act.type === 'BROWSE' || act.type === 'COMMENT')"
                      @click="$router.push('/article_comment/' + act.targetId)" style="cursor: pointer;"
                      class="log-content">
                      {{ act.content }}
                    </span>
                    <span v-else>{{ act.content }}</span>
                  </el-timeline-item>
                </el-timeline>
                <el-empty v-else description="暂无足迹" />
              </el-scrollbar>
            </el-tab-pane>

            <el-tab-pane name="articles" label="我的文章">
              <el-scrollbar max-height="500px">
                <div v-if="myArticles.length > 0">
                  <div v-for="art in myArticles" :key="art.id" class="list-item"
                    @click="$router.push('/article_comment/' + art.id)">
                    <span class="item-title">{{ art.title }}</span>
                    <span class="item-date">{{ fmtDate(art.created) }}</span>
                  </div>
                </div>
                <el-empty v-else description="你还没发布过文章" />
              </el-scrollbar>
            </el-tab-pane>

            <el-tab-pane label="我的点赞" name="likes">
              <el-tabs type="border-card" class="inner-tabs">

                <el-tab-pane label="赞过的文章">
                  <el-scrollbar max-height="500px">
                    <div v-if="likedArticles.length > 0">
                      <div v-for="item in likedArticles" :key="item.id" class="list-item"
                        @click="$router.push('/article_comment/' + item.id)">
                        <div class="item-left">
                          <el-icon class="icon-prefix article-icon">
                            <Document />
                          </el-icon>
                          <span class="item-title">{{ item.title }}</span>
                        </div>
                        <span class="item-date">{{ fmtDate(item.created) }}</span>
                      </div>
                    </div>
                    <el-empty v-else description="暂无赞过的文章" :image-size="100" />
                  </el-scrollbar>
                </el-tab-pane>

                <el-tab-pane label="赞过的评论">
                  <el-scrollbar max-height="500px">
                    <div v-if="likedComments.length > 0">
                      <div v-for="item in likedComments" :key="item.id" class="list-item comment-list-item"
                        @click="$router.push('/article_comment/' + (item.articleId || item.refId))">
                        <div class="comment-wrapper">
                          <div class="comment-text">
                            <el-icon class="icon-prefix comment-icon">
                              <ChatLineRound />
                            </el-icon>
                            <span>{{ item.content }}</span>
                          </div>
                          <div class="comment-source">
                            原文: <span class="source-title">《{{ item.targetName || '未知文章' }}》</span>
                          </div>
                        </div>
                        <span class="item-date">{{ fmtDate(item.created) }}</span>
                      </div>
                    </div>
                    <el-empty v-else description="暂无赞过的评论" :image-size="100" />
                  </el-scrollbar>
                </el-tab-pane>

              </el-tabs>
            </el-tab-pane>

            <el-tab-pane name="comments" label="我的评论">
              <el-scrollbar max-height="500px">
                <div v-if="myComments.length > 0">
                  <div v-for="cmt in myComments" :key="cmt.id" class="list-item"
                    @click="$router.push('/article_comment/' + cmt.articleId)">

                    <div style="width: 100%;">
                      <div class="cmt-header">
                        <span v-if="cmt.type === 'COMMENT'">
                          评论了文章 <span class="highlight">《{{ cmt.targetTitle }}》</span>
                        </span>
                        <span v-else>
                          在 <span class="highlight">《{{ cmt.articleTitle }}》</span> 中回复了
                          <span class="highlight">@{{ cmt.targetUser }}</span>
                          <span style="color:#999; font-size:12px; margin-left:5px;">: {{ cmt.targetContent.substring(0,
                            10) }}...</span>
                        </span>
                        <span class="item-date" style="float: right;">{{ fmtDate(cmt.created) }}</span>
                      </div>

                      <div class="cmt-content">
                        {{ cmt.content }}
                      </div>
                    </div>

                  </div>
                </div>
                <el-empty v-else description="你还没发表过评论" />
              </el-scrollbar>
            </el-tab-pane>

            <el-tab-pane name="settings" label="资料设置">
              <el-form label-width="80px" style="max-width: 500px; margin-top: 20px;">
                <el-form-item label="头像">
                  <el-upload class="avatar-uploader" action="/api/file/upload" :show-file-list="false"
                    :on-success="handleAvatarSuccess" :before-upload="beforeAvatarUpload" name="file">
                    <img v-if="userInfoForm.avatar" :src="userInfoForm.avatar" class="avatar" />
                    <el-icon v-else class="avatar-uploader-icon">
                      <Plus />
                    </el-icon>
                  </el-upload>
                </el-form-item>

                <el-form-item label="用户名">
                  <el-input v-model="userInfoForm.username" placeholder="修改登录用户名" />
                </el-form-item>

                <el-form-item label="邮箱">
                  <el-input v-model="userInfoForm.email" placeholder="修改邮箱需验证" />
                </el-form-item>

                <el-form-item label="验证码" v-if="userInfoForm.email !== store.user.user?.email">
                  <div style="display: flex; gap: 10px; width: 100%;">
                    <el-input v-model="userInfoForm.code" placeholder="输入邮件验证码" />
                    <el-button type="primary" plain @click="sendEmailCode" :disabled="timer > 0 || sending">
                      {{ timer > 0 ? `${timer}s` : '获取验证码' }}
                    </el-button>
                  </div>
                </el-form-item>

                <el-form-item>
                  <el-button type="primary" @click="submitUpdate">保存修改</el-button>
                </el-form-item>
              </el-form>
            </el-tab-pane>
          </el-tabs>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.center-container {
  max-width: 1100px;
  margin: 30px auto;
  padding: 0 15px;
}

.user-header {
  text-align: center;
  padding: 20px 0;
}

.reg-time {
  font-size: 12px;
  color: #999;
  margin-top: 5px;
}

.right-card {
  min-height: 600px;
}

.list-item {
  display: flex;
  justify-content: space-between;
  padding: 15px;
  border-bottom: 1px solid #eee;
  cursor: pointer;
}

.list-item:hover {
  background-color: #f9f9f9;
}

.item-title {
  font-weight: bold;
  color: #333;
}

.item-date {
  font-size: 12px;
  color: #999;
}

/* 评论样式 */
.cmt-header {
  font-size: 13px;
  color: #666;
  margin-bottom: 5px;
}

.highlight {
  color: #409EFF;
  font-weight: 500;
}

.cmt-content {
  font-size: 14px;
  color: #333;
  line-height: 1.5;
}

/* 头像上传 */
.avatar-uploader .el-upload {
  border: 1px dashed var(--el-border-color);
  border-radius: 6px;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  transition: var(--el-transition-duration-fast);
}

.avatar-uploader .el-upload:hover {
  border-color: var(--el-color-primary);
}

.avatar-uploader-icon {
  font-size: 28px;
  color: #8c939d;
  width: 100px;
  height: 100px;
  text-align: center;
  line-height: 100px;
}

.avatar {
  width: 100px;
  height: 100px;
  display: block;
}

.log-content:hover {
  color: #409EFF !important;
  text-decoration: underline;
}
</style>