<script setup>
import Top from '@/components/Top.vue';
import Comment from '@/components/Comment.vue';
import { useRoute } from 'vue-router';
import { inject, reactive, nextTick } from 'vue'; // ✅ 引入 nextTick
import { ElMessageBox, ElMessage } from 'element-plus';
import { ref, computed, onMounted, watch } from 'vue'
import { useStore } from '@/stores/my'
import { marked } from 'marked'
import { LocationInformation, Close, StarFilled } from '@element-plus/icons-vue'
// === 引入 FontAwesome ===
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome';
import { library } from '@fortawesome/fontawesome-svg-core';
// 引入具体的图标：faThumbsUp(大拇指), faCommentDots(类似短信/气泡)
import { faThumbsUp, faCommentDots } from '@fortawesome/free-solid-svg-icons';

library.add(faThumbsUp, faCommentDots)

const route = useRoute();
const axios = inject('axios');
const store = useStore();

// 数据结构
let articleAndComment = reactive({
  "article": { "content": "", "likes": 0 },
  "comments": []
});

// pageParams
let pageParams = reactive({
  "page": 1,
  "rows": 5,
  "sort": "new"
})

const noMore = ref(false)
const loading = ref(false)
const disabled = computed(() => loading.value || noMore.value)

const commentContent = ref('')
const canComment = ref(false)

// === 灯箱控制 ===
const lightboxVisible = ref(false)
const lightboxImageSrc = ref('')

// === 1. 新增一个响应式变量存总数 ===
const commentTotal = ref(0) // 记录评论总数

// === 文章点赞逻辑 ===
function likeArticle() {
  if (!store.user.user) {
    ElMessage.warning("请先登录！")
    return
  }

  axios.post('/api/statistic/likeArticle?articleId=' + articleAndComment.article.id)
    .then(res => {
      if (res.data.success) {
        if (!articleAndComment.article.likes) articleAndComment.article.likes = 0;

        if (res.data.msg === "点赞成功") {
          articleAndComment.article.likes++;
        } else if (res.data.msg === "取消点赞") {
          articleAndComment.article.likes--;
        }
      } else {
        ElMessage.warning(res.data.msg);
      }
    });
}
// === 图片灯箱 ===
function handleContentClick(event) {
  if (event.target.tagName === 'IMG') {
    lightboxImageSrc.value = event.target.src
    lightboxVisible.value = true
    document.body.style.overflow = 'hidden'
  }
}

function closeLightbox() {
  lightboxVisible.value = false
  lightboxImageSrc.value = ''
  document.body.style.overflow = ''
}

// 过滤有效评论
const validComments = computed(() => {
  return articleAndComment.comments.filter(comment => comment && comment.id)
})

watch(() => store.user.user, (newUser) => {
  updateCommentPermission(newUser)
}, { immediate: true })

function updateCommentPermission(user) {
  // ✅ 修复：只要用户对象存在且有 ID，就视为已登录，允许显示评论框
  if (user && user.id) {
    canComment.value = true
  } else {
    canComment.value = false
  }
}

onMounted(() => {
  loadArticleAndComments()
})

// ✅【新增】滚动到指定评论的核心函数
function scrollToTargetComment() {
  const targetId = route.query.targetId;
  if (targetId) {
    nextTick(() => {
      // 尝试找到对应ID的元素
      const element = document.getElementById(`comment-${targetId}`);
      if (element) {
        // 平滑滚动到该元素
        element.scrollIntoView({ behavior: 'smooth', block: 'center' });

        // 可选：添加一个高亮动画效果，提示用户是这条
        element.style.transition = "background-color 0.5s";
        element.style.backgroundColor = "#fff8e1"; // 浅黄色高亮
        setTimeout(() => {
          element.style.backgroundColor = "#fff";
        }, 2000);
      }
    });
  }
}

function loadArticleAndComments() {
  axios({
    method: 'post',
    url: '/api/article/getArticleAndFirstPageCommentByArticleId?articleId=' + route.params.articleId,
    data: pageParams
  }).then((response) => {
    if (response.data.success) {
      if (response.data.map.article != null) {
        articleAndComment.article = response.data.map.article
        if (articleAndComment.article.likes == null) {
          articleAndComment.article.likes = 0;
        }

        articleAndComment.comments = response.data.map.comments || []

        // 【修复】获取后端返回的总数
        // 如果后端没返回 total (兼容旧代码)，就暂时用当前长度兜底
        if (response.data.map.total !== undefined) {
          commentTotal.value = response.data.map.total
        } else {
          commentTotal.value = articleAndComment.comments.length
        }

        if (!response.data.map.comments || response.data.map.comments.length < pageParams.rows) {
          noMore.value = true;
        }

        // ✅【新增】数据加载完成后，尝试滚动到指定评论
        scrollToTargetComment();

      } else {
        ElMessageBox.alert("无文章！", '结果')
      }
    } else {
      ElMessageBox.alert(response.data.msg, '结果')
    }
  }).catch((error) => {
    console.error(error)
    ElMessageBox.alert("系统错误！", '结果')
  })
}

