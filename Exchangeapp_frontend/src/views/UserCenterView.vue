<template>
  <el-container class="center-page">
    <el-main class="center-main">
      <section class="center-hero">
        <div class="identity-card">
          <div class="identity-avatar">
            {{ authStore.currentUser?.username?.slice(0, 1) || 'U' }}
          </div>
          <div>
            <p class="eyebrow">USER CENTER</p>
            <h1>{{ authStore.currentUser?.username || '未登录用户' }}</h1>
            <p class="sub-copy">
              管理你的收藏与积分，完成签到并持续积累社区权益。
            </p>
          </div>
        </div>

        <div class="hero-actions">
          <el-button
            type="primary"
            :loading="checkInLoading"
            :disabled="checkInCompleted || !authStore.isAuthenticated"
            @click="handleCheckIn"
          >
            {{ checkInCompleted ? '今日已签到' : '立即签到' }}
          </el-button>
          <p class="check-in-hint">
            {{ checkInCompleted ? '今日签到已完成，可明天再来。' : '每日签到可获得积分奖励。' }}
          </p>
        </div>
      </section>

      <section v-if="!authStore.isAuthenticated" class="state-panel">
        <el-empty description="登录后可查看用户中心">
          <el-button type="primary" @click="goToLogin">前往登录</el-button>
        </el-empty>
      </section>

      <template v-else>
        <section class="summary-grid">
          <div class="summary-card">
            <span>我的积分</span>
            <strong>{{ authStore.pointsBalance }}</strong>
          </div>
          <div class="summary-card">
            <span>我的收藏</span>
            <strong>{{ favoriteArticles.length }}</strong>
          </div>
          <div class="summary-card">
            <span>积分权益</span>
            <strong>{{ authStore.pointsSummary?.privileges?.length ?? 0 }}</strong>
          </div>
        </section>

        <section class="content-grid">
          <el-card class="panel-card" shadow="hover">
            <template #header>
              <div class="panel-head">
                <div>
                  <p class="panel-kicker">Favorites</p>
                  <h2>我的收藏</h2>
                </div>
                <el-button text @click="loadFavorites">刷新</el-button>
              </div>
            </template>

            <div v-if="favoriteLoading" class="panel-loading">
              <el-skeleton :rows="4" animated />
            </div>
            <el-result
              v-else-if="favoriteError"
              icon="warning"
              title="收藏加载失败"
              :sub-title="favoriteError"
            >
              <template #extra>
                <el-button @click="loadFavorites">重试</el-button>
              </template>
            </el-result>
            <el-empty v-else-if="!favoriteArticles.length" description="你还没有收藏任何资源" />
            <div v-else class="favorite-list">
              <button
                v-for="favorite in favoriteArticles"
                :key="favorite.id"
                class="favorite-item"
                type="button"
                @click="goToArticle(favorite.id)"
              >
                <div class="favorite-cover">
                  <img v-if="favorite.coverUrl" :src="favorite.coverUrl" :alt="favorite.title" />
                  <div v-else class="favorite-cover--placeholder">{{ favorite.title.slice(0, 1) }}</div>
                </div>
                <div class="favorite-body">
                  <div class="favorite-meta">
                    <span>{{ favorite.isFree ? '免费' : `${favorite.requiredPoints} 积分` }}</span>
                    <span>{{ favorite.author.username }}</span>
                  </div>
                  <h3>{{ favorite.title }}</h3>
                  <p>{{ favorite.preview }}</p>
                </div>
              </button>
            </div>
          </el-card>

          <el-card class="panel-card" shadow="hover">
            <template #header>
              <div class="panel-head">
                <div>
                  <p class="panel-kicker">Points</p>
                  <h2>我的积分</h2>
                </div>
                <el-button text @click="loadPointRecords">刷新</el-button>
              </div>
            </template>

            <div class="points-summary" v-if="authStore.pointsSummary">
              <div class="points-balance">
                <span>当前余额</span>
                <strong>{{ authStore.pointsBalance }}</strong>
              </div>
              <div class="privilege-list">
                <span class="privilege-label">已兑换权益</span>
                <el-tag
                  v-for="privilege in authStore.pointsSummary.privileges"
                  :key="privilege.privilegeKey"
                  effect="plain"
                  type="success"
                >
                  {{ privilege.privilegeKey }}
                </el-tag>
                <span v-if="!authStore.pointsSummary.privileges.length" class="empty-inline">
                  暂无权益
                </span>
              </div>
            </div>

            <div v-if="recordLoading" class="panel-loading">
              <el-skeleton :rows="4" animated />
            </div>
            <el-result
              v-else-if="recordError"
              icon="warning"
              title="积分记录加载失败"
              :sub-title="recordError"
            >
              <template #extra>
                <el-button @click="loadPointRecords">重试</el-button>
              </template>
            </el-result>
            <el-empty v-else-if="!pointRecords.length" description="暂无积分记录" />
            <div v-else class="record-list">
              <div v-for="record in pointRecords" :key="record.id" class="record-item">
                <div>
                  <h3>{{ record.description || record.source }}</h3>
                  <p>{{ formatDate(record.createdAt) }}</p>
                </div>
                <div class="record-side">
                  <strong :class="record.change >= 0 ? 'record-positive' : 'record-negative'">
                    {{ record.change >= 0 ? `+${record.change}` : record.change }}
                  </strong>
                  <span>余额 {{ record.balanceAfter }}</span>
                </div>
              </div>
            </div>
          </el-card>
        </section>
      </template>
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import { listMyFavorites, type FavoriteArticle } from '../api/favorite';
import { checkIn } from '../api/checkin';
import { getMyPointsRecords, type PointsRecord } from '../api/points';
import { useAuthStore } from '../store/auth';

