<script setup>
import { reactive, ref, inject, provide, nextTick, onMounted, watch } from 'vue'
import Editor from '@tinymce/tinymce-vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { useStore } from '@/stores/my'
import Cropper from "@/components/Cropper.vue";
import { undefine, nullZeroBlank } from "@/js/tool.js"
import { useRoute } from 'vue-router'
import { Position, MapLocation } from '@element-plus/icons-vue'
import 'leaflet/dist/leaflet.css'
import L from 'leaflet'
import { useRouter } from 'vue-router'

const store = useStore()
const route = useRoute()
const axios = inject('axios')

let type = "add"
const header = ref("发布文章")

const router = useRouter() // 2. 【新增】初始化 router

// === 分类相关 ===
const categoryOptions = ref([])

// === 标签相关 ===
const selectedCategory = ref([])
const dynamicTags = ref([])

// === 地图相关变量 ===
const mapVisible = ref(false)
let mapInstance = null
let markerInstance = null

// === 状态标志 ===
const isSubmitting = ref(false)

// 加载分类树
function loadCategoryTree() {
  axios.get('/api/category/getTree').then(res => {
    if (res.data.success) {
      categoryOptions.value = res.data.map.data
    }
  })
}

// 图片上传配置
const image_upload_handler = (blobInfo, progress) => new Promise((resolve, reject) => {
  const xhr = new XMLHttpRequest();
  xhr.withCredentials = false;
  xhr.open('POST', '/api/article/upload');
  xhr.upload.onprogress = (e) => { progress(e.loaded / e.total * 100); };
  xhr.onload = () => {
    if (xhr.status === 403) { reject({ message: 'HTTP Error: ' + xhr.status, remove: true }); return; }
    if (xhr.status < 200 || xhr.status >= 300) { reject('HTTP Error: ' + xhr.status); return; }
    const json = JSON.parse(xhr.response);
    if (!json || !json.map || typeof json.map.url != 'string') { reject('Invalid JSON: ' + xhr.responseText); return; }
    resolve(json.map.url);
  };
  xhr.onerror = () => { reject('Image upload failed. Code: ' + xhr.status); };
  const formData = new FormData();
  formData.append('file', blobInfo.blob(), blobInfo.filename());
  xhr.send(formData);
  resolve("url_placeholder");
});

const apiKey = ref('hyeykcd9rhyowt7om2q282pbdhjd8nnw7tci613prb5vgo7d')
const init = reactive({
  language: "zh_CN",
  placeholder: "在这里输入文字",
  plugins: ['image', 'code'],
  toolbar: 'image',
  images_file_types: 'jpg,jpeg,png,gif,bmp',
  images_upload_handler: image_upload_handler,
  convert_urls: false
})

// === 文章对象 ===
let article = reactive({
  "title": "",
  "tags": "",
  "content": "",
  "categories": "",
  "thumbnail": "",
  "location": ""
})

const cropper1 = ref(null)

// === 自动保存草稿 ===
const DRAFT_KEY = 'blog_publish_draft_v1'
let draftTimer = null

// 监听 article 对象变化，自动保存
watch(article, (newVal) => {
  if (isSubmitting.value) return

  if (draftTimer) clearTimeout(draftTimer)

  draftTimer = setTimeout(() => {
    const draftData = {
      title: newVal.title,
      tags: newVal.tags,
      content: newVal.content,
      categories: newVal.categories,
      location: newVal.location,
      thumbnail: newVal.thumbnail,
      selectedCategory: selectedCategory.value,
      dynamicTags: dynamicTags.value
    }
    localStorage.setItem(DRAFT_KEY, JSON.stringify(draftData))
  }, 5000)
}, { deep: true })

