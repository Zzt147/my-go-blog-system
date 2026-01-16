<script setup>
import { ref, onMounted, inject } from 'vue'
import * as echarts from 'echarts'
import { DataAnalysis, View, ChatLineRound, Trophy } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const axios = inject('axios')

// 基础统计数据
const stats = ref({
  totalArticles: 0,
  totalHits: 0,
  totalComments: 0
})

// 图表 DOM 引用
const pieChartRef = ref(null)
const barChartRef = ref(null)

// ECharts 实例
let pieChart = null
let barChart = null

// === 1. 饼图逻辑 (新代码的无限层级下钻功能) ===
let fullTreeData = [] // 完整的树形结构
let drillStack = [] // 下钻历史栈

// 将扁平路径 (Java/Spring/Core) 转换为树形结构
function buildTree(list) {
  const root = []
  list.forEach(item => {
    // 兼容 categories 为 null 的情况
    if (!item.name) item.name = "未分类"
    const parts = item.name.split('/')
    let currentNode = root

    parts.forEach((part, index) => {
      let existingNode = currentNode.find(n => n.name === part)
      if (!existingNode) {
        existingNode = { name: part, value: 0, children: [] }
        currentNode.push(existingNode)
      }
      // 在路径经过的所有节点都累加，保证父节点数值 = 子节点之和
      existingNode.value += item.value
      currentNode = existingNode.children
    })
  })
  return root
}

function initPieChart(flatData) {
  if (pieChart) pieChart.dispose()
  pieChart = echarts.init(pieChartRef.value)

  fullTreeData = buildTree(flatData)
  renderPie(fullTreeData, '创作分类分布')

  // 点击下钻
  pieChart.on('click', params => {
    if (params.componentType !== 'series') return
    const clickedNode = params.data
    if (clickedNode.children && clickedNode.children.length > 0) {
      drillStack.push({ data: fullTreeData, title: '创作分类分布' })
      fullTreeData = clickedNode.children
      renderPie(fullTreeData, clickedNode.name)
    } else {
      ElMessage.info('已到达最底层分类')
    }
  })
}

// 返回上一级
function goBack() {
  if (drillStack.length > 0) {
    const prev = drillStack.pop()
    fullTreeData = prev.data
    renderPie(fullTreeData, prev.title)
  }
}

function renderPie(data, title) {
  const isDeep = drillStack.length > 0
  const option = {
    title: {
      text: title,
      left: 'center',
      textStyle: { fontSize: 16 }
    },
    tooltip: { trigger: 'item' },
    // 动态返回按钮
    graphic: isDeep ? [{
      type: 'text',
      left: '10%',
      top: '10%',
      style: {
        text: '⬅ 返回上级',
        fill: '#409EFF',
        fontSize: 14,
        fontWeight: 'bold'
      },
      onclick: goBack
    }] : [],
    series: [{
      name: '文章数量',
      type: 'pie',
      radius: isDeep ? ['30%', '60%'] : '50%',
      data: data,
      emphasis: {
        itemStyle: {
          shadowBlur: 10,
          shadowOffsetX: 0,
          shadowColor: 'rgba(0, 0, 0, 0.5)'
        }
      },
      label: {
        show: true,
        formatter: '{b}: {c} ({d}%)'
      }
    }]
  }
  pieChart.setOption(option, { notMerge: true })
}

// === 2. 柱状图逻辑 (优先使用新代码的 tagObjs，如果不存在则使用旧代码逻辑) ===
function initBarChart(tagData) {
  if (barChart) barChart.dispose()
  barChart = echarts.init(barChartRef.value)

  // 判断数据结构：新代码返回的是 tagObjs 数组，旧代码返回的是 tags 字符串数组
  let topTags = []

  if (Array.isArray(tagData)) {
    // 情况1: 新代码的 tagObjs 数据结构 [{name: 'Java', count: 10}, ...]
    if (tagData.length > 0 && tagData[0].name && tagData[0].count !== undefined) {
      topTags = tagData
        .sort((a, b) => b.count - a.count)
        .slice(0, 15)
        .map(t => ({ name: t.name, value: t.count }))
    }
    // 情况2: 旧代码的 tags 字符串数组 ["Java,Spring", "Docker", ...]
    else {
      const tagCounts = {}
      tagData.forEach(tagStr => {
        if (!tagStr) return
        const tags = tagStr.replace(/，/g, ',').split(',')
        tags.forEach(t => {
          const cleanTag = t.trim()
          if (cleanTag) {
            tagCounts[cleanTag] = (tagCounts[cleanTag] || 0) + 1
          }
        })
      })

      topTags = Object.keys(tagCounts)
        .map(key => ({ name: key, value: tagCounts[key] }))
        .sort((a, b) => b.value - a.value)
        .slice(0, 15)
    }
  }

  // 渲染图表
  const option = {
    title: {
      text: '热门标签 TOP15',
      left: 'center'
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '10%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: topTags.map(t => t.name),
      axisLabel: {
        interval: 0,
        rotate: 30,
        fontSize: 12
      }
    },
    yAxis: {
      type: 'value',
      name: '引用次数'
    },
    series: [
      {
        name: '文章数',
        type: 'bar',
        data: topTags.map(t => t.value),
        itemStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: '#83bff6' },
            { offset: 0.5, color: '#188df0' },
            { offset: 1, color: '#188df0' }
          ])
        },
        barWidth: '50%',
        label: {
          show: true,
          position: 'top'
        }
      }
    ]
  }
  barChart.setOption(option)
}

