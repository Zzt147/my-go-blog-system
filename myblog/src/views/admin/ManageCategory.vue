<script setup>
import { ref, reactive, inject, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus, Edit, Delete, Folder, FolderOpened, Rank,
  Document, ArrowRight, DocumentAdd, Warning
} from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'

const axios = inject('axios')
const router = useRouter()
const treeData = ref([])
const loading = ref(false)

// 右侧内容数据
const rightLoading = ref(false)
const currentCategory = ref(null) // 当前选中的分类节点信息
const currentPath = ref('')
const folderList = ref([]) // 子分类列表
const articleList = ref([]) // 文章列表

// === 1. 获取树形数据 ===
function loadTree() {
  loading.value = true
  axios.get('/api/category/getTree').then(res => {
    if (res.data.success) {
      treeData.value = res.data.map.data || []
    }
    loading.value = false
  })
}

onMounted(() => {
  loadTree()
})

// === 2. 树节点点击事件 (加载右侧内容) ===
function handleNodeClick(data) {
  currentCategory.value = data
  loadRightContent(data.id)
}

// 加载右侧资源
function loadRightContent(id) {
  rightLoading.value = true
  axios.get('/api/category/getResources?id=' + id).then(res => {
    if (res.data.success) {
      const data = res.data.map.data
      folderList.value = data.folders || []
      articleList.value = data.articles || []
      currentPath.value = data.currentPath
    }
    rightLoading.value = false
  })
}

// === 3. 交互逻辑 ===

// 双击右侧文件夹 -> 进入该层级 (联动左侧树)
function enterFolder(folder) {
  // 这里需要让 el-tree 选中对应的节点，比较复杂，简单做法是刷新右侧
  // 如果想联动左侧，需要 ref 到 tree 并调用 setCurrentKey
  // 这里简化处理：更新当前选中状态并加载
  currentCategory.value = folder
  loadRightContent(folder.id)

  // 尝试展开左侧树对应的节点 (需要 el-tree ref)
  if (treeRef.value) {
    treeRef.value.setCurrentKey(folder.id)
    treeRef.value.expandNode(folder.id) // 需要 data 对象，或者 key
  }
}

// 点击右侧文章 -> 跳转到编辑页面
function editArticle(articleId) {
  // 假设你的编辑页面路由是 /admin/editArticle
  // 或者是 PublishArticle 页面带参数
  // 你的 PublishArticle 似乎是通过 store 传参的，这里适配你的逻辑
  // 假设有个 gotoEditArticle 方法或者直接 router push
  // store.articleId = articleId (如果使用 store)
  // 这里演示直接跳转
  router.push({ name: 'publishArticle', query: { id: articleId } })
  // 注意：你需要确保你的路由配置支持 query 参数，或者使用 store
}

// 【新增】在当前文件夹下创建文章
function createArticleInFolder() {
  if (!currentCategory.value) return
  // 跳转到发布页，并携带当前路径参数
  // currentPath.value 的格式是 "技术/后端/Java"
  router.push({
    name: 'publishArticle',
    query: { categoryPath: currentPath.value }
  })
}

// 增删改弹窗逻辑 (保留之前的)
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref(null)
const formData = reactive({ id: null, name: '', parentId: 0, sort: 0 })
const rules = { name: [{ required: true, message: '请输入名称', trigger: 'blur' }] }

function openDialog(type, nodeData) {
  dialogVisible.value = true
  formData.id = null; formData.name = ''; formData.sort = 0

  if (type === 'add_root') {
    dialogTitle.value = '新增顶级分类'; formData.parentId = 0
  } else if (type === 'add_sub') {
    dialogTitle.value = `在 [${nodeData.name}] 下新增`; formData.parentId = nodeData.id
  } else if (type === 'edit') {
    dialogTitle.value = '编辑分类'
    formData.id = nodeData.id; formData.name = nodeData.name
    formData.parentId = nodeData.parentId; formData.sort = nodeData.sort
  }
}

function submitForm() {
  formRef.value.validate((valid) => {
    if (valid) {
      const url = formData.id ? '/api/category/update' : '/api/category/add'
      axios.post(url, formData).then(res => {
        if (res.data.success) {
          ElMessage.success(res.data.msg)
          dialogVisible.value = false
          loadTree() // 刷新树
          if (currentCategory.value) loadRightContent(currentCategory.value.id) // 刷新右侧
        } else { ElMessage.error(res.data.msg) }
      })
    }
  })
}

function remove(data) {
  ElMessageBox.confirm(`确定删除 [${data.name}] ?`, '警告', { type: 'warning' }).then(() => {
    axios.post('/api/category/delete?id=' + data.id).then(res => {
      if (res.data.success) {
        ElMessage.success('删除成功')
        loadTree()
        if (currentCategory.value && currentCategory.value.id === data.parentId) {
          loadRightContent(data.parentId)
        }
      } else { ElMessage.error(res.data.msg) }
    })
  })
}

