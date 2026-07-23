<template>
  <section class="detail-shell">
    <section v-if="loading" class="detail-skeleton">
      <el-skeleton animated>
        <template #template>
          <el-skeleton-item variant="image" style="width: 100%; height: 360px" />
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
      <article class="detail-main">
        <section class="hero-panel">
          <div class="hero-cover-wrap">
            <img
              v-if="resource.coverUrl"
              :src="resource.coverUrl"
              :alt="resource.title"
              class="hero-cover"
              decoding="async"
            />
            <div v-else class="hero-cover hero-cover--placeholder">
              <span>{{ resource.title.slice(0, 1) }}</span>
            </div>
            <div class="hero-cover__shade"></div>

            <div class="hero-topbar">
              <button type="button" class="ghost-link" @click="goBackToList">
                返回资源广场
              </button>
              <div class="hero-topbar__meta">
                <span>{{ resource.status || 'published' }}</span>
                <strong>{{ resource.isFree ? '免费资源' : `${resource.requiredPoints || 0} 积分解锁` }}</strong>
              </div>
            </div>

            <div class="hero-body">
              <p class="hero-kicker">RESOURCE STORY</p>
              <h1>{{ resource.title }}</h1>
              <p class="hero-preview">{{ resource.preview }}</p>

              <div v-if="resource.tags?.length" class="hero-tags">
                <button
                  v-for="tag in resource.tags"
                  :key="tag"
                  type="button"
                  class="hero-tag"
                  @click="goToTag(tag)"
                >
                  {{ tag }}
                </button>
              </div>

              <div class="hero-meta-grid">
                <article class="hero-meta-card">
                  <span>作者</span>
                  <strong>{{ resource.author?.username || '匿名作者' }}</strong>
                </article>
                <article class="hero-meta-card">
                  <span>可见性</span>
                  <strong>{{ resource.isFree ? '公开阅读' : '积分门槛' }}</strong>
                </article>
                <article class="hero-meta-card">
                  <span>互动热度</span>
                  <strong>{{ likes }}</strong>
                </article>
              </div>
            </div>
          </div>
        </section>

        <section class="summary-panel">
          <div class="author-card">
            <div class="author-avatar">
              {{ resource.author?.username?.slice(0, 1) || 'U' }}
            </div>
            <div class="author-copy">
              <p class="summary-kicker">Author</p>
              <h2>{{ resource.author?.username || '匿名作者' }}</h2>
              <span>作者 ID #{{ resource.author?.id || resource.authorId }}</span>
            </div>
            <el-button
              v-if="canShowFollowButton"
              class="follow-button"
              :type="authorSocialStatus?.isFollowing ? 'default' : 'primary'"
              :loading="followSubmitting"
              @click="handleFollowToggle"
            >
              {{ authorSocialStatus?.isFollowing ? '取消关注' : '关注作者' }}
            </el-button>
          </div>

          <div class="summary-divider"></div>

          <div class="summary-grid">
            <article>
              <span>浏览</span>
              <strong>{{ resource.stats?.viewCount ?? resource.viewCount ?? 0 }}</strong>
            </article>
            <article>
              <span>评论</span>
              <strong>{{ resource.stats?.commentCount ?? resource.commentCount ?? 0 }}</strong>
            </article>
            <article>
              <span>粉丝</span>
              <strong>{{ authorSocialStatus?.followerCount ?? 0 }}</strong>
            </article>
            <article>
              <span>关注</span>
              <strong>{{ authorSocialStatus?.followingCount ?? 0 }}</strong>
            </article>
          </div>
        </section>

        <section class="reading-panel">
          <div class="section-head">
            <div>
              <p class="section-kicker">Article</p>
              <h2>资源正文</h2>
            </div>
            <button
              v-if="resource.tags?.[0]"
              type="button"
              class="ghost-link"
              @click="goToTag(resource.tags[0])"
            >
              查看同标签内容
            </button>
          </div>

          <p class="body-copy">{{ resource.content }}</p>

          <div v-if="resource.contentImages?.length" class="content-gallery">
            <img
              v-for="(image, index) in resource.contentImages"
              :key="`${image}-${index}`"
              :src="image"
              :alt="`${resource.title} 配图 ${index + 1}`"
              loading="lazy"
              decoding="async"
            />
          </div>
        </section>

        <section v-if="relatedResources.length" class="related-panel">
          <div class="section-head">
            <div>
              <p class="section-kicker">More Like This</p>
              <h2>相关资源</h2>
            </div>
            <button type="button" class="ghost-link" @click="goBackToList">
              浏览更多
            </button>
          </div>

          <div class="related-grid">
            <ResourceStoryCard
              v-for="item in relatedResources"
              :key="item.id"
              :resource="item"
              variant="compact"
              @tag="goToTag"
            />
          </div>
        </section>

        <section class="comment-panel">
          <div class="section-head">
            <div>
              <p class="section-kicker">Comments</p>
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
        <section class="sidebar-panel sidebar-panel--sticky">
          <p class="section-kicker">Access</p>
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

        <section class="sidebar-panel">
          <p class="section-kicker">Engagement</p>
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

          <el-button type="primary" plain class="action-button" @click="handleLikeResource">
            点赞资源
          </el-button>
          <el-button
            class="action-button"
            :type="isFavorited ? 'danger' : 'success'"
            plain
            :loading="favoriteSubmitting"
            @click="handleFavoriteToggle"
          >
            {{ isFavorited ? '取消收藏' : '收藏资源' }}
          </el-button>
        </section>

        <section v-if="resource.tags?.length" class="sidebar-panel">
          <p class="section-kicker">Topics</p>
          <h2>标签入口</h2>
          <div class="sidebar-tags">
            <button
              v-for="tag in resource.tags"
              :key="tag"
              type="button"
              class="sidebar-tag"
              @click="goToTag(tag)"
            >
              {{ tag }}
            </button>
          </div>
        </section>
      </aside>
    </section>

    <section v-else class="state-panel">
      <el-empty description="资源不存在或暂时不可用" />
    </section>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import {
  followAuthor,
  getArticleDetail,
  getArticleLikes,
  getAuthorSocialStatus,
  likeArticle,
  listArticles,
  unfollowAuthor,
} from '../api/article';
import { createComment, deleteComment, listComments, type Comment } from '../api/comment';
import { favoriteArticle, listMyFavorites, unfavoriteArticle } from '../api/favorite';
import { unlockArticle } from '../api/points';
import ResourceStoryCard from '../components/ResourceStoryCard.vue';
import { useAuthStore } from '../store/auth';
import type { AuthorSocialStatus, ResourceDetail, ResourceSummary } from '../types/resource';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();