const compiledContent = computed(() => {
  return articleAndComment.article.content ? marked.parse(articleAndComment.article.content) : ''
})

// === 无限滚动加载评论 ===
const load = () => {
  if (loading.value || noMore.value) return

  loading.value = true
  pageParams.page++

  axios({
    method: 'post',
    url: '/api/comment/getAPageCommentByArticleId?articleId=' + route.params.articleId,
    data: pageParams
  }).then((response) => {
    if (response.data.success) {
      let comments = response.data.map.comments
      if (comments && comments.length > 0) {
        articleAndComment.comments.push(...comments)

        // 【修复】加载更多时也更新总数（防止有人新发评论）
        if (response.data.map.total !== undefined) {
          commentTotal.value = response.data.map.total
        }

      }
    } else {
      ElMessageBox.alert(response.data.msg, '结果')
    }
    loading.value = false
  }).catch(() => {
    loading.value = false
    ElMessageBox.alert("系统错误！", '结果')
  })
}

function handleCommentSortChange(val) {
  pageParams.page = 0
  articleAndComment.comments = []
  noMore.value = false
  load()
}

// 提交评论
function submit() {
  if (!store.user.user) {
    ElMessage.warning("请先登录！")
    return
  }

  if (!commentContent.value.trim()) {
    ElMessage.warning("评论内容不能为空！")
    return
  }

  axios({
    method: 'post',
    url: '/api/comment/insert',
    data: {
      "articleId": route.params.articleId,
      "content": commentContent.value,
      "author": store.user.user.username || store.user.user.name
    }
  }).then((response) => {
    if (response.data.success) {
      ElMessage.success("评论成功")
      commentContent.value = ""

      const newComment = response.data.map.Comment || response.data.map.comment
      if (newComment) {
        articleAndComment.comments.unshift(newComment)
        // 【修复】发表成功，总数 +1
        commentTotal.value++
      } else {
        pageParams.page = 1
        pageParams.sort = 'new'
        noMore.value = false
        articleAndComment.comments = []
        loadArticleAndComments()
      }

    } else {
      ElMessage.error(response.data.msg)
    }
  }).catch(() => {
    ElMessage.error("系统错误")
  })
}
</script>

<template>
  <el-affix>
    <Top />
  </el-affix>

  <el-row>
    <el-col :span="14" :offset="5">
      <h1 style="text-align: center; margin-top: 20px; font-size: 28px;">
        {{ articleAndComment.article.title }}
      </h1>

      <div class="article-meta-bar">
        <div class="meta-author">
          <el-avatar :size="24"
            :src="articleAndComment.article.authorAvatar || 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'" />
          <span style="margin-left: 8px;">{{ articleAndComment.article.author || '未知作者' }}</span>
        </div>
        <!-- 调试信息 
          <div v-show="articleAndComment.article" style="color: red; font-size: 12px;">
            文章数据: {{ JSON.stringify(articleAndComment.article) }}<br>
            作者头像: {{ articleAndComment.article.authorAvatar }}
          </div>
          -->
        <div class="meta-center">
          <span>发布于: {{ articleAndComment.article.created }}</span>
        </div>

        <div class="meta-right">
          <span v-if="articleAndComment.article.location">
            <el-icon>
              <LocationInformation />
            </el-icon>
            {{ articleAndComment.article.location }}
          </span>
        </div>
      </div>
      <el-divider />
    </el-col>
  </el-row>

  <el-row>
    <el-col :span="14" :offset="5">
      <div v-html="compiledContent" class="markdown-body" @click="handleContentClick"></div>
    </el-col>
  </el-row>

  <el-row justify="center" style="margin: 40px 0;">
    <el-button :type="articleAndComment.article.likes > 0 ? 'primary' : 'info'" circle size="large" @click="likeArticle"
      style="width: 60px; height: 60px;">
      <font-awesome-icon :icon="faThumbsUp" style="font-size: 26px;" />
    </el-button>
    <span style="margin-left: 15px; line-height: 60px; font-size: 18px; color: #666;">
      {{ articleAndComment.article.likes || 0 }} 人觉得很赞
    </span>
  </el-row>

  <el-row style="background-color: #f7f7f7; margin-top: 20px;">
    <el-col :span="14" :offset="5">

      <div style="display: flex; justify-content: space-between; align-items: center; padding: 15px 0;">
        <h3 style="margin: 0; color: #333;">评论区</h3>
        <el-radio-group v-model="pageParams.sort" size="small" @change="handleCommentSortChange">
          <el-radio-button value="new">最新</el-radio-button>
          <el-radio-button value="hot">最热</el-radio-button>
        </el-radio-group>
      </div>

      <ul v-infinite-scroll="load" :infinite-scroll-disabled="disabled" class="infinite-list">

        <li class="infinite-list-item" v-if="canComment">
          <el-row>
            <el-col>
              <el-input v-model="commentContent" :autosize="{ minRows: 4 }" type="textarea" placeholder="写下你的评论..."
                maxlength="200" show-word-limit />
            </el-col>
          </el-row>

          <el-row justify="end">
            <el-col :xs="8" :sm="6" :md="4">
              <el-button @click="submit" type="primary" round style="margin-top: 10px;">
                提交评论
              </el-button>
            </el-col>
          </el-row>
        </li>

        <li v-for="(comment, index) in validComments" :key="comment.id" :id="'comment-' + comment.id"
          class="infinite-list-item">
          <Comment :comment="comment" :floor="pageParams.sort === 'new' ? (commentTotal - index) : null">
          </Comment>
        </li>

      </ul>

      <p v-if="loading" style="text-align: center; color: #999;">加载中...</p>
      <p v-if="noMore && validComments.length > 0" style="text-align: center; color: #999; padding-bottom: 20px;">
        没有更多评论了</p>
      <p v-if="validComments.length === 0 && !loading" style="text-align: center; color: #999; padding: 20px;">
        暂无评论，快来抢沙发吧！</p>

    </el-col>
  </el-row>

  <Teleport to="body">
    <transition name="fade">
      <div v-if="lightboxVisible" class="lightbox-overlay" @click="closeLightbox">
        <div class="lightbox-close" @click.stop="closeLightbox">
          <el-icon :size="40" color="#fff">
            <Close />
          </el-icon>
        </div>
        <img :src="lightboxImageSrc" class="lightbox-img" @click.stop />
      </div>
    </transition>
  </Teleport>