// 拖拽排序 (保留)
const handleDrop = (draggingNode, dropNode, dropType, ev) => {
  let newParentId = dropType === 'inner' ? dropNode.data.id : dropNode.data.parentId
  const parentNode = dropType === 'inner' ? dropNode : dropNode.parent
  const siblings = parentNode.level === 0 ? treeData.value : parentNode.data.children || []
  const updateList = siblings.map((item, index) => ({
    id: item.id, parentId: newParentId, name: item.name, sort: index + 1
  }))
  axios.post('/api/category/updateBatch', updateList).then(res => {
    if (res.data.success) ElMessage.success('排序已更新')
    else loadTree()
  })
}

const defaultProps = { children: 'children', label: 'name' }
const treeRef = ref(null)

// 注入 store 以便跳转编辑 (假设你使用了 pinia)
import { useStore } from '@/stores/my'
const store = useStore()
function gotoEdit(article) {
  store.articleId = article.id
  // 切换组件或路由
  // 假设 ManageArticle 里有 changeView 逻辑，这里如果是独立路由则用 router.push
  // router.push('/admin/publishArticle') 
  // 由于你的 PublishArticle 是组件切换模式，这里可能无法直接跳，建议让用户去文章管理页编辑
  ElMessage.info(`请前往"文章管理"页面编辑: ${article.title}`)
}

const deleteDialogVisible = ref(false)
const deleteTarget = ref(null) // 当前要删除的分类对象

// 点击删除按钮触发
function handleDeleteClick(folder) {
  deleteTarget.value = folder
  deleteDialogVisible.value = true
}

// 执行删除
function confirmDelete(mode) {
  if (!deleteTarget.value) return

  const modeText = mode === 2 ? "删除分类及其下所有文章" : "仅删除分类（文章移至父级）";

  // 再次确认防止手误
  // ElMessageBox.confirm(`确认执行操作：${modeText} ?`, '最终确认', { 
  //   type: mode === 2 ? 'error' : 'warning',
  //   confirmButtonText: '确定执行',
  //   cancelButtonText: '取消'
  // }).then(() => {
  // 发送请求
  const url = `/api/category/delete?id=${deleteTarget.value.id}&mode=${mode}`;
  axios.post(url).then(res => {
    if (res.data.success) {
      ElMessage.success("操作成功");
      deleteDialogVisible.value = false;
      loadTree(); // 刷新左侧树
      // 如果当前正好在看这个被删的分类，回退到父级或根目录
      if (currentCategory.value && currentCategory.value.id === deleteTarget.value.id) {
        currentCategory.value = null;
        folderList.value = [];
        articleList.value = [];
      } else if (currentCategory.value) {
        // 否则刷新当前视图
        loadRightContent(currentCategory.value.id);
      }
    } else {
      ElMessage.error(res.data.msg);
    }
  });
  // }).catch(() => {});
}

</script>

<template>
  <div class="category-manage">
    <el-card shadow="never" class="main-card">
      <el-container style="height: 600px;">

        <el-aside width="280px" class="left-panel">
          <div class="panel-header">
            <span>📚 目录结构</span>
            <el-button link type="primary" :icon="Plus" @click="openDialog('add_root')">新建根目录</el-button>
          </div>
          <el-scrollbar>
            <el-tree ref="treeRef" v-loading="loading" :data="treeData" node-key="id" default-expand-all draggable
              :expand-on-click-node="false" :props="defaultProps" highlight-current @node-click="handleNodeClick"
              @node-drop="handleDrop">
              <template #default="{ data }">
                <div class="custom-tree-node">
                  <span class="node-label">
                    <el-icon class="icon-folder">
                      <Folder />
                    </el-icon>
                    {{ data.name }}
                  </span>
                  <span class="node-actions">
                    <el-icon @click.stop="openDialog('add_sub', data)" title="新增子级">
                      <Plus />
                    </el-icon>
                    <el-icon @click.stop="openDialog('edit', data)" title="编辑">
                      <Edit />
                    </el-icon>
                  </span>
                </div>
              </template>
            </el-tree>
          </el-scrollbar>
        </el-aside>

        <el-main class="right-panel" v-loading="rightLoading">
          <div class="content-header" v-if="currentCategory">
            <div class="path-info">
              <el-icon>
                <FolderOpened />
              </el-icon>
              <span class="path-text">{{ currentPath || currentCategory.name }}</span>
            </div>
            <div class="header-actions">
              <el-button type="success" size="small" :icon="DocumentAdd" @click="createArticleInFolder">新建文章</el-button>
              <el-button type="primary" size="small" :icon="Plus"
                @click="openDialog('add_sub', currentCategory)">新建文件夹</el-button>
            </div>
          </div>
          <el-empty v-else description="请选择左侧分类查看内容" />

          <div v-if="currentCategory" class="resource-list">

            <div v-for="folder in folderList" :key="'folder-' + folder.id" class="resource-item folder-item"
              @dblclick="enterFolder(folder)">

              <div class="item-icon folder-icon"><el-icon>
                  <Folder />
                </el-icon></div>
              <div class="item-info">
                <div class="item-name">{{ folder.name }}</div>
                <div class="item-meta">子分类</div>
              </div>

              <div class="item-actions">
                <el-button link :icon="Edit" @click="openDialog('edit', folder)">重命名</el-button>
                <el-button link type="danger" :icon="Delete" @click="handleDeleteClick(folder)"></el-button>
              </div>
            </div>

            <div v-for="article in articleList" :key="'article-' + article.id" class="resource-item article-item">
              <div class="item-icon article-icon"><el-icon>
                  <Document />
                </el-icon></div>
              <div class="item-info">
                <div class="item-name" :title="article.title">{{ article.title }}</div>
                <div class="item-meta">{{ article.created }}</div>
              </div>
              <div class="item-actions">
                <el-tag size="small">文章</el-tag>
              </div>
            </div>

            <div v-if="folderList.length === 0 && articleList.length === 0" class="empty-tip">
              此文件夹为空
            </div>

          </div>
        </el-main>

      </el-container>
    </el-card>


    <el-dialog v-model="deleteDialogVisible" title="删除分类确认" width="450px" align-center>
      <div style="display: flex; gap: 15px; margin-bottom: 20px;">
        <el-icon :size="40" color="#F56C6C">
          <Warning />
        </el-icon>
        <div>
          <p style="font-size: 16px; font-weight: bold; margin: 0 0 10px 0;">
            您正在删除分类：[{{ deleteTarget?.name }}]
          </p>
          <p style="color: #666; margin: 0;">请选择一种删除方式：</p>
        </div>
      </div>

      <div class="delete-options">
        <div class="delete-option-item" @click="confirmDelete(1)">
          <div class="opt-title">1. 仅删除此分类</div>
          <div class="opt-desc">保留名下文章，文章将自动移动到父级分类。</div>
          <el-button type="primary" plain size="small" style="margin-top:8px;">执行此操作</el-button>
        </div>

        <div class="delete-option-item danger" @click="confirmDelete(2)">
          <div class="opt-title">⚠️ 2. 删除分类及文章</div>
          <div class="opt-desc">高危操作！该分类下的所有文章都将被永久删除。</div>
          <el-button type="danger" size="small" style="margin-top:8px;">确认销毁</el-button>
        </div>
      </div>
    </el-dialog>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="400px">
      <el-form :model="formData" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="名称" prop="name"><el-input v-model="formData.name" /></el-form-item>
        <el-form-item label="排序" prop="sort"><el-input-number v-model="formData.sort" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.category-manage {
  padding: 20px;
}

