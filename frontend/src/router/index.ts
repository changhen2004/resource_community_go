import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  { path: '/', name: 'Home', component: () => import('../views/HomeView.vue') },
  { path: '/center', name: 'Center', component: () => import('../views/UserCenterView.vue') },
  { path: '/resources', name: 'Resources', component: () => import('../views/ResourceListView.vue') },
  {
    path: '/resources/create',
    name: 'CreateResource',
    component: () => import('../views/CreateResourceView.vue'),
  },
  {
    path: '/resources/:id',
    name: 'ResourceDetail',
    component: () => import('../views/ResourceDetailView.vue'),
  },
  { path: '/login', name: 'Login', component: () => import('../views/auth/LoginView.vue') },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/auth/RegisterView.vue'),
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