const resource = ref<ResourceDetail | null>(null);
const relatedResources = ref<ResourceSummary[]>([]);
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
const favoriteSubmitting = ref(false);
const isFavorited = ref(false);
const followSubmitting = ref(false);
const authorSocialStatus = ref<AuthorSocialStatus | null>(null);

const resourceID = String(route.params.id);

const shouldShowUnlockButton = computed(() => {
  if (!resource.value) {
    return false;
  }
  return !!authStore.isAuthenticated && !resource.value.isFree && !resource.value.isUnlocked;
});

const primaryTag = computed(() => resource.value?.tags?.[0]?.trim() || '');
const authorID = computed(() => resource.value?.author?.id || resource.value?.authorId || 0);
const canShowFollowButton = computed(() => {
  if (!authStore.isAuthenticated || !authorID.value) {
    return false;
  }
  return authStore.currentUser?.userID !== authorID.value;
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

const goBackToList = () => {
  router.push({ name: 'Resources' });
};

const goToTag = (tag: string) => {
  router.push({
    name: 'Resources',
    query: tag ? { tag } : {},
  });
};

const fetchResource = async () => {
  resource.value = await getArticleDetail(resourceID);
};

const fetchAuthorSocialStatus = async () => {
  if (!authorID.value) {
    authorSocialStatus.value = null;
    return;
  }

  try {
    authorSocialStatus.value = await getAuthorSocialStatus(authorID.value);
  } catch (error) {
    console.error('Failed to load author social status:', error);
    authorSocialStatus.value = null;
  }
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

const fetchRelatedResources = async () => {
  try {
    const items = primaryTag.value
      ? await listArticles({ page: 1, pageSize: 4, sort: 'hot', tag: primaryTag.value })
      : await listArticles({ page: 1, pageSize: 4, sort: 'hot' });

    relatedResources.value = items.filter((item) => String(item.id) !== resourceID).slice(0, 3);
  } catch (error) {
    console.error('Failed to load related resources:', error);
    relatedResources.value = [];
  }
};

const syncFavoriteState = async () => {
  if (!authStore.isAuthenticated || !resource.value) {
    isFavorited.value = false;
    return;
  }

  try {
    const favorites = await listMyFavorites();
    isFavorited.value = favorites.some((favorite) => favorite.id === resource.value?.id);
  } catch (error) {
    console.error('Failed to sync favorite state:', error);
  }
};

const fetchPageData = async () => {
  loading.value = true;
  errorMessage.value = '';

  try {
    await fetchResource();
    await Promise.all([fetchLikes(), fetchComments(), fetchRelatedResources(), fetchAuthorSocialStatus()]);
    await syncFavoriteState();
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

const handleFavoriteToggle = async () => {
  if (!authStore.isAuthenticated) {
    ElMessage.error('请先登录后再收藏');
    return;
  }

  favoriteSubmitting.value = true;
  try {
    if (isFavorited.value) {
      await unfavoriteArticle(resourceID);
      isFavorited.value = false;
      ElMessage.success('已取消收藏');
    } else {
      await favoriteArticle(resourceID);
      isFavorited.value = true;
      ElMessage.success('收藏成功');
    }

    await fetchResource();
  } catch (error) {
    console.error('Failed to update favorite state:', error);
    ElMessage.error('收藏操作失败，请稍后再试。');
  } finally {
    favoriteSubmitting.value = false;
  }
};

const handleFollowToggle = async () => {
  if (!authStore.isAuthenticated || !authorID.value) {
    ElMessage.error('请先登录后再关注作者');
    return;
  }

  followSubmitting.value = true;
  try {
    const response = authorSocialStatus.value?.isFollowing
      ? await unfollowAuthor(authorID.value)
      : await followAuthor(authorID.value);
    authorSocialStatus.value = response.status;
    ElMessage.success(response.message === 'unfollowed' ? '已取消关注' : '关注成功');
  } catch (error) {
    console.error('Failed to update follow state:', error);
    ElMessage.error('关注操作失败，请稍后再试。');
  } finally {
    followSubmitting.value = false;
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
.detail-shell {
  padding: 28px 24px 64px;
}

.detail-layout {
  display: grid;
  grid-template-columns: minmax(0, 1.65fr) minmax(280px, 0.78fr);
  gap: 24px;
  max-width: 1440px;
  margin: 0 auto;
}

.detail-main,
.detail-sidebar {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.detail-skeleton,
.state-panel,
.hero-panel,
.summary-panel,
.reading-panel,
.related-panel,
.comment-panel,
.sidebar-panel {
  border: 1px solid rgba(56, 61, 64, 0.08);
  border-radius: 26px;
  background: rgba(251, 250, 245, 0.96);
  box-shadow: 0 8px 28px rgba(45, 51, 54, 0.06);
}

.detail-skeleton,
.state-panel,
.hero-panel,
.summary-panel,
.reading-panel,
.related-panel,
.comment-panel {
  contain: layout paint;
}

.sidebar-panel {
  contain: layout;
}

.detail-skeleton,
.state-panel {
  max-width: 1440px;
  margin: 0 auto;
  padding: 24px;
}

.skeleton-stack {
  display: grid;
  gap: 14px;
  margin-top: 20px;
}

.hero-cover-wrap {
  position: relative;
  min-height: 360px;
  border-radius: 26px;
  overflow: hidden;
  background:
    linear-gradient(140deg, rgba(19, 63, 69, 0.92), rgba(124, 99, 55, 0.86)),
    #244;
}

.hero-cover {
  display: block;
  width: 100%;
  height: auto;
}

.hero-cover--placeholder {
  display: grid;
  place-items: center;
  aspect-ratio: 16 / 9.5;
  width: 100%;
  color: rgba(247, 243, 233, 0.95);
  font-size: clamp(64px, 10vw, 110px);
  font-weight: 700;
}

.hero-cover__shade {
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, rgba(16, 21, 24, 0.12), rgba(16, 21, 24, 0.7));
}

.hero-topbar,
.section-head {
  display: flex;
  gap: 16px;
  align-items: center;
  justify-content: space-between;
}

.hero-topbar {
  position: absolute;
  top: 0;
  right: 0;
  left: 0;
  z-index: 2;
  padding: 28px;
}

.hero-body {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  z-index: 2;
  padding: 0 28px 28px;
  color: #f7f4eb;
}

.ghost-link,
.hero-tag,
.sidebar-tag {
  border: 0;
  cursor: pointer;
}

.ghost-link {
  padding: 10px 14px;
  border-radius: 999px;
  background: rgba(251, 250, 245, 0.88);
  color: #1a3a40;
  font-size: 13px;
  font-weight: 700;
}

.hero-topbar__meta {
  display: inline-flex;
  gap: 10px;
  align-items: center;
  padding: 10px 14px;
  border-radius: 999px;
  background: rgba(16, 21, 24, 0.66);
}

.hero-topbar__meta span,
.hero-topbar__meta strong {
  font-size: 12px;
}

.hero-kicker,
.summary-kicker,
.section-kicker {
  margin: 0;
  color: #c9b589;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.hero-body h1 {
  margin: 0;
  color: #f7f4eb;
  font-size: clamp(34px, 5vw, 58px);
  line-height: 1.04;
}

.hero-preview {
  max-width: 760px;
  margin: 16px 0 0;
  color: rgba(247, 243, 233, 0.88);
  line-height: 1.8;
}

.hero-tags,
.sidebar-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 20px;
}

.hero-tag,
.sidebar-tag {
  padding: 8px 12px;
  border-radius: 999px;
  background: rgba(251, 250, 245, 0.15);
  color: #f7f4eb;
  font-size: 12px;
}

.sidebar-tag {
  background: #edf2f1;
  color: #1d4046;
}

.hero-meta-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
  margin-top: 24px;
}

.hero-meta-card {
  padding: 18px;
  border-radius: 20px;
  background: rgba(251, 250, 245, 0.2);
}

.hero-meta-card span {
  display: block;
  color: rgba(247, 243, 233, 0.72);
  font-size: 12px;
}

.hero-meta-card strong {
  display: block;
  margin-top: 10px;
  color: #f7f4eb;
  font-size: 24px;
}

.summary-panel,
.reading-panel,
.related-panel,
.comment-panel,
.sidebar-panel {
  padding: 24px;
}

.summary-panel {
  display: grid;
  grid-template-columns: auto 1px minmax(0, 1fr);
  gap: 20px;
  align-items: center;
}

.author-card {
  display: flex;
  gap: 14px;
  align-items: center;
}

.author-avatar {
  display: grid;
  width: 58px;
  height: 58px;
  place-items: center;
  border-radius: 50%;
  background: linear-gradient(135deg, #183f44, #7c6337);
  color: #f7f4eb;
  font-size: 24px;
  font-weight: 700;
}

.author-card h2 {
  margin: 6px 0 0;
  color: #152f35;
  font-size: 24px;
}

.author-card span {
  color: #6b767a;
  font-size: 13px;
}

.author-copy {
  min-width: 0;
}

.follow-button {
  margin-left: 8px;
  border-radius: 999px;
  font-weight: 700;
}

.summary-divider {
  width: 1px;
  height: 100%;
  background: rgba(56, 61, 64, 0.08);
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

.summary-grid article,
.points-card,
.stat-item {
  padding: 18px;
  border-radius: 20px;
  background: linear-gradient(180deg, #f5efe0, #f8f6f0);
}

.summary-grid span,
.points-card span,
.stat-item span {
  display: block;
  color: #7a6b55;
  font-size: 12px;
}

.summary-grid strong,
.points-card strong,
.stat-item strong {
  display: block;
  margin-top: 10px;
  color: #152f35;
  font-size: 28px;
}

.section-head h2,
.sidebar-panel h2 {
  margin: 12px 0 0;
  color: #152f35;
  font-size: 28px;
}

.body-copy {
  margin: 20px 0 0;
  color: #4f5b60;
  line-height: 1.95;
  white-space: pre-wrap;
}

.content-gallery {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
  margin-top: 24px;
}

.content-gallery img {
  width: 100%;
  min-height: 240px;
  border-radius: 20px;
  object-fit: cover;
}

.related-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 18px;
  margin-top: 22px;
}

.comment-editor {
  display: grid;
  gap: 12px;
  margin-top: 20px;
}

.comment-editor__actions {
  display: flex;
  gap: 12px;
  align-items: center;
  justify-content: space-between;
}

.comment-editor__hint,
.gate-copy,
.comment-item__time {
  color: #6b767a;
  font-size: 13px;
}

.comment-loading {
  padding: 12px 0 0;
}

.comment-list {
  display: grid;
  gap: 14px;
  margin-top: 20px;
}

.comment-item {
  padding: 16px 18px;
  border: 1px solid rgba(56, 61, 64, 0.08);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.72);
}

.comment-item__head {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 10px;
}

.comment-item__head strong {
  display: block;
  color: #1d3b41;
}

.comment-item__time {
  display: block;
  margin-top: 4px;
}

.comment-item__content {
  margin: 0;
  color: #4f5b60;
  line-height: 1.8;
  white-space: pre-wrap;
}

.sidebar-panel--sticky {
  position: sticky;
  top: 108px;
}

.gate-status {
  margin: 16px 0 0;
  font-size: 28px;
  font-weight: 700;
}

.gate-status--free {
  color: #2b7d57;
}

.gate-status--unlocked {
  color: #95611d;
}

.gate-status--locked {
  color: #b34d35;
}

.gate-status--muted {
  color: #6b767a;
}

.gate-copy {
  margin: 12px 0 0;
  line-height: 1.8;
}

.points-card,
.stats-grid {
  margin-top: 18px;
}

.unlock-button,
.action-button {
  width: 100%;
  margin-top: 18px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

@media (max-width: 1180px) {
  .detail-layout {
    grid-template-columns: 1fr;
  }

  .sidebar-panel--sticky {
    position: static;
  }

  .related-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 820px) {
  .detail-shell {
    padding: 18px 12px 40px;
  }

  .hero-cover-wrap {
    min-height: 420px;
  }

  .hero-body {
    padding: 0 18px 18px;
  }

  .hero-topbar {
    padding: 18px;
  }

  .hero-meta-grid,
  .summary-panel,
  .summary-grid,
  .content-gallery,
  .related-grid,
  .stats-grid {
    grid-template-columns: 1fr;
  }

  .summary-panel {
    gap: 18px;
  }

  .summary-divider {
    display: none;
  }

  .hero-topbar,
  .section-head,
  .comment-editor__actions {
    flex-direction: column;
    align-items: stretch;
  }

  .hero-topbar__meta {
    justify-content: space-between;
  }

  .summary-panel,
  .reading-panel,
  .related-panel,
  .comment-panel,
  .sidebar-panel {
    padding: 18px;
    border-radius: 22px;
  }
}
</style>