// === 自动获取定位 ===
function autoLocate() {
  if (!navigator.geolocation) {
    ElMessage.error('您的浏览器不支持地理定位')
    return
  }

  const loadingMsg = ElMessage({
    message: '正在尝试获取定位...',
    type: 'info',
    duration: 0,
    showClose: true
  })

  navigator.geolocation.getCurrentPosition(async (position) => {
    const { latitude, longitude } = position.coords
    try {
      const res = await fetch(`https://nominatim.openstreetmap.org/reverse?format=json&lat=${latitude}&lon=${longitude}&zoom=18&accept-language=zh-CN&addressdetails=1`)
      const data = await res.json()

      if (data.display_name) {
        let fullAddress = data.display_name
        if (fullAddress.length > 150) {
          fullAddress = fullAddress.substring(0, 150) + "..."
        }
        article.location = fullAddress
        loadingMsg.close()
        ElMessage.success('定位成功: ' + fullAddress)
      } else {
        const addr = data.address || {}
        const locationParts = []
        if (addr.state) locationParts.push(addr.state)
        if (addr.city) locationParts.push(addr.city)
        if (addr.county) locationParts.push(addr.county)
        if (addr.town) locationParts.push(addr.town)
        if (addr.road) locationParts.push(addr.road)

        if (locationParts.length > 0) {
          article.location = locationParts.join('')
          loadingMsg.close()
          ElMessage.success('定位成功: ' + article.location)
        } else {
          loadingMsg.close()
          ElMessage.error('无法获取详细地址信息')
        }
      }
    } catch (e) {
      loadingMsg.close()
      ElMessage.error('获取地址名称失败，请手动输入')
    }
  }, (err) => {
    loadingMsg.close()
    ElMessage.error('自动定位失败，请检查浏览器权限或使用地图选取')
  }, {
    enableHighAccuracy: true,
    timeout: 15000,
    maximumAge: 0
  })
}

// === 打开地图手动选取 ===
function openMap() {
  mapVisible.value = true
  nextTick(() => {
    initMap()
  })
}

function initMap() {
  if (mapInstance) return
  const defaultLat = 39.9042
  const defaultLng = 116.4074
  mapInstance = L.map('map-container').setView([defaultLat, defaultLng], 4)
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; OpenStreetMap contributors'
  }).addTo(mapInstance)

  mapInstance.on('click', async (e) => {
    const { lat, lng } = e.latlng
    if (markerInstance) {
      markerInstance.setLatLng(e.latlng)
    } else {
      markerInstance = L.marker(e.latlng).addTo(mapInstance)
    }
    try {
      const res = await fetch(`https://nominatim.openstreetmap.org/reverse?format=json&lat=${lat}&lon=${lng}&zoom=18&accept-language=zh-CN&addressdetails=1`)
      const data = await res.json()

      if (data.display_name) {
        let fullAddress = data.display_name
        if (fullAddress.length > 150) {
          fullAddress = fullAddress.substring(0, 150) + "..."
        }
        article.location = fullAddress
        ElMessage.success(`已选中: ${fullAddress}`)
      } else {
        const addr = data.address || {}
        const locationParts = []
        if (addr.state) locationParts.push(addr.state)
        if (addr.city) locationParts.push(addr.city)
        if (addr.county) locationParts.push(addr.county)
        if (addr.town) locationParts.push(addr.town)
        if (addr.road) locationParts.push(addr.road)

        if (locationParts.length > 0) {
          article.location = locationParts.join('')
          ElMessage.success(`已选中: ${article.location}`)
        } else {
          ElMessage.warning('地址解析失败，请直接输入')
        }
      }
    } catch (err) {
      ElMessage.warning('地址解析失败，请直接输入')
    }
  })
}

// === 发布文章 ===
async function publishArticle() {
  isSubmitting.value = true

  let thumbnail = cropper1.value.getThumbnail()
  if (undefine(thumbnail) || nullZeroBlank(thumbnail) || thumbnail.indexOf("/api") != 0) {
    article.thumbnail = ""
  } else {
    article.thumbnail = thumbnail
  }

  if (selectedCategory.value && selectedCategory.value.length > 0) {
    article.categories = selectedCategory.value.join('/')
  } else {
    article.categories = "默认分类"
  }

  // === 关键修改：标签处理逻辑 ===
  // 1. 去重
  const uniqueTags = [...new Set(dynamicTags.value)]
  // 2. 格式化每个标签：确保以#开头，去掉末尾空格，用空格连接
  const formattedTags = uniqueTags.map(tag => {
    let t = tag.trim()
    // 确保以#开头
    if (!t.startsWith('#')) {
      t = '#' + t
    }
    // 去掉末尾的空格（如果有）
    t = t.replace(/\s+$/, '')
    return t
  })
  // 3. 用空格连接，而不是逗号
  article.tags = formattedTags.join(' ')

  try {
    const response = await axios({
      method: 'post',
      url: '/api/article/publishArticle?type=' + type,
      data: article,
      timeout: 3000000
    })

    // 成功后的处理：先清理数据，再显示成功消息
    clearAllData()

    ElMessageBox.alert(response.data.msg, '结果', {
      confirmButtonText: '确定',
      callback: () => {
        // 确定后的回调：再次清理确保数据消失
        clearAllData()
        window.scrollTo(0, 0)
      }
    })
  } catch (error) {
    ElMessageBox.alert("系统错误！", '结果')
  } finally {
    isSubmitting.value = false
  }
}