const router = useRouter();
const authStore = useAuthStore();

const favoriteArticles = ref<FavoriteArticle[]>([]);
const pointRecords = ref<PointsRecord[]>([]);
const favoriteLoading = ref(false);
const recordLoading = ref(false);
const checkInLoading = ref(false);
const favoriteError = ref('');
const recordError = ref('');
const checkInDate = ref<string | null>(localStorage.getItem('daily_check_in_date'));

const today = new Date().toISOString().slice(0, 10);

const checkInCompleted = computed(() => checkInDate.value === today);

const loadFavorites = async () => {
  if (!authStore.isAuthenticated) {
    favoriteArticles.value = [];
    return;
  }

  favoriteLoading.value = true;
  favoriteError.value = '';

  try {
    favoriteArticles.value = await listMyFavorites();
  } catch (error) {
    console.error('Failed to load favorites:', error);
    favoriteError.value = '请稍后重试。';
  } finally {
    favoriteLoading.value = false;
  }
};

const loadPointRecords = async () => {
  if (!authStore.isAuthenticated) {
    pointRecords.value = [];
    return;
  }

  recordLoading.value = true;
  recordError.value = '';

  try {
    pointRecords.value = await getMyPointsRecords();
  } catch (error) {
    console.error('Failed to load point records:', error);
    recordError.value = '请稍后重试。';
  } finally {
    recordLoading.value = false;
  }
};

const handleCheckIn = async () => {
  if (!authStore.isAuthenticated || checkInCompleted.value) {
    return;
  }

  checkInLoading.value = true;
  try {
    const response = await checkIn();
    checkInDate.value = today;
    localStorage.setItem('daily_check_in_date', today);
    await Promise.all([authStore.refreshSummary(), loadPointRecords()]);
    ElMessage.success(response.message || '签到成功');
  } catch (error) {
    console.error('Check-in failed:', error);
    ElMessage.error('签到失败，请稍后再试。');
  } finally {
    checkInLoading.value = false;
  }
};

const goToLogin = () => {
  router.push({ name: 'Login' });
};

const goToArticle = (id: number) => {
  router.push({ name: 'ResourceDetail', params: { id } });
};

const formatDate = (value: string) => {
  return new Date(value).toLocaleString('zh-CN', {
    hour12: false,
  });
};

onMounted(async () => {
  if (!authStore.isAuthenticated) {
    return;
  }

  await Promise.all([loadFavorites(), loadPointRecords()]);
});
</script>

