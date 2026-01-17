<script setup>
import { reactive, ref, inject, onMounted, nextTick } from 'vue'
import { useStore } from '@/stores/my'
import { ElMessage } from 'element-plus'
import { dateFormat } from '../js/tool'
import { Sugar, LocationInformation } from '@element-plus/icons-vue'

// 引入组件
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome';
// 引入具体图标
import { faThumbsUp, faCommentDots } from '@fortawesome/free-solid-svg-icons';

// 接收父组件传递的参数
const props = defineProps(['comment', 'floor'])

const store = useStore()
const axios = inject('axios')

// --- 子评论(回复)相关数据 ---
const replies = ref([])
const replyPage = ref(1)
const replyRows = ref(5)
const replyTotal = ref(0)
const showAllReplies = ref(false)

// 回复输入框控制
const showReplyInput = ref(false)
const replyContent = ref('')
const replyPlaceholder = ref('回复层主...')
const currentTargetUid = ref(null)

// 回复输入框 Ref
const replyInputRef = ref(null)

// 加载子评论
function loadReplies() {
  axios.get(`/api/reply/getReplies?commentId=${props.comment.id}&page=${replyPage.value}&rows=${replyRows.value}&_t=${new Date().getTime()}`)
    .then(res => {
      if (res.data.success) {
        replies.value = res.data.map.replies || []
        replyTotal.value = res.data.map.total || 0
      }
    })
    .catch(err => console.error("加载回复失败", err))
}

// 展开/折叠回复
function toggleReplies() {
  if (!showAllReplies.value) {
    replyRows.value = 100
    loadReplies()
  } else {
    replyRows.value = 5
    loadReplies()
  }
  showAllReplies.value = !showAllReplies.value
}

// 点击"回复"按钮 (准备回复)
async function prepareReply(targetUser) {
  // 1. 检查登录
  if (!store.user.user) {
    ElMessage.warning("请先登录！")
    return
  }

  // 2. 设置目标
  if (targetUser) {
    currentTargetUid.value = targetUser.id
    replyPlaceholder.value = `回复 @${targetUser.username}:`
  } else {
    currentTargetUid.value = null // 回复层主
    replyPlaceholder.value = `回复 @${props.comment.author}:`
  }

  // 3. 显示输入框
  showReplyInput.value = true
  replyContent.value = ""

  // 等待 DOM 更新后自动聚焦
  await nextTick()
  if (replyInputRef.value) {
    replyInputRef.value.focus()
  }
}

// 发送回复
function sendReply() {
  if (!replyContent.value.trim()) {
    ElMessage.warning("请输入回复内容")
    return
  }

  let param = {
    content: replyContent.value,
    commentId: props.comment.id,
    userId: store.user.user.id,
    toUid: currentTargetUid.value
  }

  axios.post('/api/reply/insert', param)
    .then(res => {
      if (res.data.success) {
        ElMessage.success("回复成功")
        showReplyInput.value = false
        replyContent.value = ""
        loadReplies()
      } else {
        ElMessage.error(res.data.msg || "回复失败")
      }
    })
    .catch(() => ElMessage.error("系统繁忙"))
}

// 初始化加载
onMounted(() => {
  if (props.comment.id) {
    loadReplies()
  }
})

// 通用的点赞函数
function likeTargetComment(commentObj, type) {
  if (!store.user.user) {
    ElMessage.warning("请先登录")
    return
  }

  let url = ''
  if (type === 'REPLY') {
      url = '/api/reply/likeReply?replyId=' + commentObj.id
  } else {
      url = '/api/comment/likeComment?commentId=' + commentObj.id
  }

  axios.post(url)
    .then(res => {
      if (res.data.success) {
        if (!commentObj.likes) commentObj.likes = 0
        if (res.data.msg === "点赞成功") {
          commentObj.likes++
        } else if (res.data.msg === "取消点赞") {
          commentObj.likes--
        }
      } else {
        ElMessage.warning(res.data.msg)
      }
    })
}
</script>