</template>

<style scoped>
.infinite-list {
  padding: 0;
  margin: 0;
  list-style: none;
}

.infinite-list-item {
  background: #fff;
  margin-bottom: 10px;
  margin-top: 10px;
  padding: 20px;
  border-radius: 4px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.markdown-body {
  line-height: 1.8;
  color: #333;
}

:deep(.markdown-body img) {
  max-width: 100%;
  cursor: zoom-in;
  transition: transform 0.2s;
}

:deep(.markdown-body img:hover) {
  opacity: 0.9;
}

.article-meta-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: #999;
  margin-bottom: 20px;
  padding: 0 10px;
  font-size: 14px;
}

.meta-author {
  display: flex;
  align-items: center;
  color: #333;
  font-weight: bold;
}

.meta-info {
  display: flex;
  align-items: center;
}

/* 灯箱样式 */
.lightbox-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.85);
  z-index: 9999;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: zoom-out;
}

.lightbox-img {
  max-width: 90%;
  max-height: 90%;
  object-fit: contain;
  box-shadow: 0 0 20px rgba(0, 0, 0, 0.5);
  cursor: default;
  border-radius: 4px;
  animation: zoomIn 0.3s ease;
}

.lightbox-close {
  position: absolute;
  top: 30px;
  right: 30px;
  cursor: pointer;
  padding: 10px;
  background: rgba(0, 0, 0, 0.3);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.3s;
}

.lightbox-close:hover {
  background: rgba(255, 255, 255, 0.2);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

@keyframes zoomIn {
  from {
    transform: scale(0.8);
    opacity: 0;
  }

  to {
    transform: scale(1);
    opacity: 1;
  }
}

/* 【修改】元数据栏布局样式 */
.article-meta-bar {
  position: relative;
  /* 为绝对定位做参照 */
  display: flex;
  justify-content: space-between;
  /* 左右贴边 */
  align-items: center;
  color: #999;
  margin-bottom: 20px;
  padding: 0 10px;
  font-size: 14px;
  height: 30px;
}

.meta-left {
  display: flex;
  align-items: center;
  z-index: 2;
  /* 防止被覆盖 */
}

.meta-center {
  position: absolute;
  /* 绝对定位到中间 */
  left: 50%;
  transform: translateX(-50%);
  color: #666;
  white-space: nowrap;
  z-index: 1;
}

.meta-right {
  display: flex;
  align-items: center;
  z-index: 2;
}

/* ... 其他样式 ... */

.markdown-body {
  line-height: 1.8;
  color: #333;
}

/* 【修改】图片自适应样式 */
:deep(.markdown-body img) {
  max-width: 100%;
  /* 宽度不超过容器 */
  height: auto;
  /* 高度自动，保持比例 */
  display: block;
  /* 防止行内元素间隙 */
  margin: 10px auto;
  /* 居中显示 */
  cursor: zoom-in;
  transition: transform 0.2s;
  border-radius: 4px;
  /* 可选：圆角 */
}
</style>