<script setup>
import { ref, inject, onMounted, onUnmounted, computed } from 'vue'
import { useStore } from '@/stores/my'
import { ElMessageBox, ElMessage } from 'element-plus'
import { Search, Bell } from '@element-plus/icons-vue' // 引入 Bell 图标
import { useRouter } from 'vue-router'

const size = ref(30)
const router = useRouter()
const store = useStore()
const axios = inject('axios')
const toLogin = inject("toLogin")
const toHome = inject("toHome")
const toAdminMain = inject('toAdminMain')

const userName = ref("")
const isLogined = ref(false)
const unreadCount = ref(0) // 未读消息数
const notifications = ref([]) // 消息列表

// === ✅【新增】判断是否是管理员 ===
const isAdmin = computed(() => {
  const u = store.user.user

  // 方式 A：根据 ID 判断 (推荐，适用于个人博客)
  // 假设数据库中 ID 为 1 的用户是站长
  if (u && u.id === 1) {
    return true
  }

  // 方式 B：如果你后端登录接口返回了 roles 或 authorities
  // if (u && u.roles && u.roles.includes('admin')) {
  //   return true
  // }

  return false
})

// 检查是否已登录
if (store.user.user != null) {
  userName.value = "hi " + store.user.user.username
  isLogined.value = true
}

// === 新增：获取未读数 ===
function getUnreadCount() {
  if (!isLogined.value) return
  axios.get('/api/notification/unreadCount').then(res => {
    if (res.data.success) {
      // 后端返回结构是 res.data.map.count
      unreadCount.value = res.data.map.count || 0
    }
  })
}

// === 新增：获取消息列表 ===
function getNotifications() {
  if (!isLogined.value) return
  // ✅ 改为调用新的 Go 接口 (POST /getAPageNotification)
  axios.post('/api/notification/getAPageNotification', {
    page: 1,
    rows: 5 // 下拉框只显示前 5 条
  }).then(res => {
    if (res.data.success) {
      // 后端返回结构是 res.data.map.data
      notifications.value = res.data.map.data || []
    }
  })
}

// === 新增：一键已读 ===
// 全部已读
function markAllRead() {
  axios.post('/api/notification/readAll').then(res => {
    if (res.data.success) {
      unreadCount.value = 0
      // 🔴 修改前: item.isRead = true
      // 🟢 修改后: item.status = 1
      notifications.value.forEach(item => item.status = 1)
    }
  })
}

// === ✅【新增】点击通知跳转 ===
function readNotification(item) {
  // 1. 如果有 commentId，说明后端已经支持精准定位
  const targetId = item.commentId;

  // 2. 跳转到文章详情页，并携带 targetId 参数
  // 即使 targetId 为 null，也会正常跳转到文章页顶部
  router.push({
    path: `/article_comment/${item.articleId}`,
    query: { targetId: targetId }
  });

  // 3. (可选) 这里可以顺便调用接口把该条设为已读，或者依靠用户查看列表时已经获取最新状态
  // item.isRead = true; 
}

// === 新增：定时轮询未读数 (可选，比如每30秒查一次) ===
let timer = null
onMounted(() => {
  if (isLogined.value) {
    getUnreadCount()
    // timer = setInterval(getUnreadCount, 30000) 
  }
})
onUnmounted(() => {
  if (timer) clearInterval(timer)
})

function toExit() {
  axios({
    method: 'post',
    url: '/api/logout'
  }).then((response) => {
    ElMessageBox.alert(response.data.msg, '结果')
    store.user.user = null
    isLogined.value = false
    userName.value = ""
    toHome()
  }).catch((error) => {
    ElMessageBox.alert("系统错误！", '结果')
  })
}

function goToLogin() {
  if (toLogin) {
    toLogin()
  } else {
    router.push('/login')
  }
}

// === ✅【新增】跳转到个人中心 ===
function toPersonalCenter() {
  router.push('/personal_center')
}
</script>

<template>
  <el-row class="top" align="middle">
    <el-col :span="3"></el-col>
    <el-col :span="12">
      <a @click="toHome">CrazyStone个人博客小站</a>
    </el-col>
    <el-col :span="6">
      <el-space :size="size">

        <el-popover v-if="isLogined" placement="bottom" :width="300" trigger="click" @show="getNotifications">
          <template #reference>
            <el-badge :value="unreadCount" :max="99" :hidden="unreadCount === 0" class="item">
              <el-icon class="searchIcon" :size="20" style="cursor: pointer; margin-top: 5px;">
                <Bell />
              </el-icon>
            </el-badge>
          </template>

          <div class="notification-box">
            <div class="header">
              <span>消息通知</span>
              <el-button link type="primary" size="small" @click="markAllRead">全部已读</el-button>
            </div>
            <el-divider style="margin: 10px 0" />
            <div v-if="notifications.length === 0" style="text-align: center; color: #999;">暂无消息</div>
            <ul v-else class="msg-list">
              <li v-for="item in notifications" :key="item.id" :class="{ unread: item.status === 0 }"
                @click="readNotification(item)">
                <div class="msg-title">
                  <el-tag size="small" :type="item.type === 'COMMENT' ? 'success' : 'warning'">
                    {{ item.type === 'COMMENT' ? '评论' : '回复' }}
                  </el-tag>
                  <span class="sender">{{ item.senderName }}</span>
                </div>
                <div class="msg-content">{{ item.content }}</div>
                <div class="msg-time">{{ item.created }}</div>
              </li>
            </ul>
          </div>
        </el-popover>
        <a @click="goToLogin" v-if="!isLogined">登录</a>
        <a @click="toAdminMain" v-if="isAdmin">后台管理</a>
        <a @click="toExit" v-if="isLogined">退出</a>
        <a @click="toPersonalCenter" v-if="isLogined">{{ userName }}</a>
        <router-link title="查询" :to="{ path: '/search' }" style="text-decoration: none;">
          <el-icon>
            <Search class="searchIcon" />
          </el-icon>
        </router-link>
      </el-space>
    </el-col>
    <el-col :span="3"></el-col>
  </el-row>
</template>

<style scoped>
* {
  font-size: 16px;
}

.top {
  background: #5f9ea0;
  height: 80px;
  color: #fff;
}

.top a {
  color: #fff;
  text-decoration: none;
  cursor: pointer;
  font-size: 20px;
}

a:hover {
  color: #10007A;
}

.searchIcon {
  color: white;
}

.searchIcon:hover {
  color: #10007A;
}

/* 消息列表样式 */
.notification-box .header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.msg-list {
  list-style: none;
  padding: 0;
  max-height: 300px;
  overflow-y: auto;
}

/* ✅【修改】增加鼠标手势和悬停效果 */
.msg-list li {
  padding: 10px 0;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
  transition: background-color 0.2s;
}

.msg-list li:hover {
  background-color: #f5f7fa;
}

.msg-list li.unread .msg-content {
  font-weight: bold;
  color: #333;
}

.msg-title {
  font-size: 12px;
  color: #666;
  margin-bottom: 5px;
}

.sender {
  margin-left: 5px;
  font-weight: bold;
}

.msg-content {
  font-size: 14px;
  color: #666;
  margin-bottom: 5px;
}

.msg-time {
  font-size: 12px;
  color: #999;
}
</style>