<style scoped>
.center-page {
  min-height: 100%;
  background:
    radial-gradient(circle at top left, rgba(255, 231, 205, 0.9), transparent 26%),
    linear-gradient(180deg, #f7f1ea 0%, #ede0d2 100%);
}

.center-main {
  max-width: 1180px;
  margin: 0 auto;
  padding: 28px 24px 52px;
}

.center-hero,
.summary-card,
.panel-card,
.state-panel {
  border: 1px solid rgba(92, 53, 34, 0.08);
  border-radius: 26px;
  background: rgba(255, 251, 247, 0.86);
  box-shadow: 0 18px 40px rgba(84, 53, 37, 0.08);
  backdrop-filter: blur(16px);
}

.center-hero {
  display: flex;
  justify-content: space-between;
  gap: 24px;
  align-items: center;
  padding: 28px;
  margin-bottom: 24px;
}

.identity-card {
  display: flex;
  gap: 18px;
  align-items: center;
}

.identity-avatar {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #8f4b30, #3c1f15);
  color: #fff8f1;
  font-size: 32px;
  font-weight: 700;
}

.eyebrow {
  margin: 0 0 8px;
  color: #956646;
  font-size: 12px;
  letter-spacing: 0.2em;
  font-weight: 700;
}

.identity-card h1 {
  margin: 0;
  font-size: clamp(30px, 4vw, 42px);
  color: #2e1a12;
}

.sub-copy {
  margin: 10px 0 0;
  color: #6f584b;
  line-height: 1.7;
}

.hero-actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
  align-items: flex-end;
}

.check-in-hint {
  margin: 0;
  color: #866a58;
  font-size: 13px;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 18px;
  margin-bottom: 24px;
}

.summary-card {
  padding: 22px;
}

.summary-card span {
  display: block;
  color: #8f725f;
  font-size: 12px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.summary-card strong {
  display: block;
  margin-top: 12px;
  font-size: 34px;
  color: #2e1a12;
}

.content-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.1fr) minmax(0, 1fr);
  gap: 24px;
}

.panel-head {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: center;
}

.panel-kicker {
  margin: 0;
  color: #8f725f;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.14em;
}

.panel-head h2 {
  margin: 6px 0 0;
  color: #2e1a12;
}

.panel-loading {
  padding: 6px 0 2px;
}

.favorite-list,
.record-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.favorite-item {
  display: grid;
  grid-template-columns: 112px minmax(0, 1fr);
  gap: 14px;
  border: 0;
  background: rgba(247, 240, 232, 0.92);
  border-radius: 20px;
  padding: 12px;
  text-align: left;
  cursor: pointer;
}

.favorite-cover {
  width: 112px;
  height: 92px;
  border-radius: 16px;
  overflow: hidden;
  background: linear-gradient(135deg, #a45f3d, #4a281d);
}

.favorite-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.favorite-cover--placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #fff8f1;
  font-size: 32px;
  font-weight: 700;
}

.favorite-meta {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  color: #8b705c;
  font-size: 12px;
}

.favorite-body h3,
.record-item h3 {
  margin: 8px 0 6px;
  color: #2e1a12;
}

.favorite-body p,
.record-item p {
  margin: 0;
  color: #6f584b;
  line-height: 1.6;
}

.points-summary {
  margin-bottom: 18px;
  padding: 18px;
  border-radius: 22px;
  background: rgba(247, 240, 232, 0.92);
}

.points-balance span,
.privilege-label {
  display: block;
  color: #8f725f;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.12em;
}

.points-balance strong {
  display: block;
  margin-top: 10px;
  font-size: 32px;
  color: #2e1a12;
}

.privilege-list {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
  margin-top: 16px;
}

.empty-inline {
  color: #8b705c;
  font-size: 13px;
}

.record-item {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
  padding: 16px;
  border-radius: 18px;
  background: rgba(247, 240, 232, 0.92);
}

.record-side {
  display: flex;
  flex-direction: column;
  gap: 6px;
  align-items: flex-end;
  color: #7f6657;
  font-size: 13px;
}

.record-positive {
  color: #2a7b53;
}

.record-negative {
  color: #a13f2c;
}

.state-panel {
  padding: 28px;
}

@media (max-width: 960px) {
  .center-hero,
  .content-grid,
  .summary-grid {
    grid-template-columns: 1fr;
    flex-direction: column;
    align-items: stretch;
  }

  .hero-actions {
    align-items: stretch;
  }

  .favorite-item {
    grid-template-columns: 1fr;
  }

  .favorite-cover {
    width: 100%;
    height: 180px;
  }

  .record-item {
    flex-direction: column;
    align-items: flex-start;
  }

  .record-side {
    align-items: flex-start;
  }
}
</style>