const gotoArticleManage = inject("gotoArticleManage", () => {
  router.push('/')
})

let isShowCropper = ref(true)

function freshCropper() {
  isShowCropper.value = false
  nextTick(() => { isShowCropper.value = true; })
}
provide("freshCropper", freshCropper)

// === 清理所有数据 ===
function clearAllData() {
  // 清除定时器
  if (draftTimer) {
    clearTimeout(draftTimer)
    draftTimer = null
  }

  // 清除本地存储草稿
  localStorage.removeItem(DRAFT_KEY)

  // 清空表单数据
  clearFormData()
}

function clearFormData() {
  // 清空文章对象
  article.title = ""
  article.tags = ""
  article.content = ""
  article.categories = ""
  article.location = ""
  article.thumbnail = ""

  // 清空组件数据
  selectedCategory.value = []
  dynamicTags.value = []

  // 清空Cropper
  if (cropper1.value) {
    cropper1.value.clearData()
  }
}

// === 标签输入处理：添加#前缀和空格 ===
const handleTagInput = (tag) => {
  if (tag && !tag.startsWith('#')) {
    return `#${tag} `
  }
  return tag
}

// === 处理标签变化：确保格式正确 ===
const handleTagsChange = () => {
  dynamicTags.value = dynamicTags.value.map(tag => {
    let processedTag = tag.trim()

    // 确保以#开头
    if (!processedTag.startsWith('#')) {
      processedTag = `#${processedTag}`
    }

    // 确保末尾有空格（用于显示分隔）
    if (!processedTag.endsWith(' ')) {
      processedTag = `${processedTag} `
    }

    return processedTag
  })
}

