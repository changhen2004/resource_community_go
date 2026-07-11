<template>
  <el-container class="detail-page">
    <el-main class="detail-main">
      <section v-if="loading" class="detail-skeleton">
        <el-skeleton animated>
          <template #template>
            <el-skeleton-item variant="image" style="width: 100%; height: 320px" />
            <div class="skeleton-stack">
              <el-skeleton-item variant="h1" style="width: 54%" />
              <el-skeleton-item variant="text" style="width: 92%" />
              <el-skeleton-item variant="text" style="width: 88%" />
              <el-skeleton-item variant="text" style="width: 76%" />
            </div>
          </template>
        </el-skeleton>
      </section>

      <section v-else-if="errorMessage" class="state-panel">
        <el-result icon="warning" title="资源加载失败" :sub-title="errorMessage">
          <template #extra>
            <el-button type="primary" @click="fetchPageData">重新加载</el-button>
          </template>
        </el-result>
      </section>

      <section v-else-if="resource" class="detail-layout">
        <article class="detail-content">
          <div class="detail-hero">
            <img
              v-if="resource.coverUrl"
              :src="resource.coverUrl"
              :alt="resource.title"
              class="detail-cover"
            />
            <div v-else class="detail-cover detail-cover--placeholder">
              <span>{{ resource.title.slice(0, 1) }}</span>
            </div>

            <div class="hero-overlay">
              <div class="hero-meta">
                <el-tag effect="dark" type="warning">{{ resource.status || 'published' }}</el-tag>
                <el-tag effect="dark" :type="resource.isFree ? 'success' : 'danger'">
                  {{ resource.isFree ? '免费资源' : `${resource.requiredPoints || 0} 积分解锁` }}
                </el-tag>
              </div>

              <h1>{{ resource.title }}</h1>
              <p class="hero-preview">{{ resource.preview }}</p>

              <div class="hero-tags" v-if="resource.tags?.length">
                <el-tag
                  v-for="tag in resource.tags"
                  :key="tag"
                  effect="plain"
                  type="info"
                  class="hero-tag"
                >
                  {{ tag }}
                </el-tag>
              </div>
            </div>
          </div>

          <div class="author-strip">
            <div class="author-card">
              <div class="author-avatar">
                {{ resource.author?.username?.slice(0, 1) || 'U' }}
              </div>
              <div>
                <p class="author-label">作者</p>
                <h3>{{ resource.author?.username || '匿名作者' }}</h3>
              </div>
            </div>

            <div class="author-extra">
              <span>作者 ID #{{ resource.author?.id || resource.authorId }}</span>
              <span>资源类型 {{ resource.isFree ? '公开可读' : '积分门槛' }}</span>
            </div>
          </div>

          <div class="content-body">
            <p class="body-copy">{{ resource.content }}</p>

            <div v-if="resource.contentImages?.length" class="content-gallery">
              <img
                v-for="(image, index) in resource.contentImages"
                :key="`${image}-${index}`"
                :src="image"
                :alt="`${resource.title} 配图 ${index + 1}`"
              />
            </div>
          </div>

          <section class="comment-panel">
            <div class="panel-head">
              <div>
                <p class="panel-kicker">Comments</p>
                <h2>评论区</h2>
              </div>
              <el-button text :loading="commentLoading" @click="fetchComments">刷新</el-button>
            </div>

            <div v-if="authStore.isAuthenticated" class="comment-editor">
              <el-input
                v-model="commentForm"
                type="textarea"
                :rows="4"
                maxlength="1000"
                show-word-limit
                placeholder="写下你的评论，分享你对这个资源的看法。"
              />
              <div class="comment-editor__actions">
                <span class="comment-editor__hint">登录用户可以参与讨论并推动内容热度上升。</span>
                <el-button
                  type="primary"
                  :loading="commentSubmitting"
                  :disabled="!commentForm.trim()"
                  @click="handleCreateComment"
                >
                  发布评论
                </el-button>
              </div>
            </div>
            <el-alert
              v-else
              type="info"
              :closable="false"
              show-icon
              title="登录后可以发表评论并参与互动"
            />

            <div v-if="commentLoading" class="comment-loading">
              <el-skeleton :rows="3" animated />
            </div>
            <el-result
              v-else-if="commentErrorMessage"
              icon="warning"
              title="评论加载失败"
              :sub-title="commentErrorMessage"
            >
              <template #extra>
                <el-button @click="fetchComments">重试</el-button>
              </template>
            </el-result>
            <el-empty v-else-if="!comments.length" description="还没有评论，来发表第一条观点" />
            <div v-else class="comment-list">
              <article v-for="comment in comments" :key="comment.id" class="comment-item">
                <div class="comment-item__head">
                  <div>
                    <strong>{{ comment.author.username }}</strong>
                    <span class="comment-item__time">{{ formatDate(comment.createdAt) }}</span>
                  </div>
                  <el-button
                    v-if="canDeleteComment(comment.userId)"
                    text
                    type="danger"
                    :loading="deletingCommentId === comment.id"
                    @click="handleDeleteComment(comment.id)"
                  >
                    删除
                  </el-button>
                </div>
                <p class="comment-item__content">{{ comment.content }}</p>
              </article>
            </div>
          </section>
        </article>

        <aside class="detail-sidebar">
          <section class="panel panel--sticky">
            <h2>访问规则</h2>
            <p class="gate-status" :class="gateState.className">
              {{ gateState.label }}
            </p>
            <p class="gate-copy">{{ gateState.description }}</p>

            <div class="points-card">
              <span>当前积分</span>
              <strong>{{ authStore.pointsBalance }}</strong>
            </div>

            <el-button
              v-if="shouldShowUnlockButton"
              type="primary"
              class="unlock-button"
              :loading="unlocking"
              @click="handleUnlock"
            >
              使用 {{ resource.requiredPoints || 0 }} 积分解锁
            </el-button>

            <el-button
              v-else-if="!authStore.isAuthenticated"
              class="unlock-button"
              @click="redirectToLogin"
            >
              登录后查看权限
            </el-button>
          </section>

          <section class="panel">
            <h2>互动统计</h2>
            <div class="stats-grid">
              <div class="stat-item">
                <span>点赞</span>
                <strong>{{ likes }}</strong>
              </div>
              <div class="stat-item">
                <span>浏览</span>
                <strong>{{ resource.stats?.viewCount ?? resource.viewCount ?? 0 }}</strong>
              </div>
              <div class="stat-item">
                <span>评论</span>
                <strong>{{ resource.stats?.commentCount ?? resource.commentCount ?? 0 }}</strong>
              </div>
              <div class="stat-item">
                <span>收藏</span>
                <strong>{{ resource.stats?.favoriteCount ?? resource.favoriteCount ?? 0 }}</strong>
              </div>
            </div>

            <el-button
              type="primary"
              plain
              class="action-button"
              @click="handleLikeResource"
            >
              点赞资源
            </el-button>
          </section>
        </aside>
      </section>

      <section v-else class="state-panel">
        <el-empty description="资源不存在或暂时不可用" />
      </section>
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import { getArticleDetail, getArticleLikes, likeArticle } from '../api/article';
import { createComment, deleteComment, listComments, type Comment } from '../api/comment';
import { unlockArticle } from '../api/points';
import { useAuthStore } from '../store/auth';
import type { ResourceDetail } from '../types/resource';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();

