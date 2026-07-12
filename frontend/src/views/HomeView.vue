<template>
  <section class="home-shell">
    <div class="hero">
      <p class="eyebrow">Resource Community</p>
      <h1 class="title">围绕优质资源分享的内容社区</h1>
      <p class="description">
        在这里发布图文资源、设置积分门槛、参与评论收藏，并通过签到和互动积累自己的社区资产。
      </p>

      <div class="actions">
        <el-button type="primary" size="large" @click="router.push({ name: 'Resources' })">
          进入资源广场
        </el-button>
        <el-button size="large" type="success" @click="router.push({ name: 'CreateResource' })">
          立即发布
        </el-button>
        <el-button size="large" plain @click="router.push({ name: 'Center' })">
          查看用户中心
        </el-button>
      </div>
    </div>

    <div class="feature-grid">
      <article class="feature-card">
        <h2>内容发现</h2>
        <p>支持分页、排序、关键词搜索和标签筛选，快速找到值得解锁的内容。</p>
      </article>
      <article class="feature-card">
        <h2>积分机制</h2>
        <p>签到、发布和互动都可以沉淀积分，积分可用于解锁资源和兑换权益。</p>
      </article>
      <article class="feature-card">
        <h2>用户沉淀</h2>
        <p>用户中心聚合收藏、积分、签到状态，逐步形成基础自运营闭环。</p>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';

const router = useRouter();
</script>

<style scoped>
.home-shell {
  min-height: calc(100vh - 60px);
  padding: 48px 24px 64px;
  background:
    radial-gradient(circle at top left, rgba(32, 178, 170, 0.18), transparent 32%),
    linear-gradient(180deg, #f5fbfa 0%, #eef3f8 100%);
}

.hero {
  max-width: 920px;
  margin: 0 auto 40px;
  padding: 48px;
  border: 1px solid rgba(18, 52, 86, 0.08);
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.82);
  box-shadow: 0 24px 60px rgba(36, 59, 85, 0.12);
  backdrop-filter: blur(12px);
}

.eyebrow {
  margin: 0 0 12px;
  color: #1d7f73;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.title {
  margin: 0;
  color: #17324d;
  font-size: clamp(32px, 5vw, 56px);
  line-height: 1.08;
}

.description {
  max-width: 680px;
  margin: 20px 0 0;
  color: #4b6177;
  font-size: 18px;
  line-height: 1.75;
}

.actions {
  display: flex;
  gap: 16px;
  margin-top: 32px;
  flex-wrap: wrap;
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 20px;
  max-width: 920px;
  margin: 0 auto;
}

.feature-card {
  padding: 28px;
  border-radius: 24px;
  background: #ffffff;
  box-shadow: 0 16px 40px rgba(36, 59, 85, 0.08);
}

.feature-card h2 {
  margin: 0 0 10px;
  color: #17324d;
  font-size: 22px;
}

.feature-card p {
  margin: 0;
  color: #5f7184;
  line-height: 1.75;
}

@media (max-width: 640px) {
  .home-shell {
    padding: 24px 16px 48px;
  }

  .hero {
    padding: 28px 22px;
    border-radius: 20px;
  }

  .description {
    font-size: 16px;
  }
}
</style>
