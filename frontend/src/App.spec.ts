import { mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import { describe, expect, it, vi } from 'vitest';
import { nextTick, reactive } from 'vue';
import App from './App.vue';
import { useAuthStore } from './store/auth';

const push = vi.fn();
const currentRoute = reactive({
  name: 'Home',
  query: {},
});

vi.mock('vue-router', () => ({
  useRoute: () => currentRoute,
  useRouter: () => ({
    push,
  }),
}));

describe('App navigation', () => {
  it('shows a create resource entry for authenticated users', async () => {
    const pinia = createPinia();
    setActivePinia(pinia);

    const authStore = useAuthStore();
    authStore.accessToken = 'token';

    const wrapper = mount(App, {
      global: {
        plugins: [pinia],
        stubs: {
          'router-view': true,
          'el-container': { template: '<div><slot /></div>' },
          'el-header': { template: '<header><slot /></header>' },
          'el-main': { template: '<main><slot /></main>' },
          'el-menu': { template: '<nav><slot /></nav>' },
          'el-menu-item': { template: '<button><slot /></button>' },
        },
      },
    });

    await nextTick();

    expect(wrapper.text()).toContain('发布资源');
  });

  it('does not show a create resource entry for guests', async () => {
    const pinia = createPinia();
    setActivePinia(pinia);

    const wrapper = mount(App, {
      global: {
        plugins: [pinia],
        stubs: {
          'router-view': true,
          'el-container': { template: '<div><slot /></div>' },
          'el-header': { template: '<header><slot /></header>' },
          'el-main': { template: '<main><slot /></main>' },
          'el-menu': { template: '<nav><slot /></nav>' },
          'el-menu-item': { template: '<button><slot /></button>' },
        },
      },
    });

    await nextTick();

    expect(wrapper.text()).not.toContain('发布资源');
  });
});