const resource = ref<ResourceDetail | null>(null);
const likes = ref<number>(0);
const loading = ref(false);
const unlocking = ref(false);
const errorMessage = ref('');
const comments = ref<Comment[]>([]);
const commentForm = ref('');
const commentLoading = ref(false);
const commentSubmitting = ref(false);
const deletingCommentId = ref<number | null>(null);
const commentErrorMessage = ref('');

const resourceID = String(route.params.id);

const shouldShowUnlockButton = computed(() => {
  if (!resource.value) {
    return false;
  }
  return !!authStore.isAuthenticated && !resource.value.isFree && !resource.value.isUnlocked;
});

const gateState = computed(() => {
  if (!resource.value) {
    return {
      label: '资源状态未知',
      description: '请稍后重试。',
      className: 'gate-status--muted',
    };
  }

  if (resource.value.isFree) {
    return {
      label: '免费阅读',
      description: '这是公开资源，所有用户都可以直接浏览。',
      className: 'gate-status--free',
    };
  }

  if (resource.value.isUnlocked) {
    return {
      label: '已解锁',
      description: '你已获得访问权限，可继续查看全部资源内容。',
      className: 'gate-status--unlocked',
    };
  }

  if (!authStore.isAuthenticated) {
    return {
      label: '登录后解锁',
      description: `该资源需要 ${resource.value.requiredPoints || 0} 积分，登录后可查看是否满足条件。`,
      className: 'gate-status--locked',
    };
  }

  return {
    label: '需要积分解锁',
    description: `当前资源访问门槛为 ${resource.value.requiredPoints || 0} 积分，解锁后不会重复扣分。`,
    className: 'gate-status--locked',
  };
});

const fetchResource = async () => {
  resource.value = await getArticleDetail(resourceID);
};

const fetchLikes = async () => {
  const response = await getArticleLikes(resourceID);
  likes.value = response.likes;
};

