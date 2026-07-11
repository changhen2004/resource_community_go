<template>
  <el-container class="resource-page">
    <el-main class="resource-main">
      <section class="resource-hero">
        <div>
          <p class="eyebrow">RESOURCE DISCOVERY</p>
          <h1>发现值得收藏的社区内容</h1>
          <p class="hero-copy">
            按标签、排序和关键词快速筛选资源，找到最新发布和最受欢迎的内容。
          </p>
        </div>
        <div class="hero-stats">
          <div class="stat-card">
            <span>当前页</span>
            <strong>{{ query.page }}</strong>
          </div>
          <div class="stat-card">
            <span>本页结果</span>
            <strong>{{ resources.length }}</strong>
          </div>
        </div>
      </section>

      <section class="toolbar-card">
        <div class="toolbar-top">
          <el-input
            v-model="draftKeyword"
            class="search-box"
            clearable
            placeholder="搜索标题关键词"
            @keyup.enter="applyFilters"
          >
            <template #append>
              <el-button @click="applyFilters">搜索</el-button>
            </template>
          </el-input>

          <el-segmented
            v-model="query.sort"
            :options="sortOptions"
            @change="onSortChange"
          />
        </div>

        <div class="toolbar-bottom">
          <div class="filter-group">
            <span class="filter-label">标签筛选</span>
            <el-tag
              :type="query.tag === '' ? 'primary' : 'info'"
              class="filter-tag"
              effect="light"
              @click="selectTag('')"
            >
              全部
            </el-tag>
            <el-tag
              v-for="tag in availableTags"
              :key="tag"
              :type="query.tag === tag ? 'primary' : 'info'"
              class="filter-tag"
              effect="light"
              @click="selectTag(tag)"
            >
              {{ tag }}
            </el-tag>
          </div>

          <el-select
            v-model="query.pageSize"
            class="page-size-select"
            placeholder="每页数量"
            @change="onPageSizeChange"
          >
            <el-option
              v-for="size in pageSizeOptions"
              :key="size"
              :label="`${size} 条 / 页`"
              :value="size"
            />
          </el-select>
        </div>
      </section>

      <section v-if="loading" class="state-panel skeleton-grid">
        <el-skeleton v-for="index in skeletonCount" :key="index" animated class="resource-card">
          <template #template>
            <el-skeleton-item variant="h3" style="width: 48%" />
            <el-skeleton-item variant="text" style="width: 92%" />
            <el-skeleton-item variant="text" style="width: 82%" />
            <div class="skeleton-footer">
              <el-skeleton-item variant="button" style="width: 96px; height: 32px" />
              <el-skeleton-item variant="text" style="width: 88px" />
            </div>
          </template>
        </el-skeleton>
      </section>

      <section v-else-if="errorMessage" class="state-panel">
        <el-result icon="warning" title="加载失败" :sub-title="errorMessage">
          <template #extra>
            <el-button type="primary" @click="fetchResources">重新加载</el-button>
          </template>
        </el-result>
      </section>

      <section v-else-if="!resources.length" class="state-panel">
        <el-empty description="当前筛选条件下暂无内容">
          <el-button @click="resetFilters">重置筛选</el-button>
        </el-empty>
      </section>

      <section v-else class="resource-grid">
        <el-card
          v-for="resource in resources"
          :key="resource.id"
          class="resource-card"
          shadow="hover"
        >
          <div class="resource-card__meta">
            <span class="resource-status">{{ resource.status || 'published' }}</span>
            <span class="resource-points">
              {{ resource.isFree ? '免费' : `${resource.requiredPoints || 0} 积分` }}
            </span>
          </div>

          <h2>{{ resource.title }}</h2>
          <p class="resource-preview">{{ resource.preview }}</p>

          <div class="resource-tags" v-if="resource.tags?.length">
            <el-tag
              v-for="tag in resource.tags"
              :key="tag"
              size="small"
              effect="plain"
              @click="selectTag(tag)"
            >
              {{ tag }}
            </el-tag>
          </div>

          <div class="resource-footer">
            <div class="resource-stats">
              <span>热度 {{ resource.likeCount ?? 0 }}</span>
              <span>浏览 {{ resource.viewCount ?? 0 }}</span>
              <span>评论 {{ resource.commentCount ?? 0 }}</span>
            </div>
            <el-button text type="primary" @click="viewDetail(resource.id)">阅读更多</el-button>
          </div>
        </el-card>
      </section>

      <section class="pagination-bar">
        <el-button :disabled="query.page === 1 || loading" @click="changePage(query.page - 1)">
          上一页
        </el-button>
        <div class="page-indicator">
          <span>第 {{ query.page }} 页</span>
          <small v-if="resources.length < query.pageSize">已到达最后一页</small>
        </div>
        <el-button
          :disabled="loading || resources.length < query.pageSize"
          @click="changePage(query.page + 1)"
        >
          下一页
        </el-button>
      </section>
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import { listArticles } from '../api/article';
import type { ResourceSummary } from '../types/resource';

const router = useRouter();
const resources = ref<ResourceSummary[]>([]);
const loading = ref(false);
const errorMessage = ref('');
const draftKeyword = ref('');
const query = reactive({
  page: 1,
  pageSize: 6,
  sort: 'latest' as 'latest' | 'hot',
  keyword: '',
  tag: '',
});

const sortOptions = [
  { label: '最新发布', value: 'latest' },
  { label: '热度优先', value: 'hot' },
];

const pageSizeOptions = [6, 12, 18];