<template>
  <div class="comment-item">
    <div class="main-comment">
      <div class="user-avatar">
        <img :src="comment.avatar || '/api/images/default.png'" alt="avatar">
      </div>
      <div class="content-box">
        <div class="user-info">
          <span class="username">{{ comment.author }}</span>
          <span class="floor-tag" v-if="floor">#{{ floor }}楼</span>
        </div>
        <div class="comment-text">{{ comment.content }}</div>

        <div class="comment-actions">
          <div class="actions-group">
            <span class="time-text">{{ dateFormat(comment.created, 'yyyy-MM-dd HH:mm:ss') }}</span>
            <span class="location-text" v-if="comment.location" style="color: #999; font-size: 12px;">
              <el-icon size="12" style="margin-right: 2px">
                <LocationInformation />
              </el-icon>
              {{ comment.location }}
            </span>
          </div>

          <div class="actions-group">
            <span class="action-item like-action" @click="likeTargetComment(comment, 'COMMENT')">
              <font-awesome-icon :icon="faThumbsUp" :class="{ 'liked': comment.likes > 0 }" />
              <span class="action-text">{{ comment.likes || 0 }}</span>
            </span>
            <span class="action-item reply-action" @click="prepareReply(null)">
              <font-awesome-icon :icon="faCommentDots" style="margin-right: 4px;" />
              <span class="action-text">回复</span>
            </span>
          </div>
        </div>

      </div>
    </div>

    <div class="sub-reply-container" v-if="replies.length > 0 || showReplyInput">
      <div v-for="reply in replies" :key="reply.id" class="reply-item">
        <div class="reply-line">
          <img class="mini-avatar" :src="reply.avatar || '/api/images/default.png'" />
          <span class="reply-user">{{ reply.username }}</span>
          <span v-if="reply.targetName" style="color: #409EFF; margin: 0 4px; font-size: 12px;">
            回复 @{{ reply.targetName }} :
          </span>
          <span v-else class="reply-colon"> : </span>
        </div>

        <div class="reply-text">{{ reply.content }}</div>

        <div class="reply-actions">
          <div class="actions-group">
            <span class="time-text">{{ dateFormat(reply.created, 'yyyy-MM-dd HH:mm:ss') }}</span>
            <span class="location-text" v-if="reply.location" style="color: #999; font-size: 12px;">
              <el-icon size="12" style="margin-right: 2px">
                <LocationInformation />
              </el-icon>
              {{ reply.location }}
            </span>
          </div>

          <div class="actions-group">
            <span class="action-btn" @click="likeTargetComment(reply, 'REPLY')" :class="{ 'liked': reply.likes > 0 }"
              style="display:inline-flex; align-items:center;">
              <font-awesome-icon :icon="faThumbsUp" style="margin-right: 4px;" />
              {{ reply.likes || 0 }}
            </span>
            <span class="action-btn" @click="prepareReply({ id: reply.userId, username: reply.username })">
              <font-awesome-icon :icon="faCommentDots" style="margin-right: 4px;" />
              回复
            </span>
          </div>
        </div>
      </div>

      <div v-if="replyTotal > 5" class="expand-btn" @click="toggleReplies">
        {{ showAllReplies ? '收起回复' : `查看剩余 ${replyTotal - replies.length} 条回复` }}
      </div>

      <div v-if="showReplyInput" class="reply-input-box">
        <el-input ref="replyInputRef" v-model="replyContent" :placeholder="replyPlaceholder" size="small"
          style="margin-bottom: 5px;" />
        <div style="text-align: right;">
          <el-button size="small" @click="showReplyInput = false">取消</el-button>
          <el-button type="primary" size="small" @click="sendReply">发送</el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.comment-item {
  padding: 15px 0;
  border-bottom: 1px solid #f0f0f0;
}

.main-comment {
  display: flex;
  gap: 15px;
}

.user-avatar img {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
}

.content-box {
  flex: 1;
}

.user-info {
  margin-bottom: 5px;
  display: flex;
  justify-content: space-between;
}

.username {
  font-weight: bold;
  color: #333;
  font-size: 14px;
}

.floor-tag {
  color: #999;
  font-size: 12px;
}

.comment-text {
  font-size: 14px;
  line-height: 1.6;
  color: #333;
  margin-bottom: 8px;
  word-break: break-all;
}

.action-btn {
  margin-left: 15px;
  cursor: pointer;
  color: #999;
  display: inline-flex;
  align-items: center;
}

.action-btn:hover {
  color: #409EFF;
}

/* 子回复样式 */
.sub-reply-container {
  margin-top: 10px;
  margin-left: 55px;
  /* 缩进 */
  background-color: #f9f9f9;
  padding: 10px;
  border-radius: 4px;
}

.reply-item {
  margin-bottom: 8px;
  padding-bottom: 8px;
  border-bottom: 1px dashed #eee;
}

.reply-line {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  font-size: 13px;
  margin-bottom: 4px;
}

.mini-avatar {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  margin-right: 6px;
  vertical-align: middle;
}

.reply-user {
  font-weight: 500;
  color: #333;
}

.reply-text {
  font-size: 12px;
  line-height: 1.6;
  color: #333;
  margin-bottom: 8px;
  word-break: break-all;
  padding-left: 26px;
}

.reply-actions {
  font-size: 12px;
  color: #aaa;
  padding-left: 30px;
  /* 对齐文字 */
}

.expand-btn {
  font-size: 12px;
  color: #409EFF;
  cursor: pointer;
  margin-top: 5px;
}

.reply-input-box {
  margin-top: 10px;
  padding: 10px;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 4px;
}

.comment-actions {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-top: 8px;
  color: #9499a0;
  font-size: 14px;
  user-select: none;
}

.action-item {
  cursor: pointer;
  display: flex;
  align-items: center;
  transition: all 0.2s;
}

.action-item:hover {
  color: #409EFF;
}

.liked {
  color: #409EFF;
}

.action-text {
  margin-left: 5px;
}

/* ✅ 【修改】主评论操作栏样式 */
.comment-actions {
  display: flex;
  justify-content: space-between;
  /* 左右两端对齐 */
  align-items: center;
  margin-top: 8px;
  color: #9499a0;
  font-size: 14px;
  user-select: none;
}

/* ✅ 【修改】子回复操作栏样式 */
.reply-actions {
  font-size: 12px;
  color: #aaa;
  padding-left: 30px;
  /* 增加 Flex 布局以支持两端对齐 */
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* ✅ 【新增】通用分组样式，用于控制组内元素间距 */
.actions-group {
  display: flex;
  align-items: center;
  gap: 20px;
  /* 控制 时间-位置 或 点赞-回复 之间的间距 */
}

.action-item,
.action-btn {
  cursor: pointer;
  display: flex;
  align-items: center;
  transition: all 0.2s;
}

.action-item:hover,
.action-btn:hover {
  color: #409EFF;
}

.liked {
  color: #409EFF;
}

.action-text {
  margin-left: 5px;
}

.action-btn {
  margin-left: 0;
  /* 重置可能存在的 margin */
}
</style>