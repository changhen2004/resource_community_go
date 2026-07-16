<template>
  <section class="auth-shell">
    <div class="auth-layout">
      <article class="auth-story">
        <p class="auth-kicker">WELCOME BACK</p>
        <h1>欢迎回到资源社区</h1>
        <p class="auth-copy">
          登录后继续管理你的资源发布、积分解锁记录、评论收藏和个人内容沉淀。
        </p>

        <div class="auth-highlights">
          <span>关键词搜索</span>
          <span>标签导航</span>
          <span>积分体系</span>
          <span>评论互动</span>
        </div>
      </article>

      <el-form :model="form" class="auth-form" @submit.prevent="login">
        <div class="auth-form__head">
          <p class="auth-kicker">LOGIN</p>
          <h2>登录账号</h2>
          <span>继续访问你的资源空间</span>
        </div>

        <el-form-item label="用户名" label-width="80px">
          <el-input v-model="form.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="密码" label-width="80px">
          <el-input v-model="form.password" type="password" placeholder="请输入密码" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" native-type="submit" class="auth-submit">登录</el-button>
        </el-form-item>
      </el-form>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import { useAuthStore } from '../../store/auth';

const form = ref({
  username: '',
  password: '',
});

const authStore = useAuthStore();
const router = useRouter();

const login = async () => {
  try {
    await authStore.login(form.value.username, form.value.password);
    router.push({ name: 'Resources' });
  } catch {
    ElMessage.error('登录失败，请检查用户名和密码。');
  }
};
</script>

<style scoped>
.auth-shell {
  min-height: calc(100vh - 92px);
  padding: 28px 24px 48px;
}

.auth-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(360px, 420px);
  gap: 24px;
  max-width: 1280px;
  margin: 0 auto;
}

.auth-story,
.auth-form {
  border: 1px solid rgba(56, 61, 64, 0.08);
  border-radius: 26px;
  background: rgba(251, 250, 245, 0.9);
  box-shadow: 0 20px 44px rgba(45, 51, 54, 0.08);
  backdrop-filter: blur(14px);
}

.auth-story {
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  min-height: 520px;
  padding: 32px;
  background:
    linear-gradient(140deg, rgba(19, 63, 69, 0.92), rgba(124, 99, 55, 0.86)),
    rgba(251, 250, 245, 0.9);
  color: #f7f3e9;
}

.auth-kicker {
  margin: 0;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.auth-story h1,
.auth-form__head h2 {
  margin: 14px 0 0;
}

.auth-story h1 {
  font-size: clamp(34px, 5vw, 58px);
  line-height: 1.04;
}

.auth-copy {
  max-width: 520px;
  margin: 18px 0 0;
  line-height: 1.8;
  color: rgba(247, 243, 233, 0.86);
}

.auth-highlights {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  margin-top: 28px;
}

.auth-highlights span {
  padding: 8px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.12);
  font-size: 13px;
}

.auth-form {
  align-self: center;
  padding: 28px 24px 12px;
}

.auth-form__head {
  margin-bottom: 18px;
}

.auth-form__head span {
  color: #667378;
  font-size: 14px;
}

.auth-submit {
  width: 100%;
}

@media (max-width: 920px) {
  .auth-layout {
    grid-template-columns: 1fr;
  }

  .auth-story {
    min-height: auto;
  }
}

@media (max-width: 720px) {
  .auth-shell {
    padding: 18px 12px 36px;
  }

  .auth-story,
  .auth-form {
    padding: 22px 18px;
    border-radius: 22px;
  }
}
</style>