const fetchComments = async () => {
  commentLoading.value = true;
  commentErrorMessage.value = '';

  try {
    comments.value = await listComments(resourceID);
  } catch (error) {
    console.error('Failed to load comments:', error);
    commentErrorMessage.value = '评论列表加载失败，请稍后重试。';
  } finally {
    commentLoading.value = false;
  }
};

const fetchPageData = async () => {
  loading.value = true;
  errorMessage.value = '';

  try {
    await Promise.all([fetchResource(), fetchLikes(), fetchComments()]);
  } catch (error) {
    console.error('Failed to load resource detail:', error);
    errorMessage.value = '详情内容加载失败，请稍后重试。';
  } finally {
    loading.value = false;
  }
};

const handleCreateComment = async () => {
  if (!authStore.isAuthenticated) {
    ElMessage.error('请先登录后再评论');
    return;
  }

  const content = commentForm.value.trim();
  if (!content) {
    ElMessage.error('评论内容不能为空');
    return;
  }

  commentSubmitting.value = true;
  try {
    await createComment(resourceID, { content });
    commentForm.value = '';
    await Promise.all([fetchComments(), fetchResource()]);
    ElMessage.success('评论发布成功');
  } catch (error) {
    console.error('Failed to create comment:', error);
    ElMessage.error('评论发布失败，请稍后再试。');
  } finally {
    commentSubmitting.value = false;
  }
};

const canDeleteComment = (userId: number) => authStore.currentUser?.userID === userId;

const handleDeleteComment = async (commentId: number) => {
  deletingCommentId.value = commentId;
  try {
    await deleteComment(commentId);
    await Promise.all([fetchComments(), fetchResource()]);
    ElMessage.success('评论已删除');
  } catch (error) {
    console.error('Failed to delete comment:', error);
    ElMessage.error('删除评论失败，请稍后再试。');
  } finally {
    deletingCommentId.value = null;
  }
};

const handleLikeResource = async () => {
  if (!authStore.isAuthenticated) {
    ElMessage.error('请先登录后再点赞');
    return;
  }

  try {
    const response = await likeArticle(resourceID);
    likes.value = response.likes;
  } catch (error) {
    console.error('Error liking resource:', error);
    ElMessage.error('点赞失败，请稍后再试。');
  }
};

const handleUnlock = async () => {
  if (!resource.value || !authStore.isAuthenticated) {
    return;
  }

  unlocking.value = true;
  try {
    const response = await unlockArticle(resourceID);
    await authStore.refreshSummary();
    await fetchResource();
    ElMessage.success(response.message || '资源解锁成功');
  } catch (error) {
    console.error('Error unlocking resource:', error);
    ElMessage.error('解锁失败，请确认积分是否充足。');
  } finally {
    unlocking.value = false;
  }
};

const redirectToLogin = () => {
  router.push({ name: 'Login' });
};

const formatDate = (value: string) =>
  new Date(value).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });

onMounted(fetchPageData);
</script>

<style scoped>
.detail-page {
  min-height: 100%;
  background:
    radial-gradient(circle at top right, rgba(255, 235, 205, 0.88), transparent 30%),
    linear-gradient(180deg, #f8f0e7 0%, #efe1d1 100%);
}

.detail-main {
  max-width: 1200px;
  margin: 0 auto;
  padding: 28px 24px 52px;
}

.detail-layout {
  display: grid;
  grid-template-columns: minmax(0, 1.7fr) minmax(280px, 0.86fr);
  gap: 24px;
}

.detail-skeleton,
.state-panel,
.panel,
.detail-cover,
.author-strip,
.content-body,
.comment-panel {
  border: 1px solid rgba(89, 48, 29, 0.09);
  border-radius: 26px;
  background: rgba(255, 251, 247, 0.84);
  box-shadow: 0 18px 42px rgba(84, 53, 37, 0.08);
  backdrop-filter: blur(16px);
}

.detail-skeleton,
.state-panel {
  padding: 24px;
}

.skeleton-stack {
  display: grid;
  gap: 14px;
  margin-top: 20px;
}

.detail-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.detail-hero {
  position: relative;
  overflow: hidden;
  border-radius: 30px;
}

.detail-cover {
  width: 100%;
  height: 360px;
  object-fit: cover;
  display: block;
}

.detail-cover--placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  background:
    linear-gradient(135deg, rgba(153, 78, 45, 0.95), rgba(55, 27, 19, 0.96));
  color: #fff8f1;
  font-size: 110px;
  font-weight: 700;
}

.hero-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  justify-content: end;
  padding: 28px;
  background: linear-gradient(180deg, rgba(24, 13, 9, 0.05), rgba(24, 13, 9, 0.78));
  color: #fffaf5;
}