onMounted(() => {
  loadCategoryTree()

  // 1. 处理分类回显
  if (route.query.categoryPath) {
    const pathStr = route.query.categoryPath
    selectedCategory.value = pathStr.split('/')
    article.categories = pathStr
  }

  // 2. 检查是否有草稿
  const draftStr = localStorage.getItem(DRAFT_KEY)
  if (draftStr) {
    const draft = JSON.parse(draftStr)
    if (draft.title || draft.content) {
      ElMessageBox.confirm(
        '检测到您上次有未发布的草稿，是否恢复？',
        '恢复草稿',
        {
          confirmButtonText: '恢复内容',
          cancelButtonText: '丢弃',
          type: 'info',
        }
      ).then(() => {
        // 恢复数据
        article.title = draft.title || ""
        article.content = draft.content || ""
        article.tags = draft.tags || ""
        article.location = draft.location || ""
        article.categories = draft.categories || ""
        article.thumbnail = draft.thumbnail || ""

        // 恢复分类选择框
        if (draft.selectedCategory && draft.selectedCategory.length > 0) {
          selectedCategory.value = draft.selectedCategory
        } else if (draft.categories) {
          selectedCategory.value = draft.categories.split('/')
        }

        // 恢复标签（处理空格分隔的标签字符串）
        if (draft.dynamicTags && draft.dynamicTags.length > 0) {
          dynamicTags.value = draft.dynamicTags
        } else if (draft.tags) {
          // 将空格分隔的标签字符串转换为数组（每个标签加空格）
          const tagsArray = draft.tags.split(' ').filter(item => item.trim() !== '')
          dynamicTags.value = tagsArray.map(tag => {
            if (!tag.startsWith('#')) {
              return `#${tag} `
            }
            if (!tag.endsWith(' ')) {
              return `${tag} `
            }
            return tag
          })
        }

        // 恢复缩略图
        if (article.thumbnail && article.thumbnail.indexOf("/api") == 0 && cropper1.value) {
          cropper1.value.setThumbnail(article.thumbnail)
        }

        ElMessage.success('草稿已恢复')
      }).catch(() => {
        localStorage.removeItem(DRAFT_KEY)
        ElMessage.info('已丢弃草稿')
      })
    }
  }

  // 3. 处理编辑模式
  if (store.articleId > 0) {
    type = "edit"
    header.value = "编辑文章"
    axios({
      method: 'post',
      url: '/api/article/selectById?id=' + store.articleId
    }).then((response) => {
      if (response.data.success) {
        let nowArticle = response.data.map.article
        article.id = nowArticle.id
        article.title = nowArticle.title
        article.tags = nowArticle.tags
        article.content = nowArticle.content
        article.thumbnail = nowArticle.thumbnail
        article.location = nowArticle.location || ""

        if (nowArticle.categories) {
          article.categories = nowArticle.categories
          selectedCategory.value = nowArticle.categories.split('/')
        }

        // 处理标签回显：将空格分隔的标签字符串转换为数组
        if (article.tags) {
          // 假设数据库存储格式为 "#tag1 #tag2"
          const tagsArray = article.tags.split(' ').filter(item => item.trim() !== '')
          dynamicTags.value = tagsArray.map(tag => {
            if (!tag.startsWith('#')) {
              return `#${tag} `
            }
            if (!tag.endsWith(' ')) {
              return `${tag} `
            }
            return tag
          })
        }

        if (!undefine(article.thumbnail) && !nullZeroBlank(article.thumbnail) && article.thumbnail.indexOf("/api") == 0) {
          cropper1.value.setThumbnail(article.thumbnail)
        }
      } else {
        ElMessageBox.alert(response.data.msg, '结果')
      }
      store.articleId = 0
    }).catch((error) => {
      ElMessageBox.alert("系统错误！", '结果')
      store.articleId = 0
    })
  }
})
</script>

<template>
  <el-row>
    <el-col :span="24">
      <h4>{{ header }}</h4>
    </el-col>
  </el-row>

  <el-row :gutter="20" style="margin-bottom: 20px;">
    <el-col :span="8">
      <el-input v-model="article.title" placeholder="请输入文章标题（必须）" clearable />
    </el-col>

    <el-col :span="8">
      <el-cascader v-model="selectedCategory" :options="categoryOptions"
        :props="{ checkStrictly: true, value: 'name', label: 'name' }" placeholder="请选择文章分类" clearable
        style="width: 100%;" />
    </el-col>

    <el-col :span="8">
      <el-input-tag v-model="dynamicTags" placeholder="输入标签后回车" aria-label="输入标签后回车" :max="5"
        :before-tag-add="handleTagInput" @change="handleTagsChange" />
    </el-col>
  </el-row>

  <el-row :gutter="20" style="margin-bottom: 20px;">
    <el-col :span="12">
      <el-input v-model="article.location" placeholder="发布地点 (可自动获取或手动选择)" clearable>
        <template #prepend>
          <el-icon>
            <MapLocation />
          </el-icon>
        </template>
        <template #append>
          <el-button :icon="Position" @click="autoLocate" title="自动定位" />
          <el-button :icon="MapLocation" @click="openMap" title="在地图上选择" />
        </template>
      </el-input>
    </el-col>
  </el-row>

  <el-row>
    <el-col :span="24">
      <div id="editor">
        <editor v-model="article.content" :init="init" :api-key="apiKey" />
      </div>
    </el-col>
  </el-row>
  <el-row>
    <el-col :span="24">
      <Cropper ref="cropper1" v-if="isShowCropper" />
    </el-col>
  </el-row>
  <el-row>
    <el-col :span="24">
      <div align="right">
        <el-button @click="gotoArticleManage">返回列表</el-button>
        <el-button type="primary" @click="publishArticle" :loading="isSubmitting">保存文章</el-button>
      </div>
    </el-col>
  </el-row>

  <el-dialog v-model="mapVisible" title="点击地图选择位置" width="600px" append-to-body>
    <div id="map-container" style="height: 400px; width: 100%;"></div>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="mapVisible = false">确定</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<style scoped>
#map-container {
  z-index: 1;
}
</style>/