<template>
  <section class="auth-shell">
    <div class="auth-layout">
      <article class="auth-story">
        <p class="auth-kicker">START HERE</p>
        <h1>创建你的资源社区账号</h1>
        <p class="auth-copy">
          注册后即可发布资源、参与收藏评论、通过签到积累积分，并解锁需要门槛的内容。
        </p>

        <div class="auth-highlights">
          <span>发布资源</span>
          <span>签到积分</span>
          <span>收藏管理</span>
          <span>评论互动</span>
        </div>
      </article>

      <el-form :model="form" class="auth-form" @submit.prevent="register">
        <div class="auth-form__head">
          <p class="auth-kicker">REGISTER</p>
          <h2>注册账号</h2>
          <span>创建后即可进入资源广场</span>
        </div>

      <el-form-item label="用户名" label-width="80px">
        <el-input v-model="form.username" placeholder="请输入用户名" />
      </el-form-item>
      <el-form-item label="密码" label-width="80px">
        <el-input v-model="form.password" type="password" placeholder="请输入密码" />
      </el-form-item>
        <el-button type="primary" native-type="submit" class="auth-submit">注册</el-button>
      </el-form>
    </div>
  </section>
</template>

  <script setup lang="ts">
  import { ref } from 'vue';
  import { useRouter } from 'vue-router';
  import { useAuthStore } from '../../store/auth';
  import { ElMessage } from 'element-plus';

  const form = ref({
    username: '',
    password: '',
  });

  const authStore = useAuthStore();
  const router = useRouter();

  const register = async () => {
    try {
      await authStore.register(form.value.username, form.value.password);
      router.push({ name: 'Resources' });
    } catch {
      ElMessage.error('注册失败，请确保密码在6位以上并包含数字和字母！！！');
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
    linear-gradient(140deg, rgba(124, 99, 55, 0.9), rgba(19, 63, 69, 0.92)),
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
  padding: 28px 24px 24px;
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