.hero-meta {
  display: flex;
  gap: 10px;
  margin-bottom: 14px;
}

.hero-overlay h1 {
  margin: 0;
  font-size: clamp(30px, 4.6vw, 50px);
  line-height: 1.05;
}

.hero-preview {
  max-width: 760px;
  margin: 14px 0 0;
  font-size: 15px;
  line-height: 1.75;
  color: rgba(255, 248, 240, 0.92);
}

.hero-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 18px;
}

.hero-tag {
  background: rgba(255, 255, 255, 0.12);
  border-color: rgba(255, 255, 255, 0.18);
  color: #fff6ef;
}

.author-strip {
  display: flex;
  justify-content: space-between;
  gap: 18px;
  align-items: center;
  padding: 20px 24px;
}

.author-card {
  display: flex;
  gap: 14px;
  align-items: center;
}

.author-avatar {
  width: 52px;
  height: 52px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #8f4b30, #3f2016);
  color: #fff8f1;
  font-size: 22px;
  font-weight: 700;
}

.author-label {
  margin: 0 0 4px;
  font-size: 12px;
  letter-spacing: 0.12em;
  color: #8c6e5b;
  text-transform: uppercase;
}

.author-card h3 {
  margin: 0;
  color: #2e1a12;
}

.author-extra {
  display: flex;
  flex-direction: column;
  gap: 6px;
  text-align: right;
  font-size: 13px;
  color: #7f6557;
}

.content-body {
  padding: 24px;
}

.comment-panel {
  padding: 24px;
}

.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
}

.panel-kicker {
  margin: 0 0 6px;
  font-size: 12px;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: #9f7358;
}

.panel-head h2 {
  margin: 0;
  color: #2e1a12;
}

.comment-editor {
  display: grid;
  gap: 12px;
}

.comment-editor__actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.comment-editor__hint {
  font-size: 13px;
  color: #8c6e5b;
}

.comment-loading {
  padding: 8px 0;
}

.comment-list {
  display: grid;
  gap: 14px;
  margin-top: 20px;
}

.comment-item {
  border: 1px solid rgba(111, 76, 56, 0.12);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.68);
  padding: 16px 18px;
}

.comment-item__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
}

.comment-item__head strong {
  display: block;
  color: #4b2818;
}

.comment-item__time {
  display: block;
  margin-top: 4px;
  font-size: 12px;
  color: #8f6d58;
}

.comment-item__content {
  margin: 0;
  line-height: 1.8;
  color: #5e3b29;
  white-space: pre-wrap;
}

.body-copy {
  margin: 0;
  color: #503b31;
  line-height: 1.95;
  white-space: pre-wrap;
}

.content-gallery {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
  margin-top: 22px;
}

.content-gallery img {
  width: 100%;
  min-height: 220px;
  border-radius: 20px;
  object-fit: cover;
}

.detail-sidebar {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.panel {
  padding: 22px;
}

.panel--sticky {
  position: sticky;
  top: 18px;
}

.panel h2 {
  margin: 0 0 14px;
  color: #2e1a12;
}

.gate-status {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
}

.gate-status--free {
  color: #2a7b53;
}

.gate-status--unlocked {
  color: #8b5a18;
}

.gate-status--locked {
  color: #a13f2c;
}

.gate-status--muted {
  color: #7a6153;
}

.gate-copy {
  margin: 12px 0 0;
  color: #6c5447;
  line-height: 1.75;
}

.points-card {
  margin-top: 18px;
  padding: 18px;
  border-radius: 20px;
  background: rgba(255, 240, 221, 0.78);
}

.points-card span {
  display: block;
  color: #8c6e5b;
  font-size: 12px;
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.points-card strong {
  display: block;
  margin-top: 10px;
  font-size: 32px;
  color: #2e1a12;
}

.unlock-button,
.action-button {
  width: 100%;
  margin-top: 18px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.stat-item {
  padding: 16px 12px;
  border-radius: 18px;
  background: rgba(245, 236, 228, 0.8);
  text-align: center;
}

.stat-item span {
  display: block;
  color: #8c6e5b;
  font-size: 12px;
}

.stat-item strong {
  display: block;
  margin-top: 8px;
  font-size: 24px;
  color: #2e1a12;
}

@media (max-width: 980px) {
  .detail-layout,
  .content-gallery,
  .stats-grid {
    grid-template-columns: 1fr;
  }

  .author-strip {
    flex-direction: column;
    align-items: start;
  }

  .author-extra {
    text-align: left;
  }

  .panel--sticky {
    position: static;
  }

  .comment-editor__actions {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
