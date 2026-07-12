<template>
  <el-container>
    <el-header>
      <el-menu
        :default-active="activeIndex"
        class="el-menu-demo"
        mode="horizontal"
        :ellipsis="true"
        @select="handleSelect"
      >
        <el-menu-item index="home">首页</el-menu-item>
        <el-menu-item index="resources">资源广场</el-menu-item>
        <el-menu-item v-if="authStore.isAuthenticated" index="create">发布资源</el-menu-item>
        <el-menu-item v-if="authStore.isAuthenticated" index="center">用户中心</el-menu-item>
        <el-menu-item v-if="!authStore.isAuthenticated" index="login">登录</el-menu-item>
        <el-menu-item v-if="!authStore.isAuthenticated" index="register">注册</el-menu-item>
        <el-menu-item v-if="authStore.isAuthenticated" index="points" disabled>
          积分 {{ authStore.pointsBalance }}
        </el-menu-item>
        <el-menu-item v-if="authStore.isAuthenticated" index="logout">退出</el-menu-item>
      </el-menu>
    </el-header>
    <el-main>
      <router-view />
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from './store/auth';

const routeNameMap: Record<string, string> = {
  home: 'Home',
  resources: 'Resources',
  create: 'CreateResource',
  center: 'Center',
  login: 'Login',
  register: 'Register',
};

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();
const activeIndex = ref(route.name?.toString() || 'home');

watch(route, (newRoute) => {
  activeIndex.value = newRoute.name?.toString() || 'home';
});

const handleSelect = async (key: string) => {
  if (key === 'logout') {
    await authStore.logout();
    router.push({ name: 'Home' });
    return;
  }

  const routeName = routeNameMap[key];
  if (routeName) {
    router.push({ name: routeName });
  }
};
</script>

<style scoped>
.el-menu-demo {
  line-height: 60px;
}
</style>