onMounted(() => {
  // 1. 获取仪表盘基础数据 (包含分类统计)
  axios.get('/api/statistic/dashboard').then(res => {
    if (res.data.success) {
      const map = res.data.map
      stats.value.totalArticles = map.totalArticles
      stats.value.totalHits = map.totalHits
      stats.value.totalComments = map.totalComments

      // 初始化饼图
      initPieChart(map.categoryStats || [])
    }
  })

  // 2. 获取所有标签数据
  // 优先使用新接口，它返回 tagObjs；如果不存在则回退到旧接口逻辑
  axios.get('/api/article/getAllTags').then(res => {
    if (res.data.success) {
      const map = res.data.map
      // 优先使用新代码的 tagObjs 字段
      const tagData = map.tagObjs || map.tags || []
      initBarChart(tagData)
    }
  }).catch(() => {
    // 如果失败，使用旧接口作为备用
    axios.get('/api/article/tags').then(res => {
      if (res.data.success) {
        initBarChart(res.data.map.tags || [])
      }
    })
  })

  // 窗口大小改变时自动重绘
  window.addEventListener('resize', () => {
    pieChart && pieChart.resize()
    barChart && barChart.resize()
  })
})
</script>

<template>
  <div class="dashboard-container">
    <!-- 统计卡片 (采用旧代码的界面设计) -->
    <el-row :gutter="20" class="stat-cards">
      <el-col :span="8">
        <el-card shadow="hover" class="card-item">
          <div class="card-content">
            <el-icon :size="48" color="#409EFF">
              <DataAnalysis />
            </el-icon>
            <div class="text-info">
              <div class="label">总文章数</div>
              <div class="value">{{ stats.totalArticles }}</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card shadow="hover" class="card-item">
          <div class="card-content">
            <el-icon :size="48" color="#67C23A">
              <View />
            </el-icon>
            <div class="text-info">
              <div class="label">总阅读量</div>
              <div class="value">{{ stats.totalHits }}</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card shadow="hover" class="card-item">
          <div class="card-content">
            <el-icon :size="48" color="#E6A23C">
              <ChatLineRound />
            </el-icon>
            <div class="text-info">
              <div class="label">总评论数</div>
              <div class="value">{{ stats.totalComments }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 图表区域 (采用旧代码的响应式布局) -->
    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :xs="24" :sm="12">
        <el-card shadow="hover">
          <div ref="pieChartRef" style="width: 100%; height: 400px;"></div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12">
        <el-card shadow="hover">
          <div ref="barChartRef" style="width: 100%; height: 400px;"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 底部励志语 (采用旧代码的设计) -->
    <el-row style="margin-top: 20px;">
      <el-col :span="24">
        <el-card shadow="never" style="text-align: center; background: #fdf6ec; color: #e6a23c;">
          <h3>
            <el-icon>
              <Trophy />
            </el-icon>
            坚持写作是一种修行，继续加油！
          </h3>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.dashboard-container {
  padding: 10px;
}

.stat-cards {
  margin-bottom: 20px;
}

.card-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 20px;
}

.text-info {
  text-align: right;
}

.label {
  color: #909399;
  font-size: 14px;
}

.value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
  margin-top: 5px;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .card-content {
    flex-direction: column;
    text-align: center;
    padding: 15px;
  }

  .text-info {
    text-align: center;
    margin-top: 10px;
  }

  .value {
    font-size: 20px;
  }
}
</style>