const availableTags = computed(() => {
  const tagSet = new Set<string>();
  resources.value.forEach((resource) => {
    resource.tags?.forEach((tag) => {
      if (tag.trim()) {
        tagSet.add(tag);
      }
    });
  });
  return Array.from(tagSet);
});

const skeletonCount = computed(() => query.pageSize);

const fetchResources = async () => {
  loading.value = true;
  errorMessage.value = '';

  try {
    resources.value = await listArticles({
      page: query.page,
      pageSize: query.pageSize,
      sort: query.sort,
      keyword: query.keyword || undefined,
      tag: query.tag || undefined,
    });
  } catch (error) {
    console.error('Failed to load resources:', error);
    errorMessage.value = '资源列表加载失败，请稍后重试。';
  } finally {
    loading.value = false;
  }
};

const applyFilters = async () => {
  query.page = 1;
  query.keyword = draftKeyword.value.trim();
  await fetchResources();
};

const resetFilters = async () => {
  draftKeyword.value = '';
  query.page = 1;
  query.pageSize = 6;
  query.sort = 'latest';
  query.keyword = '';
  query.tag = '';
  await fetchResources();
};

const selectTag = async (tag: string) => {
  query.tag = tag;
  query.page = 1;
  await fetchResources();
};

const onSortChange = async () => {
  query.page = 1;
  await fetchResources();
};

const onPageSizeChange = async () => {
  query.page = 1;
  await fetchResources();
};

const changePage = async (page: number) => {
  if (page < 1 || loading.value) {
    return;
  }
  query.page = page;
  await fetchResources();
};

const viewDetail = (id: number) => {
  router.push({ name: 'ResourceDetail', params: { id } });
};

onMounted(fetchResources);
</script>

<style scoped>
.resource-page {
  min-height: 100%;
  background:
    radial-gradient(circle at top left, rgba(255, 227, 191, 0.95), transparent 24%),
    radial-gradient(circle at top right, rgba(255, 171, 145, 0.28), transparent 18%),
    linear-gradient(180deg, #f7f0e7 0%, #efe2d0 100%);
}

.resource-main {
  max-width: 1180px;
  margin: 0 auto;
  padding: 32px 24px 56px;
}

.resource-hero {
  display: grid;
  grid-template-columns: minmax(0, 1.8fr) minmax(280px, 0.9fr);
  gap: 24px;
  align-items: end;
  margin-bottom: 24px;
}

.eyebrow {
  margin: 0 0 12px;
  letter-spacing: 0.22em;
  font-size: 12px;
  font-weight: 700;
  color: #9d4f2e;
}

.resource-hero h1 {
  margin: 0;
  font-size: clamp(34px, 5vw, 52px);
  line-height: 1.04;
  color: #2e1a12;
}

.hero-copy {
  max-width: 680px;
  margin: 16px 0 0;
  font-size: 16px;
  line-height: 1.75;
  color: #6d5548;
}

.hero-stats {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.stat-card,
.toolbar-card,
.resource-card,
.state-panel {
  border: 1px solid rgba(93, 52, 34, 0.08);
  border-radius: 24px;
  background: rgba(255, 251, 247, 0.85);
  box-shadow: 0 16px 40px rgba(84, 53, 37, 0.09);
  backdrop-filter: blur(18px);
}

.stat-card {
  padding: 18px 20px;
}

.stat-card span {
  display: block;
  color: #8d725d;
  font-size: 12px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.stat-card strong {
  display: block;
  margin-top: 10px;
  font-size: 28px;
  color: #2e1a12;
}

.toolbar-card {
  margin-bottom: 24px;
  padding: 20px;
}

.toolbar-top,
.toolbar-bottom {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: center;
}

.toolbar-bottom {
  margin-top: 18px;
  flex-wrap: wrap;
}

.search-box {
  max-width: 560px;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.filter-label {
  color: #7a6153;
  font-size: 13px;
  font-weight: 600;
}

.filter-tag {
  cursor: pointer;
}

.page-size-select {
  width: 140px;
}

.state-panel {
  min-height: 280px;
  padding: 28px;
}

.skeleton-grid,
.resource-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20px;
}

.resource-card {
  padding: 22px;
  transition:
    transform 0.18s ease,
    box-shadow 0.18s ease;
}

.resource-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 20px 48px rgba(84, 53, 37, 0.14);
}

.resource-card h2 {
  margin: 14px 0 10px;
  font-size: 24px;
  line-height: 1.2;
  color: #2e1a12;
}

.resource-card__meta,
.resource-footer,
.resource-stats {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
}

.resource-card__meta {
  color: #8d725d;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.resource-preview {
  margin: 0;
  color: #6d5548;
  line-height: 1.7;
  min-height: 76px;
}

.resource-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 16px;
}

.resource-footer {
  margin-top: 18px;
  padding-top: 16px;
  border-top: 1px solid rgba(93, 52, 34, 0.08);
}

.resource-stats {
  color: #7a6153;
  font-size: 13px;
}

.pagination-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  margin-top: 24px;
  padding: 0 4px;
}

.page-indicator {
  display: flex;
  flex-direction: column;
  align-items: center;
  color: #654d40;
}

.page-indicator small {
  margin-top: 4px;
  color: #9b816f;
}

.skeleton-footer {
  display: flex;
  justify-content: space-between;
  margin-top: 18px;
}

@media (max-width: 900px) {
  .resource-main {
    padding: 24px 16px 40px;
  }

  .resource-hero,
  .skeleton-grid,
  .resource-grid {
    grid-template-columns: 1fr;
  }

  .toolbar-top,
  .toolbar-bottom,
  .pagination-bar {
    flex-direction: column;
    align-items: stretch;
  }

  .page-indicator {
    order: -1;
  }
}
</style>