.main-card {
  padding: 0;
}

.el-container {
  border: 1px solid #f0f2f5;
}

/* 左侧样式 */
.left-panel {
  border-right: 1px solid #eee;
  background: #fafafa;
  display: flex;
  flex-direction: column;
}

.panel-header {
  padding: 15px;
  border-bottom: 1px solid #eee;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: bold;
}

.custom-tree-node {
  display: flex;
  flex: 1;
  align-items: center;
  justify-content: space-between;
  font-size: 14px;
  padding-right: 10px;
}

.node-label {
  display: flex;
  align-items: center;
  gap: 6px;
}

.node-actions {
  display: none;
  gap: 8px;
  color: #666;
}

.custom-tree-node:hover .node-actions {
  display: flex;
}

.node-actions .el-icon:hover {
  color: #409EFF;
  cursor: pointer;
}

/* 右侧样式 */
.right-panel {
  background: #fff;
  padding: 0;
  display: flex;
  flex-direction: column;
}

.content-header {
  padding: 15px 20px;
  border-bottom: 1px solid #eee;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
}

.path-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  color: #303133;
}

.path-text {
  font-weight: bold;
}

.resource-list {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.resource-item {
  display: flex;
  align-items: center;
  padding: 12px;
  border: 1px solid #ebeef5;
  border-radius: 6px;
  transition: all 0.2s;
  cursor: default;
}

.resource-item:hover {
  border-color: #409EFF;
  background: #ecf5ff;
}

.folder-item {
  background: #fdfdfd;
}

/* 文件夹背景微白 */

.item-icon {
  width: 40px;
  height: 40px;
  display: flex;
  justify-content: center;
  align-items: center;
  font-size: 24px;
  border-radius: 4px;
  margin-right: 15px;
}

.folder-icon {
  color: #E6A23C;
  background: #fcf6ec;
}

.article-icon {
  color: #409EFF;
  background: #ecf5ff;
}

.item-info {
  flex: 1;
}

.item-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
}

.item-meta {
  font-size: 12px;
  color: #909399;
}

.item-actions {
  opacity: 0;
  transition: opacity 0.2s;
}

.resource-item:hover .item-actions {
  opacity: 1;
}

.empty-tip {
  text-align: center;
  color: #999;
  padding: 40px;
}

/* 【新增】删除弹窗样式 */
.delete-options {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.delete-option-item {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 15px;
  cursor: pointer;
  transition: all 0.3s;
}

.delete-option-item:hover {
  border-color: #409EFF;
  background-color: #f0f9eb;
}

.delete-option-item.danger:hover {
  border-color: #F56C6C;
  background-color: #fef0f0;
}

.opt-title {
  font-weight: bold;
  font-size: 14px;
  margin-bottom: 5px;
}

.opt-desc {
  font-size: 12px;
  color: #909399;
}
</style>