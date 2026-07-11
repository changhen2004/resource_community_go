import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router';
import HomeView from '../views/HomeView.vue';
import UserCenterView from '../views/UserCenterView.vue';
import ResourceListView from '../views/ResourceListView.vue';
import ResourceDetailView from '../views/ResourceDetailView.vue';
import LoginView from '../views/auth/LoginView.vue';
import RegisterView from '../views/auth/RegisterView.vue';

const routes: RouteRecordRaw[] = [
  { path: '/', name: 'Home', component: HomeView },
  { path: '/center', name: 'Center', component: UserCenterView },
  { path: '/resources', name: 'Resources', component: ResourceListView },
  { path: '/resources/:id', name: 'ResourceDetail', component: ResourceDetailView },
  { path: '/login', name: 'Login', component: LoginView },
  { path: '/register', name: 'Register', component: RegisterView },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
